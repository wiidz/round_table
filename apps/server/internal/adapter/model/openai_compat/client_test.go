package openai_compat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"round_table/apps/server/internal/adapter/model"
)

func TestClient_Complete_stream(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if !req.Stream {
			t.Fatal("expected stream=true")
		}
		if req.StreamOptions == nil || !req.StreamOptions.IncludeUsage {
			t.Fatal("expected stream_options.include_usage")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		chunks := []string{
			`{"choices":[{"delta":{"content":"he"},"index":0}]}`,
			`{"choices":[{"delta":{"content":"llo"},"index":0}]}`,
			`{"choices":[{"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":11,"completion_tokens":3,"total_tokens":14}}`,
		}
		for _, c := range chunks {
			fmt.Fprintf(w, "data: %s\n\n", c)
			if ok {
				flusher.Flush()
			}
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer srv.Close()

	var deltas strings.Builder
	c := NewClient(srv.URL, "sk-test", 0)
	got, err := c.Complete(context.Background(), model.Request{
		Model: "deepseek-chat",
		Messages: []model.Message{
			{Role: "user", Content: "hi"},
		},
		OnDelta: func(d string) { deltas.WriteString(d) },
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Content != "hello" {
		t.Fatalf("content = %q", got.Content)
	}
	if got.Usage.TotalTokens != 14 {
		t.Fatalf("usage total = %d", got.Usage.TotalTokens)
	}
	if deltas.String() != "hello" {
		t.Fatalf("deltas = %q", deltas.String())
	}
}

func TestClient_Complete_apiError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(chatResponse{
			Error: &struct {
				Message string `json:"message"`
			}{Message: "invalid key"},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "bad", 0)
	_, err := c.Complete(context.Background(), model.Request{Model: "m", Messages: []model.Message{{Role: "user", Content: "x"}}})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadStreamResponse(t *testing.T) {
	body := strings.NewReader(
		"data: {\"choices\":[{\"delta\":{\"content\":\"a\"}}]}\n\n" +
			"data: {\"choices\":[{\"delta\":{\"content\":\"b\"}}]}\n\n" +
			"data: {\"choices\":[{\"delta\":{}}],\"usage\":{\"prompt_tokens\":1,\"completion_tokens\":2,\"total_tokens\":3}}\n\n" +
			"data: [DONE]\n\n",
	)
	got, err := readStreamResponse(body, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.Content != "ab" {
		t.Fatalf("content = %q", got.Content)
	}
	if got.Usage.TotalTokens != 3 {
		t.Fatalf("usage = %d", got.Usage.TotalTokens)
	}
}
