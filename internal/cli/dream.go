package cli

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ernesto2108/anvil/internal/memory"
	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/output"
)

// cmdDream queries recent digests for the current project, asks the configured
// summarizer to extract recurring patterns, and stores the resulting insights
// in the dream_insights table with status=pending_review.
func cmdDream(cfg *config.App, args []string) {
	ctx := context.Background()

	// Parse optional --days and --retries flags (defaults: 7, 3).
	days := 7
	maxRetries := 3
	for i := 0; i < len(args)-1; i++ {
		switch args[i] {
		case "--days":
			fmt.Sscanf(args[i+1], "%d", &days) //nolint:errcheck
		case "--retries":
			fmt.Sscanf(args[i+1], "%d", &maxRetries) //nolint:errcheck
		}
	}

	workDir, _ := os.Getwd()
	project := filepath.Base(workDir)

	output.Info("Dreaming: starting for project %q (last %d days, max %d retries).", project, days, maxRetries)

	home, err := os.UserHomeDir()
	if err != nil {
		output.Error("resolve home dir: %s", err)
		os.Exit(1)
	}
	dbPath := filepath.Join(home, ".anvil", "runs.db")
	db, err := openDashboardDB(dbPath, false)
	if err != nil {
		output.Error("open store: %s", err)
		os.Exit(1)
	}
	defer func() { _ = db.Close() }()

	// Load recent digests.
	cutoff := time.Now().UTC().AddDate(0, 0, -days).Format(time.RFC3339)
	rows, err := db.QueryContext(ctx,
		`SELECT run_id, summary, decisions, edge_cases, errors
		 FROM digests
		 WHERE project=? AND created_at >= ?
		 ORDER BY created_at DESC LIMIT 20`,
		project, cutoff)
	if err != nil {
		output.Error("query digests: %s", err)
		os.Exit(1)
	}
	defer func() { _ = rows.Close() }()

	type digestRow struct {
		RunID     string
		Summary   string
		Decisions string
		EdgeCases string
		Errors    string
	}
	var digests []digestRow
	for rows.Next() {
		var d digestRow
		if scanErr := rows.Scan(&d.RunID, &d.Summary, &d.Decisions, &d.EdgeCases, &d.Errors); scanErr != nil {
			continue
		}
		digests = append(digests, d)
	}

	if len(digests) == 0 {
		output.Info("Dreaming: no digests found for project %q in last %d days.", project, days)
		return
	}

	output.Info("Dreaming: analysing %d digest(s) from last %d days...", len(digests), days)

	// Build text blob for the summarizer.
	var sb strings.Builder
	var sourceRuns []string
	for _, d := range digests {
		sourceRuns = append(sourceRuns, d.RunID)
		fmt.Fprintf(&sb, "## Run %s\n%s\nDecisions: %s\nEdge cases: %s\nErrors: %s\n\n",
			d.RunID, d.Summary, d.Decisions, d.EdgeCases, d.Errors)
	}

	// Pick a summarizer.
	probeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	summarizer, _, ok := pickSummarizer(probeCtx)
	cancel()
	if !ok {
		output.Error("Dreaming: no summarizer available (set ANTHROPIC_API_KEY or start Ollama).")
		os.Exit(1)
	}

	// Wrap blob as a fake AgentOutput so we can reuse the Summarizer interface.
	agentOutputs := []memory.AgentOutput{
		{
			Role: "dream-reflection",
			Output: "You are an Anvil dreaming process. Extract 3-7 recurring patterns, " +
				"repeated errors, and actionable insights from the following run digests. " +
				"Format each insight as a single clear sentence.\n\n" + sb.String(),
		},
	}

	var draft memory.DigestDraft
	for attempt := 1; ; attempt++ {
		draft, err = summarizer.Summarize(ctx, agentOutputs)
		if err == nil {
			break
		}
		if attempt >= maxRetries {
			output.Error("Dreaming: summarize failed after %d attempt(s): %s", attempt, err)
			os.Exit(1)
		}
		wait := time.Duration(1<<uint(attempt-1)) * time.Second
		output.Warn("Dreaming: summarize attempt %d failed (%s), retrying in %s...", attempt, err, wait)
		time.Sleep(wait)
	}

	// Persist each decision as a separate insight.
	insights := append(draft.Decisions, draft.EdgeCases...)
	if len(insights) == 0 {
		// Fallback: use the summary itself as a single insight.
		insights = []string{draft.Summary}
	}

	sourceJSON, _ := json.Marshal(sourceRuns)
	now := time.Now().UTC().Format(time.RFC3339)
	inserted := 0
	for _, ins := range insights {
		ins = strings.TrimSpace(ins)
		if ins == "" {
			continue
		}
		id := "di_" + mustHex(8)
		_, err := db.ExecContext(ctx,
			`INSERT INTO dream_insights (id, project, insight, source_runs, status, created_at)
			 VALUES (?, ?, ?, ?, 'pending_review', ?)`,
			id, project, ins, string(sourceJSON), now)
		if err != nil {
			output.Warn("insert insight: %s", err)
			continue
		}
		inserted++
	}

	output.Info("Dreaming: %d insight(s) saved with status=pending_review.", inserted)
	output.Info("Review with: anvil digests --project %s  (dashboard coming in Phase 3)", project)
}

func mustHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
