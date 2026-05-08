package capture

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/ernesto2108/anvil/internal/memory"
	"github.com/ernesto2108/anvil/internal/memory/transcript"
)

// Capture orchestrates the full pipeline for persisting a transcript as a
// memory digest:
//
//  1. ResolveProject — determine the project name.
//  2. ShouldCapture  — decide whether the session is worth saving.
//  3. Summarize      — call the LLM (or fallback to keywords-only).
//  4. SaveDigest     — persist the digest to SQLite.
//
// runID must be the run already linked to this session in the store (direct
// mode). Pass a synthetic run ID from EnsureSyntheticRun for manual captures.
//
// When summarizer is nil the fallback keywords-only path is used regardless
// of the noLLM flag.
func Capture(
	ctx context.Context,
	runID string,
	td transcript.TranscriptDigest,
	summarizer TranscriptSummarizer,
	db *sql.DB,
	projectOverride string,
	noLLM bool,
) error {
	project, err := ResolveProject(ctx, td, projectOverride, db)
	if err != nil {
		return fmt.Errorf("capture: %w", err)
	}

	ok, err := ShouldCapture(ctx, runID, td, modeFromRunID(runID), db)
	if err != nil {
		return fmt.Errorf("capture: should_capture: %w", err)
	}
	if !ok {
		return nil
	}

	var draft memory.DigestDraft
	if noLLM || summarizer == nil {
		draft = keywordsOnlyDigest(td)
	} else {
		d, sumErr := summarizer.SummarizeTranscript(ctx, td)
		if sumErr != nil {
			// Degraded path: LLM failed, fall back to keywords.
			draft = keywordsOnlyDigest(td)
		} else {
			draft = d
		}
	}

	digestID, err := newDigestID()
	if err != nil {
		return fmt.Errorf("capture: new digest id: %w", err)
	}

	now := time.Now().UTC()
	d := memory.Digest{
		ID:        digestID,
		RunID:     runID,
		Project:   project,
		Summary:   draft.Summary,
		Decisions: draft.Decisions,
		EdgeCases: draft.EdgeCases,
		Errors:    draft.Errors,
		ModelUsed: "transcript-capture",
		Source:    "direct",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := memory.SaveDigest(ctx, db, d); err != nil {
		return fmt.Errorf("capture: save digest: %w", err)
	}

	return nil
}

// modeFromRunID returns "manual" for synthetic manual run IDs and "direct"
// for everything else. The manual prefix is defined in EnsureSyntheticRun.
func modeFromRunID(runID string) string {
	if len(runID) >= 9 && runID[:9] == "d_manual_" {
		return "manual"
	}
	return "direct"
}

// newDigestID produces a fresh random digest ID with the standard `d_` prefix.
func newDigestID() (string, error) {
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "d_" + hex.EncodeToString(b[:]), nil
}
