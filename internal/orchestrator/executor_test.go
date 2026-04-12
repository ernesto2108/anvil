package orchestrator

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// ---------------------------------------------------------------------------
// Mock implementations
// ---------------------------------------------------------------------------

// mockAgentRunner implements AgentRunner.
// For each node ID it holds a list of canned responses (one per call).
// If callCount exceeds the list length the last entry is reused.
type mockAgentRunner struct {
	mu       sync.Mutex
	calls    map[string][]AgentResult
	callNums map[string]int
	// onRun is called (if non-nil) before returning a result — useful for timing tests.
	onRun func(nodeID string)
}

func newMockRunner() *mockAgentRunner {
	return &mockAgentRunner{
		calls:    make(map[string][]AgentResult),
		callNums: make(map[string]int),
	}
}

// addResult appends a canned result for nodeID.
func (m *mockAgentRunner) addResult(nodeID string, ar AgentResult) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.calls[nodeID] = append(m.calls[nodeID], ar)
}

func (m *mockAgentRunner) RunAgent(_ context.Context, node Node, _ map[string]AgentResult) (AgentResult, error) {
	m.mu.Lock()
	idx := m.callNums[node.ID]
	canned := m.calls[node.ID]
	var ar AgentResult
	if len(canned) == 0 {
		ar = AgentResult{NodeID: node.ID, Status: NodeSuccess}
	} else {
		if idx >= len(canned) {
			idx = len(canned) - 1
		}
		m.callNums[node.ID]++
		ar = canned[idx]
	}
	onRun := m.onRun
	m.mu.Unlock()

	if onRun != nil {
		onRun(node.ID)
	}
	return ar, nil
}

// mockGateHandler implements GateHandler.
// Per node ID it returns the configured decision/error.
type mockGateHandler struct {
	mu        sync.Mutex
	responses map[string]gateResponse
	defaultResp gateResponse
}

type gateResponse struct {
	decision GateDecision
	err      error
}

func newMockGateHandler() *mockGateHandler {
	return &mockGateHandler{
		responses:   make(map[string]gateResponse),
		defaultResp: gateResponse{decision: GateApproved},
	}
}

func (m *mockGateHandler) setResponse(nodeID string, d GateDecision, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.responses[nodeID] = gateResponse{decision: d, err: err}
}

func (m *mockGateHandler) RequestApproval(_ context.Context, node Node) (GateDecision, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.responses[node.ID]; ok {
		return r.decision, r.err
	}
	return m.defaultResp.decision, m.defaultResp.err
}

// mockEventSink implements EventSink — discards events (or counts them).
type mockEventSink struct {
	mu     sync.Mutex
	events []instrumentation.Event
}

func (s *mockEventSink) Emit(ev instrumentation.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, ev)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func buildLinear(t *testing.T, ids ...string) DAG {
	t.Helper()
	nodes := make([]Node, len(ids))
	for i, id := range ids {
		nodes[i] = Node{ID: id}
		if i > 0 {
			nodes[i].DependsOn = []string{ids[i-1]}
		}
	}
	dag, err := Build(nodes)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	return dag
}

func successResult(nodeID string) AgentResult {
	return AgentResult{NodeID: nodeID, Status: NodeSuccess}
}

