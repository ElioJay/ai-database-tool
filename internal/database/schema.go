package database

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (s Schema) Summary(opts SchemaOptions) string {
	if opts.MaxTables <= 0 {
		opts.MaxTables = 50
	}
	if opts.MaxColumnsPerTable <= 0 {
		opts.MaxColumnsPerTable = 20
	}
	var out []string
	for _, table := range s.Tables {
		if len(out) >= opts.MaxTables {
			out = append(out, fmt.Sprintf("... 还有 %d 张表未显示", len(s.Tables)-len(out)))
			break
		}
		if !included(table.Name, opts.Include, opts.Exclude) {
			continue
		}
		cols := make([]string, 0, min(len(table.Columns), opts.MaxColumnsPerTable))
		for i, col := range table.Columns {
			if i >= opts.MaxColumnsPerTable {
				cols = append(cols, "...")
				break
			}
			item := strings.TrimSpace(col.Name + " " + col.Type)
			if col.PrimaryKey {
				item += " PK"
			}
			cols = append(cols, item)
		}
		header := table.Name
		if table.Comment != "" {
			header += " -- " + table.Comment
		}
		out = append(out, fmt.Sprintf("%s(%s)", header, strings.Join(cols, ", ")))
	}
	if len(out) == 0 {
		return "未探测到可用表结构。"
	}
	return strings.Join(out, "\n")
}

func included(name string, include, exclude []string) bool {
	if len(include) > 0 && !matchAny(name, include) {
		return false
	}
	if matchAny(name, exclude) {
		return false
	}
	return true
}

func matchAny(name string, patterns []string) bool {
	for _, pattern := range patterns {
		if pattern == "" {
			continue
		}
		if pattern == name {
			return true
		}
		if ok, _ := filepath.Match(pattern, name); ok {
			return true
		}
	}
	return false
}
