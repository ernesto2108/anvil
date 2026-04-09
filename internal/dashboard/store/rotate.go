package store

import "fmt"

// rotate deletes the oldest runs when the total count exceeds s.maxRuns.
// Foreign keys with ON DELETE CASCADE handle cleanup of agents, files, and events.
func (s *SQLiteStore) rotate() error {
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM runs").Scan(&count); err != nil {
		return fmt.Errorf("dashboard/store: contar runs para rotación: %w", err)
	}

	if count <= s.maxRuns {
		return nil
	}

	excess := count - s.maxRuns

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("dashboard/store: iniciar transacción de rotación: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	const q = `DELETE FROM runs WHERE id IN (SELECT id FROM runs ORDER BY started_at ASC LIMIT ?)`
	if _, err := tx.Exec(q, excess); err != nil {
		return fmt.Errorf("dashboard/store: eliminar runs antiguos: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("dashboard/store: confirmar rotación: %w", err)
	}

	return nil
}
