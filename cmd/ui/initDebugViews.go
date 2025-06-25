package ui

func (u *UI) InitDebugViews() {
	buildMenu(u.Layout.MainMenu.debugMenu, u.Layout.MainMenu.MenuList)
}
