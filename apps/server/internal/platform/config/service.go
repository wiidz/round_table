package config

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// SettingsFieldState is one row in GET /api/settings.
type SettingsFieldState struct {
	Key             string          `json:"key"`
	Value           string          `json:"value,omitempty"`
	Configured      bool            `json:"configured"`
	Secret          bool            `json:"secret"`
	Editable        bool            `json:"editable"`
	RestartRequired bool            `json:"restart_required,omitempty"`
	Label           string          `json:"label"`
	Group           string          `json:"group"`
	Subsection      string          `json:"subsection,omitempty"`
	Section         string          `json:"section,omitempty"`
	Placeholder     string          `json:"placeholder,omitempty"`
	Description     string          `json:"description,omitempty"`
	InputType       string          `json:"input_type,omitempty"`
	Options         []SettingOption `json:"options,omitempty"`
	Min             *int            `json:"min,omitempty"`
	Max             *int            `json:"max,omitempty"`
}

// SettingsResponse is returned by GET/PUT /api/settings.
type SettingsResponse struct {
	Source      string                            `json:"source"`
	SecretsPath string                            `json:"secrets_path"`
	Groups      []string                          `json:"groups"`
	Subsections map[string][]SettingsSubsectionMeta `json:"subsections"`
	Fields      []SettingsFieldState              `json:"fields"`
	DiscordBots []DiscordBotState                 `json:"discord_bots,omitempty"`
	MeetPresets         []MeetPresetConfig `json:"meet_presets,omitempty"`
	MeetPresetsDefaults []MeetPresetConfig `json:"meet_presets_defaults,omitempty"`
}

// Service holds runtime config with optional SQLite app_settings overrides (P0.5).
type Service struct {
	mu    sync.RWMutex
	cfg   Config
	store SettingsStore
}

// NewService loads base config and merges app_settings when store is non-nil.
func NewService(store SettingsStore) (*Service, error) {
	cfg := loadBase()
	s := &Service{cfg: cfg, store: store}
	if store != nil {
		ctx := context.Background()
		overrides, err := store.GetAllSettings(ctx)
		if err != nil {
			return nil, fmt.Errorf("config service: load settings: %w", err)
		}
		if toSet, toDelete := migrateLegacySettings(overrides); len(toSet) > 0 || len(toDelete) > 0 {
			for k, v := range toSet {
				overrides[k] = v
			}
			if err := store.SetSettings(ctx, toSet); err != nil {
				return nil, fmt.Errorf("config service: migrate settings: %w", err)
			}
			if deleter, ok := store.(interface {
				DeleteSettings(context.Context, ...string) error
			}); ok && len(toDelete) > 0 {
				if err := deleter.DeleteSettings(ctx, toDelete...); err != nil {
					return nil, fmt.Errorf("config service: delete legacy settings: %w", err)
				}
			}
			for _, key := range toDelete {
				delete(overrides, key)
			}
		}
		if tokenMigrate := migrateDiscordTokensFromEnv(overrides, cfg); len(tokenMigrate) > 0 {
			for k, v := range tokenMigrate {
				overrides[k] = v
			}
			if err := store.SetSettings(ctx, tokenMigrate); err != nil {
				return nil, fmt.Errorf("config service: migrate discord tokens: %w", err)
			}
		}
		if botMigrate := migrateDiscordBotsToApplicationIDs(overrides); len(botMigrate) > 0 {
			for k, v := range botMigrate {
				overrides[k] = v
			}
			if err := store.SetSettings(ctx, botMigrate); err != nil {
				return nil, fmt.Errorf("config service: migrate discord bots: %w", err)
			}
		}
		if err := applySettingsMap(&s.cfg, overrides); err != nil {
			return nil, err
		}
		applyDiscordBotTokens(&s.cfg, overrides)
	}
	s.refreshMeetPresetsLocked(s.settingsOverridesLocked())
	s.refreshMeetParticipantsLocked(s.settingsOverridesLocked())
	if err := s.refreshParticipantIMBindingsLocked(context.Background()); err != nil {
		return nil, err
	}
	normalizeLoadedConfig(&s.cfg)
	return s, nil
}

func (s *Service) refreshMeetParticipantsLocked(overrides map[string]string) {
	s.cfg.Transport.Discord.MeetParticipants = meetParticipantsFromOverrides(overrides, s.cfg)
}

func (s *Service) refreshParticipantIMBindingsLocked(ctx context.Context) error {
	overrides := s.settingsOverridesLocked()
	_, hasStored := overrides[ParticipantIMBindingsSetting]
	bindings := participantIMBindingsFromOverrides(overrides)
	if !hasStored {
		bindings = migrateLegacyParticipantIMBindings(s.cfg)
	}
	s.cfg.Transport.Discord.ParticipantIMBindings = bindings
	return nil
}

func (s *Service) effectiveParticipantIMBindingsLocked() ParticipantIMBindings {
	if s.cfg.Transport.Discord.ParticipantIMBindings != nil {
		return s.cfg.Transport.Discord.ParticipantIMBindings.clone()
	}
	return participantIMBindingsFromOverrides(s.settingsOverridesLocked())
}

