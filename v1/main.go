package main

import (
	"embed"
	"net/http"
	"strings"

	"github.com/tylertravisty/rum-goggles/v1/internal/config"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create application with options
	err := wails.Run(&options.App{
		Title:  "Rum Goggles",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets:  assets,
			Handler: http.HandlerFunc(GetImage),
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 255},
		OnShutdown:       app.shutdown,
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

func GetImage(w http.ResponseWriter, r *http.Request) {
	path := strings.Replace(r.RequestURI, "wails://wails", "", 1)
	prefix, err := config.ImageDir()
	if err != nil {
		return
	}
	http.ServeFile(w, r, prefix+path)
}
