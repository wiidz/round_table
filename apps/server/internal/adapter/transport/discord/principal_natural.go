package discord

import (
	"strings"

	"round_table/apps/server/internal/adapter/profile"
	"round_table/apps/server/internal/adapter/transport"
	principalbind "round_table/apps/server/internal/adapter/transport/principal"
)

// tryPrincipalNatural handles natural-language Principal bind / whoami / unbind.
func tryPrincipalNatural(msg transport.Inbound, reg *principalbind.Registry, profiles profile.Port, loc Locale) (string, bool) {
	if reg == nil {
		return "", false
	}
	body := strings.TrimSpace(msg.Content)
	switch {
	case matchesPrincipalBindIntent(body):
		reply, err := bindPrincipalInbound(reg, profiles, msg, loc)
		return reply, err == nil && reply != ""
	case matchesPrincipalWhoamiIntent(body):
		reply, err := principalWhoamiFromInbound(reg, msg, loc)
		return reply, err == nil && reply != ""
	case matchesPrincipalUnbindIntent(body):
		reply, err := principalUnbindFromInbound(reg, msg, loc)
		return reply, err == nil && reply != ""
	default:
		return "", false
	}
}

func principalWhoamiFromInbound(reg *principalbind.Registry, msg transport.Inbound, loc Locale) (string, error) {
	scope := principalbind.ScopeKey(msg.Platform, msg.GuildID, msg.AuthorID)
	b, ok := reg.Get(scope)
	if !ok {
		return principalNotBoundText(loc, "!rt "), nil
	}
	if b.ExternalID == msg.AuthorID {
		return principalWhoamiSelfText(loc, b.PrincipalID, b.DisplayName, b.BoundAt.Format("2006-01-02 15:04 UTC")), nil
	}
	return principalWhoamiOtherText(loc, b.DisplayName, b.PrincipalID), nil
}

func principalUnbindFromInbound(reg *principalbind.Registry, msg transport.Inbound, loc Locale) (string, error) {
	scope := principalbind.ScopeKey(msg.Platform, msg.GuildID, msg.AuthorID)
	if err := reg.Unbind(scope, msg.AuthorID); err != nil {
		return principalUnbindFailedText(loc, err), nil
	}
	return principalUnbindOKText(loc), nil
}

func matchesPrincipalBindIntent(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	lower := strings.ToLower(normalizeASCIIForms(s))
	for _, exact := range []string{
		"绑定委托人", "绑定 principal", "bind principal", "principal bind",
		"绑定委托人身份", "注册委托人", "我要绑定委托人", "绑定一下委托人",
	} {
		if lower == strings.ToLower(normalizeASCIIForms(exact)) {
			return true
		}
	}
	hasBind := containsAnySubstring(s, "绑定", "注册") ||
		strings.Contains(lower, "bind") || strings.Contains(lower, "register")
	hasPrincipal := containsAnySubstring(s, "委托人") || strings.Contains(lower, "principal")
	return hasBind && hasPrincipal
}

func matchesPrincipalWhoamiIntent(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	lower := strings.ToLower(normalizeASCIIForms(s))
	for _, exact := range []string{
		"谁是委托人", "查看委托人", "委托人是谁", "我的委托人",
		"principal whoami", "whoami principal", "who is principal",
	} {
		if lower == strings.ToLower(normalizeASCIIForms(exact)) {
			return true
		}
	}
	if containsAnySubstring(s, "委托人") && containsAnySubstring(s, "谁", "查看", "查询") {
		return true
	}
	return strings.Contains(lower, "principal") && containsAnySubstring(s, "whoami", "who is", "who's")
}

func matchesPrincipalUnbindIntent(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	lower := strings.ToLower(normalizeASCIIForms(s))
	for _, exact := range []string{
		"解绑委托人", "解除委托人绑定", "解除委托人", "取消委托人绑定",
		"principal unbind", "unbind principal",
	} {
		if lower == strings.ToLower(normalizeASCIIForms(exact)) {
			return true
		}
	}
	hasUnbind := containsAnySubstring(s, "解绑", "解除", "取消绑定") ||
		strings.Contains(lower, "unbind") || strings.Contains(lower, "unregister")
	hasPrincipal := containsAnySubstring(s, "委托人") || strings.Contains(lower, "principal")
	return hasUnbind && hasPrincipal
}
