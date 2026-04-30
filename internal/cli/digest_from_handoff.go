package cli

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ernesto2108/anvil/internal/memory"
	"github.com/ernesto2108/anvil/internal/memory/ollama"
	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/output"
)

// digestFromHandoffFlags holds parsed flags for `anvil digests from-handoff`.
type digestFromHandoffFlags struct {
	path    string
	project string
}

func parseDigestFromHandoffFlags(args []string) (digestFromHandoffFlags, error) {
	f := digestFromHandoffFlags{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project", "-p":
			if i+1 >= len(args) {
				return f, fmt.Errorf("--project requires a value")
			}
			i++
			f.project = args[i]
		default:
			if f.path == "" {
				f.path = args[i]
			}
		}
	}
	if f.path == "" {
		return f, fmt.Errorf("missing handoff file path")
	}
	return f, nil
}

// cmdDigestFromHandoff parses a handoff markdown file and writes its content as
// a digest to ~/.anvil/runs.db. Used by /task-complete and the backfill script
// to populate memory from handoffs that were never run through the CLI runner.
func cmdDigestFromHandoff(_ *config.App, args []string) {
	flags, err := parseDigestFromHandoffFlags(args)
	if err != nil {
		output.Error("usage: anvil digests from-handoff <path> [--project NAME]")
		output.Error("  %s", err)
		os.Exit(1)
	}

	if flags.project == "" {
		workDir, _ := os.Getwd()
		flags.project = filepath.Base(workDir)
	}

	content, err := os.ReadFile(flags.path)
	if err != nil {
		output.Error("read handoff: %s", err)
		os.Exit(1)
	}

	draft, taskID, err := memory.ParseHandoff(string(content), flags.path)
	if err != nil {
		output.Error("parse handoff: %s", err)
		os.Exit(1)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		output.Error("resolve home: %s", err)
		os.Exit(1)
	}
	dbPath := filepath.Join(home, ".anvil", "runs.db")

	db, err := openDashboardDB(dbPath, false)
	if err != nil {
		output.Error("open store: %s", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	ctx := context.Background()
	runID := handoffRunID(taskID)
	modelUsed := "handoff-parser"

	if err := upsertSyntheticRun(ctx, db, runID, taskID, draft.Summary, flags.project); err != nil {
		output.Error("upsert synthetic run: %s", err)
		os.Exit(1)
	}

	// Auto-start Ollama if needed. KEEP_ALIVE=5m ensures models unload after
	// idle so the laptop is not held down by a loaded model when not in use.
	if err := ollama.EnsureRunning(ctx, "5m"); err != nil {
		output.Warn("ollama: %s (saving digest without embedding)", err)
	}

	// Best-effort embedding via Ollama. If the daemon is still unreachable
	// after EnsureRunning, save without embedding — search_memories has a
	// recent-digests fallback that handles unembedded rows.
	var embedding []float32
	embedder := ollama.NewClient("", "")
	if embedder.Healthy(ctx) {
		emb, embErr := embedder.Embed(ctx, draft.Summary)
		if embErr != nil {
			output.Warn("embed handoff digest: %s (saving without embedding)", embErr)
		} else {
			embedding = emb
		}
	}

	digestID, _ := newDigestIDForHandoff()
	now := time.Now().UTC()
	d := memory.Digest{
		ID:        digestID,
		RunID:     runID,
		Project:   flags.project,
		Summary:   draft.Summary,
		Decisions: draft.Decisions,
		EdgeCases: draft.EdgeCases,
		Errors:    draft.Errors,
		Embedding: embedding,
		ModelUsed: modelUsed,
		Source:    "handoff",
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := memory.UpsertDigest(ctx, db, d); err != nil {
		output.Error("upsert digest: %s", err)
		os.Exit(1)
	}

	output.Info("Digest saved (handoff): %s", taskID)
	output.Info("  run_id:    %s", runID)
	output.Info("  project:   %s", flags.project)
	output.Info("  decisions: %d", len(draft.Decisions))
	output.Info("  edges:     %d", len(draft.EdgeCases))
	output.Info("  errors:    %d", len(draft.Errors))
	if embedding == nil {
		output.Info("  embedding: absent")
	} else {
		output.Info("  embedding: %d dims", len(embedding))
	}
}

// handoffRunID returns a deterministic synthetic run_id for a given TASK-ID so
// that re-running `from-handoff` on the same handoff overwrites the digest in
// place (UPSERT semantics) instead of producing duplicates.
func handoffRunID(taskID string) string {
	return "handoff_" + taskID
}

// upsertSyntheticRun inserts a placeholder run row so the digest's FK to runs
// is satisfied. If the row already exists (re-run on same handoff), updates the
// ended_at timestamp only — preserves the original started_at as the first time
// the handoff was ingested.
func upsertSyntheticRun(ctx context.Context, db *sql.DB, runID, taskID, summary, project string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)

	const insertQ = `INSERT INTO runs (id, task_id, task_desc, status, started_at, ended_at, project, task_source)
		VALUES (?, ?, ?, 'success', ?, ?, ?, 'handoff')
		ON CONFLICT(id) DO UPDATE SET
			task_desc = excluded.task_desc,
			ended_at  = excluded.ended_at,
			project   = excluded.project`

	taskDesc := summary
	if len(taskDesc) > 500 {
		taskDesc = taskDesc[:497] + "..."
	}

	_, err := db.ExecContext(ctx, insertQ, runID, taskID, taskDesc, now, now, project)
	if err != nil {
		return fmt.Errorf("insert synthetic run: %w", err)
	}
	return nil
}

// newDigestIDForHandoff produces a digest id distinguishable from auto-generated
// ones (`d_<random>`). Format: `dh_<random>` so a quick eyeball of the table
// reveals which digests came from handoffs vs the runner.
func newDigestIDForHandoff() (string, error) {
	var b [6]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "dh_" + hex.EncodeToString(b[:]), nil
}
