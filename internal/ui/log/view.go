package logUI

import (
	"github.com/gdamore/tcell/v2"
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

	logPanel.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			l.UI.GetApp().SetFocus(l.UI.GetLayout().SplitSidebar)
		}
		return event
	})

	return logPanel
}
