package model

import "context"

// Message is one chat turn for completion APIs.
type Message struct {
	Role    string
	Content string
}

// Request asks a model provider for a single completion.
type Request struct {
	Model       string
	Messages    []Message
	Temperature float64
}

// Port abstracts LLM providers (DeepSeek, OpenAI, Anthropic adapters).
type Port interface {
	Complete(ctx context.Context, req Request) (string, error)
}
