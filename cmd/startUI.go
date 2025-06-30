package cmd

import (
	"log"
	"os"

	debugUI "github.com/BeardedWonderDev/DIS-Reader/cmd/debug"
	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/BeardedWonderDev/DIS-Reader/ui"
)

func StartUI(cfg *types.Config) {
	modules := []types.ViewModule{
		debugUI.NewDebugService(),
	}
	ui := ui.NewUI(cfg, modules)

	err := ui.Run()
	if err != nil {
		log.Printf("fatal: %v", err)
		os.Exit(1)
	}
}
