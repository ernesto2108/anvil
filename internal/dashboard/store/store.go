package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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

// ResolveRunBySession returns the run_id associated with the given session_id.
// Returns ("", nil) if no run exists for that session.
func (s *SQLiteStore) ResolveRunBySession(sessionID string) (string, error) {
	var runID string
	err := s.db.QueryRow(
		`SELECT id FROM runs WHERE session_id = ? LIMIT 1`, sessionID,
	).Scan(&runID)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("dashboard/store: resolver run por session_id: %w", err)
	}
	return runID, nil
}

// UpdateTaskDesc sets task_desc for a run only if it is currently empty or NULL.
func (s *SQLiteStore) UpdateTaskDesc(runID, desc string) error {
	const q = `UPDATE runs SET task_desc = ? WHERE id = ? AND (task_desc IS NULL OR task_desc = '')`
	_, err := s.db.Exec(q, desc, runID)
	if err != nil {
		return fmt.Errorf("dashboard/store: actualizar task_desc: %w", err)
	}
	return nil
}

// ComputeRunTotals calculates files_touched, agents_count, and duration_ms
// from the DB and updates the run row. Called when closing a run via SessionEnd.
func (s *SQLiteStore) ComputeRunTotals(runID string) error {
	const q = `
		UPDATE runs
		SET files_touched = (SELECT COUNT(DISTINCT path) FROM files WHERE run_id = ?),
		    agents_count  = (SELECT COUNT(*) FROM agents WHERE run_id = ?),
		    duration_ms   = CASE
		        WHEN started_at IS NOT NULL AND ended_at IS NOT NULL
		        THEN CAST((julianday(ended_at) - julianday(started_at)) * 86400000 AS INTEGER)
		        ELSE duration_ms
		    END
		WHERE id = ?`
	_, err := s.db.Exec(q, runID, runID, runID)
	if err != nil {
		return fmt.Errorf("dashboard/store: calcular totales de run: %w", err)
	}
	return nil
}

// CleanupStaleRuns marks runs stuck in 'running' status as 'abandoned'
// if they have no activity for more than the given minutes threshold.
// Called at dashboard startup to clean up orphaned sessions.
func (s *SQLiteStore) CleanupStaleRuns(staleMinutes int) (int64, error) {
	const q = `
		UPDATE runs
		SET status    = 'abandoned',
		    ended_at  = datetime('now'),
		    duration_ms = CAST((julianday(datetime('now')) - julianday(started_at)) * 86400000 AS INTEGER)
		WHERE status = 'running'
		  AND started_at < datetime('now', ? || ' minutes')`
	res, err := s.db.Exec(q, fmt.Sprintf("-%d", staleMinutes))
	if err != nil {
		return 0, fmt.Errorf("dashboard/store: limpiar runs huérfanos: %w", err)
	}
	return res.RowsAffected()
}

