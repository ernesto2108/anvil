package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/pkg/output"
)

// cmdMigrate opens the dashboard DB (which runs migrations automatically)
// and reports the resulting schema version.
func cmdMigrate(forceMigrate bool) {
	home, err := os.UserHomeDir()
	if err != nil {
		output.Error("migrate: no se pudo resolver home dir: %s", err)
		os.Exit(1)
	}

	dbPath := filepath.Join(home, ".anvil", "runs.db")
	fmt.Printf("  DB: %s\n", dbPath)

	db, err := openDashboardDB(dbPath, forceMigrate)
	if err != nil {
		output.Error("migrate: %s", err)
		os.Exit(1)
	}
	defer db.Close()

	var version int
	if err := db.QueryRow("SELECT version FROM schema_migrations LIMIT 1").Scan(&version); err != nil {
		output.Error("migrate: leer versión: %s", err)
		os.Exit(1)
	}

	fmt.Printf("  %s schema version: %d\n", output.Green("✓"), version)
}
