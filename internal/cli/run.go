package cli

import (
	"bufio"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/writer"
	"github.com/ernesto2108/anvil/internal/instrumentation"
	"github.com/ernesto2108/anvil/internal/memory"
	"github.com/ernesto2108/anvil/internal/memory/haiku"
	"github.com/ernesto2108/anvil/internal/memory/ollama"
	"github.com/ernesto2108/anvil/internal/orchestrator"
	"github.com/ernesto2108/anvil/internal/runner"
	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/output"
)

// storeEventSink adapts store.SQLiteStore (which implements EventWriter with
// WriteEvent(Event) error) to the orchestrator.EventSink interface (Emit(Event)).
type storeEventSink struct {
	w instrumentation.EventWriter
}

func (s *storeEventSink) Emit(ev instrumentation.Event) {
	if err := s.w.WriteEvent(ev); err != nil {
		output.Warn("event write failed: %s", err)
	}
}

// runFlags holds the parsed flags shared by all run commands.
type runFlags struct {
	task         string
	model        string
	concurrency  int
	autoApprove  bool
	forceMigrate bool
}

// parseRunFlags extracts --task, --model, --concurrency, -y, --force-migrate from args.
func parseRunFlags(args []string) (runFlags, []string) {
	f := runFlags{concurrency: 4}
	var rest []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--task", "-t":
			if i+1 < len(args) {
				i++
				f.task = args[i]
			}
		case "--model", "-m":
			if i+1 < len(args) {
				i++
				f.model = args[i]
			}
		case "--concurrency", "-c":
			if i+1 < len(args) {
				i++
				fmt.Sscanf(args[i], "%d", &f.concurrency) //nolint:errcheck
			}
		case "--auto-approve", "-y":
			f.autoApprove = true
		case "--force-migrate":
			f.forceMigrate = true
		default:
			rest = append(rest, args[i])
		}
	}
	return f, rest
}

func cmdRun(cfg *config.App, args []string) {
	flags, rest := parseRunFlags(args)

	if len(rest) == 0 {
		output.Error("usage: anvil run <pipeline.yaml> --task \"description\" [--model name] [-y]")
		os.Exit(1)
	}

	if flags.task == "" {
		output.Error("--task is required")
		os.Exit(1)
	}

	// Load pipeline: if arg looks like a path use it directly, otherwise resolve by name.
	pipelineRef := rest[0]
	var nodes []orchestrator.Node
	var err error
	if filepath.Ext(pipelineRef) == ".yaml" || filepath.Ext(pipelineRef) == ".yml" {
		nodes, err = loadPipeline(pipelineRef)
	} else {
		nodes, err = resolvePipeline(cfg, pipelineRef)
	}
	if err != nil {
		output.Error("%s", err)
		os.Exit(1)
	}

	executeRun(cfg, nodes, flags)
}

// checkClaudeAuth verifies that `claude` CLI is authenticated before running a pipeline.
//
// The probe sets ANVIL_SKIP_EMIT=1 so the global Claude Code hooks (which call
// `anvil emit`) bypass writing the throwaway "ping" session into the dashboard.
func checkClaudeAuth() {
	cmd := exec.Command("claude", "--print", "-p", "ping")
	cmd.Env = append(os.Environ(), "ANVIL_SKIP_EMIT=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if strings.Contains(msg, "Not logged in") || strings.Contains(msg, "/login") {
			output.Error("Claude CLI not authenticated. Run 'claude' first and log in, then retry.")
		} else {
			output.Error("Claude CLI check failed: %s", msg)
		}
		os.Exit(1)
	}
}

