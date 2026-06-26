package discord

import (
	"context"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"

	"round_table/apps/server/internal/adapter/transport"
)

const maxMessageLen = 2000 // Discord API limit (code points); see message.go

// Options configures a single Discord bot connection.
type Options struct {
	Token      string
	AllowDM    bool
	AllowGuild bool
	GuildID    string // empty = all guilds the bot is in
}

// Bot implements transport.Port for one Discord bot token.
type Bot struct {
	session *discordgo.Session
	selfID  string
	opts    Options
}

// New opens a Discord session (not yet connected to the gateway).
func New(opts Options) (*Bot, error) {
	if strings.TrimSpace(opts.Token) == "" {
		return nil, fmt.Errorf("discord: bot token required")
	}
	if !opts.AllowDM && !opts.AllowGuild {
		return nil, fmt.Errorf("discord: at least one of allow_dm or allow_guild must be true")
	}

	session, err := discordgo.New("Bot " + strings.TrimSpace(opts.Token))
	if err != nil {
		return nil, fmt.Errorf("discord: new session: %w", err)
	}
	session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentDirectMessages | discordgo.IntentMessageContent

	b := &Bot{session: session, opts: opts}
	if err := b.session.Open(); err != nil {
		return nil, fmt.Errorf("discord: open gateway: %w", err)
	}
	b.selfID = session.State.User.ID
	return b, nil
}

// Run listens for messages until ctx is cancelled.
func (b *Bot) Run(ctx context.Context, handler transport.MessageHandler) error {
	if handler == nil {
		return fmt.Errorf("discord: message handler required")
	}

	b.session.AddHandler(func(_ *discordgo.Session, ev *discordgo.MessageCreate) {
		if ev == nil || ev.Message == nil || ev.Author == nil {
			return
		}
		if ev.Author.ID == b.selfID || ev.Author.Bot {
			return
		}
		if !b.shouldHandle(ev) {
			return
		}

		in := transport.Inbound{
			Platform:   "discord",
			ChannelID:  ev.ChannelID,
			GuildID:    ev.GuildID,
			AuthorID:   ev.Author.ID,
			AuthorName: ev.Author.Username,
			MessageID:  ev.ID,
			Content:    strings.TrimSpace(ev.Content),
		}
		if in.Content == "" {
			return
		}

		reply, err := handler(context.Background(), in)
		if err != nil || reply == "" {
			return
		}
		_ = b.send(ev.ChannelID, reply)
	})

	<-ctx.Done()
	return b.session.Close()
}

// Send posts a message to a channel.
func (b *Bot) Send(_ context.Context, channelID, content string) error {
	return b.send(channelID, content)
}

// DisplayName returns the connected bot username (may be empty before Open).
func (b *Bot) DisplayName() string {
	if b == nil || b.session == nil || b.session.State == nil || b.session.State.User == nil {
		return ""
	}
	return b.session.State.User.Username
}

// Close disconnects the gateway session.
func (b *Bot) Close() error {
	if b == nil || b.session == nil {
		return nil
	}
	return b.session.Close()
}

func (b *Bot) send(channelID, content string) error {
	content = clipMessageRunes(content)
	if content == "" {
		return nil
	}
	_, err := b.session.ChannelMessageSend(channelID, content)
	if err != nil {
		return fmt.Errorf("discord: send message: %w", err)
	}
	return nil
}

func (b *Bot) shouldHandle(ev *discordgo.MessageCreate) bool {
	return b.acceptMessage(ev.GuildID)
}

func (b *Bot) acceptMessage(guildID string) bool {
	if guildID == "" {
		return b.opts.AllowDM
	}
	if !b.opts.AllowGuild {
		return false
	}
	if b.opts.GuildID != "" && guildID != b.opts.GuildID {
		return false
	}
	return true
}
