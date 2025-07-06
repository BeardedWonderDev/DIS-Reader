package ui

import (
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/epiclabs-io/winman"
	"github.com/rivo/tview"
)

// CreateModalDialogParam is a helper to create a modal dialog window
func (u UI) CreateModalDialog(param types.CreateModalDialogParam) *winman.WindowBase {
	wnd := winman.NewWindow().Show()

	wnd.SetTitle(param.Title)
	wnd.SetRoot(param.RootView)
	wnd.SetDraggable(param.Draggable)
	wnd.SetResizable(param.Resizeable)
	wnd.SetModal(true)
	wnd.SetBackgroundColor(u.Theme.Colors.WindowColor)

	wnd.SetRect(param.Size.X, param.Size.Y, param.Size.Width, param.Size.Height)
	wnd.AddButton(&winman.Button{
		Symbol: 'X',
		OnClick: func() {
			// Close current window and get back focus to the fallback primitive
			u.CloseModalDialog(wnd, param.FallbackFocus)
		},
	})

	u.WinMan.AddWindow(wnd)
	u.WinMan.Center(wnd)
	u.SetFocus(wnd)

	return wnd
}

// CloseModalDialog is a helper to close the modal dialog window
func (u UI) CloseModalDialog(wnd *winman.WindowBase, focus tview.Primitive) {
	u.WinMan.RemoveWindow(wnd)
	u.SetFocus(focus)
}
