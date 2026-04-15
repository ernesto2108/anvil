package query

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// DeleteRun deletes a single run and all associated data.
// Foreign keys with ON DELETE CASCADE handle cleanup of child tables.
func (r *Reader) DeleteRun(ctx context.Context, runID string) error {
	const q = `DELETE FROM runs WHERE id = ?`
	res, err := r.db.ExecContext(ctx, q, runID)
	if err != nil {
		return fmt.Errorf("query: eliminar run %s: %w", runID, err)
	}
	return expectRowsAffected(res, "runs", runID)
}

// DeleteRuns deletes multiple runs in a single transaction.
func (r *Reader) DeleteRuns(ctx context.Context, runIDs []string) error {
	if len(runIDs) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("query: iniciar transacción de eliminación: %w", err)
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
		return fmt.Errorf("query: eliminar runs en batch: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("query: confirmar eliminación: %w", err)
	}
	return nil
}

func expectRowsAffected(res sql.Result, table, id string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("query: obtener rows affected en %s: %w", table, err)
	}
	if n == 0 {
		return fmt.Errorf("query: %s no encontrado en %s", id, table)
	}
	return nil
}
