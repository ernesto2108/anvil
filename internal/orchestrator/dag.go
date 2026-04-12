package orchestrator

import "fmt"

// Build constructs a DAG from a slice of nodes. It validates that all
// DependsOn references exist and that the graph contains no cycles
// (Kahn's algorithm). Returns an error if validation fails.
func Build(nodes []Node) (DAG, error) {
	dag := DAG{
		Nodes:    make(map[string]Node, len(nodes)),
		Edges:    make(map[string][]string, len(nodes)),
		InDegree: make(map[string]int, len(nodes)),
		Order:    make([]string, 0, len(nodes)),
	}

	// Register all nodes first so we can validate DependsOn references.
	for _, n := range nodes {
		if _, exists := dag.Nodes[n.ID]; exists {
			return DAG{}, fmt.Errorf("orchestrator/dag: duplicate node id %q", n.ID)
		}
		dag.Nodes[n.ID] = n
		dag.InDegree[n.ID] = 0
	}

	// Build adjacency list and in-degree map.
	for _, n := range nodes {
		for _, dep := range n.DependsOn {
			if _, exists := dag.Nodes[dep]; !exists {
				return DAG{}, fmt.Errorf("orchestrator/dag: node %q depends on unknown node %q", n.ID, dep)
			}
			// dep -> n (dep must finish before n can run)
			dag.Edges[dep] = append(dag.Edges[dep], n.ID)
			dag.InDegree[n.ID]++
		}
	}

	// Kahn's algorithm for topological sort and cycle detection.
	queue := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if dag.InDegree[n.ID] == 0 {
			queue = append(queue, n.ID)
		}
	}

	// Work on a mutable copy of in-degrees so we don't mutate the stored map.
	inDeg := make(map[string]int, len(nodes))
	for k, v := range dag.InDegree {
		inDeg[k] = v
	}

	for len(queue) > 0 {
		// Pop front.
		cur := queue[0]
		queue = queue[1:]
		dag.Order = append(dag.Order, cur)

		for _, dependent := range dag.Edges[cur] {
			inDeg[dependent]--
			if inDeg[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	if len(dag.Order) != len(nodes) {
		return DAG{}, fmt.Errorf("orchestrator/dag: cycle detected among nodes")
	}

	return dag, nil
}

// Ready returns the IDs of nodes that are ready to run: all dependencies have
// status success or skipped, and the node's own status is pending.
// This is a pure function — it does not mutate the DAG or statuses.
func Ready(dag DAG, statuses map[string]NodeStatus) []string {
	var ready []string

	for id, node := range dag.Nodes {
		if statuses[id] != NodePending {
			continue
		}

		allDone := true
		for _, dep := range node.DependsOn {
			s := statuses[dep]
			if s != NodeSuccess && s != NodeSkipped {
				allDone = false
				break
			}
		}
		if allDone {
			ready = append(ready, id)
		}
	}

	return ready
}
