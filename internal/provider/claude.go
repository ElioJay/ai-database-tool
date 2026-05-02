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

type ClaudeProvider struct {
	BaseURL string
	APIKey  string
	Model   string
}

func (p *ClaudeProvider) Name() string { return "claude" }

func (p *ClaudeProvider) Stream(ctx context.Context, msgs []Message) (<-chan Chunk, error) {
	system := ""
	var apiMsgs []map[string]string
	for _, msg := range msgs {
		if msg.Role == "system" {
			system = msg.Content
			continue
		}
		role := msg.Role
		if role != "assistant" {
			role = "user"
		}
		apiMsgs = append(apiMsgs, map[string]string{"role": role, "content": msg.Content})
	}
	body, _ := json.Marshal(map[string]any{
		"model":      p.Model,
		"max_tokens": 4096,
		"stream":     true,
		"system":     system,
		"messages":   apiMsgs,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.BaseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", "2023-06-01")
	if p.APIKey != "" {
		req.Header.Set("x-api-key", p.APIKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		defer resp.Body.Close()
		return nil, fmt.Errorf("Claude API 返回 HTTP %d", resp.StatusCode)
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
			var event struct {
				Type  string `json:"type"`
				Delta struct {
					Text string `json:"text"`
				} `json:"delta"`
			}
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event); err != nil {
				continue
			}
			if event.Type == "content_block_delta" && event.Delta.Text != "" {
				ch <- Chunk{Delta: event.Delta.Text}
			}
			if event.Type == "message_stop" {
				ch <- Chunk{Done: true}
				return
			}
		}
		if err := scanner.Err(); err != nil {
			ch <- Chunk{Err: err}
		}
	}()
	return ch, nil
}
