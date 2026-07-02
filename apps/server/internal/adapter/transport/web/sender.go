package web

import (
	"context"

	discordtransport "round_table/apps/server/internal/adapter/transport/discord"
	"round_table/apps/server/internal/platform/config"
)

// ChannelSender delivers outbound text to a browser chat session (discord.ChannelSender compatible).
type ChannelSender struct {
	Hub        *Hub
	Role       string
	AuthorID   string
	AuthorName string
}

// NewModeratorSender returns a sender that posts as 主持人.
func NewModeratorSender(hub *Hub) *ChannelSender {
	return &ChannelSender{
		Hub:        hub,
		Role:       RoleModerator,
		AuthorID:   "moderator",
		AuthorName: "主持人",
	}
}

// Send implements discord.ChannelSender.
func (s *ChannelSender) Send(ctx context.Context, channelID, content string) error {
	if s == nil || s.Hub == nil {
		return nil
	}
	role := s.Role
	if role == "" {
		role = RoleModerator
	}
	s.Hub.SendTyping(ctx, channelID, role, s.AuthorID, s.AuthorName)
	return s.Hub.SendOutbound(ctx, channelID, Outbound{
		Role:       role,
		Content:    content,
		AuthorID:   s.AuthorID,
		AuthorName: s.AuthorName,
	})
}

// ParticipantSender posts as a meeting expert in browser chat.
type ParticipantSender struct {
	Hub           *Hub
	ParticipantID string
	DisplayName   string
}

// Send implements discord.ChannelSender.
func (s *ParticipantSender) Send(ctx context.Context, channelID, content string) error {
	if s == nil || s.Hub == nil || s.ParticipantID == "" {
		return nil
	}
	name := s.DisplayName
	if name == "" {
		name = s.ParticipantID
	}
	s.Hub.SendTyping(ctx, channelID, RoleParticipant, s.ParticipantID, name)
	return s.Hub.SendOutbound(ctx, channelID, Outbound{
		Role:       RoleParticipant,
		Content:    content,
		AuthorID:   s.ParticipantID,
		AuthorName: name,
	})
}

// NewBotPool builds a web BotPool with per-participant senders for meeting stream output.
func NewBotPool(hub *Hub, roster string) *discordtransport.BotPool {
	mod := NewModeratorSender(hub)
	byID := make(map[string]discordtransport.ChannelSender)
	for id, entry := range config.ParseMeetParticipants(roster) {
		byID[id] = &ParticipantSender{
			Hub:           hub,
			ParticipantID: id,
			DisplayName:   entry.DisplayName,
		}
	}
	return discordtransport.NewMappedBotPool(mod, byID)
}
