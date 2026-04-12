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
	scan := fmt.Sprintf("[#484f58]Scanned:[-] [#8b949e]%d keys[-] [#8b949e][[-][#f0883e]S[-][#8b949e]][-]",
		state.ScannedKeys,
	)
	monitor := fmt.Sprintf("[#484f58]Monitor:[-] [#8b949e]%s[-] [#8b949e][[-][#f0883e]M[-][#8b949e]][-]",
		utils.FormatDuration(int64(state.TotalMonitorDuration.Seconds())),
	)
	statusLog := fmt.Sprintf("[#484f58]LOG › [#8b949e::d]%s[-::-]", state.Status)
	helpHint := "[#8b949e][[-][#f0883e]?[-][#8b949e]][-] [#8b949e]help[-]"
	legend := "[#8b949e][[-][#f0883e]orange[-] [#8b949e]= shortcut][-]"

	t.View.SetText(fmt.Sprintf(" [#58a6ff::b]RedScout[-::-]  %s  %s  %s  %s  %s", scan, monitor, helpHint, legend, statusLog))
}
