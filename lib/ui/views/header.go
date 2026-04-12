package views

import (
	"fmt"

	"github.com/rivo/tview"
	"redscout/lib/ui/theme"
	"redscout/lib/ui/views/components"
	"redscout/lib/utils"
	"redscout/models"
)

type HeaderView struct {
	HeaderFlex  *tview.Flex
	system      *tview.TextView
	performance *tview.TextView
	resources   *tview.TextView
}

func NewHeaderView() *HeaderView {
	system := tview.NewTextView().SetDynamicColors(true)
	system.SetBorder(true).
		SetBorderColor(theme.ColorBorder).
		SetBackgroundColor(theme.ColorBg)

	perf := tview.NewTextView().SetDynamicColors(true)
	perf.SetBorder(true).
		SetBorderColor(theme.ColorBorder).
		SetBackgroundColor(theme.ColorBg)

	mem := tview.NewTextView().SetDynamicColors(true)
	mem.SetBorder(true).
		SetBorderColor(theme.ColorBorder).
		SetBackgroundColor(theme.ColorBg)

	headerFlex := tview.NewFlex().SetDirection(tview.FlexColumn)
	headerFlex.SetBackgroundColor(theme.ColorBg)
	headerFlex.AddItem(system, 0, 1, false)
	headerFlex.AddItem(perf, 0, 1, false)
	headerFlex.AddItem(mem, 0, 1, false)

	return &HeaderView{
		system:      system,
		performance: perf,
		resources:   mem,
		HeaderFlex:  headerFlex,
	}
}

func (header *HeaderView) Update(state *models.State) {
	header.updateSystem(state.RedisInfo)
	header.updatePerformance(state.RedisInfo)
	header.updateMemory(state.RedisInfo)
}

func (header *HeaderView) updateSystem(info *models.RedisInfo) {
	os := info.Server.OS
	if len(os) > 20 {
		os = os[:20]
	}
	text := fmt.Sprintf(" [#484f58]SYSTEM[-]\n [#58a6ff::b]Redis[-::-] [#8b949e]v%s[-] [#484f58]· %s[-]\n [#484f58]↑ %s · %d clients[-]",
		info.Server.RedisVersion,
		os,
		utils.FormatUptime(info.Server.Uptime),
		info.Clients.ConnectedClients,
	)
	header.system.SetText(text)
}

func (header *HeaderView) updatePerformance(info *models.RedisInfo) {
	totalKeys := info.Keyspace["db0"].Keys
	avgTTL := info.Keyspace["db0"].AvgTTL

	text := fmt.Sprintf(" [#484f58]PERFORMANCE[-]\n [#f0883e::b]%s[-::-] [#8b949e]keys[-] [#484f58]│[-] [#f0883e]%s[-] [#8b949e]ops[-] [#484f58]│[-] [#3fb950]%.1f%%[-] [#8b949e]hit[-] [#484f58]│[-] [#8b949e]ttl %s[-]",
		utils.FormatNumber(float64(totalKeys)),
		utils.FormatOpsPerSec(float64(info.Stats.OpsPerSec)),
		info.Computed.HitRate*100,
		utils.FormatDuration(avgTTL),
	)
	header.performance.SetText(text)
}

func (header *HeaderView) updateMemory(info *models.RedisInfo) {
	var memLine string
	if info.Memory.MaxMemory > 0 {
		memPercent := float64(info.Memory.UsedMemory) / float64(info.Memory.MaxMemory) * 100
		memBar := components.CreateProgressBar(memPercent, 100, 15)
		memLine = fmt.Sprintf("%s [#f0883e::b]%s[-::-]", memBar, info.Memory.UsedMemoryHuman)
	} else {
		memLine = fmt.Sprintf("[#f0883e::b]%s[-::-] [#484f58]· no limit[-]", info.Memory.UsedMemoryHuman)
	}

	cpuLine := ""
	if info.CPU.SystemTime == 0 && info.CPU.UserTime == 0 {
		cpuLine = " [#484f58]· cpu: n/a[-]"
	} else {
		cpuLine = fmt.Sprintf(" [#484f58]· cpu: [#8b949e]%.1f%%[-]", info.Computed.CPUUsage*100)
	}

	text := fmt.Sprintf(" [#484f58]RESOURCES[-]\n %s [#484f58]· %s[-]%s",
		memLine,
		info.Memory.MemoryPolicy,
		cpuLine,
	)
	header.resources.SetText(text)
}
