package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/internal/dashboard/writer"
	"github.com/ernesto2108/anvil/pkg/config"
)

const maxStdinBytes = 1 << 20 // 1 MB

// cmdEmit reads a Claude Code hook JSON from stdin and translates it into
// Anvil instrumentation events written to the SQLite store.
//
// Design constraints:
//   - ALWAYS exits with code 0 (errors go to stderr)
//   - Must complete in <100ms (ephemeral process, no async)
//   - stdin limited to 1MB to prevent blocking on huge PostToolUse responses
func cmdEmit(_ *config.App) {
	if err := runEmit(); err != nil {
		fmt.Fprintf(os.Stderr, "anvil emit: %s\n", err)
	}
	// Always exit 0 — a failing hook is worse than a missed event.
}

func runEmit() error {
	raw, err := io.ReadAll(io.LimitReader(os.Stdin, maxStdinBytes))
	if err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	if len(raw) == 0 {
		return fmt.Errorf("empty stdin")
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home dir: %w", err)
	}

	dbPath := filepath.Join(home, ".anvil", "runs.db")

	db, err := openDashboardDB(dbPath, false)
	if err != nil {
		return fmt.Errorf("open store: %w", err)
	}

	w := writer.New(db, 0)
	defer w.Close()

	// Set busy_timeout for concurrent access with the dashboard.
	if _, err := w.DB().Exec("PRAGMA busy_timeout=5000"); err != nil {
		return fmt.Errorf("set busy_timeout: %w", err)
	}

	return translateHook(raw, w)
}
