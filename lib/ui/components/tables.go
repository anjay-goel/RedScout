package components

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"
)

func tableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().
		Foreground(theme.Secondary).
		Bold(true).
		Padding(0, 1)
	s.Selected = lipgloss.NewStyle().
		Foreground(theme.Text).
		Background(theme.Border).
		Bold(true).
		Padding(0, 1)
	s.Cell = lipgloss.NewStyle().
		Foreground(theme.Text).
		Padding(0, 1)
	return s
}

func NamespaceColumns(width int) []table.Column {
	nameW := utils.MaxKeyDisplayLen
	colW := (width - nameW - 20) / 8
	if colW < 10 {
		colW = 10
	}
	return []table.Column{
		{Title: "Namespace", Width: nameW},
		{Title: "~Keys 1", Width: colW},
		{Title: "~Memory 2", Width: colW},
		{Title: "Avg TTL 3", Width: colW},
		{Title: "% TTL 4", Width: colW},
		{Title: "GET/s 5", Width: colW},
		{Title: "SET/s 6", Width: colW},
		{Title: "DEL/s 7", Width: colW},
		{Title: "OPS/s 8", Width: colW},
	}
}

func NamespaceRows(stats models.NamespaceMetricList) []table.Row {
	rows := make([]table.Row, len(stats))
	for i, row := range stats {
		rows[i] = table.Row{
			utils.TruncateKey(row.Namespace),
			utils.FormatNumber(float64(row.EstKeys)),
			utils.FormatBytes(row.EstMemory),
			utils.FormatDuration(row.AvgTTL),
			fmt.Sprintf("%.1f%%", row.TTLPercent*100),
			fmt.Sprintf("%.1f/s", row.Ops[models.GetOp]),
			fmt.Sprintf("%.1f/s", row.Ops[models.SetOp]),
			fmt.Sprintf("%.1f/s", row.Ops[models.DelOp]),
			fmt.Sprintf("%.1f/s", row.Ops[models.TotalOp]),
		}
	}
	return rows
}

func RenderBreadcrumb(prefix models.Key, width int) string {
	path := lipgloss.NewStyle().Foreground(theme.Orange).Render("/ root")
	if len(prefix) > 0 {
		sep := " › "
		path = lipgloss.NewStyle().Foreground(theme.Orange).Render("/ root" + sep + strings.Join(prefix, sep))
	}
	hints := lipgloss.NewStyle().Foreground(theme.Muted).Render("(") +
		lipgloss.NewStyle().Foreground(theme.Orange).Render("→") +
		lipgloss.NewStyle().Foreground(theme.Muted).Render(" expand  ") +
		lipgloss.NewStyle().Foreground(theme.Orange).Render("←") +
		lipgloss.NewStyle().Foreground(theme.Muted).Render(" back)")

	gap := width - lipgloss.Width(path) - lipgloss.Width(hints) - 2
	if gap < 1 {
		gap = 1
	}
	return " " + path + strings.Repeat(" ", gap) + hints
}

func BigKeyColumns(width int) []table.Column {
	nameW := utils.MaxKeyDisplayLen
	return []table.Column{
		{Title: "Key", Width: nameW},
		{Title: "Size 1", Width: 14},
		{Title: "Type 2", Width: 10},
		{Title: "TTL 3", Width: 14},
	}
}

func BigKeyRows(keys models.BigKeyList) []table.Row {
	rows := make([]table.Row, len(keys))
	for i, k := range keys {
		ttlStr := "-"
		if k.TTL > 0 {
			ttlStr = utils.FormatDuration(k.TTL)
		}
		rows[i] = table.Row{
			utils.TruncateKey(k.Key.String()),
			utils.FormatBytes(k.Size),
			k.Type,
			ttlStr,
		}
	}
	return rows
}

func HotKeyColumns(width int) []table.Column {
	nameW := utils.MaxKeyDisplayLen
	return []table.Column{
		{Title: "Key", Width: nameW},
		{Title: "Ops/s 1", Width: 12},
		{Title: "Command 2", Width: 10},
	}
}

func HotKeyRows(keys models.HotKeyList) []table.Row {
	rows := make([]table.Row, len(keys))
	for i, k := range keys {
		rows[i] = table.Row{
			utils.TruncateKey(k.Key.String()),
			fmt.Sprintf("%.1f/s", k.Ops),
			k.Command,
		}
	}
	return rows
}

func SlowLogColumns(width int) []table.Column {
	return []table.Column{
		{Title: "ID 1", Width: 8},
		{Title: "Timestamp 2", Width: 20},
		{Title: "Duration 3", Width: 14},
		{Title: "Command 4", Width: 12},
		{Title: "Arguments", Width: utils.MaxKeyDisplayLen},
	}
}

func SlowLogRows(logs models.SlowLogList) []table.Row {
	rows := make([]table.Row, len(logs))
	for i, l := range logs {
		cmd := ""
		var args []string
		if len(l.Args) > 0 {
			cmd = strings.ToUpper(l.Args[0])
			args = l.Args[1:]
		}
		rows[i] = table.Row{
			fmt.Sprintf("%d", l.ID),
			l.Time.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%d ms", l.Duration.Milliseconds()),
			cmd,
			utils.TruncateKey(strings.Join(args, " ")),
		}
	}
	return rows
}

func NewTable(columns []table.Column, rows []table.Row, height int) table.Model {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)
	t.SetStyles(tableStyles())
	return t
}
