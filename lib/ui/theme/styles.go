package theme

import "charm.land/lipgloss/v2"

var (
	Bg        = lipgloss.Color("#0d1117")
	Surface   = lipgloss.Color("#161b22")
	Border    = lipgloss.Color("#21262d")
	Text      = lipgloss.Color("#c9d1d9")
	Muted     = lipgloss.Color("#484f58")
	Secondary = lipgloss.Color("#8b949e")
	Blue      = lipgloss.Color("#58a6ff")
	Orange    = lipgloss.Color("#f0883e")
	Green     = lipgloss.Color("#3fb950")
	Red       = lipgloss.Color("#f85149")
)

var (
	PanelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border).
			Padding(0, 1)

	PanelLabelStyle = lipgloss.NewStyle().
			Foreground(Muted).
			Bold(true)

	TabActiveStyle = lipgloss.NewStyle().
			Foreground(Blue).
			Bold(true)

	TabInactiveStyle = lipgloss.NewStyle().
			Foreground(Secondary)

	ShortcutStyle = lipgloss.NewStyle().
			Foreground(Orange)

	HintStyle = lipgloss.NewStyle().
			Foreground(Muted)
)
