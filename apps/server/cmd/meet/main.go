package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/engine"
	"round_table/apps/server/internal/platform/bootstrap"
	"round_table/apps/server/internal/platform/config"
)

func main() {
	topic := flag.String("topic", "", "meeting topic (required)")
	goal := flag.String("goal", "", "meeting goal (optional, default derived from topic)")
	meetingID := flag.String("id", "", "meeting id (default: mtg-<timestamp>)")
	participants := flag.String("participants", "architect:Architect:system design,developer:Developer:backend", "id:Role:Expertise,...")
	confirmation := flag.String("confirmation", "skip", "confirmation mode: skip | required")
	meetingMode := flag.String("mode", "decision", "meeting mode: decision | deliberation")
	maxRounds := flag.Int("max-rounds", 0, "max debate rounds per segment, excluding pre-meeting round 0 (0 = server.yaml default)")
	minRoundsBeforeSynthesis := flag.Int("min-rounds-before-synthesis", 0, "earliest debate round to allow early synthesis in deliberation mode (0 = server.yaml default)")
	maxFreeQuestions := flag.Int("max-free-dialogue-questions", -1, "questions per participant in free dialogue after Round 1 (-1=server.yaml default, 0=disable)")
	agenda := flag.String("agenda", "", "deliberation agenda items: id:Title,id2:Title2 (optional)")
	forceSynthesisAtRound := flag.Int("force-synthesis-at-round", 0, "deliberation: Principal forces synthesis when debate round >= N (0=disabled)")
	pauseAtRound := flag.Int("pause-at-round", 0, "Principal pauses once when debate round >= N (0=disabled)")
	abortAtRound := flag.Int("abort-at-round", 0, "Principal aborts when debate round >= N (0=disabled)")
	flag.Parse()

	if *topic == "" {
		fmt.Fprintln(os.Stderr, "usage: meet -topic \"...\" [-participants id:Role:Expertise,...]")
		flag.PrintDefaults()
		os.Exit(2)
	}

	cfg := config.Load()
	id := *meetingID
	if id == "" {
		id = fmt.Sprintf("mtg-%d", time.Now().Unix())
	}

	eng, err := bootstrap.NewEngine(cfg, bootstrap.PrincipalOptions{
		ForceSynthesisAtRound: *forceSynthesisAtRound,
		ForceSynthesisReason:  "Principal 要求立即合成草案",
		PauseAtRound:          *pauseAtRound,
		AbortAtRound:          *abortAtRound,
	})
	if err != nil {
		log.Fatalf("engine: %v", err)
	}

	parts, err := parseParticipants(*participants)
	if err != nil {
		log.Fatalf("participants: %v", err)
	}
	agendaItems, err := parseAgenda(*agenda)
	if err != nil {
		log.Fatalf("agenda: %v", err)
	}
	for _, p := range parts {
		log.Printf("participant: %s (%s)", p.ID, p.Role)
	}
	for _, item := range agendaItems {
		log.Printf("agenda: %s (%s)", item.ID, item.Title)
	}

	rounds := *maxRounds
	if rounds <= 0 {
		rounds = cfg.Meeting.MaxRoundsPerSegment
	}
	minRounds := *minRoundsBeforeSynthesis
	if minRounds <= 0 {
		minRounds = cfg.Meeting.MinRoundsBeforeSynthesis
	}
	if minRounds <= 0 {
		minRounds = 2
	}
	freeQuestions := cfg.Meeting.FreeDialogueMaxQuestions
	if *maxFreeQuestions >= 0 {
		freeQuestions = *maxFreeQuestions
	}

	mode := *confirmation
	if mode == "" {
		mode = meeting.ConfirmationModeSkip
	}

	ctx := context.Background()
	log.Printf("creating meeting %s: %q", id, *topic)
	freeQ := freeQuestions
	minQ := minRounds
	if _, err := eng.CreateMeeting(ctx, engine.CreateMeetingInput{
		MeetingID:                id,
		Topic:                    *topic,
		Goal:                     *goal,
		MeetingMode:              *meetingMode,
		ConfirmationMode:         mode,
		MaxRoundsPerSegment:      rounds,
		MinRoundsBeforeSynthesis: &minQ,
		FreeDialogueMaxQuestions: &freeQ,
		Agenda:                   agendaItems,
		Participants:             parts,
	}); err != nil {
		log.Fatalf("CreateMeeting: %v", err)
	}

	log.Printf("running meeting (model=%s, mode=%s, max_debate_rounds=%d, min_rounds_before_synthesis=%d, agenda_items=%d, free_dialogue_questions=%d, pre-meeting=round 0)...",
		cfg.Model.DefaultModel, *meetingMode, rounds, minRounds, len(agendaItems), freeQuestions)
	final, err := eng.Run(ctx, id)
	if err != nil {
		log.Fatalf("Run: %v", err)
	}

	log.Printf("done: status=%s debate_rounds=%d pre_meeting=1 workspace=%s/%s",
		final.Status, final.DebateRoundCount(), cfg.Workspace.Root, id)
	if final.Consensus != nil {
		if final.IsDeliberation() {
			log.Printf("synthesis: resolved_by=%s", final.Consensus.ResolvedBy)
		} else {
			log.Printf("consensus: resolved_by=%s", final.Consensus.ResolvedBy)
		}
	}
	if final.Confirmation != nil {
		log.Printf("confirmation: approved=%v cycle=%d", final.Confirmation.Approved, final.ConfirmationCycle)
	}
	if final.TokenUsageTotals.CallCount > 0 {
		log.Printf("tokens: calls=%d prompt=%d completion=%d total=%d (see usage/summary.md)",
			final.TokenUsageTotals.CallCount,
			final.TokenUsageTotals.PromptTokens,
			final.TokenUsageTotals.CompletionTokens,
			final.TokenUsageTotals.TotalTokens)
	}
}

func parseParticipants(raw string) ([]engine.ParticipantInput, error) {
	var out []engine.ParticipantInput
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		p, err := parseParticipantItem(item)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("at least one participant required")
	}
	return out, nil
}

// parseParticipantItem parses id:Role[:Expertise], allowing spaces in Role.
func parseParticipantItem(item string) (engine.ParticipantInput, error) {
	first := strings.Index(item, ":")
	if first <= 0 || first >= len(item)-1 {
		return engine.ParticipantInput{}, fmt.Errorf("invalid participant %q, want id:Role[:Expertise]", item)
	}
	id := item[:first]
	rest := item[first+1:]
	last := strings.LastIndex(rest, ":")
	if last <= 0 {
		return engine.ParticipantInput{ID: id, Role: rest}, nil
	}
	return engine.ParticipantInput{
		ID:        id,
		Role:      rest[:last],
		Expertise: rest[last+1:],
	}, nil
}

func parseAgenda(raw string) ([]event.AgendaItem, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var out []event.AgendaItem
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		colon := strings.Index(item, ":")
		if colon <= 0 {
			return nil, fmt.Errorf("invalid agenda item %q, want id:Title", item)
		}
		id := strings.TrimSpace(item[:colon])
		title := strings.TrimSpace(item[colon+1:])
		if id == "" || title == "" {
			return nil, fmt.Errorf("invalid agenda item %q, want id:Title", item)
		}
		out = append(out, event.AgendaItem{ID: id, Title: title})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("agenda must contain at least one item when -agenda is set")
	}
	seen := make(map[string]bool)
	for _, item := range out {
		if seen[item.ID] {
			return nil, fmt.Errorf("duplicate agenda id %q", item.ID)
		}
		seen[item.ID] = true
	}
	return out, nil
}
