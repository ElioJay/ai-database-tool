package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aidbt-tool/aidbt/internal/config"
	"github.com/aidbt-tool/aidbt/internal/database"
	"github.com/aidbt-tool/aidbt/internal/repl"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "错误：%v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "init":
			_, err := config.RunWizard()
			return err
		case "conn":
			return runConn(args[1:])
		case "config":
			return runConfig(args[1:])
		case "version", "--version", "-v":
			fmt.Println("aidbt v0.1.0")
			return nil
		case "help", "--help", "-h":
			printUsage()
			return nil
		default:
			return runOneShot(strings.Join(args, " "))
		}
	}
	return runREPL()
}

func runREPL() error {
	cfg, cd, err := loadOrInit()
	if err != nil {
		return err
	}
	session, err := repl.NewSession(cfg, cd)
	if err != nil {
		return err
	}
	return session.Run()
}

func runOneShot(input string) error {
	cfg, cd, err := loadOrInit()
	if err != nil {
		return err
	}
	session, err := repl.NewSession(cfg, cd)
	if err != nil {
		return err
	}
	defer session.Close()
	return session.HandleQuery(context.Background(), input)
}

func loadOrInit() (*config.Config, config.ConfigDir, error) {
	cd := config.Resolve()
	cfg, err := config.Load(cd.Path)
	if err != nil {
		return nil, cd, err
	}
	if cfg.DefaultProvider == "" || cfg.DefaultConnection == "" || len(cfg.Providers) == 0 || len(cfg.Connections) == 0 {
		result, err := config.RunWizard()
		if err != nil {
			return nil, cd, err
		}
		return result.Config, result.ConfigDir, nil
	}
	return cfg, cd, nil
}

func runConn(args []string) error {
	if len(args) == 0 {
		fmt.Println("用法：aidbt conn <add|edit|delete|list|test> [name]")
		return nil
	}
	cd := config.Resolve()
	switch args[0] {
	case "add":
		_, err := config.AddConnection(cd)
		return err
	case "edit":
		name := optionalArg(args, 1)
		_, err := config.EditConnection(cd, name)
		return err
	case "delete", "remove":
		name := optionalArg(args, 1)
		_, err := config.DeleteConnection(cd, name)
		return err
	case "list":
		cfg, err := config.Load(cd.Path)
		if err != nil {
			return err
		}
		masked := config.MaskConfig(cfg)
		for name, cc := range masked.Connections {
			mark := " "
			if name == masked.DefaultConnection {
				mark = "*"
			}
			fmt.Printf("%s %s  type=%s host=%s port=%d user=%s password=%s database=%s schema=%s\n",
				mark, name, cc.Type, cc.Host, cc.Port, cc.Username, cc.Password, cc.Database, cc.Schema)
		}
		return nil
	case "test":
		return testConnection(optionalArg(args, 1))
	default:
		return fmt.Errorf("未知 conn 子命令：%s", args[0])
	}
}

func testConnection(name string) error {
	cd := config.Resolve()
	cfg, err := config.Load(cd.Path)
	if err != nil {
		return err
	}
	if name == "" {
		name = cfg.DefaultConnection
	}
	cc, ok := cfg.Connections[name]
	if !ok {
		return fmt.Errorf("连接 %q 未配置", name)
	}
	db, err := database.Open(name, database.ConnectionConfig{
		Type: cc.Type, Host: cc.Host, Port: cc.Port, Username: cc.Username, Password: cc.Password,
		Database: cc.Database, Schema: cc.Schema, DSN: cc.DSN, Include: cc.Include, Exclude: cc.Exclude,
	})
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.Ping(context.Background()); err != nil {
		return err
	}
	fmt.Printf("连接 %q 成功。\n", name)
	return nil
}

func runConfig(args []string) error {
	if len(args) == 0 || args[0] != "show" {
		fmt.Println("用法：aidbt config show")
		return nil
	}
	cd := config.Resolve()
	cfg, err := config.Load(cd.Path)
	if err != nil {
		return err
	}
	masked := config.MaskConfig(cfg)
	fmt.Printf("运行模式：%s\n配置目录：%s\n", cd.Mode, cd.Path)
	fmt.Printf("默认 Provider：%s\n默认连接：%s\n", masked.DefaultProvider, masked.DefaultConnection)
	for name, pc := range masked.Providers {
		fmt.Printf("[provider.%s] base_url=%s model=%s api_key=%s\n", name, pc.BaseURL, pc.Model, pc.APIKey)
	}
	for name, cc := range masked.Connections {
		fmt.Printf("[connection.%s] type=%s host=%s port=%d user=%s password=%s database=%s schema=%s dsn=%s\n",
			name, cc.Type, cc.Host, cc.Port, cc.Username, cc.Password, cc.Database, cc.Schema, cc.DSN)
	}
	return nil
}

func optionalArg(args []string, idx int) string {
	if len(args) > idx {
		return args[idx]
	}
	return ""
}

func printUsage() {
	fmt.Print(`
用法：aidbt [命令|自然语言问题]

命令：
  aidbt                         进入 REPL
  aidbt "查最近10个用户"         一次性自然语言查询
  aidbt init                    初始化配置
  aidbt conn add                添加数据库连接
  aidbt conn edit [name]         编辑数据库连接
  aidbt conn delete [name]       删除数据库连接
  aidbt conn list               列出数据库连接
  aidbt conn test [name]         测试数据库连接
  aidbt config show             脱敏显示配置
  aidbt version                 显示版本

REPL 命令：
  /help
  /reset
  /schema refresh
  /exit
`)
}
