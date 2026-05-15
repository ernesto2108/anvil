package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/output"
)

func cmdStats(cfg *config.App) {
	ctx := context.Background()

	home, err := os.UserHomeDir()
	if err != nil {
		output.Error("resolve home dir: %s", err)
		os.Exit(1)
	}
	dbPath := filepath.Join(home, ".anvil", "runs.db")
	db, err := openDashboardDB(dbPath, false)
	if err != nil {
		output.Error("open store: %s", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	var runs, digests, insights int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM runs`).Scan(&runs); err != nil {
		output.Error("query runs: %s", err)
		os.Exit(1)
	}
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM digests`).Scan(&digests); err != nil {
		output.Error("query digests: %s", err)
		os.Exit(1)
	}
	// dream_insights may not exist on older DBs — treat missing table as 0.
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM dream_insights`).Scan(&insights); err != nil {
		insights = 0
	}

	fmt.Printf("runs:     %d\n", runs)
	fmt.Printf("digests:  %d\n", digests)
	fmt.Printf("insights: %d\n", insights)

	_ = cfg
}
