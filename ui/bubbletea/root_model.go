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

	activeTab     int
	logs          logBuffer
	search        searchModel
	auth          authModel
	showAuthModal bool
}

func NewRootModel(cfg *types.DISUIConfig, dis types.DISReaderService) rootModel {
	search := newSearchModel(cfg, dis)
	auth := newAuthModel(cfg, dis)
	return rootModel{
		cfg:           cfg,
		dis:           dis,
		logs:          newLogBuffer(200).append("Bubble UI experimental mode enabled. Press q to quit."),
		activeTab:     0,
		search:        search,
		auth:          auth,
		showAuthModal: false,
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

	var searchCmd tea.Cmd
	m.search, searchCmd = m.search.Init()
	var authCmd tea.Cmd
	m.auth, authCmd = m.auth.Init()

	cmds := []tea.Cmd{
		newLogCmd("DIS host: %s", host),
		newLogCmd("Debug search default output: %s", outputMode),
		newLogCmd("Tab/arrow keys switch panes • ctrl+l clears logs"),
	}
	if searchCmd != nil {
		cmds = append(cmds, searchCmd)
	}
	if authCmd != nil {
		cmds = append(cmds, authCmd)
	}

	return tea.Batch(cmds...)
}

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	var searchCmd tea.Cmd
	var searchHandled bool
	searchFocused := (m.activeTab == 0) && !m.showAuthModal
	m.search, searchCmd, searchHandled = m.search.Update(msg, searchFocused)
	if searchCmd != nil {
		cmds = append(cmds, searchCmd)
	}

	var authCmd tea.Cmd
	var authHandled bool
	m.auth, authCmd, authHandled = m.auth.Update(msg, m.showAuthModal)
	if authCmd != nil {
		cmds = append(cmds, authCmd)
	}
	if searchHandled || authHandled {
		return m, tea.Batch(cmds...)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.search = m.search.SetWidth(msg.Width)
	case tea.KeyMsg:
		if m.showAuthModal {
			if msg.String() == "ctrl+c" || msg.String() == "q" {
				cmds = append(cmds, tea.Quit)
			}
			return m, tea.Batch(cmds...)
		}
		switch msg.String() {
		case "ctrl+c", "q":
			cmds = append(cmds, tea.Quit)
		case "c":
			m.showAuthModal = true
			m.auth.focused = focusAuthHost
			var focusCmd tea.Cmd
			m.auth, focusCmd = m.auth.applyFocus()
			if focusCmd != nil {
				cmds = append(cmds, focusCmd)
			}
		case "tab":
			m.activeTab = (m.activeTab + 1) % len(tabs)
		case "shift+tab":
			m.activeTab = (m.activeTab - 1 + len(tabs)) % len(tabs)
		case "left":
			if m.activeTab > 0 {
				m.activeTab--
			}
		case "right":
			if m.activeTab < len(tabs)-1 {
				m.activeTab++
			}
		case "ctrl+l":
			m.logs = newLogBuffer(200)
			cmds = append(cmds, newLogCmd("Logs cleared"))
		case "1", "2":
			idx := int(msg.Runes[0] - '1')
			if idx >= 0 && idx < len(tabs) {
				m.activeTab = idx
			}
		}
	case logMsg:
		m.logs = m.logs.append(string(msg))
	case authStateChangedMsg:
		m.search = m.search.WithAuthState(msg.authenticated)
		if msg.authenticated {
			m.showAuthModal = false
			cmds = append(cmds, newLogCmd("DIS authentication successful"))
		} else {
			cmds = append(cmds, newLogCmd("DIS authentication failed"))
		}
	case authDismissedMsg:
		m.showAuthModal = false
	case requireAuthMsg:
		m.showAuthModal = true
		m.auth.focused = focusAuthHost
		var focusCmd tea.Cmd
		m.auth, focusCmd = m.auth.applyFocus()
		if focusCmd != nil {
			cmds = append(cmds, focusCmd)
		}
		cmds = append(cmds, newLogCmd("Authentication required before running search"))
	default:
		// ignore
	}

	if len(cmds) == 0 {
		return m, nil
	}
	return m, tea.Batch(cmds...)
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
	base := lipgloss.JoinVertical(lipgloss.Left, sections...)

	if m.showAuthModal {
		overlay := modalWindowStyle.Render(m.auth.View())
		width := m.width
		if width < lipgloss.Width(overlay)+4 {
			width = lipgloss.Width(overlay) + 4
		}
		height := m.height
		if height < lipgloss.Height(overlay)+4 {
			height = lipgloss.Height(overlay) + 4
		}
		screen := lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, overlay,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceForeground(lipgloss.Color("#111111")))
		return base + "\n" + screen
	}

	return base
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
		return panelStyle.Render(m.search.View())
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

	info := fmt.Sprintf("Host: %s | Output: %s | Pane: %s | %s | %s | q to quit", host, output, tabs[m.activeTab], m.search.StatusLine(), m.auth.StatusLine())
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
