package main

import (
	"embed"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:            "TON Torrent",
		Width:            800,
		Height:           487,
		MinHeight:        487,
		MinWidth:         800,
		DisableResize:    false,
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 1},
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Mac: &mac.Options{
			Appearance: mac.NSAppearanceNameVibrantLight,
		},
		Windows:    &windows.Options{},
		OnStartup:  app.startup,
		OnDomReady: app.ready,
		OnShutdown: app.exit,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
