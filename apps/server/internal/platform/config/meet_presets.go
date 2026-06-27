package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

const MeetPresetsSetting = "ROUND_TABLE_MEET_PRESETS"

// MeetPresetConfig is one Discord preset menu entry (1–6, J1–J5).
type MeetPresetConfig struct {
	ID                    string   `json:"id"`
	Group                 string   `json:"group"` // deliberation | decision
	Icon                  string   `json:"icon"`
	NameZH                string   `json:"name_zh"`
	NameEN                string   `json:"name_en"`
	Mode                  string   `json:"mode"`
	MaxRounds             int      `json:"max_rounds"`
	Confirmation          string   `json:"confirmation"`
	FreeDialogueQuestions int    `json:"free_dialogue_questions"`
	Command               string `json:"command,omitempty"`
}

func defaultPresetCommand(id string) string {
	return id
}

func DefaultMeetPresets(cfg Config) []MeetPresetConfig {
	rounds := cfg.Meeting.MaxRoundsPerSegment
	if rounds <= 0 {
		rounds = 5
	}
	confirm := cfg.Meeting.ConfirmationMode
	if confirm == "" {
		confirm = "required"
	}
	free := cfg.Meeting.FreeDialogueMaxQuestions
	if free < 0 {
		free = 1
	}
	mode := cfg.Meeting.DefaultMode
	if mode == "" {
		mode = "deliberation"
	}

	return []MeetPresetConfig{
		{ID: "1", Group: "deliberation", NameZH: "⚡ 直接开始（默认）", NameEN: "Start now (default)", Mode: mode, MaxRounds: rounds, Confirmation: confirm, FreeDialogueQuestions: free, Command: defaultPresetCommand("1")},
		{ID: "2", Group: "deliberation", NameZH: "🌩️ 闪电研讨", NameEN: "Flash", Mode: "deliberation", MaxRounds: 1, Confirmation: "skip", FreeDialogueQuestions: 0, Command: defaultPresetCommand("2")},
		{ID: "3", Group: "deliberation", NameZH: "📐 标准研讨", NameEN: "Standard", Mode: "deliberation", MaxRounds: 3, Confirmation: "skip", FreeDialogueQuestions: 0, Command: defaultPresetCommand("3")},
		{ID: "4", Group: "deliberation", NameZH: "💬 研讨 + 自由对话", NameEN: "Deliberation + Q&A", Mode: "deliberation", MaxRounds: 3, Confirmation: "skip", FreeDialogueQuestions: 1, Command: defaultPresetCommand("4")},
		{ID: "5", Group: "deliberation", NameZH: "✅ 研讨 + 需确认", NameEN: "Deliberation + review", Mode: "deliberation", MaxRounds: 3, Confirmation: "required", FreeDialogueQuestions: 1, Command: defaultPresetCommand("5")},
		{ID: "6", Group: "deliberation", NameZH: "🔬 深度研讨", NameEN: "Deep", Mode: "deliberation", MaxRounds: 5, Confirmation: "required", FreeDialogueQuestions: 1, Command: defaultPresetCommand("6")},
		{ID: "J1", Group: "decision", NameZH: "🌩️ 闪电裁决", NameEN: "Flash", Mode: "decision", MaxRounds: 1, Confirmation: "skip", FreeDialogueQuestions: 0, Command: defaultPresetCommand("J1")},
		{ID: "J2", Group: "decision", NameZH: "⚡ 快速裁决", NameEN: "Quick", Mode: "decision", MaxRounds: 2, Confirmation: "skip", FreeDialogueQuestions: 0, Command: defaultPresetCommand("J2")},
		{ID: "J3", Group: "decision", NameZH: "📋 标准裁决", NameEN: "Standard", Mode: "decision", MaxRounds: 3, Confirmation: "skip", FreeDialogueQuestions: 0, Command: defaultPresetCommand("J3")},
		{ID: "J4", Group: "decision", NameZH: "✅ 裁决 + 需确认", NameEN: "Decision + review", Mode: "decision", MaxRounds: 3, Confirmation: "required", FreeDialogueQuestions: 1, Command: defaultPresetCommand("J4")},
		{ID: "J5", Group: "decision", NameZH: "🔬 深度裁决", NameEN: "Deep", Mode: "decision", MaxRounds: 5, Confirmation: "required", FreeDialogueQuestions: 1, Command: defaultPresetCommand("J5")},
	}
}