func executeRun(cfg *config.App, nodes []orchestrator.Node, flags runFlags) {
	// 1. Preflight: ensure claude is authenticated.
	checkClaudeAuth()

	// 1b. Preflight: gather task metadata interactively.
	tc := preflight(flags.task, flags.autoApprove)

	// 2. Build DAG.
	dag, err := orchestrator.Build(nodes)
	if err != nil {
		output.Error("build DAG: %s", err)
		os.Exit(1)
	}

	output.Info("Pipeline loaded: %d nodes, order: [%s]", len(dag.Order), strings.Join(dag.Order, " → "))

	// 3. Generate run ID.
	runID, err := instrumentation.NewRunID()
	if err != nil {
		output.Error("generate run ID: %s", err)
		os.Exit(1)
	}

	// 4. Open store for event persistence.
	home, err := os.UserHomeDir()
	if err != nil {
		output.Error("resolve home dir: %s", err)
		os.Exit(1)
	}
	dbPath := filepath.Join(home, ".anvil", "runs.db")
	db, err := openDashboardDB(dbPath, flags.forceMigrate)
	if err != nil {
		output.Error("open store: %s", err)
		os.Exit(1)
	}
	s := writer.New(db, 0)
	defer func() { _ = s.Close() }()

	// Set busy_timeout for concurrent access with dashboard.
	if _, err := s.DB().Exec("PRAGMA busy_timeout=5000"); err != nil {
		output.Warn("set busy_timeout: %s", err)
	}

	// 5. Emit run.start event so the dashboard picks it up.
	workDir, _ := os.Getwd()
	agentIDs := make([]string, len(nodes))
	for i, n := range nodes {
		agentIDs[i] = n.ID
	}
	if ev, err := instrumentation.NewEvent(runID, instrumentation.EventRunStart,
		instrumentation.RunStartPayload{
			TaskDescription: flags.task,
			Complexity:      "pipeline",
			AgentsPlanned:   agentIDs,
			Provider:        cfg.ActiveProvider(),
			TriggeredBy:     "anvil-run",
			Project:         filepath.Base(workDir),
		},
	); err == nil {
		if writeErr := s.WriteEvent(ev); writeErr != nil {
			output.Warn("write run.start event: %s", writeErr)
		}
	}

	// 6. Ensure Ollama is running (auto-start if needed, models unload after 5m idle).
	project := filepath.Base(workDir)
	if err := ollama.EnsureRunning(context.Background(), "5m"); err != nil {
		output.Warn("ollama: %s (memory features disabled)", err)
	}

	// 6b. Pre-run memory injection: search for relevant digests from previous runs.
	// Always emit a status line so the user can tell why memory was/wasn't loaded
	// instead of guessing from silence.
	const memoryThreshold = 0.35
	ollamaClient := ollama.NewClient("", "")
	var memoryCtx string
	switch {
	case !ollamaClient.Healthy(context.Background()):
		output.Warn("Memory: Ollama unreachable, skipping memory injection")
	default:
		results, err := memory.SearchSimilar(context.Background(), s.DB(), ollamaClient, flags.task, project, 3, memoryThreshold)
		switch {
		case err != nil:
			output.Warn("Memory: search failed: %s", err)
		case len(results) == 0:
			output.Info("Memory: no relevant digests above threshold %.2f", memoryThreshold)
		default:
			memoryCtx = formatMemoryContext(results)
			output.Info("Memory: injecting %d digest(s) from previous runs:", len(results))
			for _, r := range results {
				output.Info("  - %s (score %.2f) %s", r.Digest.RunID, r.Score, truncate(r.Digest.Summary, 70))
			}
		}
	}

	// 7. Wire components.
	agentRunner := runner.New(workDir, flags.model, runID, runner.TaskContext{
		Objective:     tc.Objective,
		Stack:         tc.Stack,
		Files:         tc.Files,
		Complexity:    tc.Complexity,
		MemoryContext: memoryCtx,
	})

	// Inject per-agent MCP configs if declared in anvil.yaml mcp.servers.
	if len(cfg.Manifest.MCP.Servers) > 0 {
		defs := make(map[string]map[string]runner.MCPEntry)
		for _, node := range nodes {
			agentServers := cfg.BuildAgentMCPConfig(node.Role)
			if len(agentServers) == 0 {
				continue
			}
			entries := make(map[string]runner.MCPEntry, len(agentServers))
			for name, def := range agentServers {
				entries[name] = runner.MCPEntry{
					Command: def.Command,
					Args:    def.Args,
					Env:     def.Env,
				}
			}
			defs[node.Role] = entries
		}
		if mcpByRole, err := runner.BuildMCPByRole(defs); err != nil {
			output.Warn("mcp config build failed (agents will run without MCP injection): %s", err)
		} else {
			agentRunner.MCPByRole = mcpByRole
		}
	}

	var gate orchestrator.GateHandler
	if flags.autoApprove {
		gate = orchestrator.NewAutoApproveHandler(os.Stdout)
	} else {
		gate = orchestrator.NewCLIGateHandler(os.Stdin, os.Stdout)
	}

	sink := &storeEventSink{w: s}

	// 7. Handle Ctrl+C gracefully. Created BEFORE the checkpointer so the
	// checkpointer can derive its in-flight context from the same signal
	// source — a Ctrl+C cancels any checkpoint summarizer call immediately
	// instead of blocking the shutdown for the per-checkpoint timeout.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// 6c. Wrap the sink with an incremental digest checkpointer. On each
	// agent.end event, a cumulative digest is upserted so the run has a
	// recovery point even if the conversation or process dies mid-pipeline.
	// When no summarizer is available the wrapper is a pass-through.
	probeCtx, probeCancel := context.WithTimeout(context.Background(), 10*time.Second)
	checkpointSummarizer, checkpointModel, checkpointOK := pickSummarizer(probeCtx)
	probeCancel()
	var agentSink orchestrator.EventSink = sink
	var checkpointer *digestCheckpointer
	if checkpointOK {
		checkpointer = newDigestCheckpointer(ctx, sink, s.DB(), checkpointSummarizer, ollamaClient, checkpointModel, runID, project)
		agentSink = checkpointer
		output.Info("Checkpoints: enabled (%s)", checkpointModel)
	} else {
		output.Info("Checkpoints: disabled (no summarizer available)")
	}

	exec := orchestrator.New(agentRunner, gate, agentSink, flags.concurrency)

	// 8. Execute.
	output.Info("Starting run %s", output.Cyan(runID))
	fmt.Println()

	state := exec.Execute(ctx, dag, runID)

	// 8b. Drain in-flight checkpoint work before the terminal digest runs.
	// Without this, a late-arriving checkpoint could overwrite the terminal
	// UPSERT with stale content.
	if checkpointer != nil {
		checkpointer.Wait()
	}

	// 10. Emit run.end event.
	var completed, failed int
	for _, status := range state.Statuses {
		switch status {
		case orchestrator.NodeSuccess:
			completed++
		case orchestrator.NodeFailed:
			failed++
		}
	}
	durationMs := state.FinishedAt.Sub(state.StartedAt).Milliseconds()

	if ev, err := instrumentation.NewEvent(runID, instrumentation.EventRunEnd,
		instrumentation.RunEndPayload{
			Status:          state.FinalStatus,
			DurationMs:      durationMs,
			AgentsCompleted: completed,
			AgentsFailed:    failed,
		},
	); err == nil {
		if writeErr := s.WriteEvent(ev); writeErr != nil {
			output.Warn("write run.end event: %s", writeErr)
		}
	}

	// 11. Post-run digest generation.
	if state.FinalStatus == "success" || state.FinalStatus == "partial" {
		generateDigest(ctx, s.DB(), ollamaClient, runID, project, flags.autoApprove)
	}

	// 12. Print summary.
	fmt.Println()
	statusColor := output.Green
	if state.FinalStatus != "success" {
		statusColor = output.Red
	}
	output.Info("Run %s finished: %s (%dms)", runID, statusColor(state.FinalStatus), durationMs)

	for _, nodeID := range dag.Order {
		status := state.Statuses[nodeID]
		icon := "✓"
		colorFn := output.Green
		switch status {
		case orchestrator.NodeFailed:
			icon = "✗"
			colorFn = output.Red
		case orchestrator.NodeSkipped:
			icon = "○"
			colorFn = output.Yellow
		}
		fmt.Printf("   %s %s → %s\n", colorFn(icon), nodeID, status)
	}

	if state.FinalStatus != "success" {
		os.Exit(1)
	}
}

