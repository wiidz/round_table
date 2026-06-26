package engine

import (
	"bytes"
	"encoding/json"
	"log"
	"strings"
	"testing"

	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
)

type captureLogger struct {
	buf bytes.Buffer
}

func (c *captureLogger) Logf(format string, args ...any) {
	log.New(&c.buf, "", 0).Printf(format, args...)
}

func TestLogProgress_roundStarted(t *testing.T) {
	var cap captureLogger
	e := &Engine{Progress: &cap}
	env := eventRoundStarted(1, []string{"a", "b"})
	s := meeting.State{ParticipantOrder: []string{"a", "b"}}
	e.logProgress(env, s)
	out := cap.buf.String()
	if !strings.Contains(out, "debate round 1 started") || !strings.Contains(out, "a → b") {
		t.Fatalf("got %q", out)
	}
}

func TestLogProgress_moderatorSummarized(t *testing.T) {
	var cap captureLogger
	e := &Engine{Progress: &cap}
	payload, _ := json.Marshal(event.ModeratorSummarizedPayload{
		RoundNumber: 2,
		Summary:     "## Round 2 研讨摘要\n\n要点 A",
	})
	env := event.Envelope{Type: event.TypeModeratorSummarized, Payload: payload}
	e.logProgress(env, meeting.State{CurrentRound: 2})
	out := cap.buf.String()
	if !strings.Contains(out, "moderator summary round=2") {
		t.Fatalf("missing header: %q", out)
	}
	if !strings.Contains(out, "要点 A") {
		t.Fatalf("missing body: %q", out)
	}
}

func TestDiscardProgressLogger(t *testing.T) {
	e := &Engine{Progress: DiscardProgressLogger{}}
	e.logf("should not panic")
}
