package cli

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"
)

// EnsureSyntheticRun inserts a placeholder run row for a manual capture so the
// digest's FK to runs is satisfied. The run ID uses the format
// `d_manual_<ts_unix>_<rand6hex>` (D5 — synthetic only for manual captures).
//
// If a run with the generated ID already exists (extremely unlikely given the
// random suffix), the INSERT is a no-op and the existing row is returned.
func EnsureSyntheticRun(ctx context.Context, db *sql.DB, project string, ts time.Time) (string, error) {
	runID, err := newManualRunID(ts)
	if err != nil {
		return "", fmt.Errorf("synthetic_run: generate id: %w", err)
	}

	now := ts.UTC().Format(time.RFC3339Nano)

	const q = `INSERT INTO runs (id, task_id, task_desc, status, started_at, ended_at, project, task_source)
		VALUES (?, ?, ?, 'success', ?, ?, ?, 'manual')
		ON CONFLICT(id) DO NOTHING`

	if _, err := db.ExecContext(ctx, q, runID, runID, "manual transcript capture", now, now, project); err != nil {
		return "", fmt.Errorf("synthetic_run: insert: %w", err)
	}

	return runID, nil
}

// newManualRunID generates a run ID of the form `d_manual_<ts_unix>_<rand6hex>`.
func newManualRunID(ts time.Time) (string, error) {
	var b [3]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("d_manual_%d_%s", ts.Unix(), hex.EncodeToString(b[:])), nil
}
