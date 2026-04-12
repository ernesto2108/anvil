package orchestrator

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// Executor drives the DAG-based orchestration loop with bounded parallelism,
// HITL gates, and re-planning on failure.
//
// Invariants:
//   - RunState is mutated ONLY by the main Execute goroutine.
//   - Worker goroutines ONLY write to the results channel.
//   - Gate evaluation happens BEFORE semaphore acquisition so that a blocking
//     gate does not consume a concurrency slot.
type Executor struct {
	runner         AgentRunner
	gateHandler    GateHandler
	sink           EventSink
	maxConcurrency int
}

// New creates an Executor. maxConcurrency must be >= 1.
func New(runner AgentRunner, gateHandler GateHandler, sink EventSink, maxConcurrency int) *Executor {
	if maxConcurrency < 1 {
		maxConcurrency = 1
	}
	return &Executor{
		runner:         runner,
		gateHandler:    gateHandler,
		sink:           sink,
		maxConcurrency: maxConcurrency,
	}
}

// workerResult is what a worker goroutine sends back to the main loop.
type workerResult struct {
	result AgentResult
	err    error // runner-level error (distinct from AgentResult.Error)
}

// Execute runs the DAG to completion (or until an abort/ctx cancellation).
// RunState is safe to inspect after Execute returns; it is not safe to share
// while Execute is running.
func (e *Executor) Execute(ctx context.Context, dag DAG, runID string) RunState {
	nodeIDs := make([]string, 0, len(dag.Nodes))
	for id := range dag.Nodes {
		nodeIDs = append(nodeIDs, id)
	}

	state := RunState{
		RunID:     runID,
		Statuses:  make(map[string]NodeStatus, len(dag.Nodes)),
		Gates:     make(map[string]GateState),
		Results:   make(map[string]AgentResult, len(dag.Nodes)),
		StartedAt: time.Now().UTC(),
	}
	for id := range dag.Nodes {
		state.Statuses[id] = NodePending
	}

	// Emit orchestrator.start.
	if ev, err := instrumentation.NewEvent(runID, instrumentation.EventOrchestratorStart,
		instrumentation.OrchestratorStartPayload{
			NodeCount:      len(dag.Nodes),
			MaxConcurrency: e.maxConcurrency,
			AgentIDs:       nodeIDs,
		},
	); err == nil {
		e.sink.Emit(ev)
	}

	// Cancellable context so we can abort all in-flight workers on failure.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sem := make(chan struct{}, e.maxConcurrency)
	results := make(chan workerResult, len(dag.Nodes))
	var wg sync.WaitGroup

	// retryCount tracks how many times a node has been attempted.
	retryCount := make(map[string]int, len(dag.Nodes))

	aborted := false

	for !aborted {
		// Check if all nodes are in a terminal state.
		if e.allTerminal(state.Statuses) {
			break
		}

		// Drain any completed workers first.
		drained := false
		for !drained {
			select {
			case wr := <-results:
				e.applyWorkerResult(ctx, cancel, &state, wr, retryCount, results, &wg, sem, dag, runID, &aborted)
				if aborted {
					drained = true
				}
			default:
				drained = true
			}
		}

		if aborted {
			break
		}

		// Dispatch ready nodes.
		ready := Ready(dag, state.Statuses)
		for _, nodeID := range ready {
			if aborted {
				break
			}

			node := dag.Nodes[nodeID]

			// Gate evaluation BEFORE semaphore acquisition.
			if node.Gate {
				state.Statuses[nodeID] = NodeGated
				gateStart := time.Now()

				decision, gateErr := e.gateHandler.RequestApproval(ctx, node)
				waitMs := time.Since(gateStart).Milliseconds()

				gs := GateState{
					NodeID:    nodeID,
					Decision:  decision,
					DecidedAt: time.Now().UTC(),
				}
				if gateErr != nil {
					gs.Decision = GateRejected
					gs.Reason = gateErr.Error()
				}
				state.Gates[nodeID] = gs

				// Emit gate event.
				if ev, err := instrumentation.NewEvent(runID, instrumentation.EventOrchestratorGate,
					instrumentation.OrchestratorGatePayload{
						AgentID:        nodeID,
						Decision:       string(decision),
						WaitDurationMs: waitMs,
					},
				); err == nil {
					e.sink.Emit(ev)
				}

				switch gs.Decision {
				case GateRejected:
					cancel()
					aborted = true
					state.Statuses[nodeID] = NodeFailed
				case GateSkipped:
					state.Statuses[nodeID] = NodeSkipped
					e.cascadeSkip(dag, state.Statuses, nodeID)
					continue
				case GateApproved:
					// fall through to dispatch
				}

				if aborted {
					break
				}
			}

			// Acquire semaphore slot (non-blocking; if full, leave node for
			// next iteration — it will be re-evaluated when a slot frees up).
			select {
			case sem <- struct{}{}:
			default:
				// No slot available; reset to pending so Ready() picks it up again.
				// (It was already pending — gate changed it to gated, so restore.)
				if node.Gate {
					// Gate was approved but no slot; mark as waiting so we don't
					// re-evaluate the gate.
					state.Statuses[nodeID] = NodeWaiting
				}
				continue
			}

			state.Statuses[nodeID] = NodeRunning
			wg.Add(1)

			// Snapshot upstream results for this node.
			upstream := e.snapshotUpstream(node, state.Results)

			// Emit agent.start event.
			e.emitAgentStart(runID, node)

			go func(n Node, up map[string]AgentResult) {
				defer wg.Done()
				defer func() { <-sem }()

				runCtx := ctx
				if n.Timeout > 0 {
					var tCancel context.CancelFunc
					runCtx, tCancel = context.WithTimeout(ctx, n.Timeout)
					defer tCancel()
				}

				ar, err := e.runner.RunAgent(runCtx, n, up)
				results <- workerResult{result: ar, err: err}
			}(node, upstream)
		}

		// Handle waiting nodes that need a semaphore slot.
		e.dispatchWaiting(ctx, dag, &state, sem, results, &wg)

		// If no nodes are running and no nodes are pending/waiting/gated,
		// we're either stuck or done — break.
		if !e.anyInFlight(state.Statuses) && !e.anyDispatchable(state.Statuses) {
			break
		}

		// Wait for at least one result before looping again.
		if e.anyInFlight(state.Statuses) {
			wr := <-results
			e.applyWorkerResult(ctx, cancel, &state, wr, retryCount, results, &wg, sem, dag, runID, &aborted)
		}
	}

	// Wait for all in-flight goroutines to finish.
	// Drain their results to unblock them (context is already cancelled if aborted).
	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	// Keep draining results until all goroutines complete.
drainLoop:
	for {
		select {
		case wr := <-results:
			// During drain we still apply results so RunState is accurate.
			if !aborted {
				e.applyWorkerResult(ctx, cancel, &state, wr, retryCount, results, &wg, sem, dag, runID, &aborted)
			}
		case <-waitDone:
			break drainLoop
		}
	}

	state.FinishedAt = time.Now().UTC()
	state.FinalStatus = e.computeFinalStatus(state.Statuses)

	return state
}

