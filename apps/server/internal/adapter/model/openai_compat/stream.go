package openai_compat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"round_table/apps/server/internal/adapter/model"
)

type chatStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
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

func readStreamResponse(body io.Reader, onDelta model.StreamHandler) (model.Response, error) {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	var content strings.Builder
	var usage model.Usage

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" {
			continue
		}

		var chunk chatStreamChunk
		if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
			return model.Response{}, fmt.Errorf("openai_compat: decode stream chunk: %w", err)
		}
		if chunk.Error != nil {
			return model.Response{}, fmt.Errorf("openai_compat: api error: %s", chunk.Error.Message)
		}
		if len(chunk.Choices) > 0 {
			delta := chunk.Choices[0].Delta.Content
			if delta != "" {
				content.WriteString(delta)
				if onDelta != nil {
					onDelta(delta)
				}
			}
		}
		if chunk.Usage != nil {
			usage = model.Usage{
				PromptTokens:     chunk.Usage.PromptTokens,
				CompletionTokens: chunk.Usage.CompletionTokens,
				TotalTokens:      chunk.Usage.TotalTokens,
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return model.Response{}, fmt.Errorf("openai_compat: read stream: %w", err)
	}
	if content.Len() == 0 {
		return model.Response{}, fmt.Errorf("openai_compat: empty stream content")
	}
	return model.Response{Content: strings.TrimSpace(content.String()), Usage: usage}, nil
}

func parseJSONResponse(data []byte) (model.Response, error) {
	var out chatResponse
	if err := json.Unmarshal(data, &out); err != nil {
		return model.Response{}, fmt.Errorf("openai_compat: decode response: %w", err)
	}
	if out.Error != nil {
		return model.Response{}, fmt.Errorf("openai_compat: api error: %s", out.Error.Message)
	}
	if len(out.Choices) == 0 {
		return model.Response{}, fmt.Errorf("openai_compat: empty choices")
	}
	result := model.Response{Content: strings.TrimSpace(out.Choices[0].Message.Content)}
	if out.Usage != nil {
		result.Usage = model.Usage{
			PromptTokens:     out.Usage.PromptTokens,
			CompletionTokens: out.Usage.CompletionTokens,
			TotalTokens:      out.Usage.TotalTokens,
		}
	}
	return result, nil
}

func decodeErrorBody(status int, data []byte) error {
	var out chatResponse
	if err := json.Unmarshal(data, &out); err == nil && out.Error != nil {
		return fmt.Errorf("openai_compat: http %d: %s", status, out.Error.Message)
	}
	return fmt.Errorf("openai_compat: http %d: %s", status, string(data))
}

func isEventStream(contentType string) bool {
	return strings.Contains(strings.ToLower(contentType), "text/event-stream")
}

func readNonStreamBody(data []byte) (model.Response, error) {
	return parseJSONResponse(data)
}

// readResponseBody picks SSE or JSON based on Content-Type (fallback for proxies).
func readResponseBody(contentType string, body io.Reader, onDelta model.StreamHandler) (model.Response, error) {
	if isEventStream(contentType) {
		return readStreamResponse(body, onDelta)
	}
	data, err := io.ReadAll(body)
	if err != nil {
		return model.Response{}, err
	}
	return readNonStreamBody(data)
}
