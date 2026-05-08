package capture

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ernesto2108/anvil/internal/memory/transcript"
)

// ShouldCapture decides whether a transcript is worth persisting as a digest.
//
// mode "direct" — consults SQLite: if the run accumulated at least one
// tool_use or file_touch event it is worth summarising. This avoids creating
// digests for sessions that opened and closed without any real activity (D8).
//
// mode "manual" — always returns true; the human explicitly requested capture.
func ShouldCapture(ctx context.Context, runID string, _ transcript.TranscriptDigest, mode string, db *sql.DB) (bool, error) {
	if mode == "manual" {
		return true, nil
	}

	if runID == "" {
		return false, nil
	}

	const q = `
		SELECT
			(SELECT COUNT(*) FROM tool_uses WHERE run_id = ?) +
			(SELECT COUNT(*) FROM files      WHERE run_id = ?) AS total`

	var total int
	if err := db.QueryRowContext(ctx, q, runID, runID).Scan(&total); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("capture: should_capture query: %w", err)
	}

	return total > 0, nil
}