func (s *Service) persistParticipantIMBindingsLocked(ctx context.Context, bindings ParticipantIMBindings) error {
	roster := ParticipantRosterFromConfig(s.cfg)
	bots := effectiveDiscordBots(s.cfg, s.settingsOverridesLocked())
	if err := validateParticipantIMBindings(bindings, rosterIDSet(roster), discordApplicationIDSet(bots)); err != nil {
		return err
	}
	if s.store != nil {
		if err := s.store.SetSettings(ctx, map[string]string{
			ParticipantIMBindingsSetting: formatParticipantIMBindingsJSON(bindings),
		}); err != nil {
			return err
		}
	}
	s.cfg.Transport.Discord.ParticipantIMBindings = bindings
	return nil
}

func (s *Service) refreshMeetPresetsLocked(overrides map[string]string) {
	s.cfg.Meeting.MeetPresets = meetPresetsFromOverrides(overrides, s.cfg)
}

// UpdateMeetPresets validates and persists Discord preset menu entries.
func (s *Service) UpdateMeetPresets(ctx context.Context, presets []MeetPresetConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	next := s.cfg
	normalized, err := normalizeMeetPresets(presets, next)
	if err != nil {
		return err
	}
	jsonVal := formatMeetPresetsJSON(normalized)
	if s.store != nil {
		if err := s.store.SetSettings(ctx, map[string]string{MeetPresetsSetting: jsonVal}); err != nil {
			return err
		}
	}
	next.Meeting.MeetPresets = normalized
	s.cfg = next
	return nil
}

// ResetMeetPresets removes stored overrides and restores seed preset menu entries.
func (s *Service) ResetMeetPresets(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.store != nil {
		if deleter, ok := s.store.(interface {
			DeleteSettings(context.Context, ...string) error
		}); ok {
			if err := deleter.DeleteSettings(ctx, MeetPresetsSetting); err != nil {
				return err
			}
		}
	}
	next := s.cfg
	next.Meeting.MeetPresets = DefaultMeetPresets(next)
	s.cfg = next
	return nil
}

// Current returns a snapshot of effective runtime config.
func (s *Service) Current() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

