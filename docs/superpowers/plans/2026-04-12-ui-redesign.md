# RedScout UI Redesign Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Redesign RedScout's TUI from teal/cyan to GitHub Dark aesthetic with restructured layout, contextual hints, and enriched tables.

**Architecture:** Create a theme.go color constants file, then restyle every component top-down: title bar (new), header panels (3 instead of 4), tab bar (underline style with action hints), tables (alternating rows, sort hints, enriched columns), help overlay (new). Remove shortcuts bar entirely.

**Tech Stack:** Go, tview, tcell

**Spec:** `docs/superpowers/specs/2026-04-12-ui-redesign-design.md`

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `lib/ui/theme.go` | Create | Color constants, themed cell helpers |
| `lib/ui/views/titlebar.go` | Create | Title bar component (RedScout + scan state) |
| `lib/ui/views/help_overlay.go` | Create | ? help overlay modal |
| `lib/ui/ui.go` | Modify | New layout assembly, ? key handler, remove shortcuts |
| `lib/ui/views/header.go` | Modify | 3 panels, new colors, compact layout |
| `lib/ui/views/body.go` | Modify | Remove shortcuts, new tab bar style, action hints |
| `lib/ui/views/components/namespace_table.go` | Modify | New colors, sort hints, remove Types, alternating rows |
| `lib/ui/views/components/bigkeys_table.go` | Modify | Add columns (Type, TTL, Namespace), new colors |
| `lib/ui/views/components/hotkeys_table.go` | Modify | Add columns (Command, Namespace), new colors |
| `lib/ui/views/components/slowlog_table.go` | Modify | New colors, sort hints |
| `lib/ui/views/components/progress_bar.go` | Modify | GitHub Dark colors |
| `models/special_keys.go` | Modify | Add fields to BigKey and HotKey structs |
| `lib/scanner/analytics.go` | Modify | Populate new BigKey/HotKey fields |

---

### Task 1: Create Theme Color Constants

**Files:**
- Create: `lib/ui/theme.go`

- [ ] **Step 1: Create theme.go with color constants**

```go
package ui

import "github.com/gdamore/tcell/v2"

var (
	ColorBg        = tcell.NewRGBColor(13, 17, 23)    // #0d1117
	ColorSurface   = tcell.NewRGBColor(22, 27, 34)    // #161b22
	ColorBorder    = tcell.NewRGBColor(33, 38, 45)     // #21262d
	ColorText      = tcell.NewRGBColor(201, 209, 217)  // #c9d1d9
	ColorMuted     = tcell.NewRGBColor(72, 79, 88)     // #484f58
	ColorSecondary = tcell.NewRGBColor(139, 148, 158)  // #8b949e
	ColorBlue      = tcell.NewRGBColor(88, 166, 255)   // #58a6ff
	ColorOrange    = tcell.NewRGBColor(240, 136, 62)   // #f0883e
	ColorGreen     = tcell.NewRGBColor(63, 185, 80)    // #3fb950
	ColorRed       = tcell.NewRGBColor(248, 81, 73)    // #f85149
)
```

- [ ] **Step 2: Build to verify**

Run: `go build ./...`
Expected: Success

- [ ] **Step 3: Commit**

```bash
git add lib/ui/theme.go
git commit -m "feat: add GitHub Dark theme color constants"
```

---

### Task 2: Add Fields to BigKey and HotKey Structs

**Files:**
- Modify: `models/special_keys.go`

- [ ] **Step 1: Add Type, TTL, Namespace to BigKey and Command, Namespace to HotKey**

In `models/special_keys.go`, replace the HotKey and BigKey structs:

```go
type HotKey struct {
	Key       Key
	Ops       float64
	Command   string
	Namespace string
}
```

```go
type BigKey struct {
	Key       Key
	Size      int64
	Type      string
	TTL       int64
	Namespace string
}
```

- [ ] **Step 2: Build to verify**

Run: `go build ./...`
Expected: Success

- [ ] **Step 3: Commit**

```bash
git add models/special_keys.go
git commit -m "feat: add Type/TTL/Namespace to BigKey, Command/Namespace to HotKey"
```

---

### Task 3: Populate New BigKey Fields in Scanner

**Files:**
- Modify: `lib/scanner/analytics.go` (ComputeBigKeysFromScanLog function, ~line 155)

- [ ] **Step 1: Update ComputeBigKeysFromScanLog to populate Type, TTL, Namespace**

In `lib/scanner/analytics.go`, in the `ComputeBigKeysFromScanLog` function, update the loop body where `BigKey` is constructed. Currently it reads `parts[0]` (key), `parts[1]` (memory). It should also read `parts[2]` (TTL) and `parts[3]` (type), and derive namespace. Replace the section starting at the `for scanner.Scan()` loop:

```go
	scanner := bufio.NewScanner(s.scanFile)
	h := &models.BigKeyMinHeap{}
	heap.Init(h)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 4 {
			continue
		}
		keyStr := parts[0]
		key := s.kp.NewKey(keyStr, false)

		memory, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			continue
		}
		ttl, _ := strconv.ParseInt(parts[2], 10, 64)
		keyType := parts[3]

		ns, _ := s.kp.Namespace(s.kp.NewKey(keyStr, true), models.Key{}, true)

		bk := models.BigKey{Key: key, Size: memory, Type: keyType, TTL: ttl, Namespace: ns}
		if int64(h.Len()) < s.Config.TopK {
			heap.Push(h, bk)
		} else if h.Len() > 0 && (*h)[0].Size < memory {
			heap.Pop(h)
			heap.Push(h, bk)
		}
	}
```

- [ ] **Step 2: Build to verify**

Run: `go build ./...`
Expected: Success

- [ ] **Step 3: Commit**

```bash
git add lib/scanner/analytics.go
git commit -m "feat: populate Type/TTL/Namespace in BigKey during scan"
```

---

### Task 4: Populate New HotKey Fields in Scanner

**Files:**
- Modify: `lib/scanner/analytics.go` (ComputeHotKeysFromMonitorLog function, ~line 201)

- [ ] **Step 1: Track top command per key and derive namespace**

In `lib/scanner/analytics.go`, in the `ComputeHotKeysFromMonitorLog` function, change the `keyOps` map to also track command frequency. Replace the entire function body:

