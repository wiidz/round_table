package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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
	maxRounds := flag.Int("max-rounds", 0, "max debate rounds per segment, excluding pre-meeting round 0 (0 = server.yaml default)")
	maxFreeQuestions := flag.Int("max-free-dialogue-questions", -1, "questions per participant in free dialogue after Round 1 (-1=server.yaml default, 0=disable)")
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

	eng, err := bootstrap.NewEngine(cfg)
	if err != nil {
		log.Fatalf("engine: %v", err)
	}

	parts, err := parseParticipants(*participants)
	if err != nil {
		log.Fatalf("participants: %v", err)
	}
	for _, p := range parts {
		log.Printf("participant: %s (%s)", p.ID, p.Role)
	}

	rounds := *maxRounds
	if rounds <= 0 {
		rounds = cfg.Meeting.MaxRoundsPerSegment
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
	if _, err := eng.CreateMeeting(ctx, engine.CreateMeetingInput{
		MeetingID:                id,
		Topic:                    *topic,
		Goal:                     *goal,
		ConfirmationMode:         mode,
		MaxRoundsPerSegment:      rounds,
		FreeDialogueMaxQuestions: &freeQ,
		Participants:             parts,
	}); err != nil {
		log.Fatalf("CreateMeeting: %v", err)
	}

	log.Printf("running meeting (model=%s, max_debate_rounds=%d, free_dialogue_questions=%d, pre-meeting=round 0)...",
		cfg.Model.DefaultModel, rounds, freeQuestions)
	final, err := eng.Run(ctx, id)
	if err != nil {
		log.Fatalf("Run: %v", err)
	}

	log.Printf("done: status=%s debate_rounds=%d pre_meeting=1 workspace=%s/%s",
		final.Status, final.DebateRoundCount(), cfg.Workspace.Root, id)
	if final.Consensus != nil {
		log.Printf("consensus: resolved_by=%s", final.Consensus.ResolvedBy)
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
