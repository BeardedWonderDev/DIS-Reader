package logUI

import (
	"github.com/rivo/tview"
)

// InitLogList initializes the log panel on the main screen
func (l *LogPanelWriter) initLogPanelView() *tview.TextView {
	logPanel := tview.NewTextView()
	logPanel.SetDynamicColors(true)
	logPanel.SetTitle(" 📃 Logs ")
	logPanel.SetBorder(true)
	logPanel.SetWordWrap(true)
	logPanel.SetBorderPadding(1, 1, 1, 1)

	logPanel.SetScrollable(true).SetChangedFunc(func() {
		l.UI.GetApp().Draw()
	})

	return logPanel
}
