package utilsUI

import (
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/rivo/tview"
)

func BuildMenu(menuType types.MenuDefinition, menuList *tview.List) {
	menuList.Clear()
	menuList.SetTitle(menuType.Title)
	for _, item := range menuType.Items {
		menuList.AddItem(item.MainText, item.SecondaryText, item.Shortcut, item.Selected)
	}
}
