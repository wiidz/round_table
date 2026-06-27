package config

import (
	"strings"
	"testing"
)

func TestMeetingSettingValidation(t *testing.T) {
	cfg := defaults()

	tests := []struct {
		key string
		val string
		ok  bool
	}{
		{"ROUND_TABLE_LOCALE", "zh", true},
		{"ROUND_TABLE_LOCALE", "en", true},
		{"ROUND_TABLE_LOCALE", "fr", false},
		{"ROUND_TABLE_MAX_ROUNDS_PER_SEGMENT", "5", true},
		{"ROUND_TABLE_MAX_ROUNDS_PER_SEGMENT", "0", false},
		{"ROUND_TABLE_MAX_ROUNDS_PER_SEGMENT", "21", false},
		{"ROUND_TABLE_MIN_ROUNDS_BEFORE_SYNTHESIS", "2", true},
		{"ROUND_TABLE_MAX_CONFIRMATION_CYCLES", "3", true},
	}

	for _, tc := range tests {
		field := editableSettingFields()[tc.key]
		if field.apply == nil {
			t.Fatalf("unknown key %s", tc.key)
		}
		err := field.apply(&cfg, tc.val)
		if tc.ok && err != nil {
			t.Fatalf("%s=%q: unexpected error: %v", tc.key, tc.val, err)
		}
		if !tc.ok && err == nil {
			t.Fatalf("%s=%q: expected validation error", tc.key, tc.val)
		}
		if !tc.ok && err != nil && !strings.Contains(err.Error(), "：") {
			t.Fatalf("%s=%q: want friendly error, got %v", tc.key, tc.val, err)
		}
	}
}

func TestSettingsView_meetingFieldsHaveUXMetadata(t *testing.T) {
	svc, err := NewService(nil)
	if err != nil {
		t.Fatal(err)
	}
	resp := svc.SettingsView()

	want := map[string]struct {
		inputType string
		section   string
		min       int
		max       int
	}{
		"ROUND_TABLE_MAX_ROUNDS_PER_SEGMENT":      {inputType: "number", section: meetingSectionLimits, min: 1, max: 20},
		"ROUND_TABLE_MIN_ROUNDS_BEFORE_SYNTHESIS": {inputType: "number", section: meetingSectionLimits, min: 1, max: 20},
		"ROUND_TABLE_MAX_CONFIRMATION_CYCLES":     {inputType: "number", section: meetingSectionLimits, min: 1, max: 10},
	}

	seen := map[string]bool{}
	for _, f := range resp.Fields {
		exp, ok := want[f.Key]
		if !ok {
			continue
		}
		seen[f.Key] = true
		if f.InputType != exp.inputType {
			t.Fatalf("%s input_type = %q, want %q", f.Key, f.InputType, exp.inputType)
		}
		if f.Section != exp.section {
			t.Fatalf("%s section = %q, want %q", f.Key, f.Section, exp.section)
		}
		if exp.inputType == "number" {
			if f.Min == nil || *f.Min != exp.min {
				t.Fatalf("%s min = %v, want %d", f.Key, f.Min, exp.min)
			}
			if f.Max == nil || *f.Max != exp.max {
				t.Fatalf("%s max = %v, want %d", f.Key, f.Max, exp.max)
			}
		}
		if f.Description == "" {
			t.Fatalf("%s missing description", f.Key)
		}
	}
	for key := range want {
		if !seen[key] {
			t.Fatalf("field %s missing from settings view", key)
		}
	}
	if len(resp.MeetPresets) != 11 {
		t.Fatalf("meet_presets = %d, want 11", len(resp.MeetPresets))
	}
}

