package ui

import (
	"encoding/base64"
	"fmt"

	"github.com/BeardedWonderDev/DIS-Reader/cmd/entity"
)

func (u *UI) startupSequence() {
	u.loadStartupUI()
	u.showAuthModal()
}

func (u *UI) loadStartupUI() {
	buildMenu(u.Layout.MainMenu.rootMenu, u.Layout.MainMenu.MenuList)
	u.App.SetRoot(u.WinMan, true)

	u.PrintLog(entity.Log{
		Content: fmt.Sprintf("✨ Welcome to DIS Reader v%s", entity.APP_VERSION),
		Type:    entity.LOG_INFO,
	})

	banner, _ := base64.StdEncoding.DecodeString(entity.BANNER)
	u.PrintOutput(entity.Output{
		Content:     string(banner),
		WithHeader:  false,
		CursorAtEnd: false,
	})
}
