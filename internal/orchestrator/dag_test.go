package orchestrator

import (
	"strings"
	"testing"
)

// containsAll checks that every element of want appears in got (order-independent).
func containsAll(got, want []string) bool {
	set := make(map[string]struct{}, len(got))
	for _, v := range got {
		set[v] = struct{}{}
	}
	for _, v := range want {
		if _, ok := set[v]; !ok {
			return false
		}
	}
	return true
}

// isValidTopologicalOrder verifies that for every node in order, all its
// DependsOn nodes appear at an earlier index.
func isValidTopologicalOrder(dag DAG, order []string) bool {
	pos := make(map[string]int, len(order))
	for i, id := range order {
		pos[id] = i
	}
	for id, node := range dag.Nodes {
		for _, dep := range node.DependsOn {
			if pos[dep] >= pos[id] {
				return false
			}
		}
	}
	return true
}

func Test_Build(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		nodes     []Node
		wantErr   bool
		errSubstr string
		validate  func(t *testing.T, dag DAG)
	}{
		{
			name: "linear chain A→B→C builds correctly",
			nodes: []Node{
				{ID: "A"},
				{ID: "B", DependsOn: []string{"A"}},
				{ID: "C", DependsOn: []string{"B"}},
			},
			validate: func(t *testing.T, dag DAG) {
				t.Helper()
				if len(dag.Order) != 3 {
					t.Fatalf("want 3 nodes in Order, got %d", len(dag.Order))
				}
				if !isValidTopologicalOrder(dag, dag.Order) {
					t.Errorf("Order is not a valid topological sort: %v", dag.Order)
				}
				// A must be first, C must be last.
				if dag.Order[0] != "A" {
					t.Errorf("expected A first in Order, got %q", dag.Order[0])
				}
				if dag.Order[2] != "C" {
					t.Errorf("expected C last in Order, got %q", dag.Order[2])
				}
			},
		},
		{
			name: "single node with no deps builds fine",
			nodes: []Node{
				{ID: "solo"},
			},
			validate: func(t *testing.T, dag DAG) {
				t.Helper()
				if len(dag.Order) != 1 || dag.Order[0] != "solo" {
					t.Errorf("unexpected Order: %v", dag.Order)
				}
			},
		},
		{
			name: "parallel nodes: A and B independent, C depends on both",
			nodes: []Node{
				{ID: "A"},
				{ID: "B"},
				{ID: "C", DependsOn: []string{"A", "B"}},
			},
			validate: func(t *testing.T, dag DAG) {
				t.Helper()
				if len(dag.Order) != 3 {
					t.Fatalf("want 3 nodes in Order, got %d", len(dag.Order))
				}
				if !isValidTopologicalOrder(dag, dag.Order) {
					t.Errorf("Order is not a valid topological sort: %v", dag.Order)
				}
				// C must be last; A and B must both precede it.
				if dag.Order[2] != "C" {
					t.Errorf("expected C last in Order, got %q", dag.Order[2])
				}
			},
		},
		{
			name: "cycle detection: A→B→C→A returns error mentioning cycle",
			nodes: []Node{
				{ID: "A", DependsOn: []string{"C"}},
				{ID: "B", DependsOn: []string{"A"}},
				{ID: "C", DependsOn: []string{"B"}},
			},
			wantErr:   true,
			errSubstr: "cycle",
		},
		{
			name: "missing dependency: node references non-existent ID",
			nodes: []Node{
				{ID: "A"},
				{ID: "B", DependsOn: []string{"nonexistent"}},
			},
			wantErr:   true,
			errSubstr: "unknown node",
		},
		{
			name: "duplicate node ID returns error",
			nodes: []Node{
				{ID: "A"},
				{ID: "A"},
			},
			wantErr:   true,
			errSubstr: "duplicate",
		},
	}

	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			dag, err := Build(tc.nodes)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tc.errSubstr != "" && !strings.Contains(err.Error(), tc.errSubstr) {
					t.Errorf("error %q does not contain %q", err.Error(), tc.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.validate != nil {
				tc.validate(t, dag)
			}
		})
	}
}

func Test_Ready(t *testing.T) {
	t.Parallel()

	// Build a shared DAG: A and B are independent, C depends on A and B.
	dag, err := Build([]Node{
		{ID: "A"},
		{ID: "B"},
		{ID: "C", DependsOn: []string{"A", "B"}},
	})
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	tests := []struct {
		name     string
		statuses map[string]NodeStatus
		wantIDs  []string
	}{
		{
			name: "all pending: A and B are ready",
			statuses: map[string]NodeStatus{
				"A": NodePending,
				"B": NodePending,
				"C": NodePending,
			},
			wantIDs: []string{"A", "B"},
		},
		{
			name: "A success, B pending: B and C not ready yet",
			statuses: map[string]NodeStatus{
				"A": NodeSuccess,
				"B": NodePending,
				"C": NodePending,
			},
			wantIDs: []string{"B"},
		},
		{
			name: "A and B success: C is ready",
			statuses: map[string]NodeStatus{
				"A": NodeSuccess,
				"B": NodeSuccess,
				"C": NodePending,
			},
			wantIDs: []string{"C"},
		},
		{
			name: "skipped dep counts as resolved: A skipped, B success → C ready",
			statuses: map[string]NodeStatus{
				"A": NodeSkipped,
				"B": NodeSuccess,
				"C": NodePending,
			},
			wantIDs: []string{"C"},
		},
		{
			name: "node already running is not returned as ready",
			statuses: map[string]NodeStatus{
				"A": NodeRunning,
				"B": NodePending,
				"C": NodePending,
			},
			wantIDs: []string{"B"},
		},
		{
			name: "node already succeeded is not returned as ready",
			statuses: map[string]NodeStatus{
				"A": NodeSuccess,
				"B": NodeSuccess,
				"C": NodeSuccess,
			},
			wantIDs: []string{},
		},
		{
			name: "failed dep does not make C ready",
			statuses: map[string]NodeStatus{
				"A": NodeFailed,
				"B": NodeSuccess,
				"C": NodePending,
			},
			wantIDs: []string{},
		},
	}

	for _, tc := range tests {

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := Ready(dag, tc.statuses)

			if len(tc.wantIDs) == 0 {
				if len(got) != 0 {
					t.Errorf("expected no ready nodes, got %v", got)
				}
				return
			}

			if !containsAll(got, tc.wantIDs) {
				t.Errorf("Ready() = %v, want all of %v", got, tc.wantIDs)
			}
			if len(got) != len(tc.wantIDs) {
				t.Errorf("Ready() returned %d nodes, want %d: got %v", len(got), len(tc.wantIDs), got)
			}
		})
	}
}
