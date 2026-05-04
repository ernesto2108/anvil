package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// SaveDigest inserts a new digest into SQLite. Mirrors the embedding into
// vec_digests inside the same transaction (see UpsertDigest for the rationale).
func SaveDigest(ctx context.Context, db *sql.DB, d Digest) error {
	decisions, edgeCases, errors, embBlob, err := marshalDigest(d)
	if err != nil {
		return err
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	if d.CreatedAt.IsZero() {
		d.CreatedAt, _ = time.Parse(time.RFC3339Nano, now)
	}
	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = d.CreatedAt
	}

	source := d.Source
	if source == "" {
		source = "auto"
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("memory: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const q = `INSERT INTO digests (id, run_id, project, summary, decisions, edge_cases, errors, embedding, model_used, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	if _, err := tx.ExecContext(ctx, q,
		d.ID, d.RunID, d.Project, d.Summary,
		string(decisions), string(edgeCases), string(errors),
		embBlob, d.ModelUsed, source,
		d.CreatedAt.Format(time.RFC3339Nano),
		d.UpdatedAt.Format(time.RFC3339Nano),
	); err != nil {
		return fmt.Errorf("memory: insert digest: %w", err)
	}

	if err := upsertVecDigest(ctx, tx, d.ID, d.Project, embBlob); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("memory: commit save: %w", err)
	}
	return nil
}

// UpsertDigest inserts a new digest or updates the existing one for the same
// run_id. Used for incremental checkpoints — each agent completion rewrites
// the run's digest with the latest cumulative summary, so if the run crashes
// mid-flight the last checkpoint survives.
//
// On conflict, id and created_at are preserved (the row keeps its identity and
// original insertion time). summary/decisions/edge_cases/errors/embedding/
// model_used/updated_at are all replaced.
//
// Double-write: the embedding (when present and well-formed) is also mirrored
// into the vec_digests virtual table so SearchSimilar can use sqlite-vec's
// KNN index instead of an in-Go linear scan. Both writes happen in the same
// transaction — if vec_digests fails, the whole upsert rolls back.
func UpsertDigest(ctx context.Context, db *sql.DB, d Digest) error {
	decisions, edgeCases, errors, embBlob, err := marshalDigest(d)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	if d.CreatedAt.IsZero() {
		d.CreatedAt = now
	}
	d.UpdatedAt = now

	source := d.Source
	if source == "" {
		source = "auto"
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("memory: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const upsertQ = `INSERT INTO digests (id, run_id, project, summary, decisions, edge_cases, errors, embedding, model_used, source, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(run_id) DO UPDATE SET
			summary    = excluded.summary,
			decisions  = excluded.decisions,
			edge_cases = excluded.edge_cases,
			errors     = excluded.errors,
			embedding  = excluded.embedding,
			model_used = excluded.model_used,
			source     = excluded.source,
			updated_at = excluded.updated_at`

	if _, err := tx.ExecContext(ctx, upsertQ,
		d.ID, d.RunID, d.Project, d.Summary,
		string(decisions), string(edgeCases), string(errors),
		embBlob, d.ModelUsed, source,
		d.CreatedAt.Format(time.RFC3339Nano),
		d.UpdatedAt.Format(time.RFC3339Nano),
	); err != nil {
		return fmt.Errorf("memory: upsert digest: %w", err)
	}

	// Resolve the actual stored row (id may differ from d.ID if a previous
	// upsert won the conflict). The vec_digests entry must reference the
	// surviving digest_id, not the input one.
	var storedID string
	if err := tx.QueryRowContext(ctx, `SELECT id FROM digests WHERE run_id = ?`, d.RunID).Scan(&storedID); err != nil {
		return fmt.Errorf("memory: resolve stored digest id: %w", err)
	}

	if err := upsertVecDigest(ctx, tx, storedID, d.Project, embBlob); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("memory: commit upsert: %w", err)
	}
	return nil
}

// upsertVecDigest mirrors a digest's embedding into vec_digests. A nil or
// empty blob means "no embedding yet" — we delete any stale row instead of
// inserting garbage. sqlite-vec validates the byte length matches the column
// dimension at insert time, so callers don't need to pre-validate.
func upsertVecDigest(ctx context.Context, tx *sql.Tx, digestID, project string, embBlob []byte) error {
	if len(embBlob) == 0 {
		if _, err := tx.ExecContext(ctx, `DELETE FROM vec_digests WHERE digest_id = ?`, digestID); err != nil {
			return fmt.Errorf("memory: delete vec_digests for %q: %w", digestID, err)
		}
		return nil
	}

	// vec0 doesn't support ON CONFLICT, so we delete-then-insert. Both
	// statements run inside the caller's transaction.
	if _, err := tx.ExecContext(ctx, `DELETE FROM vec_digests WHERE digest_id = ?`, digestID); err != nil {
		return fmt.Errorf("memory: delete vec_digests for %q: %w", digestID, err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO vec_digests(digest_id, project, embedding) VALUES (?, ?, ?)`,
		digestID, project, embBlob,
	); err != nil {
		return fmt.Errorf("memory: insert vec_digests for %q: %w", digestID, err)
	}
	return nil
}

// marshalDigest encodes the JSON-array fields and embedding blob used by both
// SaveDigest and UpsertDigest.
func marshalDigest(d Digest) (decisions, edgeCases, errors []byte, embBlob []byte, err error) {
	decisions, err = json.Marshal(d.Decisions)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("memory: marshal decisions: %w", err)
	}
	edgeCases, err = json.Marshal(d.EdgeCases)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("memory: marshal edge_cases: %w", err)
	}
	errors, err = json.Marshal(d.Errors)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("memory: marshal errors: %w", err)
	}
	if d.Embedding != nil {
		embBlob = EncodeEmbedding(d.Embedding)
	}
	return decisions, edgeCases, errors, embBlob, nil
}