```go
func (s *Scanner) ComputeHotKeysFromMonitorLog() error {
	s.muMonitor.Lock()
	defer s.muMonitor.Unlock()

	_, err := s.monitorFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(s.monitorFile)

	type keyStats struct {
		ops      int64
		commands map[string]int64
	}

	keyData := make(map[string]*keyStats)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		keyStr := parts[0]
		cmd := parts[1]

		stats, exists := keyData[keyStr]
		if !exists {
			stats = &keyStats{commands: make(map[string]int64)}
			keyData[keyStr] = stats
		}
		stats.ops++
		stats.commands[cmd]++
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	duration := s.State.TotalMonitorDuration.Seconds()
	if duration == 0 {
		duration = 1
	}

	h := &models.HotKeyMinHeap{}
	heap.Init(h)
	for k, stats := range keyData {
		opsPerSec := float64(stats.ops) / duration

		// Find top command
		topCmd := ""
		topCmdCount := int64(0)
		for cmd, count := range stats.commands {
			if count > topCmdCount {
				topCmd = cmd
				topCmdCount = count
			}
		}

		ns, _ := s.kp.Namespace(s.kp.NewKey(k, true), models.Key{}, true)

		hk := models.HotKey{
			Key:       s.kp.NewKey(k, false),
			Ops:       opsPerSec,
			Command:   strings.ToUpper(topCmd),
			Namespace: ns,
		}
		if int64(h.Len()) < s.Config.TopK {
			heap.Push(h, hk)
		} else if h.Len() > 0 && (*h)[0].Ops < opsPerSec {
			heap.Pop(h)
			heap.Push(h, hk)
		}
	}
	result := make(models.HotKeyList, h.Len())
	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).(models.HotKey)
	}
	s.State.HotKeys = result
	s.State.Updates <- s.State
	return nil
}
```

- [ ] **Step 2: Build to verify**

Run: `go build ./...`
Expected: Success

- [ ] **Step 3: Commit**

```bash
git add lib/scanner/analytics.go
git commit -m "feat: track top command and namespace for HotKey"
```

---

### Task 5: Restyle Progress Bar

**Files:**
- Modify: `lib/ui/views/components/progress_bar.go`

- [ ] **Step 1: Update progress bar to use GitHub Dark colors**

Replace the entire `CreateProgressBar` function in `lib/ui/views/components/progress_bar.go`:

```go
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

	// GitHub Dark: orange for fill, muted for empty
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
```

- [ ] **Step 2: Build to verify**

Run: `go build ./...`
Expected: Success

- [ ] **Step 3: Commit**

```bash
git add lib/ui/views/components/progress_bar.go
git commit -m "feat: restyle progress bar with GitHub Dark colors"
```

---

### Task 6: Create Title Bar Component

**Files:**
- Create: `lib/ui/views/titlebar.go`

- [ ] **Step 1: Create titlebar.go**

```go
package views

import (
	"fmt"
	"redscout/lib/utils"
	"redscout/models"

	"github.com/rivo/tview"
)

type TitleBar struct {
	View *tview.TextView
}

func NewTitleBar() *TitleBar {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	tv.SetBackgroundColor(ui.ColorBg)
	return &TitleBar{View: tv}
}

func (t *TitleBar) Update(state *models.State) {
	status := "[#3fb950]●[-]"
	if !state.ScanComplete {
		status = "[#f0883e]⠋[-]"
	}

	right := fmt.Sprintf("[#8b949e]%d scanned · %s monitored[-] %s  [#f0883e]?[-] [#484f58]help[-]",
		state.ScannedKeys,
		utils.FormatDuration(int64(state.TotalMonitorDuration.Seconds())),
		status,
	)

	// Pad with spaces to push right side to the edge
	t.View.SetText(fmt.Sprintf("[#58a6ff::b]RedScout[-::-]  %s", right))
}
```

Wait — `ui.ColorBg` won't work from the `views` package since `theme.go` is in the `ui` package. We need to either move theme to its own package or use hex colors directly in tview markup. Since tview supports `[#rrggbb]` syntax for dynamic colors, and tcell colors are needed for `SetBackgroundColor`, let's put color constants in a shared location.

- [ ] **Step 1 (revised): Move theme constants to a new package**

Create `lib/ui/theme/theme.go` instead:

```go
package theme

import "github.com/gdamore/tcell/v2"

var (
	ColorBg        = tcell.NewRGBColor(13, 17, 23)
	ColorSurface   = tcell.NewRGBColor(22, 27, 34)
	ColorBorder    = tcell.NewRGBColor(33, 38, 45)
	ColorText      = tcell.NewRGBColor(201, 209, 217)
	ColorMuted     = tcell.NewRGBColor(72, 79, 88)
	ColorSecondary = tcell.NewRGBColor(139, 148, 158)
	ColorBlue      = tcell.NewRGBColor(88, 166, 255)
	ColorOrange    = tcell.NewRGBColor(240, 136, 62)
	ColorGreen     = tcell.NewRGBColor(63, 185, 80)
	ColorRed       = tcell.NewRGBColor(248, 81, 73)
)
```

Update Task 1's file path from `lib/ui/theme.go` to `lib/ui/theme/theme.go`.

- [ ] **Step 2: Create titlebar.go**

```go
package views

import (
	"fmt"
	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"

	"github.com/rivo/tview"
)

type TitleBar struct {
	View *tview.TextView
}

func NewTitleBar() *TitleBar {
	tv := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	tv.SetBackgroundColor(theme.ColorBg)
	return &TitleBar{View: tv}
}

func (t *TitleBar) Update(state *models.State) {
	status := "[#3fb950]●[-]"
	if !state.ScanComplete {
		status = "[#f0883e]⠋[-]"
	}

	right := fmt.Sprintf("[#8b949e]%d scanned · %s monitored[-] %s  [#f0883e]?[-] [#484f58]help[-]",
		state.ScannedKeys,
		utils.FormatDuration(int64(state.TotalMonitorDuration.Seconds())),
		status,
	)

	t.View.SetText(fmt.Sprintf("[#58a6ff::b]RedScout[-::-]  %s", right))
}
```

