package components

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"
)

func RenderHeader(info *models.RedisInfo, width int) string {
	panelWidth := (width - 4) / 3

	system := renderPanel("SYSTEM", renderSystemContent(info), panelWidth)
	perf := renderPanel("PERFORMANCE", renderPerfContent(info), panelWidth)
	resources := renderPanel("RESOURCES", renderResourcesContent(info), panelWidth)

	return lipgloss.JoinHorizontal(lipgloss.Top, system, " ", perf, " ", resources)
}

func renderPanel(label string, content string, width int) string {
	labelStr := theme.PanelLabelStyle.Render(label)
	style := theme.PanelStyle.Width(width)
	return style.Render(labelStr + "\n" + content)
}

func renderSystemContent(info *models.RedisInfo) string {
	redis := lipgloss.NewStyle().Foreground(theme.Blue).Bold(true).Render("Redis")
	version := lipgloss.NewStyle().Foreground(theme.Secondary).Render("v" + info.Server.RedisVersion)
	rest := lipgloss.NewStyle().Foreground(theme.Muted).Render(
		fmt.Sprintf("· %s · %s · %d clients",
			info.Server.OS,
			utils.FormatDuration(info.Server.Uptime),
			info.Clients.ConnectedClients,
		),
	)
	return fmt.Sprintf("%s %s %s", redis, version, rest)
}

func renderPerfContent(info *models.RedisInfo) string {
	totalKeys := info.Keyspace["db0"].Keys
	avgTTL := info.Keyspace["db0"].AvgTTL

	keys := lipgloss.NewStyle().Foreground(theme.Orange).Bold(true).Render(utils.FormatNumber(float64(totalKeys)))
	keysLabel := lipgloss.NewStyle().Foreground(theme.Secondary).Render(" keys")
	sep := lipgloss.NewStyle().Foreground(theme.Muted).Render(" │ ")
	ops := lipgloss.NewStyle().Foreground(theme.Orange).Render(utils.FormatOpsPerSec(float64(info.Stats.OpsPerSec)))
	hit := lipgloss.NewStyle().Foreground(theme.Green).Render(fmt.Sprintf("%.1f%%", info.Computed.HitRate*100))
	hitLabel := lipgloss.NewStyle().Foreground(theme.Secondary).Render(" hit")
	ttl := lipgloss.NewStyle().Foreground(theme.Secondary).Render("ttl " + utils.FormatDuration(avgTTL))

	return keys + keysLabel + sep + ops + sep + hit + hitLabel + sep + ttl
}

func renderResourcesContent(info *models.RedisInfo) string {
	var memLine string
	if info.Memory.MaxMemory > 0 {
		memPercent := float64(info.Memory.UsedMemory) / float64(info.Memory.MaxMemory) * 100
		memBar := RenderProgressBar(memPercent, 100, 15)
		mem := lipgloss.NewStyle().Foreground(theme.Orange).Bold(true).Render(info.Memory.UsedMemoryHuman)
		memLine = memBar + " " + mem
	} else {
		mem := lipgloss.NewStyle().Foreground(theme.Orange).Bold(true).Render(info.Memory.UsedMemoryHuman)
		noLimit := lipgloss.NewStyle().Foreground(theme.Muted).Render(" · no limit")
		memLine = mem + noLimit
	}

	policy := lipgloss.NewStyle().Foreground(theme.Muted).Render(" · " + info.Memory.MemoryPolicy)

	cpuLine := ""
	if info.CPU.SystemTime == 0 && info.CPU.UserTime == 0 {
		cpuLine = lipgloss.NewStyle().Foreground(theme.Muted).Render(" · cpu: n/a")
	} else {
		cpuLine = lipgloss.NewStyle().Foreground(theme.Muted).Render(" · cpu: ") +
			lipgloss.NewStyle().Foreground(theme.Secondary).Render(fmt.Sprintf("%.1f%%", info.Computed.CPUUsage*100))
	}

	return memLine + policy + cpuLine
}
