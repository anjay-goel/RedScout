package screens

import (
	"charm.land/bubbles/v2/table"
	"charm.land/lipgloss/v2"
	tea "charm.land/bubbletea/v2"
	"redscout/lib/ui/components"
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
	tableHeight := m.Height - 8
	if tableHeight < 10 {
		tableHeight = 20 // sensible default before window size is known
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

	return screen
}

func (m *MainModel) SetTab(tab components.Tab) {
	m.ActiveTab = tab
	m.rebuildTable()
}

func (m *MainModel) RefreshData(state *models.State) {
	m.State = state
	m.rebuildTable()
}
