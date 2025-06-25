package debugUI

import utilsUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/utils"

func (d *Debug) InitDebugViews() {
	utilsUI.BuildMenu(d.GetMainMenu(), d.UI.GetLayout().MainMenu.MenuList)
}
