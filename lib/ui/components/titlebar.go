package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"
)

func RenderTitleBar(state *models.State, width int) string {
	brand := lipgloss.NewStyle().Foreground(theme.Blue).Bold(true).Render("RedScout")

	legend := lipgloss.NewStyle().Foreground(theme.Muted).Render("(") +
		lipgloss.NewStyle().Foreground(theme.Orange).Render("orange") +
		lipgloss.NewStyle().Foreground(theme.Muted).Render(" = shortcut)")

	status := lipgloss.NewStyle().Foreground(theme.Green).Render("●")
	if !state.ScanComplete {
		status = lipgloss.NewStyle().Foreground(theme.Orange).Render("⠋")
	}

	info := lipgloss.NewStyle().Foreground(theme.Secondary).Render(
		fmt.Sprintf("%d scanned · %s monitored",
			state.ScannedKeys,
			utils.FormatDuration(int64(state.TotalMonitorDuration.Seconds())),
		),
	)

	helpHint := lipgloss.NewStyle().Foreground(theme.Orange).Render("?") +
		lipgloss.NewStyle().Foreground(theme.Muted).Render(" help")

	left := fmt.Sprintf("%s  %s", brand, legend)
	right := fmt.Sprintf("%s %s  %s", info, status, helpHint)

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + right
}
