package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSaveLoadAndMaskConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		DefaultProvider:   "openai",
		DefaultConnection: "dev",
		Providers: map[string]ProviderConfig{
			"openai": {BaseURL: "https://api.openai.com/v1", APIKey: "sk-1234567890", Model: "gpt-4o"},
		},
		Connections: map[string]ConnectionConfig{
			"dev": {Type: "mysql", Host: "127.0.0.1", Port: 3306, Username: "root", Password: "secret-value", Database: "app"},
		},
		UI: UIConfig{Stream: true, Color: "auto", MaxRows: 100},
	}

	if err := Save(cfg, dir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.DefaultProvider != "openai" || loaded.DefaultConnection != "dev" {
		t.Fatalf("loaded defaults = %q/%q", loaded.DefaultProvider, loaded.DefaultConnection)
	}
	if loaded.Connections["dev"].Password != "secret-value" {
		t.Fatalf("password was not preserved")
	}
	if got := MaskSecret("secret-value"); got != "secr...alue" {
		t.Fatalf("MaskSecret() = %q", got)
	}
	if got := MaskConfig(loaded).Connections["dev"].Password; got != "secr...alue" {
		t.Fatalf("masked connection password = %q", got)
	}
	if _, err := os.Stat(filepath.Join(dir, "config.toml")); err != nil {
		t.Fatalf("config.toml missing: %v", err)
	}
}

func TestLoadAppliesEnvironmentOverrides(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{
		DefaultProvider:   "openai",
		DefaultConnection: "dev",
		Providers: map[string]ProviderConfig{
			"openai": {BaseURL: "https://api.openai.com/v1", APIKey: "old-key", Model: "old-model"},
		},
		Connections: map[string]ConnectionConfig{
			"dev": {Type: "mysql", Host: "127.0.0.1", Port: 3306, Username: "root", Password: "old-pass", Database: "app"},
		},
	}
	if err := Save(cfg, dir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	t.Setenv("AIDBT_PROVIDER", "openai")
	t.Setenv("AIDBT_CONNECTION", "dev")
	t.Setenv("AIDBT_OPENAI_API_KEY", "new-key")
	t.Setenv("AIDBT_OPENAI_MODEL", "new-model")
	t.Setenv("AIDBT_DEV_PASSWORD", "new-pass")

	loaded, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if loaded.Providers["openai"].APIKey != "new-key" {
		t.Fatalf("API key override not applied")
	}
	if loaded.Providers["openai"].Model != "new-model" {
		t.Fatalf("model override not applied")
	}
	if loaded.Connections["dev"].Password != "new-pass" {
		t.Fatalf("password override not applied")
	}
}