- [ ] **Step 3: Build to verify**

Run: `go build ./...`
Expected: Success

- [ ] **Step 4: Commit**

```bash
git add lib/ui/theme/theme.go lib/ui/views/titlebar.go
git commit -m "feat: add theme package and title bar component"
```

---

### Task 7: Restyle Header to 3 Panels

**Files:**
- Modify: `lib/ui/views/header.go`

- [ ] **Step 1: Rewrite header.go with 3 panels and GitHub Dark styling**

Replace the entire file:

```go
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
	text := fmt.Sprintf(" [#484f58]SYSTEM[-]\n [#58a6ff::b]Redis[-::-] [#8b949e]v%s[-] [#484f58]· %s · %s · %d clients[-]",
		info.Server.RedisVersion,
		info.Server.OS,
		utils.FormatDuration(info.Server.Uptime),
		info.Clients.ConnectedClients,
	)
	header.system.SetText(text)
}

func (header *HeaderView) updatePerformance(info *models.RedisInfo) {
	totalKeys := info.Keyspace["db0"].Keys
	avgTTL := info.Keyspace["db0"].AvgTTL

	text := fmt.Sprintf(" [#484f58]PERFORMANCE[-]\n [#f0883e::b]%s[-::-] [#8b949e]keys[-] [#484f58]│[-] [#f0883e]%s[-] [#484f58]│[-] [#3fb950]%.1f%%[-] [#8b949e]hit[-] [#484f58]│[-] [#8b949e]ttl %s[-]",
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

	text := fmt.Sprintf(" [#484f58]MEMORY[-]\n %s [#484f58]· %s[-]%s",
		memLine,
		info.Memory.MemoryPolicy,
		cpuLine,
	)
	header.resources.SetText(text)
}
```

- [ ] **Step 2: Build to verify**

Run: `go build ./...`
Expected: Success

- [ ] **Step 3: Commit**

```bash
git add lib/ui/views/header.go
git commit -m "feat: restyle header to 3 panels with GitHub Dark theme"
```

---

### Task 8: Restyle Tab Bar and Remove Shortcuts Bar

**Files:**
- Modify: `lib/ui/views/body.go`

- [ ] **Step 1: Remove Shortcuts field, restyle tab bar with underline active indicator and action hints**

Replace the entire `body.go`:

```go
package views

import (
	"fmt"

	"github.com/rivo/tview"
	"redscout/lib/ui/theme"
	"redscout/lib/ui/views/components"
	"redscout/models"
)

type Tab string

const (
	TabNamespace Tab = "namespace"
	TabSlowLog   Tab = "slowlog"
	TabBigKeys   Tab = "bigkeys"
	TabHotKeys   Tab = "hotkeys"
)

type BodyView struct {
	ContentFlex *tview.Flex
	namespace   *components.Namespace
	slowLog     *components.SlowLogTable
	activeView  Tab
	TabBar      *tview.TextView
	app         *tview.Application
	bigKeyTable *tview.Table
	hotKeyTable *tview.Table
}

func NewBodyView(app *tview.Application) *BodyView {
	view := &BodyView{
		app:         app,
		ContentFlex: newContentFlex(),
		namespace:   components.NewNamespace(),
		slowLog:     components.NewSlowLogTable(),
		activeView:  TabNamespace,
		TabBar:      newTabBar(),
		bigKeyTable: components.NewBigKeyTable(),
		hotKeyTable: components.NewHotKeyTable(),
	}
	view.SetActiveView(TabNamespace)
	return view
}

var namespaceSortKeyMap = map[rune]string{
	'1': "Keys",
	'2': "Memory",
	'3': "TTL",
	'4': "% TTL",
	'5': "Get",
	'6': "Set",
	'7': "Del",
	'8': "Total Ops",
}

var slowLogSortKeyMap = map[rune]string{
	'1': "ID",
	'2': "Timestamp",
	'3': "Duration",
	'4': "Command",
}

func newContentFlex() *tview.Flex {
	f := tview.NewFlex().SetDirection(tview.FlexColumn)
	f.SetBackgroundColor(theme.ColorBg)
	return f
}

func newTabBar() *tview.TextView {
	tb := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignLeft)
	tb.SetBackgroundColor(theme.ColorBg)
	return tb
}

func tabLabel(name string, active bool) string {
	if active {
		return fmt.Sprintf("[#58a6ff::b]%s[-::-]", name)
	}
	return fmt.Sprintf("[#8b949e]%s[-]", name)
}

func (b *BodyView) updateTabBar() {
	tabs := fmt.Sprintf(" %s  %s  %s  %s",
		tabLabel("Namespaces", b.activeView == TabNamespace),
		tabLabel("Slow Log", b.activeView == TabSlowLog),
		tabLabel("Big Keys", b.activeView == TabBigKeys),
		tabLabel("Hot Keys", b.activeView == TabHotKeys),
	)
	// Right-aligned action hints — use enough padding
	actions := "[#f0883e]S[-] [#484f58]scan[-]  [#f0883e]M[-] [#484f58]monitor[-]"
	b.TabBar.SetText(fmt.Sprintf("%s                    %s", tabs, actions))
}

func (b *BodyView) SetActiveView(view Tab) {
	b.activeView = view
	b.updateTabBar()

	switch view {
	case TabNamespace:
		b.ContentFlex.Clear().AddItem(b.namespace.Flex, 0, 2, true)
		b.namespace.Table.Select(1, 0)
		b.app.SetFocus(b.namespace.Table)
	case TabSlowLog:
		b.ContentFlex.Clear().AddItem(b.slowLog.Table, 0, 2, true)
		b.slowLog.Table.Select(1, 0)
		b.app.SetFocus(b.slowLog.Table)
	case TabBigKeys:
		b.ContentFlex.Clear().AddItem(b.bigKeyTable, 0, 2, true)
		b.bigKeyTable.Select(1, 0)
		b.app.SetFocus(b.bigKeyTable)
	case TabHotKeys:
		b.ContentFlex.Clear().AddItem(b.hotKeyTable, 0, 2, true)
		b.hotKeyTable.Select(1, 0)
		b.app.SetFocus(b.hotKeyTable)
	}
}

func (b *BodyView) ToggleView() {
	switch b.activeView {
	case TabNamespace:
		b.SetActiveView(TabSlowLog)
	case TabSlowLog:
		b.SetActiveView(TabBigKeys)
	case TabBigKeys:
		b.SetActiveView(TabHotKeys)
	default:
		b.SetActiveView(TabNamespace)
	}
}

func (b *BodyView) Update(data *models.State) {
	b.slowLog.Update(data.SlowLogs)
	b.namespace.Update(data.CurrentPrefix, data.NamespaceStats)
	components.UpdateBigKeyTable(b.bigKeyTable, data.BigKeys)
	components.UpdateHotKeyTable(b.hotKeyTable, data.HotKeys)
}

func (b *BodyView) HandleInput(inp rune, state *models.State) {
	if inp == 'T' || inp == 't' {
		b.ToggleView()
		return
	}
	if inp == 'B' || inp == 'b' {
		b.SetActiveView(TabBigKeys)
		return
	}
	if inp == 'H' || inp == 'h' {
		b.SetActiveView(TabHotKeys)
		return
	}
	if inp == 'N' || inp == 'n' {
		b.SetActiveView(TabNamespace)
		return
	}
	if inp == 'L' || inp == 'l' {
		b.SetActiveView(TabSlowLog)
		return
	}
	if inp > '8' || inp < '1' {
		return
	}
	key := ""
	if b.activeView == TabNamespace {
		key = namespaceSortKeyMap[inp]
		if key == "" {
			return
		}
		state.NamespaceStats.Sort(key)
	} else if b.activeView == TabSlowLog {
		key = slowLogSortKeyMap[inp]
		if key == "" {
			return
		}
		state.SlowLogs.Sort(key)
	}
	b.Update(state)
}

func (b *BodyView) ActiveView() Tab {
	return b.activeView
}

func (b *BodyView) NamespaceTable() *tview.Table {
	return b.namespace.Table
}
```

