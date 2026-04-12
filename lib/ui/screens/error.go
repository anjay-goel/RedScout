package screens

import (
	"charm.land/lipgloss/v2"
	"redscout/lib/ui/theme"
)

func RenderError(msg string, width, height int) string {
	title := lipgloss.NewStyle().Foreground(theme.Red).Bold(true).Render("ERROR")
	body := lipgloss.NewStyle().Foreground(theme.Text).Render(msg)
	retry := lipgloss.NewStyle().Foreground(theme.Green).Render("R") +
		lipgloss.NewStyle().Foreground(theme.Text).Render("etry")
	quit := lipgloss.NewStyle().Foreground(theme.Red).Render("Q") +
		lipgloss.NewStyle().Foreground(theme.Text).Render("uit")

	content := title + "\n\n" + body + "\n\n" + retry + " / " + quit

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Padding(2, 4).
		Align(lipgloss.Center)

	box := boxStyle.Render(content)
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, box)
}
