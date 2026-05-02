package repl

import (
	"fmt"
	"strings"
)

type MetaResult struct {
	Handled       bool
	ShouldExit    bool
	ResetHistory  bool
	RefreshSchema bool
	ShowHelp      bool
}

func HandleMeta(input string) MetaResult {
	input = strings.TrimSpace(input)
	if !strings.HasPrefix(input, "/") {
		return MetaResult{}
	}
	parts := strings.Fields(strings.TrimPrefix(input, "/"))
	if len(parts) == 0 {
		return MetaResult{Handled: true}
	}
	switch strings.ToLower(parts[0]) {
	case "exit", "quit":
		return MetaResult{Handled: true, ShouldExit: true}
	case "reset":
		return MetaResult{Handled: true, ResetHistory: true}
	case "schema":
		if len(parts) > 1 && strings.EqualFold(parts[1], "refresh") {
			return MetaResult{Handled: true, RefreshSchema: true}
		}
		fmt.Println("用法：/schema refresh")
		return MetaResult{Handled: true}
	case "help":
		return MetaResult{Handled: true, ShowHelp: true}
	default:
		fmt.Printf("未知命令：%s（输入 /help 查看帮助）\n", input)
		return MetaResult{Handled: true}
	}
}

func printHelp() {
	fmt.Print(`
使用方式：直接输入中文数据库需求，aidbt 会生成 SQL，确认后执行。

REPL 命令：
  /help             显示帮助
  /exit / /quit     退出
  /reset            清空当前对话历史
  /schema refresh   重新探测当前连接表结构
`)
}
