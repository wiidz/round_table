package openai_compat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"round_table/apps/server/internal/adapter/model"
)

// Client calls OpenAI-compatible chat/completions endpoints (DeepSeek, OpenAI, etc.).
type Client struct {
	BaseURL    string
	APIKey     string
	HTTPClient *http.Client
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	Stream      bool          `json:"stream"`
	StreamOptions *streamOptions `json:"stream_options,omitempty"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// NewClient returns a client with the given base URL and API key.
func NewClient(baseURL, apiKey string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		BaseURL: baseURL,
		APIKey:  apiKey,
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
	}
}

var _ model.Port = (*Client)(nil)

// Complete implements model.Port using SSE streaming (OpenAI-compatible stream=true).
func (c *Client) Complete(ctx context.Context, req model.Request) (model.Response, error) {
	if c.APIKey == "" {
		return model.Response{}, fmt.Errorf("openai_compat: api key required")
	}
	msgs := make([]chatMessage, len(req.Messages))
	for i, m := range req.Messages {
		msgs[i] = chatMessage{Role: m.Role, Content: m.Content}
	}
	body, err := json.Marshal(chatRequest{
		Model:       req.Model,
		Messages:    msgs,
		Temperature: req.Temperature,
		Stream:      true,
		StreamOptions: &streamOptions{IncludeUsage: true},
	})
	if err != nil {
		return model.Response{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return model.Response{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.HTTPClient.Do(httpReq)
	if err != nil {
		return model.Response{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		data, _ := io.ReadAll(resp.Body)
		return model.Response{}, decodeErrorBody(resp.StatusCode, data)
	}

	return readResponseBody(resp.Header.Get("Content-Type"), resp.Body, req.OnDelta)
}