// pickSummarizer selects the summarizer used for digest generation.
// Haiku (remote) is preferred if ANTHROPIC_API_KEY is set; otherwise Ollama
// (local) is used if healthy. Returns (nil, "", false) when neither is
// available.
func pickSummarizer(ctx context.Context) (memory.Summarizer, string, bool) {
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		client := haiku.NewClient(apiKey, "")
		return client, client.Model(), true
	}
	s := ollama.NewSummarizer("", "")
	if !ollama.NewClient("", "").Healthy(ctx) {
		return nil, "", false
	}
	return s, s.Model(), true
}

// summarizeAndUpsert summarizes the current agent outputs for runID and
// upserts the resulting digest, overwriting any previous checkpoint for the
// same run. No user interaction. Returns (nil, nil) when there is nothing to
// summarize (no agent outputs yet).
//
// Used by both the incremental checkpointer (after each agent.end) and the
// terminal digest path (after run.end), so the cumulative digest of the last
// checkpoint is always consistent with what the user approves at the end.
func summarizeAndUpsert(
	ctx context.Context,
	db *sql.DB,
	summarizer memory.Summarizer,
	embedder memory.Embedder,
	modelUsed, runID, project string,
) (*memory.DigestDraft, error) {
	agentOutputs, err := loadAgentOutputs(ctx, db, runID)
	if err != nil {
		return nil, fmt.Errorf("load agent outputs: %w", err)
	}
	if len(agentOutputs) == 0 {
		return nil, nil
	}

	draft, err := summarizer.Summarize(ctx, agentOutputs)
	if err != nil {
		return nil, fmt.Errorf("summarize: %w", err)
	}

	var embedding []float32
	if embedder != nil {
		emb, embErr := embedder.Embed(ctx, draft.Summary)
		if embErr != nil {
			output.Warn("embed digest: %s", embErr)
			// Continue without embedding — the digest is still searchable by project.
		} else {
			embedding = emb
		}
	}

	// Reuse existing digest id when upserting so row identity is preserved
	// across checkpoints. If no row exists, mint a fresh id.
	existing, err := memory.GetDigestByRunID(ctx, db, runID)
	if err != nil {
		return nil, fmt.Errorf("lookup existing digest: %w", err)
	}
	digestID := ""
	var createdAt time.Time
	if existing != nil {
		digestID = existing.ID
		createdAt = existing.CreatedAt
	} else {
		digestID, _ = newDigestID()
		createdAt = time.Now().UTC()
	}

	d := memory.Digest{
		ID:        digestID,
		RunID:     runID,
		Project:   project,
		Summary:   draft.Summary,
		Decisions: draft.Decisions,
		EdgeCases: draft.EdgeCases,
		Errors:    draft.Errors,
		Embedding: embedding,
		ModelUsed: modelUsed,
		CreatedAt: createdAt,
	}

	if err := memory.UpsertDigest(ctx, db, d); err != nil {
		return nil, fmt.Errorf("upsert digest: %w", err)
	}
	return &draft, nil
}

