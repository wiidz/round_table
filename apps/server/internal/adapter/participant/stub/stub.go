package stub

import (
	"context"

	"round_table/apps/server/internal/adapter/participant"
)

// Participant returns fixed responses for engine integration tests.
type Participant struct {
	Content string
	Stance  string
}

var _ participant.Port = (*Participant)(nil)

// Respond implements participant.Port.
func (p *Participant) Respond(ctx context.Context, _, _ string, _ string) (participant.Response, error) {
	if err := ctx.Err(); err != nil {
		return participant.Response{}, err
	}
	stance := p.Stance
	if stance == "" {
		stance = "agree"
	}
	content := p.Content
	if content == "" {
		content = "同意当前方案。"
	}
	return participant.Response{
		Content: content,
		Stance:  stance,
	}, nil
}
