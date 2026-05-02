package provider

import (
	"fmt"
	"strings"

	"github.com/aidbt-tool/aidbt/internal/config"
)

func Build(name string, pc config.ProviderConfig) (Provider, error) {
	switch strings.ToLower(name) {
	case "claude":
		baseURL := pc.BaseURL
		if baseURL == "" {
			baseURL = "https://api.anthropic.com"
		}
		return &ClaudeProvider{BaseURL: strings.TrimRight(baseURL, "/"), APIKey: pc.APIKey, Model: pc.Model}, nil
	case "openai":
		baseURL := pc.BaseURL
		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}
		return &OpenAIProvider{BaseURL: strings.TrimRight(baseURL, "/"), APIKey: pc.APIKey, Model: pc.Model, ProviderID: "openai"}, nil
	case "deepseek":
		baseURL := pc.BaseURL
		if baseURL == "" {
			baseURL = "https://api.deepseek.com/v1"
		}
		return &OpenAIProvider{BaseURL: strings.TrimRight(baseURL, "/"), APIKey: pc.APIKey, Model: pc.Model, ProviderID: "deepseek"}, nil
	case "ollama":
		baseURL := pc.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434/v1"
		}
		return &OpenAIProvider{BaseURL: strings.TrimRight(baseURL, "/"), APIKey: pc.APIKey, Model: pc.Model, ProviderID: "ollama"}, nil
	default:
		if pc.BaseURL == "" {
			return nil, fmt.Errorf("provider %q 需要配置 base_url", name)
		}
		return &OpenAIProvider{BaseURL: strings.TrimRight(pc.BaseURL, "/"), APIKey: pc.APIKey, Model: pc.Model, ProviderID: name}, nil
	}
}
