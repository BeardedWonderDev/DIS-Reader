package bubbletea

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

var tabs = []string{"Batch Search", "SQLite Viewer"}

// rootModel drives the experimental Bubble Tea UI skeleton. It will be expanded
// with Bubble components as features come online.
type rootModel struct {
	cfg *types.DISUIConfig
	dis types.DISReaderService

	width  int
	height int
	ready  bool

	activeTab int
	logs      logBuffer
}

func NewRootModel(cfg *types.DISUIConfig, dis types.DISReaderService) rootModel {
	return rootModel{
		cfg:       cfg,
		dis:       dis,
		logs:      newLogBuffer(200).append("Bubble UI experimental mode enabled. Press q to quit."),
		activeTab: 0,
	}
}

func (m rootModel) Init() tea.Cmd {
	host := "<unset>"
	if m.cfg != nil && m.cfg.DIS != nil && m.cfg.DIS.Host != "" {
		host = m.cfg.DIS.Host
	}
	outputMode := "sqlite"
	if m.cfg != nil {
		if mode := strings.TrimSpace(m.cfg.DebugSearch.DefaultOutputMode); mode != "" {
			outputMode = mode
		}
	}

	return tea.Batch(
		newLogCmd("DIS host: %s", host),
		newLogCmd("Debug search default output: %s", outputMode),
		newLogCmd("Tab/arrow keys switch panes • ctrl+l clears logs"),
	)
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % len(tabs)
			return m, nil
		case "shift+tab":
			m.activeTab = (m.activeTab - 1 + len(tabs)) % len(tabs)
			return m, nil
		case "left":
			if m.activeTab > 0 {
				m.activeTab--
			}
			return m, nil
		case "right":
			if m.activeTab < len(tabs)-1 {
				m.activeTab++
			}
			return m, nil
		case "ctrl+l":
			m.logs = newLogBuffer(200)
			return m, newLogCmd("Logs cleared")
		case "1", "2":
			idx := int(msg.Runes[0] - '1')
			if idx >= 0 && idx < len(tabs) {
				m.activeTab = idx
			}
			return m, nil
		}
	case logMsg:
		m.logs = m.logs.append(string(msg))
		return m, nil
	}

	return m, nil
}

func (m rootModel) View() string {
	if !m.ready {
		return "Initializing Bubble UI..."
	}

	title := titleStyle.Render(fmt.Sprintf("%s v%s — Bubble UI Preview", m.cfg.AppName, m.cfg.AppVersion))
	nav := m.renderTabs()
	main := m.renderActivePane()
	status := m.renderStatusBar()
	logs := m.renderLogs()

	sections := []string{title, nav, main, status, logs}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m rootModel) renderTabs() string {
	items := make([]string, 0, len(tabs))
	for i, tab := range tabs {
		label := fmt.Sprintf("%d. %s", i+1, tab)
		if i == m.activeTab {
			items = append(items, tabActiveStyle.Render(label))
		} else {
			items = append(items, tabInactiveStyle.Render(label))
		}
	}
	return tabBarStyle.Render(strings.Join(items, "  "))
}

func (m rootModel) renderActivePane() string {
	switch m.activeTab {
	case 0:
		return panelStyle.Render(
			placeholderStyle.Render(
				"Batch Debug Search workspace will live here.\n" +
					"Upcoming tasks:\n" +
					"  • Auth-aware search form built with Bubble text inputs.\n" +
					"  • Progress + event stream wired to RunDebugSearch.\n" +
					"  • Inline log/status lines for long-running scans.",
			),
		)
	case 1:
		return panelStyle.Render(
			placeholderStyle.Render(
				"SQLite data browser coming soon.\n" +
					"Plan:\n" +
					"  • Detect newly generated debug-search SQLite files.\n" +
					"  • Browse schemas/tables with Bubble lists.\n" +
					"  • Inspect rows with paginated table view.",
			),
		)
	default:
		return panelStyle.Render("Unknown pane")
	}
}

func (m rootModel) renderStatusBar() string {
	host := "<unset>"
	if m.cfg != nil && m.cfg.DIS != nil && m.cfg.DIS.Host != "" {
		host = m.cfg.DIS.Host
	}
	output := "sqlite"
	if m.cfg != nil {
		if path := strings.TrimSpace(m.cfg.DebugSearch.DefaultOutputPath); path != "" {
			output = path
		}
	}

	info := fmt.Sprintf("Host: %s | Output: %s | Pane: %s | q to quit", host, output, tabs[m.activeTab])
	return statusBarStyle.Render(info)
}

func (m rootModel) renderLogs() string {
	lines := m.logs.tail(8)
	if len(lines) == 0 {
		lines = []string{"No log entries yet."}
	}
	body := logBodyStyle.Render(strings.Join(lines, "\n"))
	return logPanelTitleStyle.Render("Activity") + "\n" + body
}
