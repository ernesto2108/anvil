package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ernesto2108/anvil/internal/memory/capture"
	"github.com/ernesto2108/anvil/internal/memory/haiku"
	"github.com/ernesto2108/anvil/internal/memory/transcript"
	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/output"
)

// captureFlags holds parsed flags for `anvil capture`.
type captureFlags struct {
	path    string
	project string
	noLLM   bool
	dryRun  bool
}

func parseCaptureFlags(args []string) (captureFlags, error) {
	f := captureFlags{}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--project", "-p":
			if i+1 >= len(args) {
				return f, fmt.Errorf("--project requires a value")
			}
			i++
			f.project = args[i]
		case "--no-llm":
			f.noLLM = true
		case "--dry-run":
			f.dryRun = true
		default:
			if f.path == "" {
				f.path = args[i]
			}
		}
	}
	if f.path == "" {
		return f, fmt.Errorf("missing transcript path")
	}
	return f, nil
}

// cmdCapture implements `anvil capture <transcript-path>`.
// It parses the JSONL transcript, creates a synthetic run, and persists a
// memory digest.
//
// Flags:
//
//	--project NAME  override the project name
//	--no-llm        skip Haiku and use keywords-only summarisation
//	--dry-run       print what would be captured without writing to the DB
func cmdCapture(_ *config.App, args []string) {
	flags, err := parseCaptureFlags(args)
	if err != nil {
		output.Error("usage: anvil capture <transcript-path> [--project NAME] [--no-llm] [--dry-run]")
		output.Error("  %s", err)
		os.Exit(1)
	}

	td, err := transcript.ParseFile(flags.path)
	if err != nil {
		output.Error("parse transcript: %s", err)
		os.Exit(1)
	}

	if flags.dryRun {
		printDryRun(td, flags.project)
		return
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

	ts := td.StartedAt
	if ts.IsZero() {
		ts = time.Now().UTC()
	}

	// Resolve project early so we can pass it to EnsureSyntheticRun.
	project, err := capture.ResolveProject(ctx, td, flags.project, db)
	if err != nil {
		output.Error("resolve project: %s", err)
		os.Exit(1)
	}

	runID, err := EnsureSyntheticRun(ctx, db, project, ts)
	if err != nil {
		output.Error("ensure synthetic run: %s", err)
		os.Exit(1)
	}

	var summarizer capture.TranscriptSummarizer
	if !flags.noLLM {
		if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
			summarizer = haiku.NewClient(apiKey, "")
		}
	}

	if err := capture.Capture(ctx, runID, td, summarizer, db, project, flags.noLLM); err != nil {
		output.Error("capture: %s", err)
		os.Exit(1)
	}

	output.Info("Digest captured from transcript: %s", flags.path)
	output.Info("  run_id:    %s", runID)
	output.Info("  project:   %s", project)
	output.Info("  turns:     %d", td.TurnCount)
	output.Info("  decisions: %d", len(td.Decisions))
	output.Info("  tools:     %d", len(td.ToolsUsed))
}

// printDryRun prints what would be captured without writing anything.
func printDryRun(td transcript.TranscriptDigest, projectOverride string) {
	project := projectOverride
	if project == "" && td.Cwd != "" {
		project = filepath.Base(td.Cwd)
	}
	output.Info("[dry-run] Would capture:")
	output.Info("  session_id: %s", td.SessionID)
	output.Info("  project:    %s", project)
	output.Info("  cwd:        %s", td.Cwd)
	output.Info("  turns:      %d", td.TurnCount)
	output.Info("  tools:      %v", td.ToolsUsed)
	output.Info("  decisions:  %d", len(td.Decisions))
	output.Info("  has_errors: %v", td.HasErrors)
	if !td.StartedAt.IsZero() {
		output.Info("  started_at: %s", td.StartedAt.Format(time.RFC3339))
	}
	if !td.EndedAt.IsZero() {
		output.Info("  ended_at:   %s", td.EndedAt.Format(time.RFC3339))
	}
}
