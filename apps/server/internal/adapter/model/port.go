package model

import "context"

// Message is one chat turn for completion APIs.
type Message struct {
	Role    string
	Content string
}

// Usage is token consumption reported by the provider.
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// Request asks a model provider for a single completion.
type Request struct {
	Model       string
	Messages    []Message
	Temperature float64
}

// Response is a completion result with optional usage stats.
type Response struct {
	Content string
	Usage   Usage
}

// Port abstracts LLM providers (DeepSeek, OpenAI, Anthropic adapters).
type Port interface {
	Complete(ctx context.Context, req Request) (Response, error)
}
