package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type ProviderConfig struct {
	BaseURL string `toml:"base_url"`
	APIKey  string `toml:"api_key"`
	Model   string `toml:"model"`
}

type ConnectionConfig struct {
	Type     string   `toml:"type"`
	Host     string   `toml:"host"`
	Port     int      `toml:"port"`
	Username string   `toml:"username"`
	Password string   `toml:"password"`
	Database string   `toml:"database"`
	Schema   string   `toml:"schema"`
	DSN      string   `toml:"dsn"`
	Include  []string `toml:"include"`
	Exclude  []string `toml:"exclude"`
}

type UIConfig struct {
	Stream  bool   `toml:"stream"`
	Color   string `toml:"color"`
	MaxRows int    `toml:"max_rows"`
}

type Config struct {
	DefaultProvider   string                      `toml:"default_provider"`
	DefaultConnection string                      `toml:"default_connection"`
	Providers         map[string]ProviderConfig   `toml:"providers"`
	Connections       map[string]ConnectionConfig `toml:"connections"`
	UI                UIConfig                    `toml:"ui"`
}

func NewDefault() *Config {
	return &Config{
		Providers:   make(map[string]ProviderConfig),
		Connections: make(map[string]ConnectionConfig),
		UI:          UIConfig{Stream: true, Color: "auto", MaxRows: 100},
	}
}

func Load(configDir string) (*Config, error) {
	path := filepath.Join(configDir, "config.toml")
	cfg := NewDefault()

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil
	}
	if _, err := toml.DecodeFile(path, cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	normalize(cfg)
	applyEnvOverrides(cfg)
	return cfg, nil
}

func Save(cfg *Config, configDir string) error {
	if cfg == nil {
		cfg = NewDefault()
	}
	normalize(cfg)
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}
	path := filepath.Join(configDir, "config.toml")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("打开配置文件失败: %w", err)
	}
	defer f.Close()
	return toml.NewEncoder(f).Encode(cfg)
}

func normalize(cfg *Config) {
	if cfg.Providers == nil {
		cfg.Providers = make(map[string]ProviderConfig)
	}
	if cfg.Connections == nil {
		cfg.Connections = make(map[string]ConnectionConfig)
	}
	if cfg.UI.Color == "" {
		cfg.UI.Color = "auto"
	}
	if cfg.UI.MaxRows <= 0 {
		cfg.UI.MaxRows = 100
	}
}

func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("AIDBT_PROVIDER"); v != "" {
		cfg.DefaultProvider = v
	}
	if v := os.Getenv("AIDBT_CONNECTION"); v != "" {
		cfg.DefaultConnection = v
	}
	for name, pc := range cfg.Providers {
		key := envName(name)
		if v := os.Getenv("AIDBT_" + key + "_API_KEY"); v != "" {
			pc.APIKey = v
		}
		if v := os.Getenv("AIDBT_" + key + "_MODEL"); v != "" {
			pc.Model = v
		}
		if v := os.Getenv("AIDBT_" + key + "_BASE_URL"); v != "" {
			pc.BaseURL = v
		}
		cfg.Providers[name] = pc
	}
	for name, cc := range cfg.Connections {
		key := envName(name)
		if v := os.Getenv("AIDBT_" + key + "_PASSWORD"); v != "" {
			cc.Password = v
		}
		if v := os.Getenv("AIDBT_" + key + "_DSN"); v != "" {
			cc.DSN = v
		}
		cfg.Connections[name] = cc
	}
}

func (c *Config) CurrentProvider() (string, ProviderConfig, error) {
	if c == nil {
		return "", ProviderConfig{}, fmt.Errorf("配置为空")
	}
	pc, ok := c.Providers[c.DefaultProvider]
	if !ok {
		return "", ProviderConfig{}, fmt.Errorf("provider %q 未配置", c.DefaultProvider)
	}
	return c.DefaultProvider, pc, nil
}

func (c *Config) CurrentConnection() (string, ConnectionConfig, error) {
	if c == nil {
		return "", ConnectionConfig{}, fmt.Errorf("配置为空")
	}
	cc, ok := c.Connections[c.DefaultConnection]
	if !ok {
		return "", ConnectionConfig{}, fmt.Errorf("连接 %q 未配置", c.DefaultConnection)
	}
	return c.DefaultConnection, cc, nil
}

func MaskConfig(cfg *Config) *Config {
	if cfg == nil {
		return nil
	}
	out := *cfg
	out.Providers = make(map[string]ProviderConfig, len(cfg.Providers))
	for name, pc := range cfg.Providers {
		pc.APIKey = MaskSecret(pc.APIKey)
		out.Providers[name] = pc
	}
	out.Connections = make(map[string]ConnectionConfig, len(cfg.Connections))
	for name, cc := range cfg.Connections {
		cc.Password = MaskSecret(cc.Password)
		cc.DSN = MaskDSN(cc.DSN)
		out.Connections[name] = cc
	}
	return &out
}

func MaskSecret(secret string) string {
	if secret == "" {
		return ""
	}
	if len(secret) <= 8 {
		return "***"
	}
	return secret[:4] + "..." + secret[len(secret)-4:]
}

func MaskDSN(dsn string) string {
	if dsn == "" {
		return ""
	}
	at := strings.LastIndex(dsn, "@")
	colon := strings.LastIndex(dsn[:max(0, at)], ":")
	if at > 0 && colon > 0 {
		return dsn[:colon+1] + "***" + dsn[at:]
	}
	return "***"
}

func envName(name string) string {
	replacer := strings.NewReplacer("-", "_", ".", "_", " ", "_")
	return strings.ToUpper(replacer.Replace(name))
}
