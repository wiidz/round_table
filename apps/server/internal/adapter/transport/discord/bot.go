package discord

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"

	"round_table/apps/server/internal/adapter/transport"
	"round_table/apps/server/internal/platform/config"
)

const maxMessageLen = 2000 // Discord API limit (code points); see message.go

// Options configures a single Discord bot connection.
type Options struct {
	Token      string
	AllowDM    bool
	AllowGuild bool
	GuildID    string // empty = all guilds the bot is in
	Locale     string // en | zh — network / reconnect notices

	// OnGatewayResumed is called when discordgo resumes the gateway session after disconnect.
	OnGatewayResumed func()

	// SendOnly opens a REST client without connecting to the Discord gateway (outbound only).
	SendOnly bool
}

// Bot implements transport.Port for one Discord bot token.
type Bot struct {
	session      *discordgo.Session
	selfID       string
	appID        string
	opts         Options
	loc          Locale
	hostGuard    func() bool
	claimMessage func(messageID string) bool
	runOnce      sync.Once
	opened       bool
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

	b := &Bot{session: session, opts: opts, loc: ParseLocale(opts.Locale)}
	if opts.SendOnly {
		return b, nil
	}
	if err := b.session.Open(); err != nil {
		return nil, fmt.Errorf("discord: open gateway: %w", err)
	}
	b.opened = true
	b.selfID = session.State.User.ID
	b.appID = config.ApplicationIDFromToken(opts.Token)
	return b, nil
}

// Run listens for messages until ctx is cancelled.
func (b *Bot) Run(ctx context.Context, handler transport.MessageHandler) error {
	if handler == nil {
		return fmt.Errorf("discord: message handler required")
	}
	if b.opts.SendOnly {
		return fmt.Errorf("discord: send-only bot cannot run message handler")
	}

	b.runOnce.Do(func() {
		if b.opts.OnGatewayResumed != nil {
			b.session.AddHandler(func(_ *discordgo.Session, _ *discordgo.Resumed) {
				b.opts.OnGatewayResumed()
			})
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
			if b.claimMessage != nil && !b.claimMessage(ev.ID) {
				return
			}
			if b.hostGuard != nil && !b.hostGuard() {
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

// Token returns the bot token used for this connection.
func (b *Bot) Token() string {
	if b == nil {
		return ""
	}
	return strings.TrimSpace(b.opts.Token)
}

// ApplicationID returns the connected bot's Discord Application snowflake.
func (b *Bot) ApplicationID() string {
	if b == nil {
		return ""
	}
	return b.appID
}

// SetHostGuard registers a callback that decides whether this bot may handle commands.
// Must be called before Run. The callback should reload host config when invoked.
func (b *Bot) SetHostGuard(fn func() bool) {
	if b == nil {
		return
	}
	b.hostGuard = fn
}

// SetMessageClaim registers cross-process inbound deduplication. Must be called before Run.
func (b *Bot) SetMessageClaim(fn func(messageID string) bool) {
	if b == nil {
		return
	}
	b.claimMessage = fn
}

// SetOnGatewayResumed registers a callback when discordgo resumes after disconnect.
// Must be called before Run.
func (b *Bot) SetOnGatewayResumed(fn func()) {
	if b == nil {
		return
	}
	b.opts.OnGatewayResumed = fn
}

// Close disconnects the gateway session.
func (b *Bot) Close() error {
	if b == nil || b.session == nil || !b.opened {
		return nil
	}
	return b.session.Close()
}

func (b *Bot) send(channelID, content string) error {
	content = clipMessageRunes(content)
	if content == "" {
		return nil
	}
	err := retrySend(func() error {
		_, sendErr := b.session.ChannelMessageSend(channelID, content)
		return sendErr
	})
	if err != nil {
		b.notifySendFailed(channelID)
		return fmt.Errorf("discord: send message: %w", err)
	}
	return nil
}

func (b *Bot) notifySendFailed(channelID string) {
	warn := networkSendFailedText(b.loc)
	_, _ = b.session.ChannelMessageSend(channelID, warn)
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
