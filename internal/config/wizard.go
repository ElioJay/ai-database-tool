package config

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type WizardResult struct {
	ConfigDir ConfigDir
	Config    *Config
}

var providerDefaults = map[string]ProviderConfig{
	"claude":   {BaseURL: "https://api.anthropic.com", Model: "claude-sonnet-4-6"},
	"openai":   {BaseURL: "https://api.openai.com/v1", Model: "gpt-4o"},
	"deepseek": {BaseURL: "https://api.deepseek.com/v1", Model: "deepseek-chat"},
	"ollama":   {BaseURL: "http://localhost:11434/v1", Model: "qwen2.5-coder:7b"},
}

func RunWizard() (*WizardResult, error) {
	cd := Resolve()
	cfg := NewDefault()
	scanner, readLine := newScanner()

	fmt.Println("欢迎使用 aidbt！我们来完成首次配置。")
	fmt.Println()
	providerName, pc := askProvider(scanner, readLine)
	cfg.DefaultProvider = providerName
	cfg.Providers[providerName] = pc

	fmt.Println()
	connName, conn := askConnection(scanner, readLine, "dev")
	cfg.DefaultConnection = connName
	cfg.Connections[connName] = conn

	if err := Save(cfg, cd.Path); err != nil {
		return nil, err
	}
	fmt.Printf("\n配置已保存到 %s\n", cd.Path)
	return &WizardResult{ConfigDir: cd, Config: cfg}, nil
}

func AddConnection(cd ConfigDir) (*Config, error) {
	cfg, err := Load(cd.Path)
	if err != nil {
		return nil, err
	}
	scanner, readLine := newScanner()
	name, conn := askConnection(scanner, readLine, "dev")
	if _, exists := cfg.Connections[name]; exists {
		fmt.Printf("连接 %q 已存在，是否覆盖？[y/N] ", name)
		scanner.Scan()
		if !isYes(scanner.Text()) {
			fmt.Println("已取消。")
			return cfg, nil
		}
	}
	cfg.Connections[name] = conn
	if cfg.DefaultConnection == "" {
		cfg.DefaultConnection = name
	} else {
		fmt.Printf("是否将 %s 设为默认连接？[y/N] ", name)
		scanner.Scan()
		if isYes(scanner.Text()) {
			cfg.DefaultConnection = name
		}
	}
	if err := Save(cfg, cd.Path); err != nil {
		return nil, err
	}
	fmt.Printf("连接 %q 已保存。\n", name)
	return cfg, nil
}

func EditConnection(cd ConfigDir, name string) (*Config, error) {
	cfg, err := Load(cd.Path)
	if err != nil {
		return nil, err
	}
	if name == "" {
		name = cfg.DefaultConnection
	}
	old, ok := cfg.Connections[name]
	if !ok {
		return nil, fmt.Errorf("连接 %q 未配置", name)
	}
	_, readLine := newScanner()
	fmt.Printf("正在编辑连接 %q，直接回车保留当前值。\n", name)
	old.Type = readLine("数据库类型 mysql/oracle/dm", old.Type)
	old.Host = readLine("Host", old.Host)
	old.Port = readInt(readLine("Port", strconv.Itoa(old.Port)), old.Port)
	old.Username = readLine("用户名", old.Username)
	if newPass := readLine("密码（留空保留当前密码）", ""); newPass != "" {
		old.Password = newPass
	}
	old.Database = readLine("数据库/服务名", old.Database)
	old.Schema = readLine("Schema（可空）", old.Schema)
	cfg.Connections[name] = old
	if err := Save(cfg, cd.Path); err != nil {
		return nil, err
	}
	fmt.Printf("连接 %q 已更新。\n", name)
	return cfg, nil
}