- [ ] **Step 2: Build to verify**

Run: `go build ./...`
Expected: May fail due to references to `b.Shortcuts` in `ui.go`. That's expected — we'll fix `ui.go` in the next task.

- [ ] **Step 3: Commit**

```bash
git add lib/ui/views/body.go
git commit -m "feat: restyle tab bar, remove shortcuts bar"
```

---

### Task 9: Create Help Overlay

**Files:**
- Create: `lib/ui/views/help_overlay.go`

- [ ] **Step 1: Create help overlay component**

```go
package views

import (
	"github.com/rivo/tview"
	"redscout/lib/ui/theme"
)

func NewHelpOverlay() *tview.Flex {
	helpText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter)
	helpText.SetBackgroundColor(theme.ColorSurface)
	helpText.SetBorderColor(theme.ColorBorder)
	helpText.SetBorder(true)

	helpText.SetText(
		"[#58a6ff::b]Keyboard Shortcuts[-::-]\n" +
			"[#484f58]press ? or Esc to close[-]\n\n" +
			"[#f0883e]NAVIGATION[-]\n" +
			"[#f0883e]N[-] [#8b949e]Namespaces[-]    [#f0883e]L[-] [#8b949e]Slow Log[-]    [#f0883e]B[-] [#8b949e]Big Keys[-]    [#f0883e]H[-] [#8b949e]Hot Keys[-]    [#f0883e]T[-] [#8b949e]Next tab[-]\n\n" +
			"[#f0883e]ACTIONS[-]\n" +
			"[#f0883e]S[-] [#8b949e]Run SCAN[-]    [#f0883e]M[-] [#8b949e]Run MONITOR[-]    [#f0883e]Q[-] [#8b949e]Quit[-]\n\n" +
			"[#f0883e]NAMESPACE[-]\n" +
			"[#f0883e]→/Enter[-] [#8b949e]Drill down[-]    [#f0883e]←/Bksp[-] [#8b949e]Level up[-]    [#f0883e]1-8[-] [#8b949e]Sort columns[-]\n",
	)

	// Center the help overlay
	overlay := tview.NewFlex().SetDirection(tview.FlexRow)
	overlay.SetBackgroundColor(theme.ColorBg)
	overlay.AddItem(nil, 0, 1, false)
	inner := tview.NewFlex().SetDirection(tview.FlexColumn)
	inner.AddItem(nil, 0, 1, false)
	inner.AddItem(helpText, 80, 0, false)
	inner.AddItem(nil, 0, 1, false)
	overlay.AddItem(inner, 15, 0, false)
	overlay.AddItem(nil, 0, 1, false)

	return overlay
}
```

- [ ] **Step 2: Build to verify**

Run: `go build ./...`
Expected: Success

- [ ] **Step 3: Commit**

```bash
git add lib/ui/views/help_overlay.go
git commit -m "feat: add help overlay modal for ? key"
```

---

### Task 10: Rewrite Main UI Layout Assembly

**Files:**
- Modify: `lib/ui/ui.go`

- [ ] **Step 1: Rewrite ui.go with new layout: title bar, 3-panel header, tab bar, content. Add ? key handler, help overlay via tview.Pages. Remove shortcuts bar. Update loading/disclaimer/error screens with GitHub Dark colors.**

Replace the entire file:

