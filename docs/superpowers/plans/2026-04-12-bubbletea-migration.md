# Bubbletea Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace the tview-based UI with bubbletea (MVU pattern) while preserving all functionality and the GitHub Dark design.

**Architecture:** Single root `tea.Model` with sub-models for each screen (disclaimer, loading, main). Main screen composes sub-models for title bar, header panels, tab bar, and table content. Styling via lipgloss. Tables via bubbles/table. State updates arrive via tea.Cmd polling the scanner's channel.

**Tech Stack:** charm.land/bubbletea/v2, charm.land/lipgloss/v2, charm.land/bubbles/v2

**Design Reference:** `docs/superpowers/specs/2026-04-12-ui-redesign-design.md` + `Desktop/mockup.png`

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `lib/ui/theme/styles.go` | Create | lipgloss styles (replaces theme.go tcell colors) |
| `lib/ui/model.go` | Create | Root tea.Model, screen state machine, top-level Update/View |
| `lib/ui/msgs.go` | Create | Custom tea.Msg types (state updates, scan complete, etc.) |
| `lib/ui/screens/disclaimer.go` | Create | Disclaimer screen model |
| `lib/ui/screens/loading.go` | Create | Loading screen with spinner + progress |
| `lib/ui/screens/error.go` | Create | Error screen model |
| `lib/ui/screens/main.go` | Create | Main screen: composes header, tabs, content |
| `lib/ui/components/titlebar.go` | Create | Title bar view (RedScout + scan state + ? help) |
| `lib/ui/components/header.go` | Create | 3-panel header (System, Performance, Resources) |
| `lib/ui/components/tabs.go` | Create | Tab bar with shortcut hints |
| `lib/ui/components/tables.go` | Create | Table models for namespace, bigkeys, hotkeys, slowlog |
| `lib/ui/components/help.go` | Create | Help overlay (? modal) |
| `lib/ui/components/progress.go` | Create | Progress bar helper |
| `main.go` | Modify | Use bubbletea program instead of tview app |
| `lib/ui/theme/theme.go` | Delete | Replaced by styles.go |
| `lib/ui/ui.go` | Delete | Replaced by model.go |
| `lib/ui/views/` | Delete | Entire directory replaced by screens/ and components/ |

---

### Task 1: Add bubbletea dependencies

**Files:**
- Modify: `go.mod`

- [ ] **Step 1: Install dependencies**

```bash
go get charm.land/bubbletea/v2
go get charm.land/lipgloss/v2
go get charm.land/bubbles/v2
```

- [ ] **Step 2: Verify**

Run: `go mod tidy`
Expected: go.mod and go.sum updated

- [ ] **Step 3: Commit**

```bash
git add go.mod go.sum
git commit -m "deps: add bubbletea v2, lipgloss v2, bubbles v2"
```

---

### Task 2: Create lipgloss styles

**Files:**
- Create: `lib/ui/theme/styles.go`

- [ ] **Step 1: Create styles.go with GitHub Dark color palette and reusable styles**

```go
package theme

import "charm.land/lipgloss/v2"

// GitHub Dark color palette
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

// Panel styles
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
```

- [ ] **Step 2: Build to verify**

Run: `go build ./lib/ui/theme/...`
Expected: Success

- [ ] **Step 3: Commit**

```bash
git add lib/ui/theme/styles.go
git commit -m "feat: add lipgloss styles for bubbletea migration"
```

---

### Task 3: Create message types

**Files:**
- Create: `lib/ui/msgs.go`

- [ ] **Step 1: Define custom messages for state updates**

```go
package ui

import (
	"redscout/models"
)

// StateUpdateMsg is sent when scanner state changes
type StateUpdateMsg struct {
	State *models.State
}

// ScanCompleteMsg is sent when initial scan finishes
type ScanCompleteMsg struct {
	State *models.State
}

// ErrorMsg is sent on scanner errors
type ErrorMsg struct {
	Err error
}
```

- [ ] **Step 2: Build**

Run: `go build ./lib/ui/...`

- [ ] **Step 3: Commit**

```bash
git add lib/ui/msgs.go
git commit -m "feat: add bubbletea message types"
```

