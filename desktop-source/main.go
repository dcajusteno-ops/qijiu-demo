package main

import (
	"embed"
	"net/http"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:            "ComfyUI Manager",
		Width:            1440,
		Height:           900,
		MinWidth:         900,
		MinHeight:        600,
		BackgroundColour: &options.RGBA{R: 18, G: 18, B: 18, A: 255},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: http.HandlerFunc(app.serveImage),
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
