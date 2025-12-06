//go:build gui

package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed frontend/*
var assets embed.FS

func main() {
	err := wails.Run(&options.App{
		Title:         "DIS Agent Control",
		Width:         960,
		Height:        720,
		DisableResize: false,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 12, G: 18, B: 32, A: 255},
	})
	if err != nil {
		log.Fatal(err)
	}
}
