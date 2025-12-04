package main

import (
	"github.com/BeardedWonderDev/DIS-Reader/disreader"
	"github.com/BeardedWonderDev/DIS-Reader/ui"
)

func main() {
	cfg := NewConfig()

	disReader, err := disreader.NewDISReaderService(cfg.DIS, nil)
	if err != nil {
		panic(err)
	}
	defer disReader.Shutdown()

	ui.Run(cfg, disReader)
}
