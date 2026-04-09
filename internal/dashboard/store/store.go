package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	"github.com/ernesto2108/anvil/internal/instrumentation"
)

const defaultMaxRuns = 500

// SQLiteStore persists instrumentation events to a local SQLite database.
type SQLiteStore struct {
	db      *sql.DB
	maxRuns int
}

// New opens (or creates) a SQLite database at dbPath, applies migrations from
// migrationsPath, and returns a ready-to-use SQLiteStore. maxRuns controls how
// many completed runs are retained before the oldest are pruned; values <= 0
// default to 500.
func New(dbPath string, migrationsPath string, maxRuns int) (*SQLiteStore, error) {
	if maxRuns <= 0 {
		maxRuns = defaultMaxRuns
	}

	db, err := openDB(dbPath)
	if err != nil {
		return nil, err
	}

	if err := RunMigrations(db, migrationsPath); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &SQLiteStore{db: db, maxRuns: maxRuns}, nil
}

// NewFS opens (or creates) a SQLite database at dbPath, applies migrations from
// the provided filesystem (typically embed.FS), and returns a ready-to-use
// SQLiteStore. Use this variant when migrations are embedded in the binary.
// migrationsSubPath is the sub-directory inside the fs.FS where migrations live
// (pass "migrations" if the embed root is the repo root, or "." if the embed
// root is already the migrations dir).
// maxRuns controls how many completed runs are retained before the oldest are
// pruned; values <= 0 default to 500.
func NewFS(dbPath string, migrations fs.FS, migrationsSubPath string, maxRuns int) (*SQLiteStore, error) {
	if maxRuns <= 0 {
		maxRuns = defaultMaxRuns
	}

	db, err := openDB(dbPath)
	if err != nil {
		return nil, err
	}

	subFS := migrations
	if migrationsSubPath != "" && migrationsSubPath != "." {
		subFS, err = fs.Sub(migrations, migrationsSubPath)
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("dashboard/store: resolver sub-filesystem de migraciones %q: %w", migrationsSubPath, err)
		}
	}

	if err := RunMigrationsFS(db, subFS); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &SQLiteStore{db: db, maxRuns: maxRuns}, nil
}

// openDB creates the database file (with restrictive permissions) and opens a
// *sql.DB with WAL journal mode and foreign keys enabled. The caller is
// responsible for closing the returned *sql.DB on error.
func openDB(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("dashboard/store: crear directorio de base de datos: %w", err)
	}

	// Pre-create the file with 0600 permissions to avoid a window where the
	// file exists with umask-derived permissions (typically 0644).
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		f, err := os.OpenFile(dbPath, os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			return nil, fmt.Errorf("dashboard/store: crear archivo de base de datos: %w", err)
		}
		_ = f.Close()
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("dashboard/store: abrir base de datos: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("dashboard/store: configurar journal_mode=WAL: %w", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("dashboard/store: configurar foreign_keys=ON: %w", err)
	}

	return db, nil
}

// WriteEvent implements instrumentation.EventWriter. It always appends a row to
// the raw events table and then updates the structured tables (runs, agents,
// files) according to the event type. Every event is written inside a
// transaction so structured and raw rows are always consistent.
func (s *SQLiteStore) WriteEvent(ev instrumentation.Event) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("dashboard/store: iniciar transacción: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	// Dispatch to structured tables first (runs must exist before events FK).
	switch ev.EventType {
	case instrumentation.EventRunStart:
		if err := s.handleRunStart(tx, ev); err != nil {
			return err
		}

	case instrumentation.EventRunEnd:
		if err := s.handleRunEnd(tx, ev); err != nil {
			return err
		}

	case instrumentation.EventAgentStart:
		if err := s.handleAgentStart(tx, ev); err != nil {
			return err
		}

	case instrumentation.EventAgentEnd:
		if err := s.handleAgentEnd(tx, ev); err != nil {
			return err
		}

	case instrumentation.EventAgentError:
		if err := s.handleAgentError(tx, ev); err != nil {
			return err
		}

	case instrumentation.EventFileTouched:
		if err := s.handleFileTouched(tx, ev); err != nil {
			return err
		}

	case instrumentation.EventQAScore:
		if err := s.handleQAScore(tx, ev); err != nil {
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
		return fmt.Errorf("dashboard/store: insertar en events: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("dashboard/store: confirmar transacción: %w", err)
	}

	// Rotate outside the per-event transaction — it opens its own.
	if ev.EventType == instrumentation.EventRunEnd {
		if err := s.rotate(); err != nil {
			return err
		}
	}

	return nil
}

// Close closes the underlying database connection.
func (s *SQLiteStore) Close() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("dashboard/store: cerrar base de datos: %w", err)
	}
	return nil
}

