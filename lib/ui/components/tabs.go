package components

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"redscout/lib/ui/theme"
)

type Tab int

const (
	TabNamespace Tab = iota
	TabSlowLog
	TabBigKeys
	TabHotKeys
)

type TabDef struct {
	Name string
	Key  string
}

var TabDefs = []TabDef{
	{"Namespaces", "N"},
	{"Slow Log", "L"},
	{"Big Keys", "B"},
	{"Hot Keys", "H"},
}

func RenderTabBar(active Tab, width int) string {
	var parts []string
	for i, t := range TabDefs {
		key := lipgloss.NewStyle().Foreground(theme.Orange).Render(t.Key)
		var name string
		if Tab(i) == active {
			name = lipgloss.NewStyle().Foreground(theme.Blue).Bold(true).Render(t.Name)
		} else {
			name = lipgloss.NewStyle().Foreground(theme.Secondary).Render(t.Name)
		}
		parts = append(parts, fmt.Sprintf("%s %s", name, key))
	}

	result := " "
	for i, p := range parts {
		if i > 0 {
			result += "  "
		}
		result += p
	}
	return result
}
