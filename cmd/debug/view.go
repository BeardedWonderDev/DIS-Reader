package debugUI

import utilsUI "github.com/BeardedWonderDev/DIS-Reader/ui/utils"

func (d *Debug) InitDebugViews() {
	utilsUI.BuildMenu(d.GetMainMenu(), d.UI.GetLayout().MainMenu.MenuList)
}
