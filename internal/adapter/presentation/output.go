package presentation

import (
	"fmt"
	"io"
	"strings"

	"github.com/zawa-kyo/shelvia/internal/application"
)

// Writes user-facing command output.
func Write(out io.Writer, output application.Output) {
	if output.Message != "" {
		fmt.Fprintln(out, output.Message)
		return
	}
	if len(output.Table.Headers) == 0 {
		return
	}
	writeTable(out, output.Table)
}

// Writes a compact user-facing error.
func WriteError(out io.Writer, err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(out, err.Error())
}

func writeTable(out io.Writer, table application.Table) {
	widths := columnWidths(table)
	writeRow(out, table.Headers, widths)
	writeSeparator(out, widths)
	for _, row := range table.Rows {
		writeRow(out, row, widths)
	}
}

func columnWidths(table application.Table) []int {
	widths := make([]int, len(table.Headers))
	for i, header := range table.Headers {
		widths[i] = len([]rune(header))
	}
	for _, row := range table.Rows {
		for i := range widths {
			value := ""
			if i < len(row) {
				value = printableCell(row[i])
			}
			if width := len([]rune(value)); width > widths[i] {
				widths[i] = width
			}
		}
	}
	return widths
}

func writeRow(out io.Writer, row []string, widths []int) {
	cells := make([]string, len(widths))
	for i := range widths {
		value := ""
		if i < len(row) {
			value = printableCell(row[i])
		}
		cells[i] = padRight(value, widths[i])
	}
	fmt.Fprintln(out, strings.Join(cells, "  "))
}

func writeSeparator(out io.Writer, widths []int) {
	parts := make([]string, len(widths))
	for i, width := range widths {
		parts[i] = strings.Repeat("-", width)
	}
	fmt.Fprintln(out, strings.Join(parts, "  "))
}

func printableCell(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return truncate(value, 24)
}

func truncate(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	if limit <= 3 {
		return string(runes[:limit])
	}
	return string(runes[:limit-3]) + "..."
}

func padRight(value string, width int) string {
	padding := width - len([]rune(value))
	if padding <= 0 {
		return value
	}
	return value + strings.Repeat(" ", padding)
}
