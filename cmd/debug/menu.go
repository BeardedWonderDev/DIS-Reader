package debugUI

import "github.com/BeardedWonderDev/DIS-Reader/types"

func (d *Debug) GetMainMenu() types.MenuDefinition {
	return types.MenuDefinition{
		Title: " Debug Menu ",
		Items: []types.MenuItem{
			{
				MainText:      "Start Batch Search",
				SecondaryText: "",
				Shortcut:      's',
				Selected: func() {
					d.showDebugModal()
				},
			},
			{
				MainText:      "Main Menu",
				SecondaryText: "",
				Shortcut:      'm',
				Selected:      d.UI.RevertToMainMenu,
			},
		},
	}
}
