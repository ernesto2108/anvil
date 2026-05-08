package capture

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/internal/memory/transcript"
)

// ResolveProject determines the project name for a capture using a chain of
// strategies:
//
//  1. override — caller-supplied name (--project flag), always wins.
//  2. .git guard on td.Cwd — if cwd has a .git directory/file, use Base(cwd).
//  3. run lookup — find the project of the run linked to td.SessionID in the DB.
//  4. fallback — Base(cwd) without the .git check (best effort).
//
// Names that appear in the denylist are skipped so "projects", "src", etc.
// never become a project identifier.
func ResolveProject(ctx context.Context, td transcript.TranscriptDigest, override string, db *sql.DB) (string, error) {
	if override != "" {
		return override, nil
	}

	// Strategy 2: git root.
	if td.Cwd != "" {
		name := filepath.Base(td.Cwd)
		if !transcript.IsGeneric(name) {
			if hasGit(td.Cwd) {
				return name, nil
			}
		}
	}

	// Strategy 3: run lookup by session ID.
	if td.SessionID != "" && db != nil {
		name, err := projectFromSession(ctx, db, td.SessionID)
		if err != nil {
			// Non-fatal — fall through to next strategy.
			_ = err
		}
		if name != "" && !transcript.IsGeneric(name) {
			return name, nil
		}
	}

	// Strategy 4: fallback to Base(cwd).
	if td.Cwd != "" {
		name := filepath.Base(td.Cwd)
		if name != "" && name != "." {
			return name, nil
		}
	}

	return "", fmt.Errorf("capture: cannot resolve project (no cwd, no session match, no override)")
}

// hasGit returns true if cwd contains a .git entry (directory or file for
// git worktrees).
func hasGit(cwd string) bool {
	_, err := os.Stat(filepath.Join(cwd, ".git"))
	return err == nil
}

// projectFromSession queries the runs table for the project of the most
// recent run linked to the given session ID.
func projectFromSession(ctx context.Context, db *sql.DB, sessionID string) (string, error) {
	const q = `SELECT project FROM runs WHERE session_id = ? ORDER BY started_at DESC LIMIT 1`
	var project sql.NullString
	if err := db.QueryRowContext(ctx, q, sessionID).Scan(&project); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("capture: resolve project from session: %w", err)
	}
	return project.String, nil
}
