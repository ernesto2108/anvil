//go:build dashboard

package dashboard

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// Run launches the Wails native window. Called from the CLI command.
// The assets FS must contain the built frontend (frontend/dist) and is
// provided by the caller to keep the embed directive at the cmd level.
func Run(assets embed.FS, s Store) error {
	app := NewApp(s)

	return wails.Run(&options.App{
		Title:     "Anvil Dashboard",
		Width:     1280,
		Height:    800,
		MinWidth:  900,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 255, G: 255, B: 255, A: 1},
		OnStartup:        app.Startup,
		Bind: []interface{}{
			app,
		},
	})
}
