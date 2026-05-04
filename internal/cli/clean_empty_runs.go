package cli

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/output"
)

// cmdCleanEmptyRuns is a one-shot maintenance command that removes runs with
// no real activity (no user prompts, no tool uses, no created tasks) from the
// SQLite store. Use case: clean up phantom runs left over from before lazy
// run creation was implemented (sessions where the user opened claude and
// exited without interacting).
//
// Dry-run by default. Pass --force to actually delete.
func cmdCleanEmptyRuns(_ *config.App, args []string) {
	force := false
	for _, a := range args {
		if a == "--force" || a == "-f" {
			force = true
		}
	}

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
	defer db.Close() //nolint:errcheck

	candidates, err := emptyRunIDs(db)
	if err != nil {
		output.Error("query empty runs: %s", err)
		os.Exit(1)
	}

	if len(candidates) == 0 {
		fmt.Println("No empty runs found. Store is clean.")
		return
	}

	fmt.Printf("Found %d empty run(s) (no prompts, tool uses, or tasks).\n", len(candidates))
	if !force {
		fmt.Println()
		fmt.Println("Dry run — nothing deleted. Re-run with --force to remove them.")
		return
	}

	deleted, err := deleteRuns(db, candidates)
	if err != nil {
		output.Error("delete runs: %s", err)
		os.Exit(1)
	}
	fmt.Printf("Deleted %d empty run(s).\n", deleted)
}

// emptyRunIDs returns the IDs of runs with no user_prompt, no tool_use, and
// no task_created event. These are runs whose session opened but never
// produced real activity.
//
// Orchestrator-owned runs (parent_run_id IS NOT NULL or with descendants) are
// preserved unconditionally — even if their direct events are empty, the
// child runs may carry the real work.
func emptyRunIDs(db *sql.DB) ([]string, error) {
	const q = `
		SELECT r.id
		FROM runs r
		WHERE r.parent_run_id IS NULL
		  AND NOT EXISTS (
		    SELECT 1 FROM runs c WHERE c.parent_run_id = r.id
		  )
		  AND NOT EXISTS (
		    SELECT 1 FROM prompts p WHERE p.run_id = r.id
		  )
		  AND NOT EXISTS (
		    SELECT 1 FROM tool_uses t WHERE t.run_id = r.id
		  )
		  AND NOT EXISTS (
		    SELECT 1 FROM tasks tk WHERE tk.run_id = r.id
		  )`

	rows, err := db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// deleteRuns removes the given run IDs. Foreign keys with ON DELETE CASCADE
// take care of agents, files, events, tool_uses, prompts, tasks, and
// run_projects. Each delete runs in its own statement so a single bad row
// does not poison the whole batch.
func deleteRuns(db *sql.DB, ids []string) (int, error) {
	deleted := 0
	for _, id := range ids {
		res, err := db.Exec(`DELETE FROM runs WHERE id = ?`, id)
		if err != nil {
			return deleted, fmt.Errorf("delete %s: %w", id, err)
		}
		n, _ := res.RowsAffected()
		deleted += int(n)
	}
	return deleted, nil
}