func DeleteConnection(cd ConfigDir, name string) (*Config, error) {
	cfg, err := Load(cd.Path)
	if err != nil {
		return nil, err
	}
	if name == "" {
		name = cfg.DefaultConnection
	}
	if _, ok := cfg.Connections[name]; !ok {
		return nil, fmt.Errorf("连接 %q 未配置", name)
	}
	if len(cfg.Connections) == 1 {
		return nil, fmt.Errorf("仅剩一个连接，无法删除")
	}
	scanner, _ := newScanner()
	fmt.Printf("确定删除连接 %q？[y/N] ", name)
	scanner.Scan()
	if !isYes(scanner.Text()) {
		fmt.Println("已取消。")
		return cfg, nil
	}
	delete(cfg.Connections, name)
	if cfg.DefaultConnection == name {
		for n := range cfg.Connections {
			cfg.DefaultConnection = n
			break
		}
	}
	if err := Save(cfg, cd.Path); err != nil {
		return nil, err
	}
	fmt.Printf("连接 %q 已删除。\n", name)
	return cfg, nil
}

func newScanner() (*bufio.Scanner, func(prompt, defaultVal string) string) {
	scanner := bufio.NewScanner(os.Stdin)
	readLine := func(prompt, defaultVal string) string {
		if defaultVal != "" {
			fmt.Printf("%s（回车使用默认 %s）：\n> ", prompt, defaultVal)
		} else {
			fmt.Printf("%s：\n> ", prompt)
		}
		scanner.Scan()
		val := strings.TrimSpace(scanner.Text())
		if val == "" {
			return defaultVal
		}
		return val
	}
	return scanner, readLine
}

func askProvider(scanner *bufio.Scanner, readLine func(string, string) string) (string, ProviderConfig) {
	fmt.Println("选择 AI provider：")
	fmt.Println("  1) Claude")
	fmt.Println("  2) OpenAI")
	fmt.Println("  3) DeepSeek")
	fmt.Println("  4) Ollama")
	fmt.Println("  5) 自定义 OpenAI 兼容接口")
	fmt.Print("> ")
	scanner.Scan()
	choice := strings.TrimSpace(scanner.Text())
	name := "claude"
	switch choice {
	case "2":
		name = "openai"
	case "3":
		name = "deepseek"
	case "4":
		name = "ollama"
	case "5":
		name = readLine("Provider 名称", "custom")
	}
	pc := providerDefaults[name]
	pc.BaseURL = readLine("Base URL", pc.BaseURL)
	if name != "ollama" {
		pc.APIKey = readLine("API Key", "")
	} else {
		fmt.Println("Ollama 本地模式无需 API Key。")
	}
	pc.Model = readLine("模型名称", pc.Model)
	return name, pc
}

func askConnection(scanner *bufio.Scanner, readLine func(string, string) string, defaultName string) (string, ConnectionConfig) {
	fmt.Println("配置数据库连接：")
	name := readLine("连接名称", defaultName)
	fmt.Println("数据库类型：")
	fmt.Println("  1) MySQL")
	fmt.Println("  2) Oracle")
	fmt.Println("  3) 达梦")
	fmt.Print("> ")
	scanner.Scan()
	dbType := "mysql"
	switch strings.TrimSpace(scanner.Text()) {
	case "2":
		dbType = "oracle"
	case "3":
		dbType = "dm"
	}
	defaultPort := 3306
	if dbType == "oracle" {
		defaultPort = 1521
	}
	if dbType == "dm" {
		defaultPort = 5236
	}
	conn := ConnectionConfig{
		Type:     dbType,
		Host:     readLine("Host", "127.0.0.1"),
		Port:     readInt(readLine("Port", strconv.Itoa(defaultPort)), defaultPort),
		Username: readLine("用户名", ""),
		Password: readLine("密码", ""),
	}
	if dbType == "oracle" {
		conn.Database = readLine("服务名/service", "")
		conn.Schema = readLine("Schema（默认使用用户名）", "")
	} else {
		conn.Database = readLine("数据库名", "")
		conn.Schema = readLine("Schema（可空）", "")
	}
	return name, conn
}

func readInt(s string, fallback int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return fallback
	}
	return n
}

func isYes(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	return s == "y" || s == "yes"
}