// expectRowsAffected returns an error when an UPDATE affects zero rows.
func expectRowsAffected(res sql.Result, table, id string) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("dashboard/store: obtener rows affected en %s: %w", table, err)
	}
	if n == 0 {
		return fmt.Errorf("dashboard/store: %s no encontrado en %s", id, table)
	}
	return nil
}

// --- private handlers -------------------------------------------------------

func (s *SQLiteStore) handleRunStart(tx *sql.Tx, ev instrumentation.Event) error {
	var p instrumentation.RunStartPayload
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		return fmt.Errorf("dashboard/store: deserializar RunStartPayload: %w", err)
	}

	const q = `
		INSERT INTO runs (id, task_id, task_desc, status, complexity, provider, started_at)
		VALUES (?, ?, ?, 'running', ?, ?, ?)`

	if _, err := tx.Exec(q,
		ev.RunID,
		p.TaskID,
		p.TaskDescription,
		p.Complexity,
		p.Provider,
		ev.Timestamp.Format("2006-01-02T15:04:05.999999999Z07:00"),
	); err != nil {
		return fmt.Errorf("dashboard/store: insertar en runs (run.start): %w", err)
	}
	return nil
}

func (s *SQLiteStore) handleRunEnd(tx *sql.Tx, ev instrumentation.Event) error {
	var p instrumentation.RunEndPayload
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		return fmt.Errorf("dashboard/store: deserializar RunEndPayload: %w", err)
	}

	const q = `
		UPDATE runs
		SET status        = ?,
		    ended_at      = ?,
		    duration_ms   = ?,
		    total_tokens  = ?,
		    files_touched = ?,
		    agents_count  = ?
		WHERE id = ?`

	res, err := tx.Exec(q,
		p.Status,
		ev.Timestamp.Format("2006-01-02T15:04:05.999999999Z07:00"),
		p.DurationMs,
		p.TotalTokens,
		p.TotalFiles,
		p.AgentsCompleted+p.AgentsFailed,
		ev.RunID,
	)
	if err != nil {
		return fmt.Errorf("dashboard/store: actualizar runs (run.end): %w", err)
	}
	if err := expectRowsAffected(res, "runs", ev.RunID); err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStore) handleAgentStart(tx *sql.Tx, ev instrumentation.Event) error {
	var p instrumentation.AgentStartPayload
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		return fmt.Errorf("dashboard/store: deserializar AgentStartPayload: %w", err)
	}

	dependsOnJSON, err := json.Marshal(p.DependsOn)
	if err != nil {
		return fmt.Errorf("dashboard/store: serializar depends_on: %w", err)
	}

	const q = `
		INSERT INTO agents (run_id, agent_id, agent_role, sequence, depends_on, model, status, started_at)
		VALUES (?, ?, ?, ?, ?, ?, 'running', ?)`

	if _, err := tx.Exec(q,
		ev.RunID,
		p.AgentID,
		p.AgentRole,
		p.Sequence,
		string(dependsOnJSON),
		p.Model,
		ev.Timestamp.Format("2006-01-02T15:04:05.999999999Z07:00"),
	); err != nil {
		return fmt.Errorf("dashboard/store: insertar en agents (agent.start): %w", err)
	}
	return nil
}

