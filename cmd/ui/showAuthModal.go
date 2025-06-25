package ui

import (
	"github.com/BeardedWonderDev/DIS-Reader/cmd/entity"
	"github.com/epiclabs-io/winman"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	txtServerURL *tview.InputField
	txtUsername  *tview.InputField
	txtPassword  *tview.InputField
	btnConnect   *tview.Button
)

func (u *UI) ShowAuthModal() {
	txtServerURL = tview.NewInputField()
	txtServerURL.SetBackgroundColor(u.Theme.Colors.WindowColor)
	txtServerURL.SetFieldStyle(u.Theme.Style.FieldStyle)
	txtServerURL.SetText(u.Config.Host)
	txtServerURL.SetLabel("DIS IP Address: ")

	txtUsername = tview.NewInputField()
	txtUsername.SetBackgroundColor(u.Theme.Colors.WindowColor)
	txtUsername.SetFieldStyle(u.Theme.Style.FieldStyle)
	txtUsername.SetText(u.Config.User)
	txtUsername.SetLabel("DIS Username: ")

	txtPassword = tview.NewInputField()
	txtPassword.SetBackgroundColor(u.Theme.Colors.WindowColor)
	txtPassword.SetFieldStyle(u.Theme.Style.FieldStyle)
	txtPassword.SetText(u.Config.Password)
	txtPassword.SetLabel("DIS Password: ")

	btnConnect = tview.NewButton("Connect")
	btnConnect.SetStyle(u.Theme.Style.ButtonStyle)

	layout := tview.NewGrid()
	layout.SetBorderPadding(1, 1, 1, 1)
	layout.SetBackgroundColor(u.Theme.Colors.WindowColor)
	layout.AddItem(txtServerURL, 0, 0, 1, 2, 0, 0, true)
	layout.AddItem(txtUsername, 1, 0, 1, 1, 0, 0, false)
	layout.AddItem(txtPassword, 1, 1, 1, 1, 0, 0, false)
	layout.AddItem(btnConnect, 2, 0, 1, 2, 0, 0, false)

	wnd := u.CreateModalDialog(CreateModalDialogParam{
		title:         " Connect to DIS ",
		rootView:      layout,
		draggable:     true,
		size:          winSize{0, 0, 70, 10},
		fallbackFocus: u.Layout.MenuList,
	})

	u.ShowAuthModal_SetInputCapture(wnd)
}

func (u *UI) ShowAuthModal_SetInputCapture(wnd *winman.WindowBase) {

	txtServerURL.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			u.WinMan.RemoveWindow(wnd)
			u.SetFocus(u.Layout.MenuList)
			return nil

		case tcell.KeyTAB:
			u.SetFocus(txtUsername)

		case tcell.KeyEnter:
			u.SetFocus(txtUsername)
			return nil
		}

		return event
	})

	txtUsername.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			u.WinMan.RemoveWindow(wnd)
			u.SetFocus(u.Layout.MenuList)
			return nil

		case tcell.KeyTAB:
			u.SetFocus(txtPassword)

		case tcell.KeyEnter:
			u.SetFocus(txtPassword)
			return nil
		}

		return event
	})

	txtPassword.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			u.WinMan.RemoveWindow(wnd)
			u.SetFocus(u.Layout.MenuList)
			return nil

		case tcell.KeyTAB:
			u.SetFocus(btnConnect)

		case tcell.KeyEnter:
			u.SetFocus(btnConnect)
			return nil
		}

		return event
	})

	btnConnect.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			u.SetFocus(txtServerURL)
			return nil
		case tcell.KeyEnter:
			u.saveAndTestConnection(wnd)
			return nil
		}

		return event
	})

	btnConnect.SetSelectedFunc(func() {
		u.saveAndTestConnection(wnd)
	})
}

func (u *UI) saveAndTestConnection(wnd *winman.WindowBase) {
	btnConnect.SetLabel("Connecting...")

	go func() {
		u.PrintLog(entity.Log{
			Content: "🌏 Verifying DIS Connection to [blue]" + txtServerURL.GetText() + ", connecting...",
			Type:    entity.LOG_INFO,
		})

		u.Config.Host = txtServerURL.GetText()
		u.Config.User = txtUsername.GetText()
		u.Config.Password = txtPassword.GetText()

		// TODO - TEST CONNECTION

		// Remove the window and restore focus to menu list
		u.CloseModalDialog(wnd, u.Layout.MenuList)
	}()
}
