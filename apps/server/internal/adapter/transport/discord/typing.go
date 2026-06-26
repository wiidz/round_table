package discord

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// TypingSender triggers Discord's native "… is typing" indicator for a bot account.
// See https://discord.com/developers/docs/resources/channel#trigger-typing-indicator
type TypingSender interface {
	StartTyping(channelID string) (stop func())
}

// Discord typing indicators expire after ~10s; refresh before that.
const typingRefreshInterval = 7 * time.Second

// StartTyping pulses ChannelTyping until stop is called.
func (b *Bot) StartTyping(channelID string) (stop func()) {
	if b == nil || b.session == nil || channelID == "" {
		return func() {}
	}
	session := b.session
	stopCh := make(chan struct{})
	var once sync.Once
	go runTypingLoop(session, channelID, stopCh)
	return func() {
		once.Do(func() { close(stopCh) })
	}
}

func runTypingLoop(session *discordgo.Session, channelID string, stop <-chan struct{}) {
	if session == nil || channelID == "" {
		<-stop
		return
	}
	pulse := func() {
		_ = session.ChannelTyping(channelID)
	}
	pulse()
	ticker := time.NewTicker(typingRefreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			pulse()
		}
	}
}