---

### Task 4: Create progress bar helper

**Files:**
- Create: `lib/ui/components/progress.go`

- [ ] **Step 1: Create a simple progress bar renderer using lipgloss**

```go
package components

import (
	"fmt"

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

	bar := ""
	for range filled {
		bar += "█"
	}
	for range empty {
		bar += "░"
	}

	return filledStyle.Render(bar[:filled]) +
		emptyStyle.Render(bar[filled:]) +
		pctStyle.Render(fmt.Sprintf(" %.1f%%", value))
}
```

Wait — the string slicing above is wrong since we're building `bar` as a single string. Let me fix:

```go
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
```

- [ ] **Step 2: Commit**

```bash
git add lib/ui/components/progress.go
git commit -m "feat: add lipgloss progress bar component"
```

---

### Task 5: Create title bar component

**Files:**
- Create: `lib/ui/components/titlebar.go`

- [ ] **Step 1: Create title bar view function**

```go
package components

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"
)

func RenderTitleBar(state *models.State, width int) string {
	brand := lipgloss.NewStyle().Foreground(theme.Blue).Bold(true).Render("RedScout")

	legend := lipgloss.NewStyle().Foreground(theme.Muted).Render("(") +
		lipgloss.NewStyle().Foreground(theme.Orange).Render("orange") +
		lipgloss.NewStyle().Foreground(theme.Muted).Render(" = shortcut)")

	status := lipgloss.NewStyle().Foreground(theme.Green).Render("●")
	if !state.ScanComplete {
		status = lipgloss.NewStyle().Foreground(theme.Orange).Render("⠋")
	}

	info := lipgloss.NewStyle().Foreground(theme.Secondary).Render(
		fmt.Sprintf("%d scanned · %s monitored",
			state.ScannedKeys,
			utils.FormatDuration(int64(state.TotalMonitorDuration.Seconds())),
		),
	)

	helpHint := lipgloss.NewStyle().Foreground(theme.Orange).Render("?") +
		lipgloss.NewStyle().Foreground(theme.Muted).Render(" help")

	left := fmt.Sprintf("%s  %s", brand, legend)
	right := fmt.Sprintf("%s %s  %s", info, status, helpHint)

	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}

	return left + strings.Repeat(" ", gap) + right
}
```

Add `"strings"` to imports.

- [ ] **Step 2: Commit**

```bash
git add lib/ui/components/titlebar.go
git commit -m "feat: add bubbletea title bar component"
```

---

### Task 6: Create header panels component

**Files:**
- Create: `lib/ui/components/header.go`

- [ ] **Step 1: Create 3-panel header view function**

```go
package components

import (
	"fmt"

	"charm.land/lipgloss/v2"
	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"
)

func RenderHeader(info *models.RedisInfo, width int) string {
	panelWidth := (width - 6) / 3 // account for borders and gaps

	system := renderPanel("SYSTEM", renderSystemContent(info), panelWidth)
	perf := renderPanel("PERFORMANCE", renderPerfContent(info), panelWidth)
	resources := renderPanel("RESOURCES", renderResourcesContent(info), panelWidth)

	return lipgloss.JoinHorizontal(lipgloss.Top, system, " ", perf, " ", resources)
}

func renderPanel(label string, content string, width int) string {
	labelStr := lipgloss.NewStyle().Foreground(theme.Muted).Bold(true).Render(label)
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
		cpuLine = lipgloss.NewStyle().Foreground(theme.Muted).Render(
			fmt.Sprintf(" · cpu: ")) +
			lipgloss.NewStyle().Foreground(theme.Secondary).Render(
				fmt.Sprintf("%.1f%%", info.Computed.CPUUsage*100))
	}

	return memLine + policy + cpuLine
}
```

- [ ] **Step 2: Commit**

```bash
git add lib/ui/components/header.go
git commit -m "feat: add bubbletea header panels component"
```

---

### Task 7: Create tab bar component

**Files:**
- Create: `lib/ui/components/tabs.go`

- [ ] **Step 1: Create tab bar view function**

