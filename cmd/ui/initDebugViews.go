package ui

func (u *UI) InitDebugViews() {
	buildMenu(u.Layout.MainMenu.DebugMenu, u.Layout.MainMenu.MenuList)
}
