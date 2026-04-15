package mapper

import (
	"testing"

	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

func stubParseDependsOn(raw string) []string {
	if raw == "" || raw == "[]" || raw == "null" {
		return nil
	}
	// Simple parser for test: just strip brackets and quotes, split by comma.
	s := raw
	if len(s) >= 2 && s[0] == '[' && s[len(s)-1] == ']' {
		s = s[1 : len(s)-1]
	}
	if s == "" {
		return nil
	}
	var result []string
	for _, p := range splitComma(s) {
		p = trimQ(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func splitComma(s string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			parts = append(parts, s[start:i])
			start = i + 1
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func trimQ(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '"') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '"') {
		s = s[:len(s)-1]
	}
	return s
}

func TestToFlowDTO_Sequential(t *testing.T) {
	rows := []entity.AgentRow{
		{AgentID: "a1", AgentRole: "developer", Status: "success"},
		{AgentID: "a2", AgentRole: "tester", Status: "success"},
		{AgentID: "a3", AgentRole: "qa", Status: "running"},
	}

	flow := ToFlowDTO(rows, stubParseDependsOn)

	if len(flow.Nodes) != 3 {
		t.Fatalf("Nodes: expected 3, got %d", len(flow.Nodes))
	}
	if len(flow.Edges) != 2 {
		t.Fatalf("Edges: expected 2, got %d", len(flow.Edges))
	}
	if flow.Edges[0].Source != "a1" || flow.Edges[0].Target != "a2" {
		t.Errorf("Edge[0]: expected a1->a2, got %s->%s", flow.Edges[0].Source, flow.Edges[0].Target)
	}
	if flow.Edges[1].Source != "a2" || flow.Edges[1].Target != "a3" {
		t.Errorf("Edge[1]: expected a2->a3, got %s->%s", flow.Edges[1].Source, flow.Edges[1].Target)
	}
}

func TestToFlowDTO_WithDependsOn(t *testing.T) {
	rows := []entity.AgentRow{
		{AgentID: "a1", AgentRole: "developer", Status: "success"},
		{AgentID: "a2", AgentRole: "developer", Status: "success"},
		{AgentID: "a3", AgentRole: "tester", Status: "success", DependsOn: `["a1","a2"]`},
	}

	flow := ToFlowDTO(rows, stubParseDependsOn)

	// 3 edges: a1→a2 (sequential fallback), a1→a3, a2→a3 (depends_on)
	if len(flow.Edges) != 3 {
		t.Fatalf("Edges: expected 3, got %d", len(flow.Edges))
	}
	// Verify a3 has both a1 and a2 as sources
	a3Sources := map[string]bool{}
	for _, e := range flow.Edges {
		if e.Target == "a3" {
			a3Sources[e.Source] = true
		}
	}
	if !a3Sources["a1"] || !a3Sources["a2"] {
		t.Errorf("expected a3 sources {a1, a2}, got %v", a3Sources)
	}
}

func TestToFlowDTO_Empty(t *testing.T) {
	flow := ToFlowDTO(nil, stubParseDependsOn)
	if len(flow.Nodes) != 0 {
		t.Errorf("Nodes: expected 0, got %d", len(flow.Nodes))
	}
	if len(flow.Edges) != 0 {
		t.Errorf("Edges: expected 0, got %d", len(flow.Edges))
	}
}