func (s *SQLiteStore) handleAgentEnd(tx *sql.Tx, ev instrumentation.Event) error {
	var p instrumentation.AgentEndPayload
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		return fmt.Errorf("dashboard/store: deserializar AgentEndPayload: %w", err)
	}

	const updateAgent = `
		UPDATE agents
		SET status        = ?,
		    ended_at      = ?,
		    duration_ms   = ?,
		    tokens_input  = ?,
		    tokens_output = ?,
		    tokens_total  = ?,
		    exit_code     = ?
		WHERE run_id = ? AND agent_id = ?`

	res, err := tx.Exec(updateAgent,
		p.Status,
		ev.Timestamp.Format("2006-01-02T15:04:05.999999999Z07:00"),
		p.DurationMs,
		p.TokensInput,
		p.TokensOutput,
		p.TokensTotal,
		p.ExitCode,
		ev.RunID,
		p.AgentID,
	)
	if err != nil {
		return fmt.Errorf("dashboard/store: actualizar agents (agent.end): %w", err)
	}
	if err := expectRowsAffected(res, "agents", p.AgentID); err != nil {
		return err
	}

	const insertFile = `INSERT INTO files (run_id, agent_id, path, operation) VALUES (?, ?, ?, ?)`
	for _, path := range p.FilesTouched {
		if _, err := tx.Exec(insertFile, ev.RunID, p.AgentID, path, "touched"); err != nil {
			return fmt.Errorf("dashboard/store: insertar en files (agent.end): %w", err)
		}
	}
	return nil
}

func (s *SQLiteStore) handleAgentError(tx *sql.Tx, ev instrumentation.Event) error {
	var p instrumentation.AgentErrorPayload
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		return fmt.Errorf("dashboard/store: deserializar AgentErrorPayload: %w", err)
	}

	const q = `
		UPDATE agents
		SET status      = 'failed',
		    ended_at    = ?,
		    duration_ms = ?,
		    error_msg   = ?
		WHERE run_id = ? AND agent_id = ?`

	res, err := tx.Exec(q,
		ev.Timestamp.Format("2006-01-02T15:04:05.999999999Z07:00"),
		p.DurationMs,
		p.Error,
		ev.RunID,
		p.AgentID,
	)
	if err != nil {
		return fmt.Errorf("dashboard/store: actualizar agents (agent.error): %w", err)
	}
	if err := expectRowsAffected(res, "agents", p.AgentID); err != nil {
		return err
	}
	return nil
}

func (s *SQLiteStore) handleFileTouched(tx *sql.Tx, ev instrumentation.Event) error {
	var p instrumentation.FileTouchedPayload
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		return fmt.Errorf("dashboard/store: deserializar FileTouchedPayload: %w", err)
	}

	const q = `INSERT INTO files (run_id, agent_id, path, operation) VALUES (?, ?, ?, ?)`
	if _, err := tx.Exec(q, ev.RunID, p.AgentID, p.Path, p.Operation); err != nil {
		return fmt.Errorf("dashboard/store: insertar en files (file.touched): %w", err)
	}
	return nil
}

func (s *SQLiteStore) handleQAScore(tx *sql.Tx, ev instrumentation.Event) error {
	var p instrumentation.QAScorePayload
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		return fmt.Errorf("dashboard/store: deserializar QAScorePayload: %w", err)
	}

	const q = `UPDATE runs SET qa_score = ? WHERE id = ?`
	res, err := tx.Exec(q, p.Score, ev.RunID)
	if err != nil {
		return fmt.Errorf("dashboard/store: actualizar runs (qa.score): %w", err)
	}
	if err := expectRowsAffected(res, "runs", ev.RunID); err != nil {
		return err
	}
	return nil
}