```go
package ui

import (
	"fmt"
	"redscout/lib/scanner"
	"redscout/lib/ui/theme"
	"redscout/lib/ui/views"
	"redscout/lib/ui/views/components"
	"redscout/models"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type AppUI struct {
	config  *models.Config
	scanner *scanner.Scanner

	app       *tview.Application
	pages     *tview.Pages
	titleBar  *views.TitleBar
	headers   *views.HeaderView
	body      *views.BodyView

	loadingTextView *tview.TextView
	initialisedLayout bool
	helpVisible       bool
}

func NewAppUI(cfg models.Config) *AppUI {
	app := tview.NewApplication()

	ui := &AppUI{
		config:            &cfg,
		app:               app,
		pages:             tview.NewPages(),
		body:              views.NewBodyView(app),
		headers:           views.NewHeaderView(),
		titleBar:          views.NewTitleBar(),
		initialisedLayout: false,
		helpVisible:       false,
	}

	return ui
}

func (ui *AppUI) createDisclaimerScreen() {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBackgroundColor(theme.ColorBg)

	disclaimer := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[#f85149]DISCLAIMER[-]\n\n" +
			"[#8b949e]RedScout will run the 'MONITOR' command on your Redis instance.[-]\n" +
			"[#f0883e]This can impact Redis performance. Use with caution on production environments.[-]\n\n" +
			"[#c9d1d9]Do you want to continue?[-]\n\n" +
			"[#3fb950]Y[-]es / [#f85149]N[-]o")
	disclaimer.SetBorder(true)
	disclaimer.SetBorderColor(theme.ColorBorder)
	disclaimer.SetBackgroundColor(theme.ColorBg)
	disclaimer.SetBorderPadding(2, 2, 2, 2)

	flex.AddItem(disclaimer, 0, 1, false)
	ui.app.SetRoot(flex, true)

	ui.app.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		switch e.Rune() {
		case 'y', 'Y', '\r':
			ui.start()
			return nil
		case 'n', 'N', 'q', 'Q':
			ui.app.Stop()
			return nil
		}
		return e
	})
}

func (ui *AppUI) createErrorScreen(errorMsg string) {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBackgroundColor(theme.ColorBg)

	errorText := tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText(fmt.Sprintf("[#f85149]ERROR[-]\n\n[#c9d1d9]%s[-]\n\n[#3fb950]R[-]etry / [#f85149]Q[-]uit", errorMsg))
	errorText.SetBorder(true)
	errorText.SetBorderColor(theme.ColorBorder)
	errorText.SetBackgroundColor(theme.ColorBg)
	errorText.SetBorderPadding(2, 2, 2, 2)

	flex.AddItem(errorText, 0, 1, false)
	ui.app.QueueUpdateDraw(func() {
		ui.app.SetRoot(flex, true)
	})

	ui.app.SetInputCapture(func(e *tcell.EventKey) *tcell.EventKey {
		switch e.Rune() {
		case 'r', 'R', '\r':
			ui.start()
			return nil
		case 'q', 'Q':
			ui.app.Stop()
			return nil
		}
		return e
	})
}

func (ui *AppUI) start() {
	ui.createLoadingScreen()
	go func() {
		s, err := scanner.NewScanner(ui.config)
		if err != nil {
			ui.createErrorScreen(fmt.Sprintf("Error initializing scanner:\n%v", err))
			return
		}
		ui.scanner = s

		go s.Start()
		go ui.stateUpdateListener()
	}()
}

func (ui *AppUI) createLoadingScreen() {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBackgroundColor(theme.ColorBg)

	ui.loadingTextView = tview.NewTextView().
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetText("[#f0883e]Analysing Redis ⠋\n\n[#c9d1d9]Initializing...[-]")
	ui.loadingTextView.SetBorder(true)
	ui.loadingTextView.SetBorderColor(theme.ColorBorder)
	ui.loadingTextView.SetBackgroundColor(theme.ColorBg)
	ui.loadingTextView.SetBorderPadding(2, 2, 2, 2)

	flex.AddItem(ui.loadingTextView, 0, 1, false)
	ui.app.SetRoot(flex, true)

	spinner := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		i := 0

		for {
			select {
			case <-ticker.C:
				ui.app.QueueUpdateDraw(func() {
					var text string
					if ui.scanner == nil || ui.scanner.State == nil {
						text = fmt.Sprintf("[#f0883e]Analysing Redis %c\n\n[#c9d1d9]Initializing...[-]", spinner[i%len(spinner)])
					} else if ui.scanner.State.ScanComplete {
						ticker.Stop()
						return
					} else {
						var progressInfo string

						if ui.scanner.State.ScanProgress < 100 {
							scanBar := components.CreateProgressBar(ui.scanner.State.ScanProgress, 100, 40)
							progressInfo = fmt.Sprintf("\n\n[#8b949e]%s[-]\n%s\n[#c9d1d9]%d keys collected[-]", ui.scanner.State.Status, scanBar, ui.scanner.State.ScannedKeys)
						} else if ui.scanner.State.MonitorDurationTotal == 0 {
							progressInfo = "\n\n[#8b949e]Starting monitor...[-]"
						} else if ui.scanner.State.MonitorProgress < 100 {
							elapsed := time.Duration(float64(ui.scanner.State.MonitorDurationTotal) * ui.scanner.State.MonitorProgress / 100)
							monitorBar := components.CreateProgressBar(ui.scanner.State.MonitorProgress, 100, 40)
							progressInfo = fmt.Sprintf("\n\n[#8b949e]Monitor Progress:[-]\n%s\n[#c9d1d9]%v / %v[-]", monitorBar, elapsed.Round(time.Second), ui.scanner.State.MonitorDurationTotal)
						}

						text = fmt.Sprintf("[#f0883e]Analysing Redis %c[-]\n\n%s", spinner[i%len(spinner)], progressInfo)
					}
					ui.loadingTextView.SetText(text)
				})
				i++
			}
		}
	}()
}

func (ui *AppUI) createMainScreen() {
	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBackgroundColor(theme.ColorBg)

	// Title bar (1 row)
	flex.AddItem(ui.titleBar.View, 1, 0, false)

	// Header panels (4 rows — border + 2 content lines)
	flex.AddItem(ui.headers.HeaderFlex, 4, 0, false)

	// Tab bar (1 row)
	ui.body.TabBar.SetBorder(false)
	flex.AddItem(ui.body.TabBar, 1, 0, false)

	// Content area (fills remaining)
	flex.AddItem(ui.body.ContentFlex, 0, 1, true)

	// Set up pages for help overlay
	ui.pages.AddPage("main", flex, true, true)
	ui.pages.AddPage("help", views.NewHelpOverlay(), true, false)

	ui.app.SetInputCapture(ui.handleInput)
	ui.app.SetRoot(ui.pages, true)
}

func (ui *AppUI) stateUpdateListener() {
	for update := range ui.scanner.State.Updates {
		if !update.ScanComplete {
			continue
		}

		ui.app.QueueUpdateDraw(func() {
			if !ui.initialisedLayout {
				ui.createMainScreen()
				ui.initialisedLayout = true
			}
			ui.update(update)
		})
	}
}

func (ui *AppUI) Run() error {
	ui.createDisclaimerScreen()
	return ui.app.Run()
}

func (ui *AppUI) update(ctx *models.State) {
	ui.titleBar.Update(ctx)
	ui.headers.Update(ctx)
	ui.body.Update(ctx)
}

func (ui *AppUI) handleInput(e *tcell.EventKey) *tcell.EventKey {
	// Help overlay toggle
	if e.Rune() == '?' {
		if ui.helpVisible {
			ui.pages.HidePage("help")
			ui.helpVisible = false
		} else {
			ui.pages.ShowPage("help")
			ui.helpVisible = true
		}
		return nil
	}
	if e.Key() == tcell.KeyEscape && ui.helpVisible {
		ui.pages.HidePage("help")
		ui.helpVisible = false
		return nil
	}
	if ui.helpVisible {
		return nil // swallow all keys when help is shown
	}

	changed := true

	switch e.Key() {
	case tcell.KeyEnter, tcell.KeyRight:
		if ui.body.ActiveView() == "namespace" {
			row, _ := ui.body.NamespaceTable().GetSelection()
			if row <= 0 || row > len(ui.scanner.State.NamespaceStats) {
				return nil
			}
			namespace := ui.scanner.State.NamespaceStats[row-1].Namespace

			ui.scanner.DrillDownNamespace(namespace)

			return nil
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2, tcell.KeyLeft:
		if ui.body.ActiveView() == "namespace" {
			ui.scanner.LevelUpNamespace()
			return nil
		}
	}

	switch e.Rune() {
	case '1', '2', '3', '4', '5', '6', '7', '8', 't', 'T', 'n', 'N', 'l', 'L', 'b', 'B', 'h', 'H':
		ui.body.HandleInput(e.Rune(), ui.scanner.State)
	case 'q', 'Q':
		ui.app.Stop()
		ui.scanner.Close()
		return nil
	case 's', 'S':
		go func() {
			err := ui.scanner.ScanMemory()
			if err == nil {
				_ = ui.scanner.ComputeNamespaceStats()
				_ = ui.scanner.ComputeBigKeysFromScanLog()
			}
		}()
	case 'm', 'M':
		go func() {
			err := ui.scanner.MonitorOps()
			if err == nil {
				_ = ui.scanner.ComputeNamespaceStats()
				_ = ui.scanner.ComputeHotKeysFromMonitorLog()
			}
		}()
	default:
		changed = false
	}
	if changed {
		ui.update(ui.scanner.State)
	}
	return e
}
```

