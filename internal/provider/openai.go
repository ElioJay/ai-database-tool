package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type OpenAIProvider struct {
	BaseURL    string
	APIKey     string
	Model      string
	ProviderID string
}

func (p *OpenAIProvider) Name() string {
	if p.ProviderID != "" {
		return p.ProviderID
	}
	return "openai"
}

func (p *OpenAIProvider) Stream(ctx context.Context, msgs []Message) (<-chan Chunk, error) {
	var apiMsgs []map[string]string
	for _, msg := range msgs {
		apiMsgs = append(apiMsgs, map[string]string{"role": msg.Role, "content": msg.Content})
	}
	body, _ := json.Marshal(map[string]any{
		"model":       p.Model,
		"stream":      true,
		"temperature": 0,
		"messages":    apiMsgs,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if p.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+p.APIKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, fmt.Errorf("%s API 返回 HTTP %d", p.Name(), resp.StatusCode)
	}

	ch := make(chan Chunk, 8)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			raw := strings.TrimPrefix(line, "data: ")
			if raw == "[DONE]" {
				ch <- Chunk{Done: true}
				return
			}
			var event struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}
			if err := json.Unmarshal([]byte(raw), &event); err != nil {
				continue
			}
			if len(event.Choices) > 0 && event.Choices[0].Delta.Content != "" {
				ch <- Chunk{Delta: event.Choices[0].Delta.Content}
			}
		}
		if err := scanner.Err(); err != nil {
			ch <- Chunk{Err: err}
		}
	}()
	return ch, nil
}