// BackfillProjects sets the project column for runs that have project = ''
// by deriving it from the most common file path prefix in the files table.
func (s *SQLiteStore) BackfillProjects() (int64, error) {
	// For each run with empty project that has files, extract project name
	// from the first file path's directory structure.
	const q = `
		UPDATE runs
		SET project = (
			SELECT REPLACE(
				RTRIM(SUBSTR(f.path, 1, INSTR(SUBSTR(f.path, 2), '/') + 1), '/'),
				'/', ''
			)
			FROM files f
			WHERE f.run_id = runs.id
			  AND f.path LIKE '/%'
			LIMIT 1
		)
		WHERE project = ''
		  AND EXISTS (SELECT 1 FROM files WHERE run_id = runs.id AND path LIKE '/%')`

	// The above is fragile with deep paths. Use a simpler approach:
	// get the basename of the common prefix directory.
	// Actually, let's just do it in Go for reliability.
	rows, err := s.db.Query(`
		SELECT DISTINCT r.id, f.path
		FROM runs r
		JOIN files f ON f.run_id = r.id
		WHERE r.project = '' AND f.path LIKE '/%'
		ORDER BY r.id`)
	if err != nil {
		return 0, fmt.Errorf("dashboard/store: backfill projects query: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	// Collect first file path per run
	runProject := make(map[string]string)
	for rows.Next() {
		var runID, path string
		if err := rows.Scan(&runID, &path); err != nil {
			return 0, err
		}
		if _, ok := runProject[runID]; ok {
			continue // already have one
		}
		// Extract project: find common project root
		// e.g. /Users/ernesto/projects/anvil/src/foo.go → anvil
		runProject[runID] = extractProjectFromPath(path)
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	var updated int64
	for runID, project := range runProject {
		if project == "" {
			continue
		}
		res, err := s.db.Exec(`UPDATE runs SET project = ? WHERE id = ? AND project = ''`, project, runID)
		if err != nil {
			return updated, err
		}
		n, _ := res.RowsAffected()
		updated += n
	}
	return updated, nil
}

// extractProjectFromPath finds the project directory name from a file path.
// Heuristic: look for a "projects" or "repos" parent, take the next segment.
// Fallback: take the 4th path segment (usually /Users/x/projects/NAME/...).
func extractProjectFromPath(path string) string {
	parts := strings.Split(path, "/")
	for i, p := range parts {
		if (p == "projects" || p == "repos" || p == "src" || p == "workspace") && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	if len(parts) >= 5 {
		return parts[4]
	}
	return ""
}

// AgentStartedAt returns the started_at timestamp for a given agent.
// Returns zero time if not found.
func (s *SQLiteStore) AgentStartedAt(runID, agentID string) (string, bool) {
	var startedAt string
	err := s.db.QueryRow(
		`SELECT started_at FROM agents WHERE run_id = ? AND agent_id = ? LIMIT 1`,
		runID, agentID,
	).Scan(&startedAt)
	if err != nil {
		return "", false
	}
	return startedAt, true
}

// ActiveAgentID returns the agent_id of the most recently started agent
// with status "running" for the given run. Returns "" if none found.
func (s *SQLiteStore) ActiveAgentID(runID string) string {
	var agentID string
	err := s.db.QueryRow(
		`SELECT agent_id FROM agents WHERE run_id = ? AND status = 'running' ORDER BY started_at DESC LIMIT 1`,
		runID,
	).Scan(&agentID)
	if err != nil {
		return ""
	}
	return agentID
}

// DB returns the underlying *sql.DB. Exported for use by the emit command
// which needs direct access for busy_timeout pragma.
func (s *SQLiteStore) DB() *sql.DB {
	return s.db
}

// --- private handlers -------------------------------------------------------

func (s *SQLiteStore) handleRunStart(tx *sql.Tx, ev instrumentation.Event) error {
	var p instrumentation.RunStartPayload
	if err := json.Unmarshal(ev.Payload, &p); err != nil {
		return fmt.Errorf("dashboard/store: deserializar RunStartPayload: %w", err)
	}

	var sessionID *string
	if p.SessionID != "" {
		sessionID = &p.SessionID
	}

	var parentRunID *string
	if p.ParentRunID != "" {
		parentRunID = &p.ParentRunID
	}

	const q = `
		INSERT INTO runs (id, task_id, task_desc, status, complexity, provider, started_at, session_id, project, parent_run_id)
		VALUES (?, ?, ?, 'running', ?, ?, ?, ?, ?, ?)`

	if _, err := tx.Exec(q,
		ev.RunID,
		p.TaskID,
		p.TaskDescription,
		p.Complexity,
		p.Provider,
		ev.Timestamp.Format("2006-01-02T15:04:05.999999999Z07:00"),
		sessionID,
		p.Project,
		parentRunID,
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

	const maxOutputBytes = 100 * 1024 // 100 KB

	output := p.Output
	if len(output) > maxOutputBytes {
		output = output[:maxOutputBytes]
		// Retroceder si cortamos a mitad de un carácter UTF-8 multibyte.
		for len(output) > 0 && output[len(output)-1]&0xC0 == 0x80 {
			output = output[:len(output)-1]
		}
		// Si el último byte es un líder UTF-8 incompleto, eliminarlo también.
		if len(output) > 0 && output[len(output)-1]&0xC0 == 0xC0 {
			output = output[:len(output)-1]
		}
	}

	const updateAgent = `
		UPDATE agents
		SET status        = ?,
		    ended_at      = ?,
		    duration_ms   = ?,
		    tokens_input  = ?,
		    tokens_output = ?,
		    tokens_total  = ?,
		    exit_code     = ?,
		    output        = ?
		WHERE run_id = ? AND agent_id = ?`

	res, err := tx.Exec(updateAgent,
		p.Status,
		ev.Timestamp.Format("2006-01-02T15:04:05.999999999Z07:00"),
		p.DurationMs,
		p.TokensInput,
		p.TokensOutput,
		p.TokensTotal,
		p.ExitCode,
		output,
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

	var diff *string
	if p.Diff != "" {
		diff = &p.Diff
	}

	const q = `INSERT INTO files (run_id, agent_id, path, operation, diff) VALUES (?, ?, ?, ?, ?)`
	if _, err := tx.Exec(q, ev.RunID, p.AgentID, p.Path, p.Operation, diff); err != nil {
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
