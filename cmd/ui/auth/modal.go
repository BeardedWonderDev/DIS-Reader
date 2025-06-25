package authUI

import (
	"fmt"
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

func (a *Auth) ShowAuthModal() {
	fmt.Printf("%v", a.UI)
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

	wnd := a.UI.CreateModalDialog(typesUI.CreateModalDialogParam{
		Title:         " DIS Auth Config ",
		RootView:      layout,
		Draggable:     true,
		Size:          typesUI.WinSize{X: 0, Y: 0, Width: 70, Height: 10},
		FallbackFocus: a.UI.GetLayout().MainMenu.MenuList,
	})

	a.ShowAuthModal_SetInputCapture(wnd)
}

func (a *Auth) ShowAuthModal_SetInputCapture(wnd *winman.WindowBase) {

	txtServerURL.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			if a.UI.GetDIS().GetConfig().User != "" && a.UI.GetDIS().GetConfig().Password != "" {
				a.UI.GetWinMan().RemoveWindow(wnd)
				a.UI.SetFocus(a.UI.GetLayout().MainMenu.MenuList)
			} else {
				noUserPassError(a.UI.PrintLog)
			}
			return nil

		case tcell.KeyTAB:
			a.UI.SetFocus(txtUsername)
		}

		return event
	})

	txtUsername.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEscape:
			if a.UI.GetDIS().GetConfig().User != "" && a.UI.GetDIS().GetConfig().Password != "" {
				a.UI.GetWinMan().RemoveWindow(wnd)
				a.UI.SetFocus(a.UI.GetLayout().MainMenu.MenuList)
			} else {
				noUserPassError(a.UI.PrintLog)
			}
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
			if a.UI.GetDIS().GetConfig().User != "" && a.UI.GetDIS().GetConfig().Password != "" {
				a.UI.GetWinMan().RemoveWindow(wnd)
				a.UI.SetFocus(a.UI.GetLayout().MainMenu.MenuList)
			} else {
				noUserPassError(a.UI.PrintLog)
			}
			return nil

		case tcell.KeyTAB:
			txtPassword.SetText(strings.ToUpper(txtPassword.GetText()))
			a.UI.SetFocus(btnConnect)
		}

		return event
	})

	btnConnect.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			a.UI.SetFocus(txtServerURL)
			return nil
		case tcell.KeyEnter:
			a.saveAndTestConnection(wnd)
			return nil
		}

		return event
	})

	btnConnect.SetSelectedFunc(func() {
		a.saveAndTestConnection(wnd)
	})
}

func (a *Auth) saveAndTestConnection(wnd *winman.WindowBase) {
	btnConnect.SetLabel("Connecting...")

	go func() {
		a.UI.GetDIS().GetConfig().Host = txtServerURL.GetText()
		a.UI.GetDIS().GetConfig().User = txtUsername.GetText()
		a.UI.GetDIS().GetConfig().Password = txtPassword.GetText()

		if a.UI.GetDIS().GetConfig().User != "" && a.UI.GetDIS().GetConfig().Password != "" {
			a.UI.PrintLog(entity.Log{
				Content: "🌏 Verifying DIS Connection to [blue]" + txtServerURL.GetText() + ", connecting...",
				Type:    entity.LOG_INFO,
			})

			if err := a.UI.GetDIS().TestDISConnection(); err != nil {
				a.UI.PrintLog(entity.Log{
					Content: "Error Connecting to DIS, check host and credentials: " + err.Error(),
					Type:    entity.LOG_ERROR,
				})
			} else {
				a.UI.PrintLog(entity.Log{
					Content: "DIS Connection [green] Succesful",
					Type:    entity.LOG_INFO,
				})
			}

			// Remove the window and restore focus to menu list
			a.UI.CloseModalDialog(wnd, a.UI.GetLayout().MainMenu.MenuList)
		} else {
			noUserPassError(a.UI.PrintLog)
		}
	}()
}

func noUserPassError(logFunc func(param entity.Log)) {
	logFunc(entity.Log{
		Content: "A valid Username and Password must be entered",
		Type:    entity.LOG_ERROR,
	})
}
