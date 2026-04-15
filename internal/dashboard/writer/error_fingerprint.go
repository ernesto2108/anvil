package writer

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// --- Fingerprinting ---

// normalizeErrorMessage strips variable parts (UUIDs, paths, line numbers, timestamps)
// to group similar errors under the same fingerprint.
func normalizeErrorMessage(msg string) string {
	s := msg
	// Replace UUIDs
	s = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`).ReplaceAllString(s, "*")
	// Replace file paths (/Users/... or /home/...)
	s = regexp.MustCompile(`/[\w./-]+`).ReplaceAllString(s, "*")
	// Replace line numbers (e.g. :123, line 42)
	s = regexp.MustCompile(`:\d+`).ReplaceAllString(s, ":*")
	s = regexp.MustCompile(`(?i)line \d+`).ReplaceAllString(s, "line *")
	// Replace standalone numbers (IDs, exit codes in context)
	s = regexp.MustCompile(`\b\d{4,}\b`).ReplaceAllString(s, "*")
	// Replace timestamps
	s = regexp.MustCompile(`\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}[^\s]*`).ReplaceAllString(s, "*")
	// Collapse multiple asterisks
	s = regexp.MustCompile(`\*+`).ReplaceAllString(s, "*")
	return strings.TrimSpace(s)
}

// computeFingerprint returns the first 16 hex chars of the SHA-256 of the normalized message.
func computeFingerprint(msg string) string {
	normalized := normalizeErrorMessage(msg)
	h := sha256.Sum256([]byte(normalized))
	return fmt.Sprintf("%x", h[:8])
}

// --- Write operations ---

// handleErrorFingerprint extracts error details from the event payload and
// upserts the corresponding error group.
func handleErrorFingerprint(w *EventWriter, tx *sql.Tx, ev instrumentation.Event) error {
	var errorMsg, agentName string
	var exitCode *int

	switch ev.EventType {
	case instrumentation.EventAgentError:
		var p instrumentation.AgentErrorPayload
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return fmt.Errorf("dashboard/writer: deserializar AgentErrorPayload (fingerprint): %w", err)
		}
		errorMsg = p.Error
		agentName = p.AgentID
	case instrumentation.EventRunError:
		var p instrumentation.RunErrorPayload
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			return fmt.Errorf("dashboard/writer: deserializar RunErrorPayload (fingerprint): %w", err)
		}
		errorMsg = p.ErrorReason
	}

	if errorMsg == "" {
		return nil
	}

	return w.UpsertErrorGroup(tx, ev.RunID, agentName, errorMsg, exitCode, ev.Timestamp)
}

// UpsertErrorGroup creates or updates an error group based on the error message fingerprint.
// It also inserts a row into error_group_runs to link the error to the run.
// If a resolved/ignored error reappears, its status is reset to "new".
func (w *EventWriter) UpsertErrorGroup(tx *sql.Tx, runID, agentName, errorMsg string, exitCode *int, occurredAt time.Time) error {
	fp := computeFingerprint(errorMsg)
	normalized := normalizeErrorMessage(errorMsg)
	now := occurredAt.Format(time.RFC3339Nano)

	// Try to find existing group
	var groupID string
	var currentStatus string
	err := tx.QueryRow(`SELECT id, resolution_status FROM error_groups WHERE fingerprint = ?`, fp).Scan(&groupID, &currentStatus)

	if err == sql.ErrNoRows {
		// New error group
		groupID = fmt.Sprintf("eg-%s", fp)
		const insertGroup = `
			INSERT INTO error_groups (id, fingerprint, title, normalized_msg, resolution_status, first_seen_at, last_seen_at, occurrence_count, created_at, updated_at)
			VALUES (?, ?, ?, ?, 'new', ?, ?, 1, ?, ?)`
		if _, err := tx.Exec(insertGroup, groupID, fp, errorMsg, normalized, now, now, now, now); err != nil {
			return fmt.Errorf("dashboard/writer: insertar error_groups: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("dashboard/writer: buscar error_groups por fingerprint: %w", err)
	} else {
		// Existing group — update count and last_seen
		newStatus := currentStatus
		if currentStatus == "resolved" || currentStatus == "ignored" {
			newStatus = "new" // regression detected
		}
		const updateGroup = `
			UPDATE error_groups
			SET occurrence_count = occurrence_count + 1,
			    last_seen_at = ?,
			    resolution_status = ?,
			    updated_at = ?
			WHERE id = ?`
		if _, err := tx.Exec(updateGroup, now, newStatus, now, groupID); err != nil {
			return fmt.Errorf("dashboard/writer: actualizar error_groups: %w", err)
		}

		// If status changed due to regression, log it
		if newStatus != currentStatus {
			const insertHistory = `
				INSERT INTO error_group_history (error_group_id, old_status, new_status, note, created_at)
				VALUES (?, ?, ?, 'Regression detected: error reappeared in a new run', ?)`
			if _, err := tx.Exec(insertHistory, groupID, currentStatus, newStatus, now); err != nil {
				return fmt.Errorf("dashboard/writer: insertar error_group_history (regression): %w", err)
			}
		}
	}

	// Link error to run (ignore duplicate)
	const insertLink = `
		INSERT OR IGNORE INTO error_group_runs (error_group_id, run_id, agent_name, error_msg, exit_code, occurred_at)
		VALUES (?, ?, ?, ?, ?, ?)`
	if _, err := tx.Exec(insertLink, groupID, runID, agentName, errorMsg, exitCode, now); err != nil {
		return fmt.Errorf("dashboard/writer: insertar error_group_runs: %w", err)
	}

	return nil
}

// UpdateErrorResolution updates the resolution status, notes, and commit link for an error group.
func (w *EventWriter) UpdateErrorResolution(ctx context.Context, groupID, status, notes, commitLink string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)

	// Get current status for history
	var oldStatus string
	err := w.db.QueryRowContext(ctx, `SELECT resolution_status FROM error_groups WHERE id = ?`, groupID).Scan(&oldStatus)
	if err != nil {
		return fmt.Errorf("dashboard/writer: buscar error_groups para update: %w", err)
	}

	const q = `
		UPDATE error_groups
		SET resolution_status = ?,
		    notes = CASE WHEN ? != '' THEN ? ELSE notes END,
		    commit_link = CASE WHEN ? != '' THEN ? ELSE commit_link END,
		    updated_at = ?
		WHERE id = ?`
	res, err := w.db.ExecContext(ctx, q, status, notes, notes, commitLink, commitLink, now, groupID)
	if err != nil {
		return fmt.Errorf("dashboard/writer: actualizar error_groups resolución: %w", err)
	}
	if err := expectRowsAffected(res, "error_groups", groupID); err != nil {
		return err
	}

	// Insert history entry
	const insertHistory = `
		INSERT INTO error_group_history (error_group_id, old_status, new_status, note, commit_link, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`
	if _, err := w.db.ExecContext(ctx, insertHistory, groupID, oldStatus, status, notes, commitLink, now); err != nil {
		return fmt.Errorf("dashboard/writer: insertar error_group_history: %w", err)
	}

	return nil
}
