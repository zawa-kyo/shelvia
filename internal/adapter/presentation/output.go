package presentation

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/zawa-kyo/shelvia/internal/application"
)

var (
	accentColor = lipgloss.AdaptiveColor{Light: "25", Dark: "81"}
	mutedColor  = lipgloss.AdaptiveColor{Light: "244", Dark: "245"}
	headerColor = lipgloss.AdaptiveColor{Light: "238", Dark: "252"}
	borderColor = lipgloss.AdaptiveColor{Light: "#B7D8BE", Dark: "#58745F"}

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accentColor)

	summaryStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	borderStyle = lipgloss.NewStyle().
			Foreground(borderColor)

	headerStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(headerColor).
			Align(lipgloss.Center).
			Padding(0, 1)

	cellStyle = lipgloss.NewStyle().
			Padding(0, 1)
)

// Writes user-facing command output.
func Write(out io.Writer, output application.Output) {
	if output.Message != "" {
		fmt.Fprintln(out, titleStyle.Render(output.Message))
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

func writeTable(out io.Writer, result application.Table) {
	fmt.Fprintln(out, titleStyle.Render("Shelvia books"))
	fmt.Fprintln(out, summaryStyle.Render(strconv.Itoa(len(result.Rows))+" row(s)"))
	fmt.Fprintln(out)

	rendered := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(borderStyle).
		Headers(result.Headers...).
		Rows(normalizeRows(result.Rows)...).
		StyleFunc(tableStyle)

	fmt.Fprintln(out, rendered)
}

func normalizeRows(rows [][]string) [][]string {
	normalized := make([][]string, len(rows))
	for i, row := range rows {
		normalized[i] = make([]string, len(row))
		for j, value := range row {
			normalized[i][j] = printableCell(value)
		}
	}
	return normalized
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

func tableStyle(row, col int) lipgloss.Style {
	if row == table.HeaderRow {
		return headerStyle
	}
	return cellStyle
}
