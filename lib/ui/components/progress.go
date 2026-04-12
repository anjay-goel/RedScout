package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"redscout/lib/ui/theme"
)

func RenderProgressBar(value, max float64, width int) string {
	if value < 0 {
		value = 0
	}
	if value > max {
		value = max
	}

	filled := int((value / max) * float64(width))
	empty := width - filled

	filledStyle := lipgloss.NewStyle().Foreground(theme.Orange)
	emptyStyle := lipgloss.NewStyle().Foreground(theme.Border)
	pctStyle := lipgloss.NewStyle().Foreground(theme.Secondary)

	return filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", empty)) +
		pctStyle.Render(fmt.Sprintf(" %.1f%%", value))
}