// PresetMenuNameZH returns the Discord menu label (emoji + title in one string).
func PresetMenuNameZH(p MeetPresetConfig) string {
	return combinePresetDisplayName(p.Icon, p.NameZH)
}

// combinePresetDisplayName merges legacy separate icon + name into one menu label.
func combinePresetDisplayName(icon, name string) string {
	name = strings.TrimSpace(name)
	icon = strings.TrimSpace(icon)
	if name == "" {
		return icon
	}
	if icon == "" || strings.HasPrefix(name, icon) {
		return name
	}
	return icon + " " + name
}

func formatMeetPresetsJSON(presets []MeetPresetConfig) string {
	b, _ := json.Marshal(presets)
	return string(b)
}

func parseMeetPresetsJSON(raw string) ([]MeetPresetConfig, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	type row struct {
		MeetPresetConfig
		LegacyCommands []string `json:"commands"`
	}
	var rows []row
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		return nil, fmt.Errorf("meet presets: invalid json: %w", err)
	}
	out := make([]MeetPresetConfig, 0, len(rows))
	for _, r := range rows {
		p := r.MeetPresetConfig
		if strings.TrimSpace(p.Command) == "" && len(r.LegacyCommands) > 0 {
			p.Command = strings.TrimSpace(r.LegacyCommands[0])
		}
		out = append(out, p)
	}
	return out, nil
}

func validateMeetPreset(p MeetPresetConfig, maxRoundsCap int) error {
	if strings.TrimSpace(p.ID) == "" {
		return fmt.Errorf("预设缺少 id")
	}
	if _, err := parseOneOfSetting(p.Mode, []string{"deliberation", "decision"}, "会议模式"); err != nil {
		return fmt.Errorf("预设 %s: %w", p.ID, err)
	}
	if _, err := parseConfirmationModeSetting(p.Confirmation); err != nil {
		return fmt.Errorf("预设 %s: %w", p.ID, err)
	}
	cap := maxRoundsCap
	if cap <= 0 {
		cap = 20
	}
	if p.MaxRounds < 1 || p.MaxRounds > cap {
		return fmt.Errorf("预设 %s: 辩论轮次需在 1–%d 之间", p.ID, cap)
	}
	if p.FreeDialogueQuestions < 0 || p.FreeDialogueQuestions > 5 {
		return fmt.Errorf("预设 %s: 自由对话轮数需在 0–5 之间", p.ID)
	}
	if strings.TrimSpace(p.NameZH) == "" {
		return fmt.Errorf("预设 %s: 菜单名称不能为空", p.ID)
	}
	if len([]rune(strings.TrimSpace(p.NameZH))) > 40 {
		return fmt.Errorf("预设 %s: 菜单名称不能超过 40 个字符", p.ID)
	}
	if en := strings.TrimSpace(p.NameEN); en != "" && len([]rune(en)) > 48 {
		return fmt.Errorf("预设 %s: 英文名称不能超过 48 个字符", p.ID)
	}
	cmd := strings.TrimSpace(p.Command)
	if cmd == "" {
		return fmt.Errorf("预设 %s: 绑定指令不能为空", p.ID)
	}
	if len([]rune(cmd)) > 24 {
		return fmt.Errorf("预设 %s: 绑定指令不能超过 24 个字符", p.ID)
	}
	key := NormalizePresetCommandKey(cmd)
	if key == "0" {
		return fmt.Errorf("预设 %s: 指令 %q 为系统保留（0 = 自定义）", p.ID, cmd)
	}
	return nil
}

