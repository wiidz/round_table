package engine

import (
	"bytes"
	"log"
	"strings"
	"testing"

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

func TestDiscardProgressLogger(t *testing.T) {
	e := &Engine{Progress: DiscardProgressLogger{}}
	e.logf("should not panic")
}