```go
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
	Name    string
	Key     string
}

var Tabs = []TabDef{
	{"Namespaces", "N"},
	{"Slow Log", "L"},
	{"Big Keys", "B"},
	{"Hot Keys", "H"},
}

func RenderTabBar(active Tab, width int) string {
	var parts []string
	for i, t := range Tabs {
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
```

- [ ] **Step 2: Commit**

```bash
git add lib/ui/components/tabs.go
git commit -m "feat: add bubbletea tab bar component"
```

---

### Task 8: Create help overlay component

**Files:**
- Create: `lib/ui/components/help.go`

- [ ] **Step 1: Create help overlay view function**

```go
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
		Background(theme.Surface).
		Padding(1, 3)

	modal := modalStyle.Render(content)

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, modal)
}
```

- [ ] **Step 2: Commit**

```bash
git add lib/ui/components/help.go
git commit -m "feat: add bubbletea help overlay component"
```

---

### Task 9: Create table rendering components

**Files:**
- Create: `lib/ui/components/tables.go`

- [ ] **Step 1: Create table models for all 4 tabs using bubbles/table**

This is the largest component. It creates and configures a `table.Model` for each tab, and provides update functions to refresh data.

```go
package components

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"
)

func tableStyles() table.Styles {
	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().
		Foreground(theme.Secondary).
		Bold(true).
		Padding(0, 1)
	s.Selected = lipgloss.NewStyle().
		Foreground(theme.Text).
		Background(theme.Border).
		Bold(true).
		Padding(0, 1)
	s.Cell = lipgloss.NewStyle().
		Foreground(theme.Text).
		Padding(0, 1)
	return s
}

// --- Namespace Table ---

func NamespaceColumns(width int) []table.Column {
	nameW := utils.MaxKeyDisplayLen
	colW := (width - nameW - 20) / 8
	if colW < 10 {
		colW = 10
	}
	return []table.Column{
		{Title: "Namespace", Width: nameW},
		{Title: "~Keys 1", Width: colW},
		{Title: "~Memory 2", Width: colW},
		{Title: "Avg TTL 3", Width: colW},
		{Title: "% TTL 4", Width: colW},
		{Title: "GET/s 5", Width: colW},
		{Title: "SET/s 6", Width: colW},
		{Title: "DEL/s 7", Width: colW},
		{Title: "OPS/s 8", Width: colW},
	}
}

func NamespaceRows(stats models.NamespaceMetricList) []table.Row {
	rows := make([]table.Row, len(stats))
	for i, row := range stats {
		rows[i] = table.Row{
			utils.TruncateKey(row.Namespace),
			utils.FormatNumber(float64(row.EstKeys)),
			utils.FormatBytes(row.EstMemory),
			utils.FormatDuration(row.AvgTTL),
			fmt.Sprintf("%.1f%%", row.TTLPercent*100),
			fmt.Sprintf("%.1f/s", row.Ops[models.GetOp]),
			fmt.Sprintf("%.1f/s", row.Ops[models.SetOp]),
			fmt.Sprintf("%.1f/s", row.Ops[models.DelOp]),
			fmt.Sprintf("%.1f/s", row.Ops[models.TotalOp]),
		}
	}
	return rows
}

func RenderBreadcrumb(prefix models.Key, width int) string {
	path := lipgloss.NewStyle().Foreground(theme.Orange).Render("/ root")
	if len(prefix) > 0 {
		sep := " › "
		path = lipgloss.NewStyle().Foreground(theme.Orange).Render("/ root" + sep + strings.Join(prefix, sep))
	}
	hints := lipgloss.NewStyle().Foreground(theme.Muted).Render("(") +
		lipgloss.NewStyle().Foreground(theme.Orange).Render("→") +
		lipgloss.NewStyle().Foreground(theme.Muted).Render(" expand  ") +
		lipgloss.NewStyle().Foreground(theme.Orange).Render("←") +
		lipgloss.NewStyle().Foreground(theme.Muted).Render(" back)")

	gap := width - lipgloss.Width(path) - lipgloss.Width(hints) - 2
	if gap < 1 {
		gap = 1
	}
	return " " + path + strings.Repeat(" ", gap) + hints
}

// --- Big Keys Table ---

func BigKeyColumns(width int) []table.Column {
	nameW := utils.MaxKeyDisplayLen
	return []table.Column{
		{Title: "Key", Width: nameW},
		{Title: "Size 1", Width: 14},
		{Title: "Type 2", Width: 10},
		{Title: "TTL 3", Width: 14},
	}
}

func BigKeyRows(keys models.BigKeyList) []table.Row {
	rows := make([]table.Row, len(keys))
	for i, k := range keys {
		ttlStr := "-"
		if k.TTL > 0 {
			ttlStr = utils.FormatDuration(k.TTL)
		}
		rows[i] = table.Row{
			utils.TruncateKey(k.Key.String()),
			utils.FormatBytes(k.Size),
			k.Type,
			ttlStr,
		}
	}
	return rows
}

// --- Hot Keys Table ---

func HotKeyColumns(width int) []table.Column {
	nameW := utils.MaxKeyDisplayLen
	return []table.Column{
		{Title: "Key", Width: nameW},
		{Title: "Ops/s 1", Width: 12},
		{Title: "Command 2", Width: 10},
	}
}

func HotKeyRows(keys models.HotKeyList) []table.Row {
	rows := make([]table.Row, len(keys))
	for i, k := range keys {
		rows[i] = table.Row{
			utils.TruncateKey(k.Key.String()),
			fmt.Sprintf("%.1f/s", k.Ops),
			k.Command,
		}
	}
	return rows
}

// --- Slow Log Table ---

func SlowLogColumns(width int) []table.Column {
	return []table.Column{
		{Title: "ID 1", Width: 8},
		{Title: "Timestamp 2", Width: 20},
		{Title: "Duration 3", Width: 14},
		{Title: "Command 4", Width: 12},
		{Title: "Arguments", Width: utils.MaxKeyDisplayLen},
	}
}

func SlowLogRows(logs models.SlowLogList) []table.Row {
	rows := make([]table.Row, len(logs))
	for i, l := range logs {
		cmd := ""
		var args []string
		if len(l.Args) > 0 {
			cmd = strings.ToUpper(l.Args[0])
			args = l.Args[1:]
		}
		rows[i] = table.Row{
			fmt.Sprintf("%d", l.ID),
			l.Time.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%d ms", l.Duration.Milliseconds()),
			cmd,
			utils.TruncateKey(strings.Join(args, " ")),
		}
	}
	return rows
}

// NewTable creates a configured table.Model
func NewTable(columns []table.Column, rows []table.Row, height int) table.Model {
	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(height),
	)
	t.SetStyles(tableStyles())
	return t
}
```

