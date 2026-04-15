package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/ernesto2108/anvil/internal/dashboard/writer"
	"github.com/ernesto2108/anvil/internal/instrumentation"
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
func checkClaudeAuth() {
	out, err := exec.Command("claude", "--print", "--bare", "-p", "ping").CombinedOutput()
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

	// 6. Wire components.
	agentRunner := runner.New(workDir, flags.model, flags.task)

	var gate orchestrator.GateHandler
	if flags.autoApprove {
		gate = orchestrator.NewAutoApproveHandler(os.Stdout)
	} else {
		gate = orchestrator.NewCLIGateHandler(os.Stdin, os.Stdout)
	}

	sink := &storeEventSink{w: s}

	exec := orchestrator.New(agentRunner, gate, sink, flags.concurrency)

	// 7. Handle Ctrl+C gracefully.
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// 8. Execute.
	output.Info("Starting run %s", output.Cyan(runID))
	fmt.Println()

	state := exec.Execute(ctx, dag, runID)

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

	// 11. Print summary.
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
