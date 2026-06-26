package discord

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"round_table/apps/server/internal/stream"
)

type typingStubSender struct {
	id           string
	typingActive int32
	typingStarts int
}

func (s *typingStubSender) Send(_ context.Context, _, _ string) error {
	return nil
}

func (s *typingStubSender) StartTyping(_ string) (stop func()) {
	s.typingStarts++
	atomic.StoreInt32(&s.typingActive, 1)
	return func() { atomic.StoreInt32(&s.typingActive, 0) }
}

func TestChannelStream_typingLifecycle(t *testing.T) {
	designer := &typingStubSender{id: "designer"}
	pool := &BotPool{
		Default: &typingStubSender{id: "main"},
		byID:    map[string]ChannelSender{"designer": designer},
	}
	cs := &channelStream{pool: pool, channelID: "ch1", loc: LocaleZH}

	cs.Start(stream.Meta{ParticipantID: "designer", Phase: "debate"})
	if designer.typingStarts != 1 {
		t.Fatalf("typingStarts=%d", designer.typingStarts)
	}
	if atomic.LoadInt32(&designer.typingActive) != 1 {
		t.Fatal("expected typing active during stream")
	}

	cs.buf.WriteString(`{"content":"hi","stance":"none"}`)
	cs.End()
	if atomic.LoadInt32(&designer.typingActive) != 1 {
		t.Fatal("typing should stay active until CompleteTurn for dedicated bot")
	}

	cs.CompleteTurn("designer", 10, time.Second)
	if atomic.LoadInt32(&designer.typingActive) != 0 {
		t.Fatal("typing should stop after message posted")
	}
}

func TestChannelStream_typingModeratorSynthesis(t *testing.T) {
	main := &typingStubSender{id: "main"}
	pool := &BotPool{Default: main}
	cs := &channelStream{pool: pool, channelID: "ch1", loc: LocaleZH}

	cs.Start(stream.Meta{ParticipantID: "moderator", Phase: "deliberation-synthesis"})
	if main.typingStarts != 1 {
		t.Fatalf("moderator typingStarts=%d", main.typingStarts)
	}

	cs.buf.WriteString(`{"summary":"draft"}`)
	cs.End()
	if atomic.LoadInt32(&main.typingActive) != 0 {
		t.Fatal("expected typing stopped after End for default bot")
	}
}

func TestChannelStream_fallbackHeaderWithoutDedicatedBot(t *testing.T) {
	main := &captureSender{}
	pool := &BotPool{Default: main}
	cs := &channelStream{pool: pool, channelID: "ch1", loc: LocaleZH}

	cs.Start(stream.Meta{ParticipantID: "designer", Phase: "debate"})
	if len(main.messages) != 1 || !strings.Contains(main.messages[0], "策划") {
		t.Fatalf("expected header fallback, messages=%v", main.messages)
	}
}

func TestRunTypingLoop_stopsOnSignal(t *testing.T) {
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		runTypingLoop(nil, "ch1", stop)
		close(done)
	}()
	close(stop)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("typing loop did not stop")
	}
}
