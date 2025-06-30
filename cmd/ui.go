package cmd

import (
	"log"
	"os"

	"github.com/BeardedWonderDev/DIS-Reader/types"
	"github.com/BeardedWonderDev/DIS-Reader/ui"
)

func StartUI(cfg *types.Config) {
	ui := ui.NewUI(cfg)

	err := ui.Run()
	if err != nil {
		log.Printf("fatal: %v", err)
		os.Exit(1)
	}
}
