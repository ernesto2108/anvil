package store

import (
	"context"
	"fmt"
	"strings"
)

// DeleteRun deletes a single run and all associated data.
// Foreign keys with ON DELETE CASCADE handle cleanup of child tables.
func (s *SQLiteStore) DeleteRun(ctx context.Context, runID string) error {
	const q = `DELETE FROM runs WHERE id = ?`
	res, err := s.db.ExecContext(ctx, q, runID)
	if err != nil {
		return fmt.Errorf("dashboard/store: eliminar run %s: %w", runID, err)
	}
	return expectRowsAffected(res, "runs", runID)
}

// DeleteRuns deletes multiple runs in a single transaction.
// Foreign keys with ON DELETE CASCADE handle cleanup of child tables.
func (s *SQLiteStore) DeleteRuns(ctx context.Context, runIDs []string) error {
	if len(runIDs) == 0 {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("dashboard/store: iniciar transacción de eliminación: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	placeholders := make([]string, len(runIDs))
	args := make([]any, len(runIDs))
	for i, id := range runIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	q := `DELETE FROM runs WHERE id IN (` + strings.Join(placeholders, ",") + `)`
	if _, err := tx.ExecContext(ctx, q, args...); err != nil {
		return fmt.Errorf("dashboard/store: eliminar runs en batch: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("dashboard/store: confirmar eliminación: %w", err)
	}
	return nil
}
