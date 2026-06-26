package openai_compat

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"round_table/apps/server/internal/adapter/model"
)

func TestClient_Complete(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer sk-test" {
			t.Fatalf("auth = %q", got)
		}
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		if req.Model != "deepseek-chat" {
			t.Fatalf("model = %q", req.Model)
		}
		_ = json.NewEncoder(w).Encode(chatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{{Message: struct {
				Content string `json:"content"`
			}{Content: "hello"}}},
			Usage: &struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{PromptTokens: 11, CompletionTokens: 3, TotalTokens: 14},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "sk-test", 0)
	got, err := c.Complete(context.Background(), model.Request{
		Model: "deepseek-chat",
		Messages: []model.Message{
			{Role: "user", Content: "hi"},
		},
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
