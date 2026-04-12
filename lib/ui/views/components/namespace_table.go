package components

import (
	"fmt"
	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Namespace struct {
	Title *tview.TextView
	Table *tview.Table
	Flex  *tview.Flex
}

func NewNamespace() *Namespace {
	ns := &Namespace{}
	ns.Table = tview.NewTable().SetFixed(1, 0)
	ns.Table.SetSelectable(true, false)
	ns.Table.SetBorders(false)
	ns.Table.SetBackgroundColor(theme.ColorBg)
	ns.Table.SetSelectedStyle(tcell.StyleDefault.
		Background(theme.ColorBorder).
		Foreground(theme.ColorText))

	ns.Title = tview.NewTextView()
	ns.Title.SetDynamicColors(true)
	ns.Title.SetBackgroundColor(theme.ColorBg)
	ns.Title.SetText(" [#f0883e]/ root[-]")

	ns.Flex = tview.NewFlex()
	ns.Flex.SetDirection(tview.FlexRow)
	ns.Flex.SetBackgroundColor(theme.ColorBg)
	ns.Flex.SetBorderPadding(0, 0, 0, 0)
	ns.Flex.AddItem(ns.Title, 1, -1, false)
	ns.Flex.AddItem(ns.Table, 0, 1, true)

	return ns
}

func (ns *Namespace) Update(prefix models.Key, stats models.NamespaceMetricList) {
	headers := []string{"Namespace", "~Keys 1", "~Memory 2", "Avg TTL 3", "% TTL 4", "GET/s 5", "SET/s 6", "DEL/s 7", "OPS/s 8"}

	ns.Table.Clear()
	for i, h := range headers {
		align := tview.AlignLeft
		if i != 0 {
			align = tview.AlignRight
		}
		cell := tview.NewTableCell(h).
			SetTextColor(theme.ColorSecondary).
			SetBackgroundColor(theme.ColorBg).
			SetSelectable(false).
			SetAlign(align)
		ns.Table.SetCell(0, i, cell)
	}

	nsPad := utils.MaxKeyDisplayLen
	for i, row := range stats {
		nsVal := utils.TruncateKey(row.Namespace)
		values := []string{
			fmt.Sprintf("%-*s", nsPad, nsVal),
			fmt.Sprintf("%12s", utils.FormatNumber(float64(row.EstKeys))),
			fmt.Sprintf("%12s", utils.FormatBytes(row.EstMemory)),
			fmt.Sprintf("%12s", utils.FormatDuration(row.AvgTTL)),
			fmt.Sprintf("%11.1f%%", row.TTLPercent*100),
			fmt.Sprintf("%10.1f/s", row.Ops[models.GetOp]),
			fmt.Sprintf("%10.1f/s", row.Ops[models.SetOp]),
			fmt.Sprintf("%10.1f/s", row.Ops[models.DelOp]),
			fmt.Sprintf("%10.1f/s", row.Ops[models.TotalOp]),
		}

		colors := []tcell.Color{
			theme.ColorText,
			theme.ColorOrange,
			theme.ColorBlue,
			theme.ColorSecondary,
			theme.ColorSecondary,
			theme.ColorSecondary,
			theme.ColorSecondary,
			theme.ColorSecondary,
			theme.ColorSecondary,
		}

		rowBg := theme.ColorBg
		if i%2 == 1 {
			rowBg = theme.ColorSurface
		}

		for j, val := range values {
			align := tview.AlignLeft
			if j != 0 {
				align = tview.AlignRight
			}
			cell := tview.NewTableCell(val).
				SetTextColor(colors[j]).
				SetAlign(align).
				SetExpansion(0).
				SetBackgroundColor(rowBg)

			ns.Table.SetCell(i+1, j, cell)
		}
	}

	ns.Table.SetFixed(1, 0)
	ns.Table.ScrollToBeginning()

	separator := " › "
	if len(prefix) == 0 {
		ns.Title.SetText(" [#f0883e]/ root[-]                                                          [#f0883e]→[-] [#484f58]drill[-]  [#f0883e]←[-] [#484f58]back[-]")
	} else {
		path := "/ root" + separator + strings.Join(prefix, separator)
		ns.Title.SetText(fmt.Sprintf(" [#f0883e]%s[-]                                                          [#f0883e]→[-] [#484f58]drill[-]  [#f0883e]←[-] [#484f58]back[-]", path))
	}
}