// GetDigestByRunID returns the digest for the given run.
// Returns (nil, nil) if not found.
func GetDigestByRunID(ctx context.Context, db *sql.DB, runID string) (*Digest, error) {
	const q = `SELECT id, run_id, project, summary, decisions, edge_cases, errors, embedding, model_used, source, created_at, updated_at
		FROM digests WHERE run_id = ? LIMIT 1`

	row := db.QueryRowContext(ctx, q, runID)
	d, err := scanDigest(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("memory: get digest by run %q: %w", runID, err)
	}
	return &d, nil
}

// ListDigestsByProject returns all digests for the given project, ordered by created_at DESC.
func ListDigestsByProject(ctx context.Context, db *sql.DB, project string) ([]Digest, error) {
	const q = `SELECT id, run_id, project, summary, decisions, edge_cases, errors, embedding, model_used, source, created_at, updated_at
		FROM digests WHERE project = ? ORDER BY created_at DESC`

	rows, err := db.QueryContext(ctx, q, project)
	if err != nil {
		return nil, fmt.Errorf("memory: list digests for project %q: %w", project, err)
	}
	defer rows.Close()

	var results []Digest
	for rows.Next() {
		d, err := scanDigest(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("memory: iterate digests: %w", err)
	}
	if results == nil {
		results = []Digest{}
	}
	return results, nil
}

// DeleteDigest removes a digest by ID, including its vec_digests mirror.
func DeleteDigest(ctx context.Context, db *sql.DB, id string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("memory: begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM vec_digests WHERE digest_id = ?`, id); err != nil {
		return fmt.Errorf("memory: delete vec_digests for %q: %w", id, err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM digests WHERE id = ?`, id); err != nil {
		return fmt.Errorf("memory: delete digest %q: %w", id, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("memory: commit delete: %w", err)
	}
	return nil
}

// scanDigest scans a single digest row from either *sql.Row or *sql.Rows.
func scanDigest(scanner interface{ Scan(dest ...any) error }) (Digest, error) {
	var (
		id        string
		runID     string
		project   string
		summary   string
		decisions string
		edgeCases string
		errors    string
		embBlob   []byte
		modelUsed string
		source    string
		createdAt string
		updatedAt string
	)

	if err := scanner.Scan(&id, &runID, &project, &summary, &decisions, &edgeCases, &errors, &embBlob, &modelUsed, &source, &createdAt, &updatedAt); err != nil {
		return Digest{}, err
	}

	d := Digest{
		ID:        id,
		RunID:     runID,
		Project:   project,
		Summary:   summary,
		ModelUsed: modelUsed,
		Source:    source,
	}

	_ = json.Unmarshal([]byte(decisions), &d.Decisions)
	_ = json.Unmarshal([]byte(edgeCases), &d.EdgeCases)
	_ = json.Unmarshal([]byte(errors), &d.Errors)

	if d.Decisions == nil {
		d.Decisions = []string{}
	}
	if d.EdgeCases == nil {
		d.EdgeCases = []string{}
	}
	if d.Errors == nil {
		d.Errors = []string{}
	}

	if embBlob != nil {
		emb, err := DecodeEmbedding(embBlob)
		if err != nil {
			// Corrupted blob: log-worthy but not fatal for reads.
			d.Embedding = nil
		} else {
			d.Embedding = emb
		}
	}

	d.CreatedAt = parseFlexTime(createdAt)
	d.UpdatedAt = parseFlexTime(updatedAt)

	return d, nil
}

// parseFlexTime parses a datetime string stored by either Go (RFC3339Nano,
// e.g. "2026-04-22T01:57:47Z") or SQLite's datetime() function
// (e.g. "2026-04-22 01:57:47"). Returns zero time if both formats fail.
func parseFlexTime(s string) time.Time {
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Time{}
}
