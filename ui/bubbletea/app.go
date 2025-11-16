package bubbletea

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

// Run starts the experimental Bubble Tea UI program.
func Run(cfg *types.DISUIConfig, dis types.DISReaderService) error {
	model := NewRootModel(cfg, dis)
	program := tea.NewProgram(model, tea.WithAltScreen())
	return program.Start()
}
