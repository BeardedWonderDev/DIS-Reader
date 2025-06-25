package logUI

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// InitLogList initializes the log panel on the main screen
func (l *Log) InitLogList() *tview.TextView {
	logPanel := tview.NewTextView()
	logPanel.SetDynamicColors(true)
	logPanel.SetTitle(" 📃 Logs ")
	logPanel.SetBorder(true)
	logPanel.SetWordWrap(true)
	logPanel.SetBorderPadding(1, 1, 1, 1)

	logPanel.SetScrollable(true).SetChangedFunc(func() {
		l.UI.GetApp().Draw()
	})

	l.initLogList_SetInputCapture(logPanel)

	return logPanel
}

func (l *Log) initLogList_SetInputCapture(logPanel *tview.TextView) {
	logPanel.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTAB:
			l.UI.GetApp().SetFocus(l.UI.GetLayout().MainMenu.MenuList)
		}
		return event
	})
}
