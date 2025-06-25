package authUI

import (
	"strings"

	"github.com/BeardedWonderDev/DIS-Reader/cmd/entity"
	typesUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/types"
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

func (a Auth) ShowAuthModal() {
	txtServerURL = tview.NewInputField()
	txtServerURL.SetBackgroundColor(a.GetUI().GetTheme().Colors.WindowColor)
	txtServerURL.SetFieldStyle(a.GetUI().GetTheme().Style.FieldStyle)
	txtServerURL.SetText(a.GetUI().GetDIS().GetConfig().Host)
	txtServerURL.SetLabel("DIS IP Address: ")

	txtUsername = tview.NewInputField()
	txtUsername.SetBackgroundColor(a.GetUI().GetTheme().Colors.WindowColor)
	txtUsername.SetFieldStyle(a.GetUI().GetTheme().Style.FieldStyle)
	txtUsername.SetText(a.GetUI().GetDIS().GetConfig().User)
	txtUsername.SetLabel("DIS Username: ")

	txtPassword = tview.NewInputField()
	txtPassword.SetBackgroundColor(a.GetUI().GetTheme().Colors.WindowColor)
	txtPassword.SetFieldStyle(a.GetUI().GetTheme().Style.FieldStyle)
	txtPassword.SetText(a.GetUI().GetDIS().GetConfig().Password)
	txtPassword.SetLabel("DIS Password: ")
	txtPassword.SetMaskCharacter('*')

	btnConnect = tview.NewButton("Connect")
	btnConnect.SetStyle(a.GetUI().GetTheme().Style.ButtonStyle)

	layout := tview.NewGrid()
	layout.SetBorderPadding(1, 1, 1, 1)
	layout.SetBackgroundColor(a.GetUI().GetTheme().Colors.WindowColor)
	layout.SetColumns(-1, 1, -1)
	layout.AddItem(txtServerURL, 0, 0, 1, 3, 0, 0, true)
	layout.AddItem(txtUsername, 1, 0, 1, 1, 0, 0, false)
	layout.AddItem(txtPassword, 1, 2, 1, 1, 0, 0, false)
	layout.AddItem(btnConnect, 2, 0, 1, 3, 0, 0, false)

	wnd := a.GetUI().CreateModalDialog(typesUI.CreateModalDialogParam{
		Title:         " DIS Auth Config ",
		RootView:      layout,
		Draggable:     true,
		Size:          typesUI.WinSize{X: 0, Y: 0, Width: 70, Height: 10},
		FallbackFocus: a.GetUI().GetLayout().MainMenu.MenuList,
	})

	a.showAuthModal_SetInputCapture(wnd)
}

func (a Auth) showAuthModal_SetInputCapture(wnd *winman.WindowBase) {

	txtServerURL.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			a.GetUI().GetWinMan().RemoveWindow(wnd)
			a.GetUI().SetFocus(a.GetUI().GetLayout().MainMenu.MenuList)
			return nil

		case tcell.KeyTAB:
			a.GetUI().SetFocus(txtUsername)
		}

		return event
	})

	txtUsername.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			a.GetUI().GetWinMan().RemoveWindow(wnd)
			a.GetUI().SetFocus(a.GetUI().GetLayout().MainMenu.MenuList)
			return nil

		case tcell.KeyTAB:
			txtUsername.SetText(strings.ToUpper(txtUsername.GetText()))
			a.GetUI().SetFocus(txtPassword)
		}

		return event
	})

	txtPassword.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			a.GetUI().GetWinMan().RemoveWindow(wnd)
			a.GetUI().SetFocus(a.GetUI().GetLayout().MainMenu.MenuList)
			return nil

		case tcell.KeyTAB:
			txtPassword.SetText(strings.ToUpper(txtPassword.GetText()))
			a.GetUI().SetFocus(btnConnect)
		}

		return event
	})

	conBtnSelected := func() {
		a.GetUI().GetDIS().GetConfig().Host = txtServerURL.GetText()
		a.GetUI().GetDIS().GetConfig().User = txtUsername.GetText()
		a.GetUI().GetDIS().GetConfig().Password = txtPassword.GetText()

		if a.GetUI().GetDIS().GetConfig().User != "" && a.GetUI().GetDIS().GetConfig().Password != "" {
			btnConnect.SetLabel("Connecting...")
			btnConnect.SetDisabled(true)

			if err := a.Authenticate(); err != nil {
				a.GetUI().PrintLog(entity.Log{
					Content: "Error Connecting to DIS, check host and credentials: " + err.Error(),
					Type:    entity.LOG_ERROR,
				})
				btnConnect.SetLabel("Connect")
				btnConnect.SetDisabled(false)
				return
			}

			// Remove the window and restore focus to menu list
			a.GetUI().CloseModalDialog(wnd, a.GetUI().GetLayout().MainMenu.MenuList)
		} else {
			a.GetUI().PrintLog(entity.Log{
				Content: "A valid Username and Password must be entered",
				Type:    entity.LOG_ERROR,
			})
		}
	}

	btnConnect.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			a.GetUI().SetFocus(txtServerURL)
			return nil
		case tcell.KeyEnter:
			conBtnSelected()
			return nil
		}

		return event
	})

	btnConnect.SetSelectedFunc(conBtnSelected)
}
