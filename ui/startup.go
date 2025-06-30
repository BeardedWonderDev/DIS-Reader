package ui

import (
	"fmt"

	utilsUI "github.com/BeardedWonderDev/DIS-Reader/ui/utils"
)

func (u *UI) startupSequence() {
	u.loadStartupUI()

	if u.DIS.GetConfig().User != "" && u.DIS.GetConfig().Password != "" {
		u.Auth.Authenticate(func(success bool, err error) {
			if err != nil {
				u.GetLogger().Error("Error Authenticating To DIS", "error", err)
			}
		})
	} else {
		u.Auth.ShowAuthModal()
	}
}

func (u *UI) loadStartupUI() {
	utilsUI.BuildMenu(u.Layout.MainMenu.RootMenu, u.Layout.MainMenu.MenuList)
	u.App.SetRoot(u.WinMan, true)

	u.GetLogger().Info(fmt.Sprintf("✨ Welcome to DIS Reader v%s", u.Config.AppVersion))
}
