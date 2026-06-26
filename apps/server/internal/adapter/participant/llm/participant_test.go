package llm

import (
	"context"
	"testing"

	"round_table/apps/server/internal/adapter/model"
	"round_table/apps/server/internal/adapter/participant"
)

type fakeModel struct {
	content string
}

func (f fakeModel) Complete(ctx context.Context, req model.Request) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if f.content != "" {
		return f.content, nil
	}
	return `{"content":"同意方案","stance":"agree","object_reason":""}`, nil
}

func TestParticipant_Respond(t *testing.T) {
	p := &Participant{Model: fakeModel{}}
	resp, err := p.Respond(context.Background(), "mtg-1", "architect", "Topic: API design\nRound: 1")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Stance != "agree" {
		t.Fatalf("stance = %q", resp.Stance)
	}
	if resp.Content == "" {
		t.Fatal("empty content")
	}
}

func TestParticipant_Respond_object(t *testing.T) {
	p := &Participant{Model: fakeModel{content: `{"content":"需要补充测试","stance":"object","object_reason":"缺少边界用例"}`}}
	resp, err := p.Respond(context.Background(), "mtg-1", "dev", "prompt")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Stance != "object" || resp.ObjectReason == "" {
		t.Fatalf("got %+v", resp)
	}
}

func TestParticipant_Respond_noModel(t *testing.T) {
	p := &Participant{}
	_, err := p.Respond(context.Background(), "mtg-1", "p1", "prompt")
	if err == nil {
		t.Fatal("expected error")
	}
}

var _ participant.Port = (*Participant)(nil)
