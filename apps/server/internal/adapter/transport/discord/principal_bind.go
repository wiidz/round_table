package discord

import (
	"strings"

	"round_table/apps/server/internal/adapter/profile"
	"round_table/apps/server/internal/adapter/transport"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
)

func bindPrincipalInbound(
	reg *principalbind.Registry,
	profiles profile.Port,
	msg transport.Inbound,
	loc Locale,
) (string, error) {
	display := strings.TrimSpace(msg.AuthorName)
	if display == "" {
		display = msg.AuthorID
	}
	scope := principalbind.ScopeKey(msg.Platform, msg.GuildID, msg.AuthorID)
	b, err := reg.Bind(scope, msg.Platform, msg.AuthorID, display)
	if err != nil {
		return principalBindFailedText(loc, err), nil
	}
	ensurePrincipalProfile(profiles, b.PrincipalID)
	return principalBindOKText(loc, b.PrincipalID, b.DisplayName, scopeLabel(loc, msg.GuildID == "")), nil
}

func ensurePrincipalProfile(profiles profile.Port, principalID string) {
	if profiles == nil || principalID == "" {
		return
	}
	_ = profiles.EnsurePrincipal(principalID)
	if store, ok := profiles.(interface {
		EnsurePrincipalPersonas(string) (profile.PrincipalPersonaManifest, error)
	}); ok {
		_, _ = store.EnsurePrincipalPersonas(principalID)
	}
}