func TestNormalizeMeetPresets(t *testing.T) {
	cfg := defaults()
	presets, err := normalizeMeetPresets([]MeetPresetConfig{
		{ID: "1", Mode: "decision", MaxRounds: 3, Confirmation: "skip", FreeDialogueQuestions: 0, NameZH: "自定义默认", NameEN: "Custom default"},
	}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if presets[0].Mode != "decision" || presets[0].MaxRounds != 3 {
		t.Fatalf("preset 1 = %+v", presets[0])
	}
	if presets[0].NameZH != "自定义默认" || presets[0].NameEN != "Custom default" {
		t.Fatalf("preset 1 names = %+v", presets[0])
	}
}

func TestNormalizeMeetPresets_rejectsEmptyName(t *testing.T) {
	cfg := defaults()
	_, err := normalizeMeetPresets([]MeetPresetConfig{
		{ID: "2", Mode: "deliberation", MaxRounds: 1, Confirmation: "skip", FreeDialogueQuestions: 0, NameZH: "   "},
	}, cfg)
	if err == nil {
		t.Fatal("expected empty name error")
	}
}

func TestCombinePresetDisplayName_legacyIcon(t *testing.T) {
	got := combinePresetDisplayName("⚡", "闪电研讨")
	if got != "⚡ 闪电研讨" {
		t.Fatalf("got=%q", got)
	}
	if combinePresetDisplayName("", "⚡ 闪电研讨") != "⚡ 闪电研讨" {
		t.Fatal("already combined")
	}
}

func TestNormalizeMeetPresets_mergesLegacyIcon(t *testing.T) {
	cfg := defaults()
	presets, err := normalizeMeetPresets([]MeetPresetConfig{
		{ID: "2", Icon: "🌩️", NameZH: "闪电研讨", Mode: "deliberation", MaxRounds: 1, Confirmation: "skip", FreeDialogueQuestions: 0},
	}, cfg)
	if err != nil {
		t.Fatal(err)
	}
	if presets[1].NameZH != "🌩️ 闪电研讨" || presets[1].Icon != "" {
		t.Fatalf("preset 2 = %+v", presets[1])
	}
}

func TestDefaultMeetPresets_seedNames(t *testing.T) {
	presets := DefaultMeetPresets(defaults())
	want := map[string]string{
		"1": "⚡ 直接开始（默认）", "2": "🌩️ 闪电研讨", "J1": "🌩️ 闪电裁决", "J5": "🔬 深度裁决",
	}
	for id, name := range want {
		for _, p := range presets {
			if p.ID == id && p.NameZH != name {
				t.Fatalf("preset %s name_zh = %q want %q", id, p.NameZH, name)
			}
		}
	}
	if presets[0].Command != "1" {
		t.Fatalf("preset 1 command = %q", presets[0].Command)
	}
}

func TestNormalizeMeetPresets_rejectsDuplicateCommands(t *testing.T) {
	cfg := defaults()
	_, err := normalizeMeetPresets([]MeetPresetConfig{
		{ID: "1", NameZH: "A", Mode: "deliberation", MaxRounds: 3, Confirmation: "skip", FreeDialogueQuestions: 0, Command: "go"},
		{ID: "2", NameZH: "B", Mode: "deliberation", MaxRounds: 1, Confirmation: "skip", FreeDialogueQuestions: 0, Command: "go"},
	}, cfg)
	if err == nil {
		t.Fatal("expected duplicate command error")
	}
}

func TestNormalizeMeetPresets_rejectsReservedZero(t *testing.T) {
	cfg := defaults()
	_, err := normalizeMeetPresets([]MeetPresetConfig{
		{ID: "2", NameZH: "B", Mode: "deliberation", MaxRounds: 1, Confirmation: "skip", FreeDialogueQuestions: 0, Command: "0"},
	}, cfg)
	if err == nil {
		t.Fatal("expected reserved 0 error")
	}
}

func TestParseMeetPresetsJSON_legacyCommandsArray(t *testing.T) {
	presets, err := parseMeetPresetsJSON(`[{"id":"2","command":"","commands":["flash","extra"],"name_zh":"x","mode":"deliberation","max_rounds":1,"confirmation":"skip","free_dialogue_questions":0}]`)
	if err != nil {
		t.Fatal(err)
	}
	if presets[0].Command != "flash" {
		t.Fatalf("command = %q", presets[0].Command)
	}
}
