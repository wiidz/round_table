package discord

import "testing"

func TestAcceptMessage(t *testing.T) {
	tests := []struct {
		name   string
		opts   Options
		guild  string
		want   bool
	}{
		{
			name:  "dm allowed",
			opts:  Options{AllowDM: true, AllowGuild: false},
			guild: "",
			want:  true,
		},
		{
			name:  "dm rejected",
			opts:  Options{AllowDM: false, AllowGuild: true},
			guild: "",
			want:  false,
		},
		{
			name:  "guild allowed",
			opts:  Options{AllowDM: false, AllowGuild: true},
			guild: "g1",
			want:  true,
		},
		{
			name:  "guild filtered",
			opts:  Options{AllowDM: false, AllowGuild: true, GuildID: "g1"},
			guild: "g2",
			want:  false,
		},
		{
			name:  "guild filter match",
			opts:  Options{AllowDM: false, AllowGuild: true, GuildID: "g1"},
			guild: "g1",
			want:  true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			b := &Bot{opts: tc.opts}
			if got := b.acceptMessage(tc.guild); got != tc.want {
				t.Fatalf("acceptMessage(%q) = %v, want %v", tc.guild, got, tc.want)
			}
		})
	}
}

func TestNew_requiresToken(t *testing.T) {
	if _, err := New(Options{AllowDM: true}); err == nil {
		t.Fatal("expected error for empty token")
	}
}
