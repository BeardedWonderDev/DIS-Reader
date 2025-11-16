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
	viewerColumnStyle  = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#4C566A")).Padding(0, 1).MarginRight(1)
	viewerCardStyle    = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#4C566A")).Padding(0, 1)

	sectionTitleStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A6E3A1")).Underline(true)
	fieldLabelStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("#82AAFF"))
	selectedFormatStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#0A192F")).Background(lipgloss.Color("#FFD166")).Padding(0, 1).Bold(true)
	buttonStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#0A192F")).Background(lipgloss.Color("#BB86FC")).Padding(0, 4).Bold(true)
	buttonFocusedStyle  = buttonStyle.Copy().Background(lipgloss.Color("#C3E88D"))
	helpStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("#9E9E9E"))
	noticeStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#C3E88D"))
	errorStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5370")).Bold(true)
	infoStyle           = lipgloss.NewStyle().Foreground(lipgloss.Color("#80CBC4"))
	sampleStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#F2F2F2"))
	modalWindowStyle    = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("#7FDBFF")).Background(lipgloss.Color("#1B1D2A")).Padding(1, 3).Width(60)
)