- [ ] **Step 2: Commit**

```bash
git add lib/ui/components/tables.go
git commit -m "feat: add bubbletea table components for all 4 tabs"
```

---

### Task 10: Create screen models (disclaimer, loading, error)

**Files:**
- Create: `lib/ui/screens/disclaimer.go`
- Create: `lib/ui/screens/loading.go`
- Create: `lib/ui/screens/error.go`

- [ ] **Step 1: Create disclaimer screen**

```go
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
```

- [ ] **Step 2: Create loading screen**

```go
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
	monitorProgress float64, monitorDuration string, monitorTotal string,
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
			fmt.Sprintf("%s / %s", monitorDuration, monitorTotal))
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
```

- [ ] **Step 3: Create error screen**

```go
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
```

- [ ] **Step 4: Commit**

```bash
git add lib/ui/screens/
git commit -m "feat: add bubbletea screen models (disclaimer, loading, error)"
```

---

### Task 11: Create main screen model

**Files:**
- Create: `lib/ui/screens/main.go`

- [ ] **Step 1: Create main screen that composes all components**

```go
package screens

import (
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	tea "charm.land/bubbletea/v2"
	"redscout/lib/ui/components"
	"redscout/lib/ui/theme"
	"redscout/models"
)

type MainModel struct {
	State       *models.State
	ActiveTab   components.Tab
	Table       table.Model
	Width       int
	Height      int
	HelpVisible bool
}

func NewMainModel(state *models.State, width, height int) MainModel {
	m := MainModel{
		State:     state,
		ActiveTab: components.TabNamespace,
		Width:     width,
		Height:    height,
	}
	m.rebuildTable()
	return m
}

func (m *MainModel) rebuildTable() {
	tableHeight := m.Height - 8 // title(1) + header(3) + tabs(1) + breadcrumb(1) + padding(2)
	if tableHeight < 5 {
		tableHeight = 5
	}

	switch m.ActiveTab {
	case components.TabNamespace:
		m.Table = components.NewTable(
			components.NamespaceColumns(m.Width),
			components.NamespaceRows(m.State.NamespaceStats),
			tableHeight,
		)
	case components.TabBigKeys:
		m.Table = components.NewTable(
			components.BigKeyColumns(m.Width),
			components.BigKeyRows(m.State.BigKeys),
			tableHeight,
		)
	case components.TabHotKeys:
		m.Table = components.NewTable(
			components.HotKeyColumns(m.Width),
			components.HotKeyRows(m.State.HotKeys),
			tableHeight,
		)
	case components.TabSlowLog:
		m.Table = components.NewTable(
			components.SlowLogColumns(m.Width),
			components.SlowLogRows(m.State.SlowLogs),
			tableHeight,
		)
	}
}

func (m MainModel) Update(msg tea.Msg) (MainModel, tea.Cmd) {
	var cmd tea.Cmd
	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}

func (m MainModel) View() string {
	if m.HelpVisible {
		return components.RenderHelpOverlay(m.Width, m.Height)
	}

	titleBar := components.RenderTitleBar(m.State, m.Width)
	header := components.RenderHeader(m.State.RedisInfo, m.Width)
	tabBar := components.RenderTabBar(m.ActiveTab, m.Width)

	var content string
	if m.ActiveTab == components.TabNamespace {
		breadcrumb := components.RenderBreadcrumb(m.State.CurrentPrefix, m.Width)
		content = breadcrumb + "\n" + m.Table.View()
	} else {
		content = m.Table.View()
	}

	screen := lipgloss.JoinVertical(lipgloss.Left,
		titleBar,
		header,
		tabBar,
		content,
	)

	return lipgloss.NewStyle().Background(theme.Bg).Render(screen)
}

func (m *MainModel) SetTab(tab components.Tab) {
	m.ActiveTab = tab
	m.rebuildTable()
}

func (m *MainModel) RefreshData(state *models.State) {
	m.State = state
	m.rebuildTable()
}
```

