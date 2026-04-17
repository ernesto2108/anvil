package cli

import (
	"context"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/ernesto2108/anvil/internal/mcp"
	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/output"
)

// cmdMCPServer starts an MCP server over stdio. It blocks until the process is
// killed or stdin is closed by the client.
//
// Usage:
//
//	anvil mcp-server
//
// Configure Claude Code to use it by adding to ~/.claude/settings.json:
//
//	"mcpServers": {
//	  "anvil": { "command": "anvil", "args": ["mcp-server"] }
//	}
//
// Optionally set ANVIL_VAULT_PATH to enable vault tools (get_backlog, get_task):
//
//	"env": { "ANVIL_VAULT_PATH": "/Users/you/projects/anvil-knowledge-base" }
func cmdMCPServer(cfg *config.App) {
	home, err := os.UserHomeDir()
	if err != nil {
		output.Error("resolve home dir: %s", err)
		os.Exit(1)
	}
	dbPath := filepath.Join(home, ".anvil", "runs.db")

	db, err := openDashboardDB(dbPath, false)
	if err != nil {
		// DB is optional — continue without it (inventory + context tools still work).
		output.Warn("open database (run/memory tools disabled): %s", err)
		db = nil
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := mcp.Serve(ctx, cfg, db); err != nil && err != context.Canceled {
		output.Error("mcp server: %s", err)
		os.Exit(1)
	}
}