- [ ] **Step 2: Build to verify**

Run: `go build ./...`
Expected: Success

- [ ] **Step 3: Commit**

```bash
git add lib/ui/ui.go
git commit -m "feat: rewrite main UI layout with GitHub Dark theme, title bar, help overlay"
```

---

### Task 11: Restyle Namespace Table

**Files:**
- Modify: `lib/ui/views/components/namespace_table.go`

- [ ] **Step 1: Rewrite with GitHub Dark colors, sort hints in headers, alternating rows, remove Types column, add breadcrumb navigation hints**

Replace the entire file:

```go
package components

import (
	"fmt"
	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Namespace struct {
	Title *tview.TextView
	Table *tview.Table
	Flex  *tview.Flex
}

func NewNamespace() *Namespace {
	ns := &Namespace{}
	ns.Table = tview.NewTable().SetFixed(1, 0)
	ns.Table.SetSelectable(true, false)
	ns.Table.SetBorders(false)
	ns.Table.SetBackgroundColor(theme.ColorBg)
	ns.Table.SetSelectedStyle(tcell.StyleDefault.
		Background(theme.ColorBorder).
		Foreground(theme.ColorText))

	ns.Title = tview.NewTextView()
	ns.Title.SetDynamicColors(true)
	ns.Title.SetBackgroundColor(theme.ColorBg)
	ns.Title.SetText(" [#f0883e]/ root[-]")

	ns.Flex = tview.NewFlex()
	ns.Flex.SetDirection(tview.FlexRow)
	ns.Flex.SetBackgroundColor(theme.ColorBg)
	ns.Flex.SetBorderPadding(0, 0, 0, 0)
	ns.Flex.AddItem(ns.Title, 1, -1, false)
	ns.Flex.AddItem(ns.Table, 0, 1, true)

	return ns
}

func (ns *Namespace) Update(prefix models.Key, stats models.NamespaceMetricList) {
	headers := []string{"Namespace", "~Keys 1", "~Memory 2", "Avg TTL 3", "% TTL 4", "GET/s 5", "SET/s 6", "DEL/s 7", "OPS/s 8"}

	ns.Table.Clear()
	for i, h := range headers {
		align := tview.AlignLeft
		if i != 0 {
			align = tview.AlignRight
		}
		cell := tview.NewTableCell(h).
			SetTextColor(theme.ColorSecondary).
			SetBackgroundColor(theme.ColorBg).
			SetSelectable(false).
			SetAlign(align)
		ns.Table.SetCell(0, i, cell)
	}

	nsPad := utils.MaxKeyDisplayLen
	for i, row := range stats {
		nsVal := utils.TruncateKey(row.Namespace)
		values := []string{
			fmt.Sprintf("%-*s", nsPad, nsVal),
			fmt.Sprintf("%12s", utils.FormatNumber(float64(row.EstKeys))),
			fmt.Sprintf("%12s", utils.FormatBytes(row.EstMemory)),
			fmt.Sprintf("%12s", utils.FormatDuration(row.AvgTTL)),
			fmt.Sprintf("%11.1f%%", row.TTLPercent*100),
			fmt.Sprintf("%10.1f/s", row.Ops[models.GetOp]),
			fmt.Sprintf("%10.1f/s", row.Ops[models.SetOp]),
			fmt.Sprintf("%10.1f/s", row.Ops[models.DelOp]),
			fmt.Sprintf("%10.1f/s", row.Ops[models.TotalOp]),
		}

		colors := []tcell.Color{
			theme.ColorText,
			theme.ColorOrange,
			theme.ColorBlue,
			theme.ColorSecondary,
			theme.ColorSecondary,
			theme.ColorSecondary,
			theme.ColorSecondary,
			theme.ColorSecondary,
			theme.ColorSecondary,
		}

		rowBg := theme.ColorBg
		if i%2 == 1 {
			rowBg = theme.ColorSurface
		}

		for j, val := range values {
			align := tview.AlignLeft
			if j != 0 {
				align = tview.AlignRight
			}
			cell := tview.NewTableCell(val).
				SetTextColor(colors[j]).
				SetAlign(align).
				SetExpansion(0).
				SetBackgroundColor(rowBg)

			ns.Table.SetCell(i+1, j, cell)
		}
	}

	ns.Table.SetFixed(1, 0)
	ns.Table.ScrollToBeginning()

	separator := " › "
	if len(prefix) == 0 {
		ns.Title.SetText(" [#f0883e]/ root[-]                                                          [#f0883e]→[-] [#484f58]drill[-]  [#f0883e]←[-] [#484f58]back[-]")
	} else {
		path := "/ root" + separator + strings.Join(prefix, separator)
		ns.Title.SetText(fmt.Sprintf(" [#f0883e]%s[-]                                                          [#f0883e]→[-] [#484f58]drill[-]  [#f0883e]←[-] [#484f58]back[-]", path))
	}
}
```

