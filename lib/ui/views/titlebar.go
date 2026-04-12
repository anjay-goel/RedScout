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
	stats := fmt.Sprintf("[#8b949e]%d keys scanned · %s monitored[-]",
		state.ScannedKeys,
		utils.FormatDuration(int64(state.TotalMonitorDuration.Seconds())),
	)
	statusLog := fmt.Sprintf("[#484f58]LOG › [#8b949e::d]%s[-::-]", state.Status)
	helpHint := "[#484f58][[-][#f0883e]?[-][#484f58]][-] [#8b949e]help[-]"
	legend := "[#484f58][[-][#f0883e]orange[-] [#484f58]= shortcut][-]"

	t.View.SetText(fmt.Sprintf(" [#58a6ff::b]RedScout[-::-]  %s  %s  %s  %s", stats, helpHint, legend, statusLog))
}