// Update persists editable settings and applies them in memory.
func (s *Service) Update(ctx context.Context, updates map[string]string) error {
	if len(updates) == 0 {
		return nil
	}
	filtered := make(map[string]string, len(updates))
	for key, val := range updates {
		if !IsEditableSettingKey(key) {
			continue
		}
		filtered[key] = val
	}
	if len(filtered) == 0 {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	next := s.cfg
	if err := applySettingsMap(&next, filtered); err != nil {
		return err
	}
	if s.store != nil {
		if err := s.store.SetSettings(ctx, filtered); err != nil {
			return err
		}
	}
	s.cfg = next
	s.refreshMeetPresetsLocked(s.settingsOverridesLocked())
	return nil
}

func (s *Service) settingsOverrides() map[string]string {
	if s.store == nil {
		return nil
	}
	overrides, err := s.store.GetAllSettings(context.Background())
	if err != nil {
		return nil
	}
	return overrides
}

// UpdateDiscordBots persists participant bot roster and tokens to SQLite app_settings.
func (s *Service) UpdateDiscordBots(ctx context.Context, update DiscordBotsUpdate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	overrides := s.settingsOverridesLocked()
	profileCache := discordBotProfilesFromOverrides(overrides)
	tokenStore := effectiveDiscordBotTokens(s.cfg, overrides)
	entries, participantTokens, boundByApp, fetchedProfiles, err := normalizeDiscordBotInputs(update.Participants, profileCache, tokenStore.Participants)
	if err != nil {
		return err
	}

	if err := validateDiscordBotExpertInputs(update.Participants); err != nil {
		return err
	}

	oldPrimary := effectivePrimaryBotID(overrides)
	newPrimary := strings.TrimSpace(update.ModeratorRoleID)
	if newPrimary == "" {
		newPrimary = oldPrimary
	}
	if err := validateModeratorRoleID(newPrimary, entries); err != nil {
		return err
	}

	next := s.cfg
	if err := applyDiscordBots(&next, entries); err != nil {
		return err
	}

	activeIDs := make([]string, 0, len(entries))
	for _, e := range entries {
		activeIDs = append(activeIDs, e.ApplicationID)
	}

	tokenStore = effectiveDiscordBotTokens(s.cfg, overrides)
	if newPrimary != oldPrimary {
		tokenStore = swapPrimaryBotTokens(tokenStore, oldPrimary, newPrimary)
	}
	tokenStore = mergeDiscordBotTokenUpdates(
		tokenStore,
		newPrimary,
		update.ModeratorToken,
		update.ModeratorRoleToken,
		participantTokens,
		activeIDs,
	)

	if err := validateUniqueDiscordBotTokens(update.Participants, tokenStore, newPrimary); err != nil {
		return err
	}

	bindings := s.effectiveParticipantIMBindingsLocked()
	seenBotBinding := make(map[string]bool, len(boundByApp))
	for appID, participantID := range boundByApp {
		seenBotBinding[appID] = true
		setDiscordBotBinding(bindings, appID, participantID)
	}
	for _, e := range entries {
		if !seenBotBinding[e.ApplicationID] {
			setDiscordBotBinding(bindings, e.ApplicationID, "")
		}
	}
	activeBotSet := discordApplicationIDSet(entries)
	for pid, binds := range bindings {
		next := binds[:0]
		for _, bind := range binds {
			appID := strings.TrimSpace(bind.ApplicationID)
			if bind.Platform == IMPlatformDiscord && appID != "" {
				if _, ok := activeBotSet[appID]; !ok {
					continue
				}
			}
			next = append(next, bind)
		}
		if len(next) == 0 {
			delete(bindings, pid)
		} else {
			bindings[pid] = next
		}
	}
	roster := ParticipantRosterFromConfig(next)
	if err := validateParticipantIMBindings(bindings, rosterIDSet(roster), activeBotSet); err != nil {
		return err
	}

	jsonVal := formatDiscordBotsJSON(entries)
	settingsUpdate := map[string]string{
		DiscordBotsSetting:             jsonVal,
		DiscordBotTokensSetting:        formatDiscordBotTokensJSON(tokenStore),
		DiscordModeratorRoleSetting:    newPrimary,
		ParticipantIMBindingsSetting: formatParticipantIMBindingsJSON(bindings),
	}

	for appID, profile := range fetchedProfiles {
		if profile.DiscordApplicationID != "" || profile.DiscordUsername != "" {
			profileCache[appID] = profile
		}
	}
	activeBotIDs := map[string]bool{ModeratorBotID: true}
	for _, e := range entries {
		activeBotIDs[e.ApplicationID] = true
	}
	pruneDiscordBotProfilesCache(profileCache, activeBotIDs)
	settingsUpdate[DiscordBotProfilesSetting] = formatDiscordBotProfilesJSON(profileCache)

	if s.store != nil {
		if err := s.store.SetSettings(ctx, settingsUpdate); err != nil {
			return err
		}
	}

	next.Secrets.DiscordBotToken = tokenStore.TokenForBot(newPrimary, newPrimary)
	next.Secrets.DiscordParticipantTokens = tokenStore.Participants
	next.Transport.Discord.ParticipantIMBindings = bindings
	s.cfg = next
	return nil
}

// RefreshDiscordBotProfiles fetches bot avatars from Discord and caches them in SQLite.
func (s *Service) RefreshDiscordBotProfiles(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	overrides := s.settingsOverridesLocked()
	states := buildDiscordBotStates(s.cfg, overrides)
	tokens := effectiveDiscordBotTokens(s.cfg, overrides)
	primaryID := effectivePrimaryBotID(overrides)
	fetched := fetchDiscordBotProfiles(states, tokens, primaryID)

	cache := discordBotProfilesFromOverrides(overrides)
	for id, profile := range fetched {
		cache[id] = profile
	}

	if s.store != nil {
		if err := s.store.SetSettings(ctx, map[string]string{
			DiscordBotProfilesSetting: formatDiscordBotProfilesJSON(cache),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) settingsOverridesLocked() map[string]string {
	if s.store == nil {
		return nil
	}
	overrides, err := s.store.GetAllSettings(context.Background())
	if err != nil {
		return nil
	}
	return overrides
}

// SettingsView builds the settings API response.
func (s *Service) SettingsView() SettingsResponse {
	cfg := s.Current()
	overrides := s.settingsOverrides()
	groups := make([]string, 0, 8)
	seen := make(map[string]bool)
	fields := make([]SettingsFieldState, 0, len(settingFields))

	for _, def := range settingFields {
		if def.Key == DiscordBotsSetting {
			continue
		}
		state := SettingsFieldState{
			Key:             def.Key,
			Secret:          def.Secret,
			Editable:        def.Editable,
			RestartRequired: def.RestartRequired,
			Label:           def.Label,
			Group:           def.Group,
			Subsection:      def.Subsection,
			Section:         def.Section,
			Placeholder:     def.Placeholder,
			Description:     def.Description,
			InputType:       def.InputType,
			Options:         def.Options,
			Min:             def.Min,
			Max:             def.Max,
		}
		if def.Secret {
			state.Configured = def.secretConfigured != nil && def.secretConfigured()
		} else if def.read != nil {
			state.Value = def.read(cfg)
			state.Configured = state.Value != ""
		}
		fields = append(fields, state)
		if !seen[def.Group] {
			seen[def.Group] = true
			groups = append(groups, def.Group)
		}
	}

	source := "yaml"
	if s.store != nil {
		source = "app_settings"
	}
	return SettingsResponse{
		Source:      source,
		SecretsPath: deployEnvPath(),
		Groups:      groups,
		Subsections: groupSubsections,
		Fields:      fields,
		DiscordBots: applyCachedDiscordBotProfiles(
			buildDiscordBotStates(cfg, overrides),
			discordBotProfilesFromOverrides(overrides),
		),
		MeetPresets:         cfg.Meeting.MeetPresets,
		MeetPresetsDefaults: DefaultMeetPresets(cfg),
	}
}
