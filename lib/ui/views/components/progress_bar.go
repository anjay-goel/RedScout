package components

import "fmt"

func CreateProgressBar(value, max float64, width int) string {
	if value < 0 {
		value = 0
	}
	if value > max {
		value = max
	}

	filled := int((value / max) * float64(width))
	empty := width - filled

	bar := "[#f0883e]"
	for i := 0; i < filled; i++ {
		bar += "█"
	}
	bar += "[#21262d]"
	for i := 0; i < empty; i++ {
		bar += "░"
	}
	bar += fmt.Sprintf("[-] [#8b949e]%.1f%%[-]", value)
	return bar
}