- [ ] **Step 2: Commit**

```bash
git add lib/ui/screens/main.go
git commit -m "feat: add bubbletea main screen model"
```

---

### Task 12: Create root model

**Files:**
- Create: `lib/ui/model.go`

- [ ] **Step 1: Create root tea.Model with screen state machine**

This is the main orchestrator. It manages which screen is active, handles keyboard input, and dispatches commands to the scanner.

```go
package ui

import (
	"fmt"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"redscout/lib/scanner"
	"redscout/lib/ui/components"
	"redscout/lib/ui/screens"
	"redscout/lib/utils"
	"redscout/models"
)

type screen int

const (
	screenDisclaimer screen = iota
	screenLoading
	screenError
	screenMain
)

type Model struct {
	config  *models.Config
	scanner *scanner.Scanner
	screen  screen
	err     string

	// Sub-models
	spinner spinner.Model
	main    screens.MainModel

	// State
	width  int
	height int
}

func NewModel(cfg models.Config) Model {
	return Model{
		config:  &cfg,
		screen:  screenDisclaimer,
		spinner: screens.NewSpinner(),
	}
}

func (m Model) Init() tea.Cmd {
	return m.spinner.Tick
}

// pollScanner polls the scanner state channel for updates
func pollScanner(s *scanner.Scanner) tea.Cmd {
	return func() tea.Msg {
		state, ok := <-s.State.Updates
		if !ok {
			return nil
		}
		if state.ScanComplete {
			return ScanCompleteMsg{State: state}
		}
		return StateUpdateMsg{State: state}
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.screen == screenMain {
			m.main.Width = msg.Width
			m.main.Height = msg.Height
			m.main.RefreshData(m.main.State)
		}
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case StateUpdateMsg:
		if m.screen == screenMain {
			m.main.RefreshData(msg.State)
		}
		return m, pollScanner(m.scanner)

	case ScanCompleteMsg:
		m.main = screens.NewMainModel(msg.State, m.width, m.height)
		m.screen = screenMain
		return m, pollScanner(m.scanner)

	case ErrorMsg:
		m.err = msg.Err.Error()
		m.screen = screenError
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}

	if m.screen == screenMain {
		var cmd tea.Cmd
		m.main, cmd = m.main.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch m.screen {
	case screenDisclaimer:
		switch key {
		case "y", "Y", "enter":
			return m.startScanner()
		case "n", "N", "q", "Q":
			return m, tea.Quit
		}

	case screenError:
		switch key {
		case "r", "R", "enter":
			return m.startScanner()
		case "q", "Q":
			return m, tea.Quit
		}

	case screenMain:
		// Help overlay
		if key == "?" {
			m.main.HelpVisible = !m.main.HelpVisible
			return m, nil
		}
		if key == "escape" && m.main.HelpVisible {
			m.main.HelpVisible = false
			return m, nil
		}
		if m.main.HelpVisible {
			return m, nil
		}

		switch key {
		case "q", "Q":
			if m.scanner != nil {
				m.scanner.Close()
			}
			return m, tea.Quit

		case "n", "N":
			m.main.SetTab(components.TabNamespace)
		case "l", "L":
			m.main.SetTab(components.TabSlowLog)
		case "b", "B":
			m.main.SetTab(components.TabBigKeys)
		case "h", "H":
			m.main.SetTab(components.TabHotKeys)
		case "t", "T":
			next := (m.main.ActiveTab + 1) % 4
			m.main.SetTab(next)

		case "s", "S":
			go func() {
				err := m.scanner.ScanMemory()
				if err == nil {
					_ = m.scanner.ComputeNamespaceStats()
					_ = m.scanner.ComputeBigKeysFromScanLog()
				}
			}()
		case "m", "M":
			go func() {
				err := m.scanner.MonitorOps()
				if err == nil {
					_ = m.scanner.ComputeNamespaceStats()
					_ = m.scanner.ComputeHotKeysFromMonitorLog()
				}
			}()

		case "enter", "right":
			if m.main.ActiveTab == components.TabNamespace {
				row := m.main.Table.Cursor()
				if row >= 0 && row < len(m.main.State.NamespaceStats) {
					ns := m.main.State.NamespaceStats[row].Namespace
					m.scanner.DrillDownNamespace(ns)
				}
			}
		case "backspace", "left":
			if m.main.ActiveTab == components.TabNamespace {
				m.scanner.LevelUpNamespace()
			}

		case "1", "2", "3", "4", "5", "6", "7", "8":
			m.handleSort(key)
		default:
			var cmd tea.Cmd
			m.main, cmd = m.main.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

var namespaceSortKeys = map[string]string{
	"1": "Keys", "2": "Memory", "3": "TTL", "4": "% TTL",
	"5": "Get", "6": "Set", "7": "Del", "8": "Total Ops",
}

var slowLogSortKeys = map[string]string{
	"1": "ID", "2": "Timestamp", "3": "Duration", "4": "Command",
}

func (m *Model) handleSort(key string) {
	if m.main.ActiveTab == components.TabNamespace {
		if sortKey, ok := namespaceSortKeys[key]; ok {
			m.main.State.NamespaceStats.Sort(sortKey)
			m.main.RefreshData(m.main.State)
		}
	} else if m.main.ActiveTab == components.TabSlowLog {
		if sortKey, ok := slowLogSortKeys[key]; ok {
			m.main.State.SlowLogs.Sort(sortKey)
			m.main.RefreshData(m.main.State)
		}
	}
}

func (m Model) startScanner() (tea.Model, tea.Cmd) {
	m.screen = screenLoading
	return m, func() tea.Msg {
		s, err := scanner.NewScanner(m.config)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		m.scanner = s
		go s.Start()
		return pollScanner(s)()
	}
}

func (m Model) View() string {
	switch m.screen {
	case screenDisclaimer:
		return screens.RenderDisclaimer(m.width, m.height)
	case screenLoading:
		status := "Initializing..."
		scanProgress := 0.0
		scannedKeys := int64(0)
		monitorProgress := 0.0
		monitorDur := "0s"
		monitorTotal := "0s"
		if m.scanner != nil && m.scanner.State != nil {
			status = m.scanner.State.Status
			scanProgress = m.scanner.State.ScanProgress
			scannedKeys = m.scanner.State.ScannedKeys
			monitorProgress = m.scanner.State.MonitorProgress
			if m.scanner.State.MonitorDurationTotal > 0 {
				elapsed := time.Duration(float64(m.scanner.State.MonitorDurationTotal) * m.scanner.State.MonitorProgress / 100)
				monitorDur = utils.FormatDuration(int64(elapsed.Seconds()))
				monitorTotal = utils.FormatDuration(int64(m.scanner.State.MonitorDurationTotal.Seconds()))
			}
		}
		return screens.RenderLoading(m.spinner, status, scanProgress, scannedKeys,
			monitorProgress, monitorDur, monitorTotal, m.width, m.height)
	case screenError:
		return screens.RenderError(m.err, m.width, m.height)
	case screenMain:
		return m.main.View()
	}
	return ""
}
```

