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
	status := "[#3fb950]●[-]"
	if !state.ScanComplete {
		status = "[#f0883e]⠋[-]"
	}

	right := fmt.Sprintf("[#8b949e]%d scanned · %s monitored[-] %s  [#f0883e]?[-] [#484f58]help[-]",
		state.ScannedKeys,
		utils.FormatDuration(int64(state.TotalMonitorDuration.Seconds())),
		status,
	)

	legend := "[#484f58]([#f0883e]orange[-] = shortcut)[-]"
	statusLog := fmt.Sprintf("[#484f58]│[-] [#8b949e]%s[-]", state.Status)
	t.View.SetText(fmt.Sprintf("[#58a6ff::b]RedScout[-::-]  %s  %s  %s", legend, statusLog, right))
}
