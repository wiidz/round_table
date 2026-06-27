package sqlite_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"round_table/apps/server/internal/adapter/storage/sqlite"
	"round_table/apps/server/internal/domain/event"
	"round_table/apps/server/internal/domain/meeting"
	"round_table/apps/server/internal/platform/config"
)

func TestStore_AppendListAndMeetingsPage(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "test.db")
	st, err := sqlite.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	topic := "SQLite 会议测试"
	payload, _ := json.Marshal(event.MeetingCreatedPayload{Topic: topic})
	created := event.Envelope{
		ID:         "mtg-1-1",
		MeetingID:  "mtg-1",
		Sequence:   1,
		Type:       event.TypeMeetingCreated,
		Version:    1,
		Payload:    payload,
		OccurredAt: time.Now().UTC(),
		Actor:      event.ActorPrincipal,
	}
	if err := st.Append(ctx, created); err != nil {
		t.Fatalf("Append MeetingCreated: %v", err)
	}

	invitedPayload, _ := json.Marshal(event.ParticipantInvitedPayload{ParticipantID: "p1"})
	if err := st.Append(ctx, event.Envelope{
		ID: "mtg-1-2", MeetingID: "mtg-1", Sequence: 2,
		Type: event.TypeParticipantInvited, Version: 1, Payload: invitedPayload,
		OccurredAt: time.Now().UTC(), Actor: event.ActorModerator,
	}); err != nil {
		t.Fatalf("Append ParticipantInvited: %v", err)
	}

	list, err := st.List(ctx, "mtg-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 {
		t.Fatalf("List len = %d, want 2", len(list))
	}
	if list[0].Type != event.TypeMeetingCreated {
		t.Fatalf("first event type = %s", list[0].Type)
	}

	page, err := st.ListMeetingsPage(ctx, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || len(page.Meetings) != 1 {
		t.Fatalf("page total=%d meetings=%d", page.Total, len(page.Meetings))
	}
	if page.Meetings[0].Topic != topic {
		t.Fatalf("topic = %q", page.Meetings[0].Topic)
	}
	if page.Meetings[0].Status != string(meeting.StatusPreparing) {
		t.Fatalf("status = %q", page.Meetings[0].Status)
	}

	dup := created
	dup.ID = "mtg-1-1-dup"
	if err := st.Append(ctx, dup); err == nil {
		t.Fatal("expected duplicate sequence error")
	}
}

func TestStore_DuplicateSequenceRejected(t *testing.T) {
	ctx := context.Background()
	st, err := sqlite.Open(filepath.Join(t.TempDir(), "dup.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	env := event.Envelope{
		ID: "m-1", MeetingID: "m", Sequence: 1,
		Type: event.TypeMeetingCreated, Version: 1,
		Payload:    []byte(`{"topic":"x"}`),
		OccurredAt: time.Now().UTC(), Actor: event.ActorSystem,
	}
	if err := st.Append(ctx, env); err != nil {
		t.Fatal(err)
	}
	env.ID = "m-1b"
	if err := st.Append(ctx, env); err == nil {
		t.Fatal("want duplicate error")
	}
}

func TestStore_SetSettings_persistsDiscordBotProfiles(t *testing.T) {
	st, err := sqlite.Open(filepath.Join(t.TempDir(), "profiles.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	ctx := context.Background()
	cache := `{"moderator":{"discord_username":"mod","avatar_url":"https://cdn.example/a.png","fetched_at":"2026-01-01T00:00:00Z"}}`
	if err := st.SetSettings(ctx, map[string]string{
		config.DiscordBotProfilesSetting: cache,
	}); err != nil {
		t.Fatal(err)
	}
	all, err := st.GetAllSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if all[config.DiscordBotProfilesSetting] != cache {
		t.Fatalf("cache not persisted: %+v", all)
	}
}