func failResult(nodeID string) AgentResult {
	return AgentResult{NodeID: nodeID, Status: NodeFailed, Error: errors.New("mock failure")}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func Test_Executor_HappyPath_LinearChain(t *testing.T) {
	t.Parallel()

	dag := buildLinear(t, "A", "B", "C")
	runner := newMockRunner()
	// All nodes succeed by default.

	exec := New(runner, newMockGateHandler(), &mockEventSink{}, 3)
	state := exec.Execute(context.Background(), dag, "run-1")

	if state.FinalStatus != "success" {
		t.Errorf("FinalStatus = %q, want %q", state.FinalStatus, "success")
	}
	for _, id := range []string{"A", "B", "C"} {
		if state.Statuses[id] != NodeSuccess {
			t.Errorf("node %q status = %q, want %q", id, state.Statuses[id], NodeSuccess)
		}
	}
}

func Test_Executor_ParallelExecution(t *testing.T) {
	t.Parallel()

	// A and B are independent; C depends on both.
	dag, err := Build([]Node{
		{ID: "A"},
		{ID: "B"},
		{ID: "C", DependsOn: []string{"A", "B"}},
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Both A and B are dispatched and must overlap in time.
	// Use a WaitGroup + channel gate: each goroutine waits until
	// the other has also started before proceeding.
	var mu sync.Mutex
	var maxSeen int
	var active int

	runner := newMockRunner()
	runner.onRun = func(_ string) {
		mu.Lock()
		active++
		if active > maxSeen {
			maxSeen = active
		}
		mu.Unlock()

		// Hold the slot long enough for the sibling node to also enter.
		time.Sleep(30 * time.Millisecond)

		mu.Lock()
		active--
		mu.Unlock()
	}

	exec := New(runner, newMockGateHandler(), &mockEventSink{}, 3)
	state := exec.Execute(context.Background(), dag, "run-parallel")

	if state.FinalStatus != "success" {
		t.Errorf("FinalStatus = %q, want %q", state.FinalStatus, "success")
	}
	mu.Lock()
	max := maxSeen
	mu.Unlock()
	if max < 2 {
		t.Errorf("expected at least 2 nodes to run concurrently, max observed = %d", max)
	}
}

func Test_Executor_Gate_Approved(t *testing.T) {
	t.Parallel()

	dag, err := Build([]Node{
		{ID: "A", Gate: true},
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	gate := newMockGateHandler()
	gate.setResponse("A", GateApproved, nil)

	exec := New(newMockRunner(), gate, &mockEventSink{}, 1)
	state := exec.Execute(context.Background(), dag, "run-gate-approved")

	if state.FinalStatus != "success" {
		t.Errorf("FinalStatus = %q, want %q", state.FinalStatus, "success")
	}
	if state.Statuses["A"] != NodeSuccess {
		t.Errorf("node A status = %q, want %q", state.Statuses["A"], NodeSuccess)
	}
	if state.Gates["A"].Decision != GateApproved {
		t.Errorf("gate decision = %q, want %q", state.Gates["A"].Decision, GateApproved)
	}
}

func Test_Executor_Gate_Rejected(t *testing.T) {
	t.Parallel()

	dag, err := Build([]Node{
		{ID: "A", Gate: true},
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	gate := newMockGateHandler()
	gate.setResponse("A", GateRejected, nil)

	exec := New(newMockRunner(), gate, &mockEventSink{}, 1)
	state := exec.Execute(context.Background(), dag, "run-gate-rejected")

	if state.FinalStatus != "failed" {
		t.Errorf("FinalStatus = %q, want %q", state.FinalStatus, "failed")
	}
}

func Test_Executor_Gate_Skipped(t *testing.T) {
	t.Parallel()

	// A (gated) → B (dependent)
	dag, err := Build([]Node{
		{ID: "A", Gate: true},
		{ID: "B", DependsOn: []string{"A"}},
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	gate := newMockGateHandler()
	gate.setResponse("A", GateSkipped, nil)

	exec := New(newMockRunner(), gate, &mockEventSink{}, 2)
	state := exec.Execute(context.Background(), dag, "run-gate-skipped")

	if state.Statuses["A"] != NodeSkipped {
		t.Errorf("node A status = %q, want %q", state.Statuses["A"], NodeSkipped)
	}
	if state.Statuses["B"] != NodeSkipped {
		t.Errorf("node B status = %q, want %q (cascade)", state.Statuses["B"], NodeSkipped)
	}
	// partial because nodes were skipped
	if state.FinalStatus != "partial" {
		t.Errorf("FinalStatus = %q, want %q", state.FinalStatus, "partial")
	}
}

func Test_Executor_AgentFailure_AbortPolicy(t *testing.T) {
	t.Parallel()

	dag, err := Build([]Node{
		{ID: "A", OnFail: FailPolicyAbort},
		{ID: "B", DependsOn: []string{"A"}},
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	runner := newMockRunner()
	runner.addResult("A", failResult("A"))

	exec := New(runner, newMockGateHandler(), &mockEventSink{}, 2)
	state := exec.Execute(context.Background(), dag, "run-abort")

	if state.FinalStatus != "failed" {
		t.Errorf("FinalStatus = %q, want %q", state.FinalStatus, "failed")
	}
	if state.Statuses["A"] != NodeFailed {
		t.Errorf("node A status = %q, want %q", state.Statuses["A"], NodeFailed)
	}
}

func Test_Executor_AgentFailure_SkipPolicy(t *testing.T) {
	t.Parallel()

	// A fails with skip → B (dependent of A) should cascade-skip.
	dag, err := Build([]Node{
		{ID: "A", OnFail: FailPolicySkip},
		{ID: "B", DependsOn: []string{"A"}},
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	runner := newMockRunner()
	runner.addResult("A", failResult("A"))

	exec := New(runner, newMockGateHandler(), &mockEventSink{}, 2)
	state := exec.Execute(context.Background(), dag, "run-skip")

	if state.Statuses["A"] != NodeSkipped {
		t.Errorf("node A status = %q, want %q", state.Statuses["A"], NodeSkipped)
	}
	if state.Statuses["B"] != NodeSkipped {
		t.Errorf("node B status = %q, want skipped (cascade)", state.Statuses["B"])
	}
	if state.FinalStatus != "partial" {
		t.Errorf("FinalStatus = %q, want %q", state.FinalStatus, "partial")
	}
}

func Test_Executor_AgentFailure_RetryPolicy(t *testing.T) {
	t.Parallel()

	dag, err := Build([]Node{
		{ID: "A", OnFail: FailPolicyRetry},
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	runner := newMockRunner()
	// First call fails, second succeeds.
	runner.addResult("A", failResult("A"))
	runner.addResult("A", successResult("A"))

	// Gate handler will be called on ReplanAsk (attempt 2 with retry policy → ask).
	// The mock gate returns GateApproved by default which triggers another retry.
	// But after the second attempt succeeds, no gate is needed. So we approve.
	gate := newMockGateHandler()
	gate.defaultResp = gateResponse{decision: GateApproved}

	exec := New(runner, gate, &mockEventSink{}, 1)
	state := exec.Execute(context.Background(), dag, "run-retry")

	if state.FinalStatus != "success" {
		t.Errorf("FinalStatus = %q, want %q", state.FinalStatus, "success")
	}
	if state.Statuses["A"] != NodeSuccess {
		t.Errorf("node A status = %q, want %q", state.Statuses["A"], NodeSuccess)
	}
}

func Test_Executor_ContextCancellation(t *testing.T) {
	t.Parallel()

	// Use a node with abort policy so that when it fails after context
	// cancellation, the executor aborts immediately without retrying.
	dag, err := Build([]Node{
		{ID: "slow", OnFail: FailPolicyAbort},
		{ID: "downstream", DependsOn: []string{"slow"}, OnFail: FailPolicyAbort},
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	runner := newMockRunner()
	runner.onRun = func(_ string) {
		// Cancel context as soon as the node starts executing.
		cancel()
		// Brief sleep to simulate work; then the node returns a failure.
		time.Sleep(5 * time.Millisecond)
	}
	// The slow node fails (simulating context-awareness in the runner).
	runner.addResult("slow", AgentResult{NodeID: "slow", Status: NodeFailed, Error: context.Canceled})

	exec := New(runner, newMockGateHandler(), &mockEventSink{}, 1)

	done := make(chan RunState, 1)
	go func() {
		done <- exec.Execute(ctx, dag, "run-cancel")
	}()

	select {
	case state := <-done:
		// The run must complete and report a non-success status.
		if state.FinalStatus == "success" {
			t.Errorf("FinalStatus = %q after cancellation, want non-success", state.FinalStatus)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Execute did not return within 10s after context cancellation")
	}
}

func Test_Executor_MaxConcurrency(t *testing.T) {
	t.Parallel()

	// 4 independent nodes, maxConcurrency=2.
	dag, err := Build([]Node{
		{ID: "N1"},
		{ID: "N2"},
		{ID: "N3"},
		{ID: "N4"},
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	var concurrent int32
	var maxConcurrent int32
	var mu sync.Mutex
	exceeded := false

	runner := newMockRunner()
	runner.onRun = func(_ string) {
		cur := atomic.AddInt32(&concurrent, 1)

		mu.Lock()
		if cur > 2 {
			exceeded = true
		}
		if cur > atomic.LoadInt32(&maxConcurrent) {
			atomic.StoreInt32(&maxConcurrent, cur)
		}
		mu.Unlock()

		time.Sleep(20 * time.Millisecond)
		atomic.AddInt32(&concurrent, -1)
	}

	exec := New(runner, newMockGateHandler(), &mockEventSink{}, 2)
	state := exec.Execute(context.Background(), dag, "run-maxconc")

	if exceeded {
		t.Errorf("concurrency exceeded limit of 2; max observed = %d", maxConcurrent)
	}
	if state.FinalStatus != "success" {
		t.Errorf("FinalStatus = %q, want %q", state.FinalStatus, "success")
	}
}
