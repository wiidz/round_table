package transport

import "context"

// Inbound is a normalized message from an external chat platform.
type Inbound struct {
	Platform  string // e.g. discord
	ChannelID string
	GuildID   string // empty for DMs
	AuthorID    string
	AuthorName  string
	MessageID string
	Content   string
}

// MessageHandler processes an inbound message and optionally returns a reply body.
type MessageHandler func(ctx context.Context, msg Inbound) (reply string, err error)

// Port is a bidirectional chat transport (Discord, Slack, …).
type Port interface {
	Run(ctx context.Context, handler MessageHandler) error
	Send(ctx context.Context, channelID, content string) error
}
