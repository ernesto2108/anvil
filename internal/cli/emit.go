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
//
// Env vars honored:
//   - ANVIL_SKIP_EMIT=1: drain stdin and exit immediately. Used by anvil's own
//     auth probe and any other internal claude invocations that should not
//     surface in the dashboard.
//   - ANVIL_PARENT_RUN_ID: when set, hook events attach to this existing run
//     instead of lazily creating a new orphan run keyed by session_id.
//   - ANVIL_AGENT_ID: when set, tool/file events use this agent_id instead of
//     querying the active agent in the DB.
func cmdEmit(_ *config.App) {
	if os.Getenv("ANVIL_SKIP_EMIT") == "1" {
		// Drain stdin so the upstream writer doesn't block on a closed pipe.
		_, _ = io.Copy(io.Discard, io.LimitReader(os.Stdin, maxStdinBytes))
		return
	}
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
	defer w.Close() //nolint:errcheck

	// Set busy_timeout for concurrent access with the dashboard.
	if _, err := w.DB().Exec("PRAGMA busy_timeout=5000"); err != nil {
		return fmt.Errorf("set busy_timeout: %w", err)
	}

	return translateHook(raw, w)
}
