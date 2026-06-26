package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"round_table/apps/server/internal/adapter/model"
	"round_table/apps/server/internal/adapter/participant"
	"round_table/apps/server/internal/adapter/profile"
)

const responseSchema = `Respond ONLY with a JSON object (no markdown fences):
{"content":"<your spoken contribution>","stance":"agree|object|abstain","object_reason":"<required when stance is object, else empty string>"}`

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

	raw, err := p.Model.Complete(ctx, model.Request{
		Model: modelName,
		Messages: []model.Message{
			{Role: "system", Content: system},
			{Role: "user", Content: prompt + "\n\n" + responseSchema},
		},
		Temperature: 0.7,
	})
	if err != nil {
		return participant.Response{}, err
	}

	out, err := parseOutput(raw)
	if err != nil {
		return participant.Response{}, fmt.Errorf("llm participant: parse response: %w", err)
	}
	stance := strings.ToLower(strings.TrimSpace(out.Stance))
	if stance == "" {
		stance = "agree"
	}
	return participant.Response{
		ParticipantID: participantID,
		Content:       strings.TrimSpace(out.Content),
		Stance:        stance,
		ObjectReason:  strings.TrimSpace(out.ObjectReason),
	}, nil
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

func parseOutput(raw string) (llmOutput, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var out llmOutput
	if err := json.Unmarshal([]byte(raw), &out); err == nil && out.Content != "" {
		return out, nil
	}
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(raw[start:end+1]), &out); err == nil && out.Content != "" {
			return out, nil
		}
	}
	return llmOutput{}, fmt.Errorf("invalid JSON: %q", raw)
}
