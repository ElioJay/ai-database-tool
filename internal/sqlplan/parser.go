package sqlplan

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ParseAIResponse(raw string) (*Plan, error) {
	raw = stripFence(strings.TrimSpace(raw))
	var plan Plan
	if err := json.Unmarshal([]byte(raw), &plan); err != nil {
		return nil, fmt.Errorf("解析 AI JSON 失败: %w", err)
	}
	plan.SQL = strings.TrimSpace(plan.SQL)
	if plan.SQL == "" {
		return nil, fmt.Errorf("AI 响应缺少 sql 字段")
	}
	plan.Local = Classify(plan.SQL)
	if plan.Local.Kind != StatementSelect || plan.Local.Risk != RiskLow {
		plan.NeedConfirm = true
	}
	return &plan, nil
}

func stripFence(raw string) string {
	if !strings.HasPrefix(raw, "```") {
		return raw
	}
	lines := strings.Split(raw, "\n")
	if len(lines) >= 3 && strings.HasPrefix(strings.TrimSpace(lines[len(lines)-1]), "```") {
		return strings.TrimSpace(strings.Join(lines[1:len(lines)-1], "\n"))
	}
	return raw
}
