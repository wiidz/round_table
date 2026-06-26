package engine

import (
	"bytes"
	"testing"

	"round_table/apps/server/internal/stream"
)

func TestStdStreamLogger_StartAndDelta(t *testing.T) {
	var buf bytes.Buffer
	l := StdStreamLogger{Out: &buf}
	meta := stream.Meta{ParticipantID: "skeptic", Phase: "debate", Detail: "turn (1/2) · round 1"}
	l.Start(meta)
	l.Delta(`{"content":"hi"}`)
	l.End()
	if got := buf.String(); got != `{"content":"hi"}`+"\n" {
		t.Fatalf("stream body = %q", got)
	}
}

func TestDiscardStreamLogger(t *testing.T) {
	l := DiscardStreamLogger{}
	l.Start(stream.Meta{ParticipantID: "p"})
	l.Delta("x")
	l.End()
}

func TestEngine_withStreamCtx_nil(t *testing.T) {
	e := &Engine{}
	ctx := e.withStreamCtx(t.Context(), stream.Meta{ParticipantID: "p"})
	if _, ok := stream.HandlersFrom(ctx); ok {
		t.Fatal("expected no handlers when Stream is nil")
	}
}
