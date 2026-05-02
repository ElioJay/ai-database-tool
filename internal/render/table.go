package render

import (
	"fmt"
	"strings"
)

type TableOptions struct {
	MaxRows      int
	MaxCellWidth int
}

func Table(headers []string, rows [][]string, opts TableOptions) string {
	if opts.MaxRows <= 0 {
		opts.MaxRows = 100
	}
	if opts.MaxCellWidth <= 0 {
		opts.MaxCellWidth = 40
	}
	limit := min(len(rows), opts.MaxRows)
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = displayLen(truncate(h, opts.MaxCellWidth))
	}
	for _, row := range rows[:limit] {
		for i := range headers {
			cell := ""
			if i < len(row) {
				cell = truncate(row[i], opts.MaxCellWidth)
			}
			widths[i] = max(widths[i], displayLen(cell))
		}
	}

	var b strings.Builder
	writeRow(&b, headers, widths, opts.MaxCellWidth)
	writeSep(&b, widths)
	for _, row := range rows[:limit] {
		writeRow(&b, row, widths, opts.MaxCellWidth)
	}
	if len(rows) > limit {
		fmt.Fprintf(&b, "仅显示前 %d 行，共 %d 行。\n", limit, len(rows))
	}
	return b.String()
}

func writeRow(b *strings.Builder, row []string, widths []int, maxCellWidth int) {
	b.WriteString("|")
	for i, width := range widths {
		cell := ""
		if i < len(row) {
			cell = truncate(row[i], maxCellWidth)
		}
		fmt.Fprintf(b, " %-*s |", width, cell)
	}
	b.WriteString("\n")
}

func writeSep(b *strings.Builder, widths []int) {
	b.WriteString("|")
	for _, width := range widths {
		b.WriteString(" ")
		b.WriteString(strings.Repeat("-", width))
		b.WriteString(" |")
	}
	b.WriteString("\n")
}

func truncate(s string, width int) string {
	rs := []rune(s)
	if len(rs) <= width {
		return s
	}
	if width <= 3 {
		return string(rs[:width])
	}
	return string(rs[:width-3]) + "..."
}

func displayLen(s string) int {
	return len([]rune(s))
}
