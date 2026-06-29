package discord

import (
	"strings"

	"round_table/apps/server/internal/platform/config"
)

func (r *MeetRunner) activeCfg() config.Config {
	if r.ConfigSvc != nil {
		return r.ConfigSvc.Current()
	}
	return r.Cfg
}

// ActiveCfg returns effective config (exported for web transport).
func (r *MeetRunner) ActiveCfg() config.Config {
	return r.activeCfg()
}

func (r *MeetRunner) dc() config.DiscordTransport {
	if r.ConfigSvc != nil {
		return r.ConfigSvc.Current().Transport.Discord
	}
	return r.Discord
}

func (r *MeetRunner) locale() Locale {
	cfg := r.activeCfg()
	loc := strings.TrimSpace(cfg.Server.Locale)
	if loc == "" {
		loc = strings.TrimSpace(cfg.Transport.Discord.Locale)
	}
	if loc == "" {
		loc = strings.TrimSpace(r.dc().Locale)
	}
	return ParseLocale(loc)
}