- [ ] **Step 2: Build**

Run: `go build ./lib/ui/...`
Note: This won't fully compile yet because main.go still uses the old tview code. That's fine — we fix main.go next.

- [ ] **Step 3: Commit**

```bash
git add lib/ui/model.go lib/ui/msgs.go
git commit -m "feat: add root bubbletea model with screen state machine"
```

---

### Task 13: Update main.go to use bubbletea

**Files:**
- Modify: `main.go`

- [ ] **Step 1: Replace tview app with bubbletea program**

Read `main.go`, then replace the UI initialization with:

```go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	tea "charm.land/bubbletea/v2"
	"redscout/lib"
	"redscout/lib/ui"
)

func main() {
	config, helpShown := lib.ParseFlags()

	if helpShown {
		flag.Usage()
		return
	}

	model := ui.NewModel(config)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		log.Fatal(err)
	}
}
```

- [ ] **Step 2: Build**

Run: `go build ./...`
Expected: May have compilation issues from old tview references. Fix any remaining import issues.

- [ ] **Step 3: Commit**

```bash
git add main.go
git commit -m "feat: switch main.go to bubbletea program"
```

---

### Task 14: Remove old tview UI code

**Files:**
- Delete: `lib/ui/ui.go`
- Delete: `lib/ui/theme/theme.go` (old tcell colors)
- Delete: `lib/ui/views/` (entire directory)