// applyWorkerResult processes a single workerResult from the results channel,
// updating RunState and potentially triggering replanning.
// ONLY called from the main Execute goroutine.
func (e *Executor) applyWorkerResult(
	ctx context.Context,
	cancel context.CancelFunc,
	state *RunState,
	wr workerResult,
	retryCount map[string]int,
	results chan workerResult,
	wg *sync.WaitGroup,
	sem chan struct{},
	dag DAG,
	runID string,
	aborted *bool,
) {
	ar := wr.result
	if ar.NodeID == "" {
		// Incomplete result from a cancelled worker — ignore.
		return
	}

	node := dag.Nodes[ar.NodeID]

	// Merge runner-level error into AgentResult if the result didn't set one.
	if wr.err != nil && ar.Error == nil {
		ar.Error = fmt.Errorf("orchestrator/executor: runner error for node %q: %w", ar.NodeID, wr.err)
	}

	state.Results[ar.NodeID] = ar

	// Emit agent.end or agent.error event with output persisted.
	e.emitAgentEnd(runID, ar)

	if ar.Status == NodeSuccess {
		state.Statuses[ar.NodeID] = NodeSuccess
		return
	}

	// Node failed or errored.
	retryCount[ar.NodeID]++
	action := EvaluateFailure(node, retryCount[ar.NodeID])

	switch action {
	case ReplanRetry:
		state.Statuses[ar.NodeID] = NodeRetrying
		upstream := e.snapshotUpstream(node, state.Results)
		wg.Add(1)
		go func(n Node, up map[string]AgentResult) {
			defer wg.Done()
			sem <- struct{}{} // blocking acquire — retry should proceed
			defer func() { <-sem }()

			runCtx := ctx
			if n.Timeout > 0 {
				var tCancel context.CancelFunc
				runCtx, tCancel = context.WithTimeout(ctx, n.Timeout)
				defer tCancel()
			}

			newAR, err := e.runner.RunAgent(runCtx, n, up)
			results <- workerResult{result: newAR, err: err}
		}(node, upstream)

	case ReplanSkip:
		state.Statuses[ar.NodeID] = NodeSkipped
		e.cascadeSkip(dag, state.Statuses, ar.NodeID)

	case ReplanAbort:
		state.Statuses[ar.NodeID] = NodeFailed
		cancel()
		*aborted = true

	case ReplanAsk:
		// Print the error so the user knows WHY the agent failed.
		if ar.Error != nil {
			fmt.Fprintf(os.Stderr, "\n[ERROR] Agent %q failed: %s\n", ar.NodeID, ar.Error)
		}
		if ar.Output != "" {
			fmt.Fprintf(os.Stderr, "[OUTPUT] %s\n", ar.Output)
		}

		// Limit ask-retries to prevent infinite loops.
		const maxAskRetries = 3
		if retryCount[ar.NodeID] >= maxAskRetries {
			fmt.Fprintf(os.Stderr, "[ABORT] Agent %q failed %d times, aborting.\n", ar.NodeID, maxAskRetries)
			state.Statuses[ar.NodeID] = NodeFailed
			cancel()
			*aborted = true
			return
		}

		// Use the gate handler as an emergency decision point.
		decision, gateErr := e.gateHandler.RequestApproval(ctx, node)
		if gateErr != nil {
			// If we can't get a decision, abort.
			state.Statuses[ar.NodeID] = NodeFailed
			cancel()
			*aborted = true
			return
		}
		switch decision {
		case GateApproved:
			// Treat as retry.
			state.Statuses[ar.NodeID] = NodeRetrying
			retryUp := e.snapshotUpstream(node, state.Results)
			wg.Add(1)
			go func(n Node, up map[string]AgentResult) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				runCtx := ctx
				if n.Timeout > 0 {
					var tCancel context.CancelFunc
					runCtx, tCancel = context.WithTimeout(ctx, n.Timeout)
					defer tCancel()
				}

				newAR, err := e.runner.RunAgent(runCtx, n, up)
				results <- workerResult{result: newAR, err: err}
			}(node, retryUp)
		case GateSkipped:
			state.Statuses[ar.NodeID] = NodeSkipped
			e.cascadeSkip(dag, state.Statuses, ar.NodeID)
		default: // GateRejected or anything unexpected
			state.Statuses[ar.NodeID] = NodeFailed
			cancel()
			*aborted = true
		}
	}
}

