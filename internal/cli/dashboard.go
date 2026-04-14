//go:build dashboard

package cli

import (
	"embed"
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/internal/dashboard"
	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/output"
)

// DashboardAssets is set by cmd/anvil/embed_dashboard.go at init time.
// It holds the embedded React frontend built with `wails build` or
// `go build -tags dashboard`.
var DashboardAssets embed.FS

func cmdDashboard(_ *config.App) {
	home, err := os.UserHomeDir()
	if err != nil {
		output.Error("resolve home dir: %s", err)
		os.Exit(1)
	}

	dbPath := filepath.Join(home, ".anvil", "runs.db")

	// Prefer filesystem migrations over embedded ones. This prevents the
	// "stale binary" problem where the dashboard binary was built before
	// new migrations were added — the DB schema would be outdated and the
	// frontend would crash on missing columns/tables.
	s, err := openStorePreferFilesystem(dbPath)
	if err != nil {
		output.Error("open dashboard store: %s", err)
		os.Exit(1)
	}
	defer func() {
		if cerr := s.Close(); cerr != nil {
			output.Error("close dashboard store: %s", cerr)
		}
	}()

	if err := dashboard.Run(DashboardAssets, s); err != nil {
		output.Error("dashboard: %s", err)
		os.Exit(1)
	}
}
