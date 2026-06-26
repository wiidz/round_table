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
	meetingID := flag.String("id", "", "meeting id (default: mtg-<timestamp>)")
	participants := flag.String("participants", "architect:Architect:system design,developer:Developer:backend", "id:Role:Expertise,...")
	confirmation := flag.String("confirmation", "skip", "confirmation mode: skip | required")
	maxRounds := flag.Int("max-rounds", 0, "max rounds per segment (0 = use server.yaml default)")
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

	rounds := *maxRounds
	if rounds <= 0 {
		rounds = cfg.Meeting.MaxRoundsPerSegment
	}

	mode := *confirmation
	if mode == "" {
		mode = meeting.ConfirmationModeSkip
	}

	ctx := context.Background()
	log.Printf("creating meeting %s: %q", id, *topic)
	if _, err := eng.CreateMeeting(ctx, engine.CreateMeetingInput{
		MeetingID:           id,
		Topic:               *topic,
		ConfirmationMode:    mode,
		MaxRoundsPerSegment: rounds,
		Participants:        parts,
	}); err != nil {
		log.Fatalf("CreateMeeting: %v", err)
	}

	log.Printf("running meeting (model=%s, max_rounds=%d)...", cfg.Model.DefaultModel, rounds)
	final, err := eng.Run(ctx, id)
	if err != nil {
		log.Fatalf("Run: %v", err)
	}

	log.Printf("done: status=%s rounds=%d workspace=%s/%s",
		final.Status, len(final.Minutes.Rounds), cfg.Workspace.Root, id)
	if final.Consensus != nil {
		log.Printf("consensus: resolved_by=%s", final.Consensus.ResolvedBy)
	}
	if final.Confirmation != nil {
		log.Printf("confirmation: approved=%v cycle=%d", final.Confirmation.Approved, final.ConfirmationCycle)
	}
}

func parseParticipants(raw string) ([]engine.ParticipantInput, error) {
	var out []engine.ParticipantInput
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		fields := strings.Split(item, ":")
		if len(fields) < 2 {
			return nil, fmt.Errorf("invalid participant %q, want id:Role[:Expertise]", item)
		}
		p := engine.ParticipantInput{ID: fields[0], Role: fields[1]}
		if len(fields) >= 3 {
			p.Expertise = fields[2]
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("at least one participant required")
	}
	return out, nil
}
