package ui

import (
	"github.com/BeardedWonderDev/DIS-Reader/cmd/entity"
	typesUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/types"
	"github.com/epiclabs-io/winman"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	txtDebugSearch *tview.InputField
	btnBatchSearch *tview.Button
)

func (u *UI) showDebugModal() {
	txtDebugSearch = tview.NewInputField()
	txtDebugSearch.SetBackgroundColor(u.Theme.Colors.WindowColor)
	txtDebugSearch.SetFieldStyle(u.Theme.Style.FieldStyle)
	txtDebugSearch.SetLabel("Search Term: ")

	txtDebugSearchDesc := tview.NewTextView()
	txtDebugSearchDesc.SetBackgroundColor(u.Theme.Colors.WindowColor)
	txtDebugSearchDesc.SetTextStyle(u.Theme.Style.TextAreaStyle)
	txtDebugSearchDesc.SetDisabled(true)
	txtDebugSearchDesc.SetText("Enter a term to search all DIS tables for\n(ie.. Part #, Invoice #, Unit#, etc).\nResults will be stored locally and visable in DIS Reader.")
	txtDebugSearchDesc.SetTextAlign(tview.AlignCenter)

	btnBatchSearch = tview.NewButton("Search")
	btnBatchSearch.SetStyle(u.Theme.Style.ButtonStyle)

	layout := tview.NewGrid()
	layout.SetBorderPadding(1, 1, 1, 1)
	layout.SetBackgroundColor(u.Theme.Colors.WindowColor)
	layout.AddItem(txtDebugSearch, 0, 0, 1, 1, 0, 0, true)
	layout.AddItem(txtDebugSearchDesc, 1, 0, 2, 1, 0, 0, false)
	layout.AddItem(btnBatchSearch, 3, 0, 1, 1, 0, 0, false)

	wnd := u.CreateModalDialog(typesUI.CreateModalDialogParam{
		Title:         " DIS Batch Debug Search ",
		RootView:      layout,
		Draggable:     true,
		Size:          typesUI.WinSize{X: 0, Y: 0, Width: 70, Height: 10},
		FallbackFocus: u.Layout.MainMenu.MenuList,
	})

	u.showDebugModal_SetInputCapture(wnd)
}

func (u *UI) showDebugModal_SetInputCapture(wnd *winman.WindowBase) {

	txtDebugSearch.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			u.WinMan.RemoveWindow(wnd)
			u.SetFocus(u.Layout.MainMenu.MenuList)
			return nil

		case tcell.KeyTAB:
			u.SetFocus(btnBatchSearch)

		case tcell.KeyEnter:
			u.SetFocus(btnBatchSearch)
			return nil
		}

		return event
	})

	btnBatchSearch.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			u.SetFocus(txtDebugSearch)
			return nil
		case tcell.KeyEnter:
			u.runBatchSearch(wnd)
			return nil
		}

		return event
	})

	btnBatchSearch.SetSelectedFunc(func() {
		u.runBatchSearch(wnd)
	})
}

func (u *UI) runBatchSearch(wnd *winman.WindowBase) {
	go func() {
		u.PrintLog(entity.Log{
			Content: "Starting Search for [blue]" + txtDebugSearch.GetText() + "...",
			Type:    entity.LOG_INFO,
		})

		// TODO - Start Batch Search

		// Load Batch Search Views
		u.InitDebugViews()

		// Remove the window and restore focus to menu list
		u.CloseModalDialog(wnd, u.Layout.MainMenu.MenuList)
	}()
}
