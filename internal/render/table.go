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
		b.WriteString(" ")
		b.WriteString(cell)
		for pad := width - displayLen(cell); pad > 0; pad-- {
			b.WriteByte(' ')
		}
		b.WriteString(" |")
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
	if displayLen(s) <= width {
		return s
	}
	if width <= 3 {
		return cutToWidth(s, width)
	}
	return cutToWidth(s, width-3) + "..."
}

func cutToWidth(s string, w int) string {
	dw := 0
	for i, r := range s {
		rw := 1
		if isWide(r) {
			rw = 2
		}
		if dw+rw > w {
			return s[:i]
		}
		dw += rw
	}
	return s
}

func isWide(r rune) bool {
	return r >= 0x1100 &&
		(r <= 0x115F ||
			r == 0x2329 || r == 0x232A ||
			(r >= 0x2E80 && r <= 0x303E) ||
			(r >= 0x3040 && r <= 0x33BF) ||
			(r >= 0x3400 && r <= 0x4DBF) ||
			(r >= 0x4E00 && r <= 0xA4CF) ||
			(r >= 0xAC00 && r <= 0xD7A3) ||
			(r >= 0xF900 && r <= 0xFAFF) ||
			(r >= 0xFE10 && r <= 0xFE19) ||
			(r >= 0xFE30 && r <= 0xFE6B) ||
			(r >= 0xFF01 && r <= 0xFF60) ||
			(r >= 0xFFE0 && r <= 0xFFE6) ||
			(r >= 0x20000 && r <= 0x2FFFD) ||
			(r >= 0x30000 && r <= 0x3FFFD))
}

func displayLen(s string) int {
	w := 0
	for _, r := range s {
		if isWide(r) {
			w += 2
		} else {
			w++
		}
	}
	return w
}
