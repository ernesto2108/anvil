package writer

import "fmt"

// rotate deletes the oldest runs when the total count exceeds w.maxRuns.
// Foreign keys with ON DELETE CASCADE handle cleanup of agents, files, and events.
func (w *EventWriter) rotate() error {
	var count int
	if err := w.db.QueryRow("SELECT COUNT(*) FROM runs").Scan(&count); err != nil {
		return fmt.Errorf("dashboard/writer: contar runs para rotación: %w", err)
	}

	if count <= w.maxRuns {
		return nil
	}

	excess := count - w.maxRuns

	tx, err := w.db.Begin()
	if err != nil {
		return fmt.Errorf("dashboard/writer: iniciar transacción de rotación: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	const q = `DELETE FROM runs WHERE id IN (SELECT id FROM runs ORDER BY started_at ASC LIMIT ?)`
	if _, err := tx.Exec(q, excess); err != nil {
		return fmt.Errorf("dashboard/writer: eliminar runs antiguos: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("dashboard/writer: confirmar rotación: %w", err)
	}

	return nil
}
