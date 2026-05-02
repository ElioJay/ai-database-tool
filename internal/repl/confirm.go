package repl

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/aidbt-tool/aidbt/internal/sqlplan"
)

func askConfirm(plan *sqlplan.Plan, connName, dbType string) bool {
	policy := sqlplan.ConfirmPolicy(plan.Local)
	fmt.Println()
	fmt.Printf("连接：%s（%s）\n", connName, dbType)
	fmt.Printf("类型：%s  风险：%s  原因：%s\n", plan.Local.Kind, plan.Local.Risk, plan.Local.Reason)
	if plan.Explanation != "" {
		fmt.Printf("解释：%s\n", plan.Explanation)
	}
	fmt.Println("SQL：")
	fmt.Println(plan.SQL)
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	if policy.PhraseRequired {
		fmt.Printf("高危操作。如确认执行，请输入 %q：\n> ", policy.Phrase)
		if scanner.Scan() {
			return strings.TrimSpace(scanner.Text()) == policy.Phrase
		}
		return false
	}
	fmt.Printf("[%s] 输入 y 确认，其他取消：\n> ", policy.ActionLabel)
	if scanner.Scan() {
		ans := strings.TrimSpace(strings.ToLower(scanner.Text()))
		return ans == "y" || ans == "yes"
	}
	return false
}