// dispatchWaiting re-dispatches nodes that were left in NodeWaiting state
// (gate approved but no semaphore slot at the time).
func (e *Executor) dispatchWaiting(
	ctx context.Context,
	dag DAG,
	state *RunState,
	sem chan struct{},
	results chan workerResult,
	wg *sync.WaitGroup,
) {
	for id, status := range state.Statuses {
		if status != NodeWaiting {
			continue
		}
		select {
		case sem <- struct{}{}:
		default:
			continue
		}
		state.Statuses[id] = NodeRunning
		node := dag.Nodes[id]
		upstream := e.snapshotUpstream(node, state.Results)
		wg.Add(1)
		go func(n Node, up map[string]AgentResult) {
			defer wg.Done()
			defer func() { <-sem }()

			runCtx := ctx
			if n.Timeout > 0 {
				var tCancel context.CancelFunc
				runCtx, tCancel = context.WithTimeout(ctx, n.Timeout)
				defer tCancel()
			}

			ar, err := e.runner.RunAgent(runCtx, n, up)
			results <- workerResult{result: ar, err: err}
		}(node, upstream)
	}
}

// cascadeSkip marks all transitive dependents of a skipped node as skipped,
// provided they haven't already reached a terminal state.
func (e *Executor) cascadeSkip(dag DAG, statuses map[string]NodeStatus, nodeID string) {
	for _, dep := range dag.Edges[nodeID] {
		if !isTerminal(statuses[dep]) {
			statuses[dep] = NodeSkipped
			e.cascadeSkip(dag, statuses, dep)
		}
	}
}

