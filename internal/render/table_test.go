package render

import (
	"strings"
	"testing"
)

func TestTableLimitsRowsAndTruncatesCells(t *testing.T) {
	out := Table(
		[]string{"id", "name"},
		[][]string{
			{"1", "short"},
			{"2", "this is a very very long value"},
			{"3", "hidden"},
		},
		TableOptions{MaxRows: 2, MaxCellWidth: 12},
	)

	if strings.Contains(out, "hidden") {
		t.Fatalf("table should not include rows beyond MaxRows:\n%s", out)
	}
	if !strings.Contains(out, "this is a...") {
		t.Fatalf("table should truncate long cells:\n%s", out)
	}
	if !strings.Contains(out, "仅显示前 2 行") {
		t.Fatalf("table should mention row limit:\n%s", out)
	}
}
