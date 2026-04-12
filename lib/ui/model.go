package ui

import (
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"redscout/lib/scanner"
	"redscout/lib/ui/components"
	"redscout/lib/ui/screens"
	"redscout/lib/utils"
	"redscout/models"
)

type activeScreen int

const (
	screenDisclaimer activeScreen = iota
	screenLoading
	screenError
	screenMain
)

type Model struct {
	config  *models.Config
	scanner *scanner.Scanner
	screen  activeScreen
	err     string

	spinner spinner.Model
	main    screens.MainModel

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

func waitForStateUpdate(s *scanner.Scanner) tea.Cmd {
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
		return m, waitForStateUpdate(m.scanner)

	case ScanCompleteMsg:
		m.main = screens.NewMainModel(msg.State, m.width, m.height)
		m.screen = screenMain
		return m, waitForStateUpdate(m.scanner)

	case ErrorMsg:
		m.err = msg.Err.Error()
		m.screen = screenError
		return m, nil

	case scannerReadyMsg:
		m.scanner = msg.scanner
		return m, waitForStateUpdate(m.scanner)

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

// scannerReadyMsg is sent when scanner is created successfully
type scannerReadyMsg struct {
	scanner *scanner.Scanner
}

func (m Model) startScanner() tea.Cmd {
	cfg := m.config
	return func() tea.Msg {
		s, err := scanner.NewScanner(cfg)
		if err != nil {
			return ErrorMsg{Err: err}
		}
		go s.Start()
		return scannerReadyMsg{scanner: s}
	}
}

func (m Model) handleKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch m.screen {
	case screenDisclaimer:
		switch key {
		case "y", "Y", "enter":
			m.screen = screenLoading
			return m, m.startScanner()
		case "n", "N", "q", "Q":
			return m, tea.Quit
		}

	case screenError:
		switch key {
		case "r", "R", "enter":
			m.screen = screenLoading
			return m, m.startScanner()
		case "q", "Q":
			return m, tea.Quit
		}

	case screenMain:
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
			if m.scanner != nil {
				go func() {
					err := m.scanner.ScanMemory()
					if err == nil {
						_ = m.scanner.ComputeNamespaceStats()
						_ = m.scanner.ComputeBigKeysFromScanLog()
					}
				}()
			}
		case "m", "M":
			if m.scanner != nil {
				go func() {
					err := m.scanner.MonitorOps()
					if err == nil {
						_ = m.scanner.ComputeNamespaceStats()
						_ = m.scanner.ComputeHotKeysFromMonitorLog()
					}
				}()
			}

		case "enter", "right":
			if m.main.ActiveTab == components.TabNamespace && m.scanner != nil {
				row := m.main.Table.Cursor()
				if row >= 0 && row < len(m.main.State.NamespaceStats) {
					ns := m.main.State.NamespaceStats[row].Namespace
					m.scanner.DrillDownNamespace(ns)
				}
			}
		case "backspace", "left":
			if m.main.ActiveTab == components.TabNamespace && m.scanner != nil {
				m.scanner.LevelUpNamespace()
			}

		case "1", "2", "3", "4", "5", "6", "7", "8":
			m.handleSort(key)

		default:
			// Forward to table for up/down navigation
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

func (m Model) View() tea.View {
	var content string
	switch m.screen {
	case screenDisclaimer:
		content = screens.RenderDisclaimer(m.width, m.height)
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
		content = screens.RenderLoading(m.spinner, status, scanProgress, scannedKeys,
			monitorProgress, monitorDur, monitorTotal, m.width, m.height)
	case screenError:
		content = screens.RenderError(m.err, m.width, m.height)
	case screenMain:
		content = m.main.View()
	}
	v := tea.NewView(content)
	v.AltScreen = true
	return v
}