- [ ] **Step 2: Build to verify**

Run: `go build ./...`
Expected: Success

- [ ] **Step 3: Commit**

```bash
git add lib/ui/views/components/namespace_table.go
git commit -m "feat: restyle namespace table with GitHub Dark, sort hints, alternating rows"
```

---

### Task 12: Restyle Big Keys Table with Enriched Columns

**Files:**
- Modify: `lib/ui/views/components/bigkeys_table.go`

- [ ] **Step 1: Rewrite with new columns (Type, TTL, Namespace), GitHub Dark colors, alternating rows**

Replace the entire file:

```go
package components

import (
	"fmt"
	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func NewBigKeyTable() *tview.Table {
	table := tview.NewTable().SetFixed(1, 0)
	table.SetSelectable(true, false)
	table.SetBorders(false)
	table.SetBorderPadding(0, 0, 1, 0)
	table.SetBackgroundColor(theme.ColorBg)
	table.SetSelectedStyle(tcell.StyleDefault.
		Background(theme.ColorBorder).
		Foreground(theme.ColorText))
	return table
}

func UpdateBigKeyTable(table *tview.Table, bigKeys models.BigKeyList) {
	headers := []string{"Key", "Size 1", "Type 2", "TTL 3", "Namespace"}
	headerColors := theme.ColorSecondary

	table.Clear()
	for i, h := range headers {
		align := tview.AlignLeft
		if i >= 1 && i <= 3 {
			align = tview.AlignRight
		}
		cell := tview.NewTableCell(h).
			SetTextColor(headerColors).
			SetBackgroundColor(theme.ColorBg).
			SetSelectable(false).
			SetAlign(align)
		table.SetCell(0, i, cell)
	}

	for i, row := range bigKeys {
		ttlStr := "-"
		if row.TTL > 0 {
			ttlStr = utils.FormatDuration(row.TTL)
		}

		values := []string{
			fmt.Sprintf("%-*s", utils.MaxKeyDisplayLen, utils.TruncateKey(row.Key.String())),
			fmt.Sprintf("%12s", utils.FormatBytes(row.Size)),
			fmt.Sprintf("%8s", row.Type),
			fmt.Sprintf("%12s", ttlStr),
			row.Namespace,
		}

		colors := []tcell.Color{
			theme.ColorText,
			theme.ColorOrange,
			theme.ColorSecondary,
			theme.ColorGreen,
			theme.ColorMuted,
		}

		rowBg := theme.ColorBg
		if i%2 == 1 {
			rowBg = theme.ColorSurface
		}

		for j, val := range values {
			align := tview.AlignLeft
			if j >= 1 && j <= 3 {
				align = tview.AlignRight
			}
			cell := tview.NewTableCell(val).
				SetTextColor(colors[j]).
				SetAlign(align).
				SetExpansion(0).
				SetBackgroundColor(rowBg)
			table.SetCell(i+1, j, cell)
		}
	}
	table.ScrollToBeginning()
}
```

- [ ] **Step 2: Build to verify**

Run: `go build ./...`
Expected: Success

- [ ] **Step 3: Commit**

```bash
git add lib/ui/views/components/bigkeys_table.go
git commit -m "feat: restyle big keys table with enriched columns and GitHub Dark"
```

---

### Task 13: Restyle Hot Keys Table with Enriched Columns

**Files:**
- Modify: `lib/ui/views/components/hotkeys_table.go`

- [ ] **Step 1: Rewrite with new columns (Command, Namespace), GitHub Dark colors, alternating rows**

Replace the entire file:

```go
package components

import (
	"fmt"
	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func NewHotKeyTable() *tview.Table {
	table := tview.NewTable().SetFixed(1, 0)
	table.SetSelectable(true, false)
	table.SetBorders(false)
	table.SetBorderPadding(0, 0, 1, 0)
	table.SetBackgroundColor(theme.ColorBg)
	table.SetSelectedStyle(tcell.StyleDefault.
		Background(theme.ColorBorder).
		Foreground(theme.ColorText))
	return table
}

func UpdateHotKeyTable(table *tview.Table, hotKeys models.HotKeyList) {
	headers := []string{"Key", "Ops/s 1", "Command 2", "Namespace"}

	table.Clear()
	for i, h := range headers {
		align := tview.AlignLeft
		if i == 1 {
			align = tview.AlignRight
		}
		cell := tview.NewTableCell(h).
			SetTextColor(theme.ColorSecondary).
			SetBackgroundColor(theme.ColorBg).
			SetSelectable(false).
			SetAlign(align)
		table.SetCell(0, i, cell)
	}

	for i, row := range hotKeys {
		values := []string{
			fmt.Sprintf("%-*s", utils.MaxKeyDisplayLen, utils.TruncateKey(row.Key.String())),
			fmt.Sprintf("%10.1f/s", row.Ops),
			fmt.Sprintf("%-8s", row.Command),
			row.Namespace,
		}

		colors := []tcell.Color{
			theme.ColorText,
			theme.ColorBlue,
			theme.ColorSecondary,
			theme.ColorMuted,
		}

		rowBg := theme.ColorBg
		if i%2 == 1 {
			rowBg = theme.ColorSurface
		}

		for j, val := range values {
			align := tview.AlignLeft
			if j == 1 {
				align = tview.AlignRight
			}
			cell := tview.NewTableCell(val).
				SetTextColor(colors[j]).
				SetAlign(align).
				SetExpansion(0).
				SetBackgroundColor(rowBg)
			table.SetCell(i+1, j, cell)
		}
	}
	table.ScrollToBeginning()
}
```

