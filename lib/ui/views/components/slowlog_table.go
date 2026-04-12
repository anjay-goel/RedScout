package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"
)

type SlowLogTable struct {
	Table *tview.Table
}

func NewSlowLogTable() *SlowLogTable {
	table := tview.NewTable().SetFixed(1, 0)
	table.SetSelectable(true, false)
	table.SetBorders(false)
	table.SetBorderPadding(0, 0, 1, 0)
	table.SetBackgroundColor(theme.ColorBg)
	table.SetSelectedStyle(tcell.StyleDefault.
		Background(theme.ColorBorder).
		Foreground(theme.ColorText))
	return &SlowLogTable{table}
}

func (sl *SlowLogTable) Update(slowLogs models.SlowLogList) {
	if len(slowLogs) == 0 {
		return
	}

	headers := []string{
		"[#8b949e]ID [#484f58][[-][#f0883e]1[-][#484f58]][-]",
		"[#8b949e]Timestamp [#484f58][[-][#f0883e]2[-][#484f58]][-]",
		"[#8b949e]Duration [#484f58][[-][#f0883e]3[-][#484f58]][-]",
		"[#8b949e]Command [#484f58][[-][#f0883e]4[-][#484f58]][-]",
		"[#8b949e]Arguments[-]",
	}

	sl.Table.Clear()
	for i, h := range headers {
		align := tview.AlignLeft
		if i == 0 || i == 2 {
			align = tview.AlignRight
		}
		cell := tview.NewTableCell(h).
			SetBackgroundColor(theme.ColorBg).
			SetSelectable(false).
			SetAlign(align)
		sl.Table.SetCell(0, i, cell)
	}

	for i, log := range slowLogs {
		command := ""
		var args []string
		if len(log.Args) > 0 {
			command = strings.ToUpper(log.Args[0])
			args = log.Args[1:]
		}

		values := []string{
			fmt.Sprintf("%d ", log.ID),
			log.Time.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%12d ms", log.Duration.Milliseconds()),
			command,
			utils.TruncateKey(strings.Join(args, " ")),
		}

		colors := []tcell.Color{
			theme.ColorText,
			theme.ColorSecondary,
			theme.ColorOrange,
			theme.ColorGreen,
			theme.ColorMuted,
		}

		rowBg := theme.ColorBg
		if i%2 == 1 {
			rowBg = theme.ColorSurface
		}

		for j, val := range values {
			align := tview.AlignLeft
			if j == 0 || j == 2 {
				align = tview.AlignRight
			}
			cell := tview.NewTableCell(val).
				SetTextColor(colors[j]).
				SetAlign(align).
				SetExpansion(0).
				SetBackgroundColor(rowBg)
			sl.Table.SetCell(i+1, j, cell)
		}
	}

	sl.Table.SetFixed(1, 0)
}