// generateDigest creates (or refines) the run-level digest after all agents
// finish. Uses Haiku if ANTHROPIC_API_KEY is set, otherwise falls back to
// Ollama (mistral). This runs AFTER the incremental checkpointer has already
// persisted per-agent checkpoints, so if the user rejects at the approval
// gate the last checkpoint still survives — work is never lost.
func generateDigest(ctx context.Context, db *sql.DB, embedder memory.Embedder, runID, project string, autoApprove bool) {
	summarizer, modelUsed, ok := pickSummarizer(ctx)
	if !ok {
		output.Warn("neither ANTHROPIC_API_KEY nor Ollama available, skipping terminal digest")
		return
	}

	// Summarize + upsert first so the on-disk checkpoint reflects the full run.
	// If the user then edits or rejects, we either re-upsert (edit) or leave
	// the current state as-is (reject) — the last checkpoint is preserved.
	draft, err := summarizeAndUpsert(ctx, db, summarizer, embedder, modelUsed, runID, project)
	if err != nil {
		output.Warn("terminal digest: %s", err)
		return
	}
	if draft == nil {
		// No agent outputs — nothing to digest.
		return
	}

	if autoApprove {
		output.Info("Digest saved (terminal, %s)", modelUsed)
		return
	}

	action := showDigestApproval(*draft)
	switch action {
	case "no":
		output.Info("Terminal digest skipped (incremental checkpoints preserved)")
		return
	case "edit":
		edited := editDigestSummary(draft.Summary)
		if edited != draft.Summary {
			draft.Summary = edited
			// Re-upsert with edited summary. Re-embed the new text.
			var embedding []float32
			if embedder != nil {
				if emb, embErr := embedder.Embed(ctx, draft.Summary); embErr == nil {
					embedding = emb
				} else {
					output.Warn("embed edited digest: %s", embErr)
				}
			}
			existing, _ := memory.GetDigestByRunID(ctx, db, runID)
			digestID := ""
			var createdAt time.Time
			if existing != nil {
				digestID = existing.ID
				createdAt = existing.CreatedAt
			} else {
				digestID, _ = newDigestID()
				createdAt = time.Now().UTC()
			}
			d := memory.Digest{
				ID:        digestID,
				RunID:     runID,
				Project:   project,
				Summary:   draft.Summary,
				Decisions: draft.Decisions,
				EdgeCases: draft.EdgeCases,
				Errors:    draft.Errors,
				Embedding: embedding,
				ModelUsed: modelUsed,
				CreatedAt: createdAt,
			}
			if err := memory.UpsertDigest(ctx, db, d); err != nil {
				output.Warn("save edited digest: %s", err)
				return
			}
		}
	}
	output.Info("Digest saved (terminal, %s)", modelUsed)
}