- [ ] **Step 2: Build to verify**

Run: `go build ./...`
Expected: Success

- [ ] **Step 3: Commit**

```bash
git add lib/ui/views/components/hotkeys_table.go
git commit -m "feat: restyle hot keys table with enriched columns and GitHub Dark"
```

---

### Task 14: Restyle Slow Log Table

**Files:**
- Modify: `lib/ui/views/components/slowlog_table.go`

- [ ] **Step 1: Rewrite with GitHub Dark colors, sort hints, alternating rows**

Replace the entire file:

```go
package components

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"redscout/lib/ui/theme"
	"redscout/lib/utils"
	"redscout/models"
)

type SlowLogTable struct {
	Table *tview.Table
}

func NewSlowLogTable() *SlowLogTable {
	table := tview.NewTable().SetFixed(1, 0)
	table.SetSelectable(true, false)
	table.SetBorders(false)
	table.SetBorderPadding(0, 0, 1, 0)
	table.SetBackgroundColor(theme.ColorBg)
	table.SetSelectedStyle(tcell.StyleDefault.
		Background(theme.ColorBorder).
		Foreground(theme.ColorText))
	return &SlowLogTable{table}
}

func (sl *SlowLogTable) Update(slowLogs models.SlowLogList) {
	if len(slowLogs) == 0 {
		return
	}

	headers := []string{"ID 1", "Timestamp 2", "Duration 3", "Command 4", "Arguments"}

	sl.Table.Clear()
	for i, h := range headers {
		align := tview.AlignLeft
		if i == 0 || i == 2 {
			align = tview.AlignRight
		}
		cell := tview.NewTableCell(h).
			SetTextColor(theme.ColorSecondary).
			SetBackgroundColor(theme.ColorBg).
			SetSelectable(false).
			SetAlign(align)
		sl.Table.SetCell(0, i, cell)
	}

	for i, log := range slowLogs {
		command := ""
		var args []string
		if len(log.Args) > 0 {
			command = strings.ToUpper(log.Args[0])
			args = log.Args[1:]
		}

		values := []string{
			fmt.Sprintf("%d ", log.ID),
			log.Time.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%12d ms", log.Duration.Milliseconds()),
			command,
			utils.TruncateKey(strings.Join(args, " ")),
		}

		colors := []tcell.Color{
			theme.ColorText,
			theme.ColorSecondary,
			theme.ColorOrange,
			theme.ColorGreen,
			theme.ColorMuted,
		}

		rowBg := theme.ColorBg
		if i%2 == 1 {
			rowBg = theme.ColorSurface
		}

		for j, val := range values {
			align := tview.AlignLeft
			if j == 0 || j == 2 {
				align = tview.AlignRight
			}
			cell := tview.NewTableCell(val).
				SetTextColor(colors[j]).
				SetAlign(align).
				SetExpansion(0).
				SetBackgroundColor(rowBg)
			sl.Table.SetCell(i+1, j, cell)
		}
	}

	sl.Table.SetFixed(1, 0)
}
```

- [ ] **Step 2: Build to verify**

Run: `go build ./...`
Expected: Success

- [ ] **Step 3: Commit**

```bash
git add lib/ui/views/components/slowlog_table.go
git commit -m "feat: restyle slow log table with GitHub Dark and sort hints"
```

---

### Task 15: Remove Unused Constants and Clean Up

**Files:**
- Modify: `lib/ui/views/components/namespace_table.go` (remove `StatsHeader` constant)
- Modify: `lib/ui/views/components/slowlog_table.go` (remove `SlowLogHeader` constant)
- Modify: `lib/ui/views/components/bigkeys_table.go` (remove `BigKeysShortcutsText` constant)
- Modify: `lib/ui/views/components/hotkeys_table.go` (remove `HotKeysShortcutsText` constant)

- [ ] **Step 1: Remove shortcut text constants from all table files**

These constants were used by the now-removed shortcuts bar. They should have been removed in the table rewrites above. Verify they're gone by grepping:

Run: `grep -r "ShortcutsText\|StatsHeader\|SlowLogHeader" lib/ui/views/components/`
Expected: No output (already removed in the full rewrites above)

- [ ] **Step 2: Final build**

Run: `go build ./...`
Expected: Success

- [ ] **Step 3: Run tests**

Run: `go test ./...`
Expected: All tests pass

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "chore: final cleanup of unused shortcut constants"
```

---

### Task 16: Manual Verification

- [ ] **Step 1: Run the app against a local or test Redis**

Run: `go run main.go -h localhost -p 6379`

Or with the Azure Redis:
`go run main.go -h <host> -p 10000 -a <password> --tls --scan-size=5000`

- [ ] **Step 2: Verify each screen**

Check:
1. Disclaimer screen — GitHub Dark colors, red title, green/red Y/N
2. Loading screen — orange spinner, progress bar with GitHub Dark colors
3. Main screen layout — title bar + 3 panels + tab bar + content (no shortcuts bar)
4. Title bar — "RedScout" left, scan state right, "? help" visible
5. Header panels — SYSTEM, PERFORMANCE, MEMORY with correct data
6. Tab bar — underline-style active tab, "S scan  M monitor" on right
7. Namespace table — sort hints in headers, alternating row colors, no Types column
8. Big Keys table — Key, Size, Type, TTL, Namespace columns
9. Hot Keys table — Key, Ops/s, Command, Namespace columns
10. Slow Log table — sort hints in headers, alternating rows
11. ? help overlay — shows/hides on ? key, Esc dismisses
12. Keyboard shortcuts — all still work (S, M, 1-8, T, N/L/B/H, →/←, Q)

- [ ] **Step 3: Final commit if any tweaks needed**
