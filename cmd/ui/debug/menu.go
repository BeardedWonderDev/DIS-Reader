package debugUI

import typesUI "github.com/BeardedWonderDev/DIS-Reader/cmd/ui/types"

func (d *Debug) GetMainMenu() typesUI.MenuDefinition {
	return typesUI.MenuDefinition{
		Title: " Debug Menu ",
		Items: []typesUI.MenuItem{
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
