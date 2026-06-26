package engine

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"round_table/apps/server/internal/stream"
)

// StreamTurnFinisher receives token usage after a participant stream ends (optional transport hook).
type StreamTurnFinisher interface {
	CompleteTurn(participantID string, tokens int, elapsed time.Duration)
}

func notifyStreamTurnComplete(stream StreamLogger, participantID string, tokens int, elapsed time.Duration) {
	if stream == nil {
		return
	}
	if f, ok := stream.(StreamTurnFinisher); ok {
		f.CompleteTurn(participantID, tokens, elapsed)
	}
}

// StreamLogger receives labeled LLM token streams (CLI / future WebSocket client).
type StreamLogger interface {
	Start(meta stream.Meta)
	Delta(delta string)
	End()
}

// DiscardStreamLogger suppresses stream output (tests).
type DiscardStreamLogger struct{}

func (DiscardStreamLogger) Start(stream.Meta) {}
func (DiscardStreamLogger) Delta(string)      {}
func (DiscardStreamLogger) End()              {}

// StdStreamLogger prints a labeled header via log and raw deltas to Out (default stderr).
type StdStreamLogger struct {
	Out io.Writer
}

func (l StdStreamLogger) out() io.Writer {
	if l.Out != nil {
		return l.Out
	}
	return os.Stderr
}

func (l StdStreamLogger) Start(meta stream.Meta) {
	if meta.Detail != "" {
		log.Printf("meet: ↳ %s · %s · %s", meta.ParticipantID, meta.Phase, meta.Detail)
		return
	}
	log.Printf("meet: ↳ %s · %s", meta.ParticipantID, meta.Phase)
}

func (l StdStreamLogger) Delta(delta string) {
	fmt.Fprint(l.out(), delta)
}

func (l StdStreamLogger) End() {
	fmt.Fprintln(l.out())
}

func (e *Engine) withStreamCtx(ctx context.Context, meta stream.Meta) context.Context {
	if e.Stream == nil {
		return ctx
	}
	return stream.WithHandlers(ctx, stream.Handlers{
		Meta:    meta,
		OnStart: e.Stream.Start,
		OnDelta: e.Stream.Delta,
		OnEnd:   e.Stream.End,
	})
}