// allTerminal returns true if every node has reached a terminal state.
func (e *Executor) allTerminal(statuses map[string]NodeStatus) bool {
	for _, s := range statuses {
		if !isTerminal(s) {
			return false
		}
	}
	return true
}

// anyInFlight returns true if at least one node is currently running or retrying.
func (e *Executor) anyInFlight(statuses map[string]NodeStatus) bool {
	for _, s := range statuses {
		if s == NodeRunning || s == NodeRetrying {
			return true
		}
	}
	return false
}

// anyDispatchable returns true if there are nodes still pending, waiting, or gated.
func (e *Executor) anyDispatchable(statuses map[string]NodeStatus) bool {
	for _, s := range statuses {
		if s == NodePending || s == NodeWaiting || s == NodeGated {
			return true
		}
	}
	return false
}

// computeFinalStatus derives the overall run status from all node statuses.
func (e *Executor) computeFinalStatus(statuses map[string]NodeStatus) string {
	for _, s := range statuses {
		if s == NodeFailed {
			return "failed"
		}
	}
	for _, s := range statuses {
		if s == NodeSkipped {
			return "partial"
		}
	}
	return "success"
}

// snapshotUpstream builds a map of completed dependency results for a node.
func (e *Executor) snapshotUpstream(node Node, results map[string]AgentResult) map[string]AgentResult {
	up := make(map[string]AgentResult, len(node.DependsOn))
	for _, dep := range node.DependsOn {
		if r, ok := results[dep]; ok {
			up[dep] = r
		}
	}
	return up
}

// emitAgentStart emits an agent.start instrumentation event.
func (e *Executor) emitAgentStart(runID string, node Node) {
	if ev, err := instrumentation.NewEvent(runID, instrumentation.EventAgentStart,
		instrumentation.AgentStartPayload{
			AgentID:   node.ID,
			AgentRole: node.Role,
			DependsOn: node.DependsOn,
		},
	); err == nil {
		e.sink.Emit(ev)
	}
}

// emitAgentEnd emits an agent.end or agent.error instrumentation event
// depending on the result status. Output is truncated to 4 KB to avoid
// bloating the events table.
func (e *Executor) emitAgentEnd(runID string, ar AgentResult) {
	output := ar.Output
	const maxOutput = 4096
	if len(output) > maxOutput {
		output = output[:maxOutput] + "…[truncated]"
	}

	if ar.Status == NodeFailed {
		errMsg := ""
		if ar.Error != nil {
			errMsg = ar.Error.Error()
		}
		if ev, err := instrumentation.NewEvent(runID, instrumentation.EventAgentError,
			instrumentation.AgentErrorPayload{
				AgentID:    ar.NodeID,
				Error:      errMsg,
				DurationMs: ar.DurationMs,
			},
		); err == nil {
			e.sink.Emit(ev)
		}
		return
	}

	if ev, err := instrumentation.NewEvent(runID, instrumentation.EventAgentEnd,
		instrumentation.AgentEndPayload{
			AgentID:      ar.NodeID,
			Status:       string(ar.Status),
			DurationMs:   ar.DurationMs,
			FilesTouched: ar.FilesTouched,
			Output:       output,
		},
	); err == nil {
		e.sink.Emit(ev)
	}
}

// isTerminal returns true for states that indicate a node will never run again.
func isTerminal(s NodeStatus) bool {
	switch s {
	case NodeSuccess, NodeFailed, NodeSkipped:
		return true
	default:
		return false
	}
}
