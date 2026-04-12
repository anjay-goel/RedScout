package screens

import (
	"charm.land/lipgloss/v2"
	"redscout/lib/ui/theme"
)

func RenderDisclaimer(width, height int) string {
	title := lipgloss.NewStyle().Foreground(theme.Red).Bold(true).Render("DISCLAIMER")
	body := lipgloss.NewStyle().Foreground(theme.Secondary).Render(
		"RedScout will run the 'MONITOR' command on your Redis instance.")
	warning := lipgloss.NewStyle().Foreground(theme.Orange).Render(
		"This can impact Redis performance. Use with caution on production environments.")
	question := lipgloss.NewStyle().Foreground(theme.Text).Render("Do you want to continue?")
	yes := lipgloss.NewStyle().Foreground(theme.Green).Render("Y") +
		lipgloss.NewStyle().Foreground(theme.Text).Render("es")
	no := lipgloss.NewStyle().Foreground(theme.Red).Render("N") +
		lipgloss.NewStyle().Foreground(theme.Text).Render("o")

	content := title + "\n\n" + body + "\n" + warning + "\n\n" + question + "\n\n" + yes + " / " + no

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Padding(2, 4).
		Align(lipgloss.Center)

	box := boxStyle.Render(content)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