// NormalizePresetCommandKey canonicalizes a preset command for lookup and uniqueness checks.
func NormalizePresetCommandKey(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = NormalizePresetASCIIForms(s)
	s = strings.ReplaceAll(s, " ", "")
	if isASCIICommand(s) {
		return strings.ToUpper(s)
	}
	return s
}

func isASCIICommand(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
}

// NormalizePresetASCIIForms maps full-width digits and J to ASCII forms.
func NormalizePresetASCIIForms(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '０', '0':
			b.WriteByte('0')
		case '１', '1':
			b.WriteByte('1')
		case '２', '2':
			b.WriteByte('2')
		case '３', '3':
			b.WriteByte('3')
		case '４', '4':
			b.WriteByte('4')
		case '５', '5':
			b.WriteByte('5')
		case '６', '6':
			b.WriteByte('6')
		case '７', '7':
			b.WriteByte('7')
		case '８', '8':
			b.WriteByte('8')
		case '９', '9':
			b.WriteByte('9')
		case 'Ｊ', 'J', 'j':
			b.WriteByte('J')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func validatePresetCommandsUnique(presets []MeetPresetConfig) error {
	seen := make(map[string]string, len(presets))
	for _, p := range presets {
		key := NormalizePresetCommandKey(p.Command)
		if prev, ok := seen[key]; ok {
			return fmt.Errorf("指令 %q 与预设 %s 冲突（已绑定到 %s）", p.Command, p.ID, prev)
		}
		seen[key] = p.ID
	}
	return nil
}

func normalizeMeetPresets(in []MeetPresetConfig, cfg Config) ([]MeetPresetConfig, error) {
	defaults := DefaultMeetPresets(cfg)
	byID := make(map[string]MeetPresetConfig, len(defaults))
	for _, d := range defaults {
		byID[d.ID] = d
	}
	cap := 20
	if cfg.Meeting.MaxRoundsPerSegment > 0 && cfg.Meeting.MaxRoundsPerSegment < cap {
		cap = cfg.Meeting.MaxRoundsPerSegment
	}

	out := make([]MeetPresetConfig, 0, len(defaults))
	for _, p := range in {
		base, ok := byID[p.ID]
		if !ok {
			return nil, fmt.Errorf("未知预设 id: %q", p.ID)
		}
		merged := base
		merged.Mode = p.Mode
		merged.MaxRounds = p.MaxRounds
		merged.Confirmation = p.Confirmation
		merged.FreeDialogueQuestions = p.FreeDialogueQuestions
		zh := combinePresetDisplayName(p.Icon, p.NameZH)
		if zh == "" {
			return nil, fmt.Errorf("预设 %s: 菜单名称不能为空", p.ID)
		}
		merged.NameZH = zh
		merged.Icon = ""
		if en := strings.TrimSpace(p.NameEN); en != "" {
			merged.NameEN = en
		}
		cmd := strings.TrimSpace(p.Command)
		if cmd == "" {
			cmd = base.Command
		}
		merged.Command = cmd
		if err := validateMeetPreset(merged, cap); err != nil {
			return nil, err
		}
		byID[p.ID] = merged
	}
	for _, d := range defaults {
		out = append(out, byID[d.ID])
	}
	if err := validatePresetCommandsUnique(out); err != nil {
		return nil, err
	}
	return out, nil
}

func meetPresetsFromOverrides(overrides map[string]string, cfg Config) []MeetPresetConfig {
	if overrides == nil {
		return DefaultMeetPresets(cfg)
	}
	raw, ok := overrides[MeetPresetsSetting]
	if !ok || strings.TrimSpace(raw) == "" {
		return DefaultMeetPresets(cfg)
	}
	parsed, err := parseMeetPresetsJSON(raw)
	if err != nil || len(parsed) == 0 {
		return DefaultMeetPresets(cfg)
	}
	normalized, err := normalizeMeetPresets(parsed, cfg)
	if err != nil {
		return DefaultMeetPresets(cfg)
	}
	return normalized
}

func applyMeetPresetsToConfig(cfg *Config, presets []MeetPresetConfig) {
	cfg.Meeting.MeetPresets = presets
}
