package authUI

import (
	"strings"

	"github.com/BeardedWonderDev/DIS-Reader/types"
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

func (a *Auth) ShowAuthModal() {
	txtServerURL = tview.NewInputField()
	txtServerURL.SetBackgroundColor(a.UI.GetTheme().Colors.WindowColor)
	txtServerURL.SetFieldStyle(a.UI.GetTheme().Style.FieldStyle)
	txtServerURL.SetText(a.UI.GetDIS().GetConfig().Host)
	txtServerURL.SetLabel("DIS IP Address: ")

	txtUsername = tview.NewInputField()
	txtUsername.SetBackgroundColor(a.UI.GetTheme().Colors.WindowColor)
	txtUsername.SetFieldStyle(a.UI.GetTheme().Style.FieldStyle)
	txtUsername.SetText(a.UI.GetDIS().GetConfig().User)
	txtUsername.SetLabel("DIS Username: ")

	txtPassword = tview.NewInputField()
	txtPassword.SetBackgroundColor(a.UI.GetTheme().Colors.WindowColor)
	txtPassword.SetFieldStyle(a.UI.GetTheme().Style.FieldStyle)
	txtPassword.SetText(a.UI.GetDIS().GetConfig().Password)
	txtPassword.SetLabel("DIS Password: ")
	txtPassword.SetMaskCharacter('*')

	btnConnect = tview.NewButton("Connect")
	btnConnect.SetStyle(a.UI.GetTheme().Style.ButtonStyle)

	layout := tview.NewGrid()
	layout.SetBorderPadding(1, 1, 1, 1)
	layout.SetBackgroundColor(a.UI.GetTheme().Colors.WindowColor)
	layout.SetColumns(-1, 1, -1)
	layout.AddItem(txtServerURL, 0, 0, 1, 3, 0, 0, true)
	layout.AddItem(txtUsername, 1, 0, 1, 1, 0, 0, false)
	layout.AddItem(txtPassword, 1, 2, 1, 1, 0, 0, false)
	layout.AddItem(btnConnect, 2, 0, 1, 3, 0, 0, false)

	wnd := a.UI.CreateModalDialog(types.CreateModalDialogParam{
		Title:         " DIS Auth Config ",
		RootView:      layout,
		Draggable:     true,
		Size:          types.WinSize{X: 0, Y: 0, Width: 70, Height: 10},
		FallbackFocus: a.UI.GetLayout().MainMenu,
	})

	a.showAuthModal_SetInputCapture(wnd)
}

func (a *Auth) showAuthModal_SetInputCapture(wnd *winman.WindowBase) {

	txtServerURL.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			a.UI.GetWinMan().RemoveWindow(wnd)
			a.UI.SetFocus(a.UI.GetLayout().MainMenu)
			return nil

		case tcell.KeyTAB:
			a.UI.SetFocus(txtUsername)
		}

		return event
	})

	txtUsername.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			a.UI.GetWinMan().RemoveWindow(wnd)
			a.UI.SetFocus(a.UI.GetLayout().MainMenu)
			return nil

		case tcell.KeyTAB:
			txtUsername.SetText(strings.ToUpper(txtUsername.GetText()))
			a.UI.SetFocus(txtPassword)
		}

		return event
	})

	txtPassword.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			a.UI.GetWinMan().RemoveWindow(wnd)
			a.UI.SetFocus(a.UI.GetLayout().MainMenu)
			return nil

		case tcell.KeyTAB:
			txtPassword.SetText(strings.ToUpper(txtPassword.GetText()))
			a.UI.SetFocus(btnConnect)
		}

		return event
	})

	conBtnSelected := func() {
		a.UI.GetDIS().GetConfig().Host = txtServerURL.GetText()
		a.UI.GetDIS().GetConfig().User = txtUsername.GetText()
		a.UI.GetDIS().GetConfig().Password = txtPassword.GetText()

		if a.UI.GetDIS().GetConfig().User != "" && a.UI.GetDIS().GetConfig().Password != "" {
			btnConnect.SetLabel("Connecting...")
			btnConnect.SetDisabled(true)

			a.Authenticate(func(success bool, err error) {
				if success {
					a.UI.CloseModalDialog(wnd, a.UI.GetLayout().MainMenu)
				} else {
					btnConnect.SetLabel("Connect")
					btnConnect.SetDisabled(false)
				}
			})
		} else {
			a.UI.GetLogger().Error("A valid Username and Password must be entered")
		}
	}

	btnConnect.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			a.UI.SetFocus(txtServerURL)
			return nil
		case tcell.KeyEnter:
			conBtnSelected()
			return nil
		}

		return event
	})

	btnConnect.SetSelectedFunc(conBtnSelected)
}
