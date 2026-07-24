package views

import (
	"fmt"
	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"

	"github.com/rivo/tview"
)

type TitleBar struct {
	View *tview.TextView
}

func NewTitleBar() *TitleBar {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	tv.SetBackgroundColor(theme.ColorBg)
	return &TitleBar{View: tv}
}

func (t *TitleBar) Update(state *models.State) {
	scan := fmt.Sprintf("[#7d8590]Scanned[-] [#8b949e][[-][#f0883e]S[-][#8b949e]]:[-] [#8b949e]%d keys[-]",
		state.ScannedKeys,
	)
	monitor := fmt.Sprintf("[#7d8590]Monitor[-] [#8b949e][[-][#f0883e]M[-][#8b949e]]:[-] [#8b949e]%s[-]",
		utils.FormatDuration(int64(state.TotalMonitorDuration.Seconds())),
	)
	statusLog := fmt.Sprintf("[#7d8590]LOG › [#8b949e::d]%s[-::-]", state.Status)
	helpHint := "[#8b949e][[-][#f0883e]?[-][#8b949e]][-] [#8b949e]help[-]"

	t.View.SetText(fmt.Sprintf(" [#58a6ff::b]RedScout[-::-]  %s  %s  %s  %s", scan, monitor, helpHint, statusLog))
}