// loadAgentOutputs queries the agents table for outputs of a given run.
func loadAgentOutputs(ctx context.Context, db *sql.DB, runID string) ([]memory.AgentOutput, error) {
	const q = `SELECT agent_id, agent_role, COALESCE(output, '') FROM agents WHERE run_id = ? AND output IS NOT NULL AND output != '' ORDER BY sequence ASC`

	rows, err := db.QueryContext(ctx, q, runID)
	if err != nil {
		return nil, fmt.Errorf("query agent outputs: %w", err)
	}
	defer rows.Close()

	var results []memory.AgentOutput
	for rows.Next() {
		var ao memory.AgentOutput
		if err := rows.Scan(&ao.AgentID, &ao.Role, &ao.Output); err != nil {
			return nil, fmt.Errorf("scan agent output: %w", err)
		}
		results = append(results, ao)
	}
	return results, rows.Err()
}

// showDigestApproval displays the draft and asks the user to approve, reject, or edit.
func showDigestApproval(draft memory.DigestDraft) string {
	fmt.Println()
	output.Info("Generated digest:")
	fmt.Println()
	fmt.Printf("  Summary: %s\n", draft.Summary)
	if len(draft.Decisions) > 0 {
		fmt.Println("  Decisions:")
		for _, d := range draft.Decisions {
			fmt.Printf("    - %s\n", d)
		}
	}
	if len(draft.EdgeCases) > 0 {
		fmt.Println("  Edge cases:")
		for _, e := range draft.EdgeCases {
			fmt.Printf("    - %s\n", e)
		}
	}
	if len(draft.Errors) > 0 {
		fmt.Println("  Errors:")
		for _, e := range draft.Errors {
			fmt.Printf("    - %s\n", e)
		}
	}
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("  Save digest? [s]í / [n]o / [e]ditar: ")
	line, _ := reader.ReadString('\n')
	line = strings.ToLower(strings.TrimSpace(line))

	switch {
	case line == "n" || line == "no":
		return "no"
	case line == "e" || line == "editar" || line == "edit":
		return "edit"
	default:
		return "yes"
	}
}

// editDigestSummary opens $EDITOR for the user to modify the summary text.
func editDigestSummary(summary string) string {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}

	tmpFile, err := os.CreateTemp("", "anvil-digest-*.md")
	if err != nil {
		output.Warn("create temp file: %s", err)
		return summary
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(summary); err != nil {
		output.Warn("write temp file: %s", err)
		return summary
	}
	tmpFile.Close()

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		output.Warn("editor failed: %s", err)
		return summary
	}

	edited, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		output.Warn("read edited file: %s", err)
		return summary
	}

	result := strings.TrimSpace(string(edited))
	if result == "" {
		return summary
	}
	return result
}

// formatMemoryContext formats search results into a prompt section.
func formatMemoryContext(results []memory.SearchResult) string {
	var b strings.Builder
	for _, r := range results {
		fmt.Fprintf(&b, "### [%s] run %s (score: %.2f)\n", r.Digest.Project, r.Digest.RunID, r.Score)
		fmt.Fprintf(&b, "Summary: %s\n", r.Digest.Summary)
		if len(r.Digest.Decisions) > 0 {
			b.WriteString("Decisions:\n")
			for _, d := range r.Digest.Decisions {
				fmt.Fprintf(&b, "  - %s\n", d)
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

// truncate clips a string to maxLen runes, appending an ellipsis when cut.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// newDigestID generates a unique digest ID.
func newDigestID() (string, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return fmt.Sprintf("d_%s_%s", time.Now().UTC().Format("20060102_150405"), hex.EncodeToString(b)), nil
}
