package main

import "testing"

func TestParseParticipantItem(t *testing.T) {
	tests := []struct {
		in          string
		id, role, exp string
	}{
		{"architect:Architect:design", "architect", "Architect", "design"},
		{"skeptic:Security Architect:security", "skeptic", "Security Architect", "security"},
		{"pragmatist:Tech Lead:delivery", "pragmatist", "Tech Lead", "delivery"},
		{"solo:Expert", "solo", "Expert", ""},
	}
	for _, tt := range tests {
		p, err := parseParticipantItem(tt.in)
		if err != nil {
			t.Fatalf("%q: %v", tt.in, err)
		}
		if p.ID != tt.id || p.Role != tt.role || p.Expertise != tt.exp {
			t.Fatalf("%q: got id=%q role=%q exp=%q", tt.in, p.ID, p.Role, p.Expertise)
		}
	}
}

func TestParseParticipants_multi(t *testing.T) {
	raw := "skeptic:Security Architect:security,pragmatist:Tech Lead:delivery"
	parts, err := parseParticipants(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 2 {
		t.Fatalf("got %d participants", len(parts))
	}
}
