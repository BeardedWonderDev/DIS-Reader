package bubbletea

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// logMsg is emitted whenever we want to append a line to the activity panel.
type logMsg string

func newLogCmd(format string, args ...interface{}) func() tea.Msg {
	text := fmt.Sprintf(format, args...)
	return func() tea.Msg {
		return logMsg(text)
	}
}
