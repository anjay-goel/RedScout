package components

import (
	"charm.land/lipgloss/v2"
	"redscout/lib/ui/theme"
)

func RenderHelpOverlay(width, height int) string {
	title := lipgloss.NewStyle().Foreground(theme.Blue).Bold(true).Render("Keyboard Shortcuts")
	subtitle := lipgloss.NewStyle().Foreground(theme.Muted).Render("  press ? or Esc to close")

	sectionStyle := lipgloss.NewStyle().Foreground(theme.Orange)
	keyStyle := lipgloss.NewStyle().Foreground(theme.Orange)
	descStyle := lipgloss.NewStyle().Foreground(theme.Secondary)

	content := title + subtitle + "\n\n" +
		sectionStyle.Render("NAVIGATION") + "\n" +
		keyStyle.Render("N") + descStyle.Render(" Namespaces    ") +
		keyStyle.Render("L") + descStyle.Render(" Slow Log    ") +
		keyStyle.Render("B") + descStyle.Render(" Big Keys    ") +
		keyStyle.Render("H") + descStyle.Render(" Hot Keys    ") +
		keyStyle.Render("T") + descStyle.Render(" Next tab") + "\n\n" +
		sectionStyle.Render("ACTIONS") + "\n" +
		keyStyle.Render("S") + descStyle.Render(" Run SCAN    ") +
		keyStyle.Render("M") + descStyle.Render(" Run MONITOR    ") +
		keyStyle.Render("Q") + descStyle.Render(" Quit") + "\n\n" +
		sectionStyle.Render("NAMESPACE") + "\n" +
		keyStyle.Render("→/Enter") + descStyle.Render(" Drill down    ") +
		keyStyle.Render("←/Bksp") + descStyle.Render(" Level up    ") +
		keyStyle.Render("1-8") + descStyle.Render(" Sort columns")

	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Padding(1, 3)

	modal := modalStyle.Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modal)
}
