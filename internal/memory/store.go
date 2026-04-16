package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// SaveDigest inserts a new digest into SQLite.
func SaveDigest(ctx context.Context, db *sql.DB, d Digest) error {
	decisions, err := json.Marshal(d.Decisions)
	if err != nil {
		return fmt.Errorf("memory: marshal decisions: %w", err)
	}
	edgeCases, err := json.Marshal(d.EdgeCases)
	if err != nil {
		return fmt.Errorf("memory: marshal edge_cases: %w", err)
	}
	errors, err := json.Marshal(d.Errors)
	if err != nil {
		return fmt.Errorf("memory: marshal errors: %w", err)
	}

	var embBlob []byte
	if d.Embedding != nil {
		embBlob = EncodeEmbedding(d.Embedding)
	}

	now := time.Now().UTC().Format(time.RFC3339Nano)
	if d.CreatedAt.IsZero() {
		d.CreatedAt, _ = time.Parse(time.RFC3339Nano, now)
	}
	if d.UpdatedAt.IsZero() {
		d.UpdatedAt = d.CreatedAt
	}

	const q = `INSERT INTO digests (id, run_id, project, summary, decisions, edge_cases, errors, embedding, model_used, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = db.ExecContext(ctx, q,
		d.ID, d.RunID, d.Project, d.Summary,
		string(decisions), string(edgeCases), string(errors),
		embBlob, d.ModelUsed,
		d.CreatedAt.Format(time.RFC3339Nano),
		d.UpdatedAt.Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("memory: insert digest: %w", err)
	}
	return nil
}

// GetDigestByRunID returns the digest for the given run.
// Returns (nil, nil) if not found.
func GetDigestByRunID(ctx context.Context, db *sql.DB, runID string) (*Digest, error) {
	const q = `SELECT id, run_id, project, summary, decisions, edge_cases, errors, embedding, model_used, created_at, updated_at
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
	const q = `SELECT id, run_id, project, summary, decisions, edge_cases, errors, embedding, model_used, created_at, updated_at
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

// DeleteDigest removes a digest by ID.
func DeleteDigest(ctx context.Context, db *sql.DB, id string) error {
	_, err := db.ExecContext(ctx, `DELETE FROM digests WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("memory: delete digest %q: %w", id, err)
	}
	return nil
}

// scanDigest scans a single digest row from either *sql.Row or *sql.Rows.
func scanDigest(scanner interface{ Scan(dest ...any) error }) (Digest, error) {
	var (
		id         string
		runID      string
		project    string
		summary    string
		decisions  string
		edgeCases  string
		errors     string
		embBlob    []byte
		modelUsed  string
		createdAt  string
		updatedAt  string
	)

	if err := scanner.Scan(&id, &runID, &project, &summary, &decisions, &edgeCases, &errors, &embBlob, &modelUsed, &createdAt, &updatedAt); err != nil {
		return Digest{}, err
	}

	d := Digest{
		ID:        id,
		RunID:     runID,
		Project:   project,
		Summary:   summary,
		ModelUsed: modelUsed,
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

	if t, err := time.Parse(time.RFC3339Nano, createdAt); err == nil {
		d.CreatedAt = t
	}
	if t, err := time.Parse(time.RFC3339Nano, updatedAt); err == nil {
		d.UpdatedAt = t
	}

	return d, nil
}
