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
	tb.SetBorderPadding(0, 1, 0, 0) // bottom padding acts as separator
	return tb
}

func tabLabel(name string, key string, active bool) string {
	shortcut := fmt.Sprintf("[#484f58][[-][#f0883e]%s[-][#484f58]][-]", key)
	if active {
		return fmt.Sprintf("[#58a6ff::b]%s[-::-] %s", name, shortcut)
	}
	return fmt.Sprintf("[#8b949e]%s[-] %s", name, shortcut)
}

func (b *BodyView) updateTabBar() {
	tabs := fmt.Sprintf(" %s  %s  %s  %s",
		tabLabel("Namespaces", "N", b.activeView == TabNamespace),
		tabLabel("Slow Log", "L", b.activeView == TabSlowLog),
		tabLabel("Big Keys", "B", b.activeView == TabBigKeys),
		tabLabel("Hot Keys", "H", b.activeView == TabHotKeys),
	)
	b.TabBar.SetText(tabs)
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
