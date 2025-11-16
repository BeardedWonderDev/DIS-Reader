package bubbletea

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#B3E5FC")).Padding(0, 1)

	tabBarStyle      = lipgloss.NewStyle().Padding(0, 1)
	tabActiveStyle   = lipgloss.NewStyle().Bold(true).Padding(0, 1).Foreground(lipgloss.Color("#0A192F")).Background(lipgloss.Color("#80CBC4"))
	tabInactiveStyle = lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("#8EACE3"))

	panelStyle         = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#4C566A")).Padding(1).MarginTop(1)
	placeholderStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#D8DEE9")).Width(60)
	statusBarStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#C3E88D")).Padding(0, 1).MarginTop(1)
	logPanelTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFB74D")).Padding(0, 1)
	logBodyStyle       = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#616161")).Padding(0, 1)
)
