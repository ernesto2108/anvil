//go:build dashboard

package cli

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/internal/dashboard"
	"github.com/ernesto2108/anvil/internal/dashboard/query"
	"github.com/ernesto2108/anvil/internal/dashboard/writer"
	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/output"
)

// DashboardAssets is set by cmd/anvil/embed_dashboard.go at init time.
// It holds the embedded React frontend built with `wails build` or
// `go build -tags dashboard`.
var DashboardAssets embed.FS

func cmdDashboard(_ *config.App, args []string) {
	var forceMigrate bool
	for _, arg := range args {
		if arg == "--force-migrate" {
			forceMigrate = true
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		output.Error("resolve home dir: %s", err)
		os.Exit(1)
	}

	dbPath := filepath.Join(home, ".anvil", "runs.db")

	db, err := openDashboardDB(dbPath, forceMigrate)
	if err != nil {
		output.Error("open dashboard store: %s", err)
		os.Exit(1)
	}

	reader := query.NewReader(db)
	w := writer.New(db, 0)
	defer func() {
		if cerr := w.Close(); cerr != nil {
			output.Error("close dashboard store: %s", cerr)
		}
	}()

	if err := dashboard.Run(DashboardAssets, reader, w); err != nil {
		output.Error("dashboard: %s", err)
		os.Exit(1)
	}
}
