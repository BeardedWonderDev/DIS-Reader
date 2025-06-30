package ui

import (
	"fmt"
)

func (u *UI) startupSequence() {
	u.App.SetRoot(u.WinMan, true)
	u.GetLogger().Info(fmt.Sprintf("✨ Welcome to %s v%s", u.Config.AppName, u.Config.AppVersion))

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
