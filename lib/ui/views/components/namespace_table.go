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
	ns.Table.SetBorderPadding(0, 0, 1, 0)
	ns.Table.SetBackgroundColor(theme.ColorBg)
	ns.Table.SetSelectedStyle(tcell.StyleDefault.
		Background(theme.ColorHighlight).
		Foreground(theme.ColorText).
		Attributes(tcell.AttrBold))

	ns.Title = tview.NewTextView()
	ns.Title.SetDynamicColors(true)
	ns.Title.SetBackgroundColor(theme.ColorBg)
	ns.Title.SetText(" [#f0883e]/ root[-]")

	ns.Flex = tview.NewFlex()
	ns.Flex.SetDirection(tview.FlexRow)
	ns.Flex.SetBackgroundColor(theme.ColorBg)
	ns.Flex.SetBorderPadding(0, 0, 0, 0)
	ns.Flex.AddItem(ns.Title, 2, -1, false)
	ns.Flex.AddItem(ns.Table, 0, 1, true)

	return ns
}

func (ns *Namespace) Update(prefix models.Key, stats models.NamespaceMetricList) {
	headers := []string{
		"[#8b949e]Namespace[-]",
		"[#8b949e]~Keys [#484f58][[-][#f0883e]1[-][#484f58]][-]",
		"[#8b949e]~Memory [#484f58][[-][#f0883e]2[-][#484f58]][-]",
		"[#8b949e]Avg TTL [#484f58][[-][#f0883e]3[-][#484f58]][-]",
		"[#8b949e]% TTL [#484f58][[-][#f0883e]4[-][#484f58]][-]",
		"[#8b949e]GET/s [#484f58][[-][#f0883e]5[-][#484f58]][-]",
		"[#8b949e]SET/s [#484f58][[-][#f0883e]6[-][#484f58]][-]",
		"[#8b949e]DEL/s [#484f58][[-][#f0883e]7[-][#484f58]][-]",
		"[#8b949e]OPS/s [#484f58][[-][#f0883e]8[-][#484f58]][-]",
		"[#8b949e]Types[-]",
	}

	ns.Table.Clear()
	for i, h := range headers {
		align := tview.AlignLeft
		if i != 0 && i != len(headers)-1 {
			align = tview.AlignRight
		}
		cell := tview.NewTableCell(h).
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
			strings.Join(row.Types, ","),
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
			theme.ColorMuted,
		}

		rowBg := theme.ColorBg
		if i%2 == 1 {
			rowBg = theme.ColorSurface
		}

		for j, val := range values {
			align := tview.AlignLeft
			if j != 0 && j != len(values)-1 {
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
	hints := "  [#484f58]([[-][#f0883e]→[-][#484f58]][-] [#484f58]expand  [[-][#f0883e]←[-][#484f58]][-] [#484f58]back)[-]"
	line := "\n[#484f58]" + strings.Repeat("─", 300) + "[-]"
	if len(prefix) == 0 {
		ns.Title.SetText(" [#f0883e]/ root[-]" + hints + line)
	} else {
		path := "/ root" + separator + strings.Join(prefix, separator)
		ns.Title.SetText(fmt.Sprintf(" [#f0883e]%s[-]%s%s", path, hints, line))
	}
}