- [ ] **Step 1: Remove old files**

```bash
rm lib/ui/ui.go
rm lib/ui/theme/theme.go
rm -rf lib/ui/views/
```

- [ ] **Step 2: Remove tview/tcell dependencies**

```bash
go mod tidy
```

- [ ] **Step 3: Build**

Run: `go build ./...`
Expected: Success — all old tview code removed, bubbletea code in place.

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "chore: remove old tview UI code"
```

---

### Task 15: Integration testing and fixes

- [ ] **Step 1: Run the app**

```bash
go run main.go -h localhost -p 6379
```

Or with Azure Redis:
```bash
go run main.go -h <host> -p 10000 -a <password> --tls --scan-size=5000
```

- [ ] **Step 2: Verify each screen**

Check:
1. Disclaimer screen — centered, GitHub Dark colors, Y/N works
2. Loading screen — spinner animates, progress bar updates, scan → monitor transition
3. Main screen — title bar + 3 panels + tabs + table
4. Tab switching — N/L/B/H/T keys work
5. Table navigation — up/down arrows, selection highlighting
6. Namespace drill-down — Enter/→ to expand, Backspace/← to go back
7. Sort — 1-8 keys sort namespace/slowlog tables
8. S/M — trigger scan/monitor
9. ? — help overlay shows/hides
10. Q — quit works
11. Window resize — layout adapts

- [ ] **Step 3: Fix any issues found during testing**

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "fix: integration fixes for bubbletea migration"
```
