package llm

import (
	"context"
	"fmt"
	"strings"

	"round_table/apps/server/internal/adapter/model"
	"round_table/apps/server/internal/adapter/participant"
	"round_table/apps/server/internal/adapter/profile"
	"round_table/apps/server/internal/stream"
)

const responseSchema = `Respond ONLY with a JSON object (no markdown fences).
Do NOT use ASCII double quotes (") inside content — use 「」 for emphasis if needed.
{"content":"<your spoken contribution>","stance":"agree|object|abstain","object_reason":"<required when stance is object, else empty string>"}`

const preMeetingSchema = `Respond ONLY with a JSON object (no markdown fences).
Do NOT use ASCII double quotes (") inside content — use 「」 for emphasis if needed.
{"content":"<your preliminary perspectives and evaluation angles>","stance":"none","object_reason":""}`

const deliberationSchema = `Respond ONLY with a JSON object (no markdown fences).
Do NOT use ASCII double quotes (") inside content — use 「」 for emphasis if needed.
{"content":"<your design contribution: ideas, constraints, trade-offs, and open questions from your role>","stance":"none","object_reason":""}`

const freeDialogueAskSchema = `Respond ONLY with a JSON object (no markdown fences).
Do NOT use ASCII double quotes (") inside content — use 「」 for emphasis if needed.
{"content":"<your question to the other participant>"}`

const freeDialogueAnswerSchema = `Respond ONLY with a JSON object (no markdown fences).
Do NOT use ASCII double quotes (") inside content — use 「」 for emphasis if needed.
{"content":"<your answer to the question>"}`

// Participant invokes an LLM with profile identity files (SOUL, AGENTS).
type Participant struct {
	Model     model.Port
	Profile   profile.Port
	ModelName string
}

var _ participant.Port = (*Participant)(nil)

type llmOutput struct {
	Content      string `json:"content"`
	Stance       string `json:"stance"`
	ObjectReason string `json:"object_reason"`
}

// Respond implements participant.Port.
func (p *Participant) Respond(ctx context.Context, _, participantID string, prompt string) (participant.Response, error) {
	if p.Model == nil {
		return participant.Response{}, fmt.Errorf("llm participant: model port required")
	}
	system, err := p.buildSystem(participantID)
	if err != nil {
		return participant.Response{}, err
	}
	modelName := p.ModelName
	if modelName == "" {
		modelName = "deepseek-chat"
	}

	schema := responseSchema
	switch {
	case isPreMeetingPrompt(prompt):
		schema = preMeetingSchema
	case isDeliberationPrompt(prompt):
		schema = deliberationSchema
	case isFreeDialogueAskPrompt(prompt):
		schema = freeDialogueAskSchema
	case isFreeDialogueAnswerPrompt(prompt):
		schema = freeDialogueAnswerSchema
	}

	onDelta, onEnd := p.streamHandlers(ctx)
	raw, err := p.Model.Complete(ctx, model.Request{
		Model: modelName,
		Messages: []model.Message{
			{Role: "system", Content: system},
			{Role: "user", Content: prompt + "\n\n" + schema},
		},
		Temperature: 0.7,
		OnDelta:     onDelta,
	})
	if err != nil {
		return participant.Response{}, err
	}
	if onEnd != nil {
		onEnd()
	}

	out, err := parseOutput(raw.Content)
	if err != nil {
		return participant.Response{}, fmt.Errorf("llm participant: parse response: %w", err)
	}
	stance := strings.ToLower(strings.TrimSpace(out.Stance))
	switch {
	case isPreMeetingPrompt(prompt):
		stance = "none"
	case isDeliberationPrompt(prompt):
		stance = "none"
	case isFreeDialogueAskPrompt(prompt), isFreeDialogueAnswerPrompt(prompt):
		stance = "none"
	case stance == "" || stance == "none":
		stance = "agree"
	}
	return participant.Response{
		ParticipantID: participantID,
		Content:       strings.TrimSpace(out.Content),
		Stance:        stance,
		ObjectReason:  strings.TrimSpace(out.ObjectReason),
		Model:         modelName,
		Usage:         raw.Usage,
	}, nil
}

func (p *Participant) streamHandlers(ctx context.Context) (model.StreamHandler, func()) {
	h, ok := stream.HandlersFrom(ctx)
	if !ok {
		return nil, nil
	}
	if h.OnStart != nil {
		h.OnStart(h.Meta)
	}
	return h.OnDelta, h.OnEnd
}

func (p *Participant) buildSystem(participantID string) (string, error) {
	var b strings.Builder
	b.WriteString("You are a RoundTable meeting participant. Stay in character.\n\n")
	if p.Profile != nil {
		if data, err := p.Profile.ReadParticipant(participantID, profile.FileSoul); err == nil {
			b.WriteString("--- SOUL.md ---\n")
			b.Write(data)
			b.WriteByte('\n')
		}
		if data, err := p.Profile.ReadParticipant(participantID, profile.FileAgents); err == nil {
			b.WriteString("\n--- AGENTS.md ---\n")
			b.Write(data)
			b.WriteByte('\n')
		}
	}
	return b.String(), nil
}

func isPreMeetingPrompt(prompt string) bool {
	for _, line := range strings.Split(prompt, "\n") {
		if strings.TrimSpace(line) == "Phase: pre-meeting" {
			return true
		}
	}
	return false
}

func isDeliberationPrompt(prompt string) bool {
	for _, line := range strings.Split(prompt, "\n") {
		if strings.TrimSpace(line) == "Phase: deliberation" {
			return true
		}
	}
	return false
}

func isFreeDialogueAskPrompt(prompt string) bool {
	for _, line := range strings.Split(prompt, "\n") {
		if strings.TrimSpace(line) == "Phase: free-dialogue-ask" {
			return true
		}
	}
	return false
}

func isFreeDialogueAnswerPrompt(prompt string) bool {
	for _, line := range strings.Split(prompt, "\n") {
		if strings.TrimSpace(line) == "Phase: free-dialogue-answer" {
			return true
		}
	}
	return false
}
