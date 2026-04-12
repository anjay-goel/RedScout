package views

import (
	"redscout/lib/ui/theme"

	"github.com/rivo/tview"
)

func NewHelpOverlay() *tview.Flex {
	helpText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	helpText.SetBackgroundColor(theme.ColorSurface)
	helpText.SetBorderColor(theme.ColorBorder)
	helpText.SetBorder(true)

	helpText.SetText(
		"[#58a6ff::b]Keyboard Shortcuts[-::-]\n" +
			"[#484f58]press ? or Esc to close[-]\n\n" +
			"[#f0883e]NAVIGATION[-]\n" +
			"[#f0883e]N[-] [#8b949e]Namespaces[-]    [#f0883e]L[-] [#8b949e]Slow Log[-]    [#f0883e]B[-] [#8b949e]Big Keys[-]    [#f0883e]H[-] [#8b949e]Hot Keys[-]    [#f0883e]T[-] [#8b949e]Next tab[-]\n\n" +
			"[#f0883e]ACTIONS[-]\n" +
			"[#f0883e]S[-] [#8b949e]Run SCAN[-]    [#f0883e]M[-] [#8b949e]Run MONITOR[-]    [#f0883e]Q[-] [#8b949e]Quit[-]\n\n" +
			"[#f0883e]NAMESPACE[-]\n" +
			"[#f0883e]→/Enter[-] [#8b949e]Drill down[-]    [#f0883e]←/Bksp[-] [#8b949e]Level up[-]    [#f0883e]1-8[-] [#8b949e]Sort columns[-]\n",
	)

	overlay := tview.NewFlex().SetDirection(tview.FlexRow)
	overlay.SetBackgroundColor(theme.ColorBg)
	overlay.AddItem(nil, 0, 1, false)
	inner := tview.NewFlex().SetDirection(tview.FlexColumn)
	inner.AddItem(nil, 0, 1, false)
	inner.AddItem(helpText, 80, 0, false)
	inner.AddItem(nil, 0, 1, false)
	overlay.AddItem(inner, 15, 0, false)
	overlay.AddItem(nil, 0, 1, false)

	return overlay
}
