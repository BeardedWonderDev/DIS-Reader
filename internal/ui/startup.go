package ui

import (
	"fmt"

	"github.com/BeardedWonderDev/DIS-Reader/types"
)

func (u *UI) StartupSequence() {
	u.App.SetRoot(u.WinMan, true)
	u.GetLogger().Info(fmt.Sprintf("✨ Welcome to %s v%s", u.Config.AppName, u.Config.AppVersion))

	if u.DIS.GetConfig().User != "" && u.DIS.GetConfig().Password != "" {
		u.Auth.Authenticate(func(success bool, err error) {
			if err != nil {
				types.LogError(u.GetLogger(), "Error authenticating to DIS", err)
			}
		})
	} else {
		u.Auth.ShowAuthModal()
	}
}
