//go:build dashboard

package dashboard

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

// Run launches the Wails native window. Called from the CLI command.
// The assets FS must contain the built frontend (frontend/dist) and is
// provided by the caller to keep the embed directive at the cmd level.
func Run(assets embed.FS, s Store) error {
	// Clean up runs stuck in 'running' for more than 10 minutes.
	if cleaned, err := s.CleanupStaleRuns(10); err != nil {
		log.Printf("dashboard: cleanup stale runs: %v", err)
	} else if cleaned > 0 {
		log.Printf("dashboard: marked %d stale runs as abandoned", cleaned)
	}

	// Backfill project names for existing runs without one.
	if filled, err := s.BackfillProjects(); err != nil {
		log.Printf("dashboard: backfill projects: %v", err)
	} else if filled > 0 {
		log.Printf("dashboard: backfilled project for %d runs", filled)
	}

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
