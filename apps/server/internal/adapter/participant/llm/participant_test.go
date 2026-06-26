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

func (f fakeModel) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if err := ctx.Err(); err != nil {
		return model.Response{}, err
	}
	if f.content != "" {
		return model.Response{Content: f.content, Usage: model.Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15}}, nil
	}
	return model.Response{Content: `{"content":"同意方案","stance":"agree","object_reason":""}`, Usage: model.Usage{PromptTokens: 20, CompletionTokens: 8, TotalTokens: 28}}, nil
}

func TestParticipant_Respond(t *testing.T) {
	p := &Participant{Model: fakeModel{}}
	resp, err := p.Respond(context.Background(), "mtg-1", "architect", "Topic: API design\nPhase: debate\nRound: 1\nYou are architect\n\n--- Discussion so far ---\n## Pre-meeting (Round 0)\n\nviews\n")
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

func TestParticipant_Respond_preMeetingPhase(t *testing.T) {
	p := &Participant{Model: fakeModel{content: `{"content":"security angle","stance":"agree","object_reason":""}`}}
	resp, err := p.Respond(context.Background(), "mtg-1", "architect", "Topic: x\nPhase: pre-meeting\nPre-meeting (Round 0)\nYou are architect")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Stance != "none" {
		t.Fatalf("stance = %q want none", resp.Stance)
	}
}

func TestIsPreMeetingPrompt_debateWithContext(t *testing.T) {
	prompt := "Topic: x\nPhase: debate\nRound: 2\n\n--- Discussion so far ---\n## Pre-meeting (Round 0)\n\nviews"
	if isPreMeetingPrompt(prompt) {
		t.Fatal("must not treat debate prompt as pre-meeting")
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
