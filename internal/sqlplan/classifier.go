package sqlplan

import (
	"strings"
	"unicode"
)

func Classify(sql string) Classification {
	normalized := normalizeSQL(sql)
	if normalized == "" {
		return Classification{Kind: StatementOther, Risk: RiskHigh, Reason: "SQL 为空", Normalized: normalized}
	}
	if hasMultipleStatements(normalized) {
		return Classification{Kind: StatementOther, Risk: RiskCritical, Reason: "检测到多语句，首版不允许一次执行多条 SQL", Multi: true, Normalized: normalized}
	}

	first := firstKeyword(normalized)
	lower := strings.ToLower(normalized)
	switch first {
	case "select", "show", "describe", "desc", "explain", "with":
		return Classification{Kind: StatementSelect, Risk: RiskLow, Reason: "只读查询", Normalized: normalized}
	case "insert", "merge":
		return Classification{Kind: StatementDML, Risk: RiskMedium, Reason: "写入数据", Normalized: normalized}
	case "update":
		if !strings.Contains(lower, " where ") {
			return Classification{Kind: StatementDML, Risk: RiskHigh, Reason: "UPDATE 未检测到 WHERE 条件", Normalized: normalized}
		}
		return Classification{Kind: StatementDML, Risk: RiskMedium, Reason: "更新数据", Normalized: normalized}
	case "delete":
		if !strings.Contains(lower, " where ") {
			return Classification{Kind: StatementDML, Risk: RiskHigh, Reason: "DELETE 未检测到 WHERE 条件", Normalized: normalized}
		}
		return Classification{Kind: StatementDML, Risk: RiskMedium, Reason: "删除数据", Normalized: normalized}
	case "drop", "truncate", "alter":
		return Classification{Kind: StatementDDL, Risk: RiskCritical, Reason: "高危 DDL", Normalized: normalized}
	case "create":
		return Classification{Kind: StatementDDL, Risk: RiskHigh, Reason: "创建对象", Normalized: normalized}
	default:
		return Classification{Kind: StatementOther, Risk: RiskHigh, Reason: "未知或不支持的 SQL 类型", Normalized: normalized}
	}
}

func ConfirmPolicy(c Classification) Policy {
	policy := Policy{ActionLabel: "执行 SQL"}
	if c.Kind == StatementSelect {
		policy.ActionLabel = "执行查询"
	}
	if c.Risk == RiskCritical {
		policy.PhraseRequired = true
		policy.Phrase = "确认执行"
	}
	return policy
}

func normalizeSQL(sql string) string {
	sql = strings.TrimSpace(sql)
	sql = strings.TrimSuffix(sql, ";")
	var lines []string
	for _, line := range strings.Split(sql, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "--") || trimmed == "" {
			continue
		}
		lines = append(lines, line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func hasMultipleStatements(sql string) bool {
	inSingle := false
	inDouble := false
	escaped := false
	count := 0
	for _, r := range sql {
		if escaped {
			escaped = false
			continue
		}
		if r == '\\' {
			escaped = true
			continue
		}
		switch r {
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
		case ';':
			if !inSingle && !inDouble {
				count++
			}
		}
	}
	return count > 0
}

func firstKeyword(sql string) string {
	sql = strings.TrimLeftFunc(sql, unicode.IsSpace)
	var b strings.Builder
	for _, r := range sql {
		if !unicode.IsLetter(r) {
			break
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}
