package writer

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/ernesto2108/anvil/internal/instrumentation"
)

const defaultMaxRuns = 500

// EventWriter persists instrumentation events to a local SQLite database.
type EventWriter struct {
	db      *sql.DB
	maxRuns int
}

// New creates an EventWriter wrapping the given *sql.DB.
// maxRuns controls how many completed runs are retained before the oldest are
// pruned; values <= 0 default to 500.
func New(db *sql.DB, maxRuns int) *EventWriter {
	if maxRuns <= 0 {
		maxRuns = defaultMaxRuns
	}
	return &EventWriter{db: db, maxRuns: maxRuns}
}

// WriteEvent implements instrumentation.EventWriter. It always appends a row to
// the raw events table and then updates the structured tables (runs, agents,
// files) according to the event type. Every event is written inside a
// transaction so structured and raw rows are always consistent.
func (w *EventWriter) WriteEvent(ev instrumentation.Event) error {
	tx, err := w.db.Begin()
	if err != nil {
		return fmt.Errorf("dashboard/writer: iniciar transacción: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Dispatch to structured tables first (runs must exist before events FK).
	switch ev.EventType {
	case instrumentation.EventRunStart:
		if err := handleRunStart(tx, ev); err != nil {
			return err
		}
	case instrumentation.EventRunEnd:
		if err := handleRunEnd(tx, ev); err != nil {
			return err
		}
	case instrumentation.EventAgentStart:
		if err := handleAgentStart(tx, ev); err != nil {
			return err
		}
	case instrumentation.EventAgentEnd:
		if err := handleAgentEnd(tx, ev); err != nil {
			return err
		}
	case instrumentation.EventAgentError:
		if err := handleAgentError(tx, ev); err != nil {
			return err
		}
		if err := handleErrorFingerprint(w, tx, ev); err != nil {
			return err
		}
	case instrumentation.EventFileTouched:
		if err := handleFileTouched(tx, ev); err != nil {
			return err
		}
	case instrumentation.EventQAScore:
		if err := handleQAScore(tx, ev); err != nil {
			return err
		}
	case instrumentation.EventToolUse:
		if err := handleToolUse(tx, ev); err != nil {
			return err
		}
	case instrumentation.EventRunError:
		if err := handleRunError(tx, ev); err != nil {
			return err
		}
		if err := handleErrorFingerprint(w, tx, ev); err != nil {
			return err
		}
	case instrumentation.EventTaskCreated:
		if err := handleTaskCreated(tx, ev); err != nil {
			return err
		}
	case instrumentation.EventTaskCompleted:
		if err := handleTaskCompleted(tx, ev); err != nil {
			return err
		}
	case instrumentation.EventCompaction, instrumentation.EventPermissionDenied:
		// Raw-only events: no structured table needed.
	case instrumentation.EventUserPrompt:
		if err := handleUserPrompt(tx, ev); err != nil {
			return err
		}
	}

	// Insert into the raw events audit trail (after structured tables so FK is satisfied).
	const insertEvent = `
		INSERT INTO events (run_id, event_id, event_type, schema_version, timestamp, payload)
		VALUES (?, ?, ?, ?, ?, ?)`

	if _, err := tx.Exec(insertEvent,
		ev.RunID,
		ev.EventID,
		ev.EventType,
		ev.SchemaVersion,
		ev.Timestamp.Format("2006-01-02T15:04:05.999999999Z07:00"),
		string(ev.Payload),
	); err != nil {
		return fmt.Errorf("dashboard/writer: insertar en events: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("dashboard/writer: confirmar transacción: %w", err)
	}

	// Rotate outside the per-event transaction — it opens its own.
	if ev.EventType == instrumentation.EventRunEnd {
		if err := w.rotate(); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the underlying database connection.
func (w *EventWriter) Close() error {
	if err := w.db.Close(); err != nil {
		return fmt.Errorf("dashboard/writer: cerrar base de datos: %w", err)
	}
	return nil
}

// DB returns the underlying *sql.DB. Exported for use by the emit command
// which needs direct access for busy_timeout pragma.
func (w *EventWriter) DB() *sql.DB {
	return w.db
}

// marshalPayload is a convenience for unmarshalling event payloads.
func marshalPayload[T any](ev instrumentation.Event) (T, error) {
	var p T
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		return p, fmt.Errorf("dashboard/writer: deserializar %T: %w", p, err)
	}
	return p, nil
}
