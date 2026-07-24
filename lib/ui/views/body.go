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
	ContentFlex  *tview.Flex
	namespace    *components.Namespace
	slowLog      *components.SlowLogTable
	activeView   Tab
	TabBar       *tview.TextView
	app          *tview.Application
	bigKeyTable  *tview.Table
	hotKeyTable  *tview.Table
	keyValueView *components.KeyValueView
	showingValue bool
}

func NewBodyView(app *tview.Application) *BodyView {
	view := &BodyView{
		app:          app,
		ContentFlex:  newContentFlex(),
		namespace:    components.NewNamespace(),
		slowLog:      components.NewSlowLogTable(),
		activeView:   TabNamespace,
		TabBar:       newTabBar(),
		bigKeyTable:  components.NewBigKeyTable(),
		hotKeyTable:  components.NewHotKeyTable(),
		keyValueView: components.NewKeyValueView(),
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
	tb.SetBorderPadding(0, 0, 0, 0)
	return tb
}

func tabLabel(name string, key string, active bool) string {
	shortcut := fmt.Sprintf("[#7d8590][[-][#f0883e]%s[-][#7d8590]][-]", key)
	if active {
		// Active tab: bright blue + bold so the current view stands out.
		return fmt.Sprintf("[#58a6ff::b]%s[-::-] %s", name, shortcut)
	}
	return fmt.Sprintf("[#8b949e]%s[-] %s", name, shortcut)
}

func (b *BodyView) updateTabBar() {
	// Vertical dividers between tabs (not on the outer edges) so the row
	// reads as a bar of selectable tabs.
	div := "[#7d8590]│[-]"
	tabs := fmt.Sprintf(" %s %s %s %s %s %s %s",
		tabLabel("Namespaces", "N", b.activeView == TabNamespace),
		div, tabLabel("Slow Log", "L", b.activeView == TabSlowLog),
		div, tabLabel("Big Keys", "B", b.activeView == TabBigKeys),
		div, tabLabel("Hot Keys", "H", b.activeView == TabHotKeys))
	b.TabBar.SetText(tabs)
}

func (b *BodyView) SetActiveView(view Tab) {
	b.activeView = view
	b.showingValue = false
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

	// Show key value viewer when key value data is available
	if data.KeyValue != nil && !b.showingValue {
		b.ShowKeyValue(data.KeyValue)
	} else if data.KeyValue != nil && b.showingValue {
		b.keyValueView.Update(data.KeyValue)
	}
}

func (b *BodyView) IsShowingValue() bool {
	return b.showingValue
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

func (b *BodyView) BigKeyTable() *tview.Table {
	return b.bigKeyTable
}

func (b *BodyView) HotKeyTable() *tview.Table {
	return b.hotKeyTable
}

func (b *BodyView) ShowKeyValue(info *models.KeyValueInfo) {
	b.showingValue = true
	b.keyValueView.Update(info)
	b.ContentFlex.Clear().AddItem(b.keyValueView.View, 0, 1, true)
	b.app.SetFocus(b.keyValueView.View)
}
