package screens

import (
	"fmt"

	"charm.land/bubbles/v2/spinner"
	"charm.land/lipgloss/v2"
	"redscout/lib/ui/components"
	"redscout/lib/ui/theme"
)

func NewSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Orange)
	return s
}

func RenderLoading(sp spinner.Model, status string, scanProgress float64, scannedKeys int64,
	monitorProgress float64, monitorDur string, monitorTotal string,
	width, height int) string {

	title := lipgloss.NewStyle().Foreground(theme.Orange).Render(
		fmt.Sprintf("%s Analysing Redis", sp.View()))

	var progressInfo string
	if scanProgress < 100 {
		bar := components.RenderProgressBar(scanProgress, 100, 40)
		statusLine := lipgloss.NewStyle().Foreground(theme.Secondary).Render(status)
		keysLine := lipgloss.NewStyle().Foreground(theme.Text).Render(
			fmt.Sprintf("%d keys collected", scannedKeys))
		progressInfo = statusLine + "\n" + bar + "\n" + keysLine
	} else if monitorTotal == "0s" {
		progressInfo = lipgloss.NewStyle().Foreground(theme.Secondary).Render("Starting monitor...")
	} else if monitorProgress < 100 {
		bar := components.RenderProgressBar(monitorProgress, 100, 40)
		statusLine := lipgloss.NewStyle().Foreground(theme.Secondary).Render("Monitor Progress:")
		timeLine := lipgloss.NewStyle().Foreground(theme.Text).Render(
			fmt.Sprintf("%s / %s", monitorDur, monitorTotal))
		progressInfo = statusLine + "\n" + bar + "\n" + timeLine
	}

	content := title + "\n\n" + progressInfo

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Padding(2, 4).
		Align(lipgloss.Center)

	box := boxStyle.Render(content)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
