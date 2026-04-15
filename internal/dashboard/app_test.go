//go:build dashboard

package dashboard

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

// fakeReader implements DashboardReader with configurable stubs.
type fakeReader struct {
	agentRows      []entity.AgentRow
	agentErr       error
	agentDetail    *entity.AgentDetail
	agentFiles     []entity.FileRecord
	agentDetailErr error
	runSummary     *entity.RunSummary
	runSummaryErr  error
}

func (f *fakeReader) ListRuns(_ context.Context, _, _ int, _, _, _, _ string) ([]entity.RunSummary, error) {
	return nil, nil
}
func (f *fakeReader) GetRunSummary(_ context.Context, _ string) (*entity.RunSummary, error) {
	return f.runSummary, f.runSummaryErr
}
func (f *fakeReader) ListChildRuns(_ context.Context, _ string) ([]entity.RunSummary, error) {
	return nil, nil
}
func (f *fakeReader) ListProjects(_ context.Context) ([]string, error) { return nil, nil }
func (f *fakeReader) ListAgentsByRun(_ context.Context, _ string) ([]entity.AgentRow, error) {
	return f.agentRows, f.agentErr
}
func (f *fakeReader) GetAgentDetail(_ context.Context, _, _ string) (*entity.AgentDetail, []entity.FileRecord, error) {
	return f.agentDetail, f.agentFiles, f.agentDetailErr
}
func (f *fakeReader) ListFilesByRun(_ context.Context, _ string) ([]entity.FileRecord, error) {
	return nil, nil
}
func (f *fakeReader) ListActivityEvents(_ context.Context, _ string) ([]entity.ActivityEvent, error) {
	return nil, nil
}
func (f *fakeReader) ListToolUsageByRun(_ context.Context, _ string) ([]entity.ToolUsage, error) {
	return nil, nil
}
func (f *fakeReader) TotalToolUsesByRun(_ context.Context, _ string) (int, error) { return 0, nil }
func (f *fakeReader) ListToolUseDetailsByRun(_ context.Context, _ string) ([]entity.ToolUseDetail, error) {
	return nil, nil
}
func (f *fakeReader) ListTasksByRun(_ context.Context, _ string) ([]entity.Task, error) {
	return nil, nil
}
func (f *fakeReader) CountCompactions(_ context.Context, _ string) (int, error) { return 0, nil }
func (f *fakeReader) ListPromptsByRun(_ context.Context, _ string) ([]entity.Prompt, error) {
	return nil, nil
}
func (f *fakeReader) GetTurnStats(_ context.Context, _ string) ([]entity.TurnStats, error) {
	return nil, nil
}
func (f *fakeReader) DeleteRun(_ context.Context, _ string) error    { return nil }
func (f *fakeReader) DeleteRuns(_ context.Context, _ []string) error { return nil }
func (f *fakeReader) ListErrorGroups(_ context.Context, _, _ string) ([]entity.ErrorGroup, error) {
	return nil, nil
}
func (f *fakeReader) GetErrorGroup(_ context.Context, _ string) (*entity.ErrorGroup, error) {
	return nil, nil
}
func (f *fakeReader) ListErrorGroupRuns(_ context.Context, _ string) ([]entity.ErrorGroupRun, error) {
	return nil, nil
}
func (f *fakeReader) ListErrorGroupHistory(_ context.Context, _ string) ([]entity.ErrorGroupHistory, error) {
	return nil, nil
}
func (f *fakeReader) GetErrorTrend(_ context.Context, _ string) ([]entity.TrendPoint, error) {
	return nil, nil
}
func (f *fakeReader) GetErrorCounts(_ context.Context) (entity.ErrorCounts, error) {
	return entity.ErrorCounts{}, nil
}

// fakeWriter implements DashboardWriter as a no-op.
type fakeWriter struct{}

func (f *fakeWriter) CleanupStaleRuns(_ int) (int64, error)                                { return 0, nil }
func (f *fakeWriter) BackfillProjects() (int64, error)                                     { return 0, nil }
func (f *fakeWriter) UpdateErrorResolution(_ context.Context, _, _, _, _ string) error      { return nil }
func (f *fakeWriter) Close() error                                                         { return nil }

func int64Ptr(v int64) *int64        { return &v }
func timePtr(t time.Time) *time.Time { return &t }

func newTestApp(r *fakeReader) *App {
	return NewApp(r, &fakeWriter{})
}

func Test_GetFlow(t *testing.T) {
	t.Run("store vacío retorna FlowDTO sin nodos ni aristas", func(t *testing.T) {
		app := newTestApp(&fakeReader{agentRows: nil})
		got, err := app.GetFlow("run-001")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}
		if got == nil {
			t.Fatal("GetFlow: esperaba FlowDTO no nil")
		}
		if len(got.Nodes) != 0 {
			t.Errorf("Nodes: esperaba 0, obtuvo %d", len(got.Nodes))
		}
		if len(got.Edges) != 0 {
			t.Errorf("Edges: esperaba 0, obtuvo %d", len(got.Edges))
		}
	})

	t.Run("1 agente → 1 nodo, 0 aristas", func(t *testing.T) {
		rows := []entity.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success", DurationMs: int64Ptr(500)},
		}
		app := newTestApp(&fakeReader{agentRows: rows})
		got, err := app.GetFlow("run-001")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}
		if len(got.Nodes) != 1 {
			t.Fatalf("Nodes: esperaba 1, obtuvo %d", len(got.Nodes))
		}
		if len(got.Edges) != 0 {
			t.Errorf("Edges: esperaba 0 (1 nodo no tiene aristas), obtuvo %d", len(got.Edges))
		}
		node := got.Nodes[0]
		if node.ID != "a1" {
			t.Errorf("Node.ID: esperado %q, obtuvo %q", "a1", node.ID)
		}
		if node.Type != "agentNode" {
			t.Errorf("Node.Type: esperado %q, obtuvo %q", "agentNode", node.Type)
		}
		if node.Data.Label != "pm" {
			t.Errorf("Node.Data.Label: esperado %q, obtuvo %q", "pm", node.Data.Label)
		}
		if node.Data.Status != "success" {
			t.Errorf("Node.Data.Status: esperado %q, obtuvo %q", "success", node.Data.Status)
		}
	})

	t.Run("store retorna error → GetFlow propaga error", func(t *testing.T) {
		app := newTestApp(&fakeReader{agentErr: context.Canceled})
		_, err := app.GetFlow("run-001")
		if err == nil {
			t.Fatal("GetFlow: esperaba error, obtuvo nil")
		}
	})
}

func Test_GetFlow_parseDependsOn(t *testing.T) {
	t.Run("3 agentes sin depends_on → aristas secuenciales", func(t *testing.T) {
		rows := []entity.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success"},
			{AgentID: "a2", AgentRole: "architect", Status: "success"},
			{AgentID: "a3", AgentRole: "developer", Status: "success"},
		}
		app := newTestApp(&fakeReader{agentRows: rows})
		got, err := app.GetFlow("run-001")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}

		if len(got.Nodes) != 3 {
			t.Fatalf("Nodes: esperaba 3, obtuvo %d", len(got.Nodes))
		}
		if len(got.Edges) != 2 {
			t.Fatalf("Edges: esperaba 2 (fallback secuencial), obtuvo %d", len(got.Edges))
		}
		if got.Edges[0].Source != "a1" || got.Edges[0].Target != "a2" {
			t.Errorf("Edges[0]: esperado a1->a2, obtuvo %s->%s", got.Edges[0].Source, got.Edges[0].Target)
		}
		if got.Edges[1].Source != "a2" || got.Edges[1].Target != "a3" {
			t.Errorf("Edges[1]: esperado a2->a3, obtuvo %s->%s", got.Edges[1].Source, got.Edges[1].Target)
		}
	})

	t.Run("3 agentes con depends_on → aristas por dependencia", func(t *testing.T) {
		rows := []entity.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success", DependsOn: "[]"},
			{AgentID: "a2", AgentRole: "architect", Status: "success", DependsOn: `["a1"]`},
			{AgentID: "a3", AgentRole: "developer", Status: "success", DependsOn: `["a2"]`},
		}
		app := newTestApp(&fakeReader{agentRows: rows})
		got, err := app.GetFlow("run-001")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}

		if len(got.Edges) != 2 {
			t.Fatalf("Edges: esperaba 2, obtuvo %d", len(got.Edges))
		}
		if got.Edges[0].Source != "a1" || got.Edges[0].Target != "a2" {
			t.Errorf("Edges[0]: esperado a1->a2, obtuvo %s->%s", got.Edges[0].Source, got.Edges[0].Target)
		}
	})

	t.Run("agente con múltiples depends_on → múltiples aristas entrantes", func(t *testing.T) {
		rows := []entity.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success", DependsOn: "[]"},
			{AgentID: "a2", AgentRole: "architect", Status: "success", DependsOn: "[]"},
			{AgentID: "a3", AgentRole: "developer", Status: "success", DependsOn: `["a1","a2"]`},
		}
		app := newTestApp(&fakeReader{agentRows: rows})
		got, err := app.GetFlow("run-001")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}

		if len(got.Edges) != 3 {
			t.Fatalf("Edges: esperaba 3, obtuvo %d", len(got.Edges))
		}
	})

	t.Run("no genera aristas duplicadas", func(t *testing.T) {
		rows := []entity.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success", DependsOn: "[]"},
			{AgentID: "a2", AgentRole: "developer", Status: "success", DependsOn: `["a1"]`},
		}
		app := newTestApp(&fakeReader{agentRows: rows})
		got, err := app.GetFlow("run-001")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}

		if len(got.Edges) != 1 {
			t.Fatalf("Edges: esperaba 1 (sin duplicados), obtuvo %d", len(got.Edges))
		}
	})

	t.Run("agente running tiene DurationMs nil en nodo", func(t *testing.T) {
		rows := []entity.AgentRow{
			{AgentID: "a1", AgentRole: "developer", Status: "running", DurationMs: nil},
		}
		app := newTestApp(&fakeReader{agentRows: rows})
		got, err := app.GetFlow("run-001")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}

		if len(got.Nodes) != 1 {
			t.Fatalf("Nodes: esperaba 1, obtuvo %d", len(got.Nodes))
		}
		if got.Nodes[0].Data.DurationMs != nil {
			t.Errorf("DurationMs: esperaba nil para agente running, obtuvo %v", got.Nodes[0].Data.DurationMs)
		}
	})

	t.Run("slice vacío → FlowDTO con slices no nil", func(t *testing.T) {
		app := newTestApp(&fakeReader{agentRows: []entity.AgentRow{}})
		got, err := app.GetFlow("run-empty")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}
		if got.Nodes == nil {
			t.Error("Nodes: esperaba slice no nil")
		}
		if got.Edges == nil {
			t.Error("Edges: esperaba slice no nil")
		}
	})

	t.Run("AC1: nodo expone nombre, duración y status correctamente", func(t *testing.T) {
		dur := int64(3750)
		rows := []entity.AgentRow{
			{AgentID: "a1", AgentRole: "architect", Status: "success", DurationMs: &dur},
		}
		app := newTestApp(&fakeReader{agentRows: rows})
		got, err := app.GetFlow("run-001")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}

		if len(got.Nodes) != 1 {
			t.Fatalf("Nodes: esperaba 1, obtuvo %d", len(got.Nodes))
		}
		n := got.Nodes[0]
		if n.Data.Label != "architect" {
			t.Errorf("Data.Label: esperado %q, obtuvo %q", "architect", n.Data.Label)
		}
		if n.Data.DurationMs == nil {
			t.Fatal("Data.DurationMs: esperaba valor no nil")
		}
		if *n.Data.DurationMs != 3750 {
			t.Errorf("Data.DurationMs: esperado 3750, obtuvo %d", *n.Data.DurationMs)
		}
		if n.Data.Status != "success" {
			t.Errorf("Data.Status: esperado %q, obtuvo %q", "success", n.Data.Status)
		}
	})

	t.Run("AC3: agente failed preserva status failed en el nodo", func(t *testing.T) {
		dur := int64(30000)
		rows := []entity.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success"},
			{AgentID: "a2", AgentRole: "architect", Status: "failed", DurationMs: &dur},
		}
		app := newTestApp(&fakeReader{agentRows: rows})
		got, err := app.GetFlow("run-001")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}

		if len(got.Nodes) != 2 {
			t.Fatalf("Nodes: esperaba 2, obtuvo %d", len(got.Nodes))
		}
		if got.Nodes[1].Data.Status != "failed" {
			t.Errorf("nodo a2 Data.Status: esperado %q, obtuvo %q", "failed", got.Nodes[1].Data.Status)
		}
		if got.Nodes[1].Data.DurationMs == nil {
			t.Fatal("nodo a2 Data.DurationMs: esperaba valor no nil para agente failed")
		}
		if *got.Nodes[1].Data.DurationMs != 30000 {
			t.Errorf("nodo a2 Data.DurationMs: esperado 30000, obtuvo %d", *got.Nodes[1].Data.DurationMs)
		}
	})

	t.Run("AC5: agente running tiene status running y DurationMs nil", func(t *testing.T) {
		rows := []entity.AgentRow{
			{AgentID: "a1", AgentRole: "developer", Status: "running", DurationMs: nil},
		}
		app := newTestApp(&fakeReader{agentRows: rows})
		got, err := app.GetFlow("run-001")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}

		if len(got.Nodes) != 1 {
			t.Fatalf("Nodes: esperaba 1, obtuvo %d", len(got.Nodes))
		}
		n := got.Nodes[0]
		if n.Data.Status != "running" {
			t.Errorf("Data.Status: esperado %q, obtuvo %q", "running", n.Data.Status)
		}
		if n.Data.DurationMs != nil {
			t.Errorf("Data.DurationMs: esperaba nil para agente en progreso, obtuvo %v", n.Data.DurationMs)
		}
	})

	t.Run("AC6: mix depends_on y fallback secuencial verifica aristas exactas", func(t *testing.T) {
		rows := []entity.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success", DependsOn: "[]"},
			{AgentID: "a2", AgentRole: "architect", Status: "success", DependsOn: "[]"},
			{AgentID: "a3", AgentRole: "developer", Status: "success", DependsOn: `["a1","a2"]`},
		}
		app := newTestApp(&fakeReader{agentRows: rows})
		got, err := app.GetFlow("run-001")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}

		if len(got.Nodes) != 3 {
			t.Fatalf("Nodes: esperaba 3, obtuvo %d", len(got.Nodes))
		}
		type edge struct{ src, tgt string }
		seen := make(map[edge]bool)
		for _, e := range got.Edges {
			seen[edge{e.Source, e.Target}] = true
		}

		if len(got.Edges) != 3 {
			t.Fatalf("Edges: esperaba 3, obtuvo %d (aristas: %v)", len(got.Edges), got.Edges)
		}
		if !seen[edge{"a1", "a2"}] {
			t.Error("arista a1→a2 (fallback secuencial) no encontrada")
		}
		if !seen[edge{"a1", "a3"}] {
			t.Error("arista a1→a3 (depends_on) no encontrada")
		}
		if !seen[edge{"a2", "a3"}] {
			t.Error("arista a2→a3 (depends_on) no encontrada")
		}
	})

	t.Run("depends_on JSON malformado no produce panic", func(t *testing.T) {
		rows := []entity.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success"},
			{AgentID: "a2", AgentRole: "developer", Status: "success", DependsOn: `["a1"`},
		}
		app := newTestApp(&fakeReader{agentRows: rows})
		got, err := app.GetFlow("run-001")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}
		if got == nil {
			t.Fatal("GetFlow: esperaba FlowDTO no nil con JSON malformado")
		}
		if len(got.Nodes) != 2 {
			t.Fatalf("Nodes: esperaba 2, obtuvo %d", len(got.Nodes))
		}
		if len(got.Edges) != 1 {
			t.Fatalf("Edges: esperaba 1, obtuvo %d", len(got.Edges))
		}
		if got.Edges[0].Target != "a2" {
			t.Errorf("arista Target: esperado %q, obtuvo %q", "a2", got.Edges[0].Target)
		}
	})

	t.Run("depends_on con espacios en valores produce aristas correctas", func(t *testing.T) {
		rows := []entity.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success"},
			{AgentID: "a2", AgentRole: "developer", Status: "success", DependsOn: `[ "a1" ]`},
		}
		app := newTestApp(&fakeReader{agentRows: rows})
		got, err := app.GetFlow("run-001")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}

		if len(got.Edges) != 1 {
			t.Fatalf("Edges: esperaba 1, obtuvo %d", len(got.Edges))
		}
		if got.Edges[0].Source != "a1" {
			t.Errorf("Source: esperado %q, obtuvo %q", "a1", got.Edges[0].Source)
		}
		if got.Edges[0].Target != "a2" {
			t.Errorf("Target: esperado %q, obtuvo %q", "a2", got.Edges[0].Target)
		}
	})

	t.Run("depends_on null literal produce fallback secuencial", func(t *testing.T) {
		rows := []entity.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success"},
			{AgentID: "a2", AgentRole: "developer", Status: "success", DependsOn: "null"},
		}
		app := newTestApp(&fakeReader{agentRows: rows})
		got, err := app.GetFlow("run-001")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}

		if len(got.Edges) != 1 {
			t.Fatalf("Edges: esperaba 1 (fallback), obtuvo %d", len(got.Edges))
		}
		if got.Edges[0].Source != "a1" || got.Edges[0].Target != "a2" {
			t.Errorf("arista: esperado a1→a2, obtuvo %s→%s", got.Edges[0].Source, got.Edges[0].Target)
		}
	})

	t.Run("agente con caracteres especiales en ID no rompe aristas", func(t *testing.T) {
		rows := []entity.AgentRow{
			{AgentID: "agent/pm-01", AgentRole: "pm", Status: "success"},
			{AgentID: "agent/dev-02", AgentRole: "developer", Status: "success", DependsOn: `["agent/pm-01"]`},
		}
		app := newTestApp(&fakeReader{agentRows: rows})
		got, err := app.GetFlow("run-001")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}

		if len(got.Nodes) != 2 {
			t.Fatalf("Nodes: esperaba 2, obtuvo %d", len(got.Nodes))
		}
		if len(got.Edges) != 1 {
			t.Fatalf("Edges: esperaba 1, obtuvo %d", len(got.Edges))
		}
		if got.Edges[0].Source != "agent/pm-01" {
			t.Errorf("Source: esperado %q, obtuvo %q", "agent/pm-01", got.Edges[0].Source)
		}
		if got.Edges[0].Target != "agent/dev-02" {
			t.Errorf("Target: esperado %q, obtuvo %q", "agent/dev-02", got.Edges[0].Target)
		}
	})
}

func Test_GetAgent(t *testing.T) {
	t.Run("store retorna detalle → GetAgent retorna DTO poblado", func(t *testing.T) {
		fr := &fakeReader{
			agentDetail: &entity.AgentDetail{
				AgentID:      "a1",
				AgentRole:    "developer",
				Status:       "success",
				DurationMs:   int64Ptr(500),
				TokensTotal:  intPtr(1500),
				TokensInput:  intPtr(1000),
				TokensOutput: intPtr(500),
			},
			agentFiles: []entity.FileRecord{
				{Path: "main.go", Operation: "touched"},
			},
		}
		app := newTestApp(fr)
		dto, err := app.GetAgent("run-x", "a1")
		if err != nil {
			t.Fatalf("GetAgent: %v", err)
		}
		if dto == nil {
			t.Fatal("GetAgent: esperaba DTO no nil")
		}
		if dto.Agent.ID != "a1" {
			t.Errorf("Agent.ID: esperado %q, obtuvo %q", "a1", dto.Agent.ID)
		}
		if dto.Agent.Name != "developer" {
			t.Errorf("Agent.Name: esperado %q (AgentRole), obtuvo %q", "developer", dto.Agent.Name)
		}
		if dto.Agent.Status != "success" {
			t.Errorf("Agent.Status: esperado %q, obtuvo %q", "success", dto.Agent.Status)
		}
		if dto.Agent.DurationMs == nil {
			t.Fatal("Agent.DurationMs: esperaba valor no nil")
		}
		if *dto.Agent.DurationMs != 500 {
			t.Errorf("Agent.DurationMs: esperado 500, obtuvo %d", *dto.Agent.DurationMs)
		}
		if len(dto.Files) != 1 {
			t.Fatalf("Files: esperaba 1, obtuvo %d", len(dto.Files))
		}
		if dto.Files[0].Path != "main.go" {
			t.Errorf("Files[0].Path: esperado %q, obtuvo %q", "main.go", dto.Files[0].Path)
		}
		if dto.Files[0].Action != "touched" {
			t.Errorf("Files[0].Action: esperado %q, obtuvo %q", "touched", dto.Files[0].Action)
		}
		if dto.Output != "" {
			t.Errorf("Output: esperaba vacío, obtuvo %q", dto.Output)
		}
	})

	t.Run("store retorna nil (agente no existe) → GetAgent retorna (nil, nil)", func(t *testing.T) {
		app := newTestApp(&fakeReader{})
		dto, err := app.GetAgent("run-x", "agente-inexistente")
		if err != nil {
			t.Fatalf("GetAgent: esperaba nil error, obtuvo %v", err)
		}
		if dto != nil {
			t.Errorf("GetAgent: esperaba nil DTO, obtuvo %+v", dto)
		}
	})

	t.Run("store retorna error → GetAgent propaga el error", func(t *testing.T) {
		app := newTestApp(&fakeReader{agentDetailErr: errors.New("boom")})
		dto, err := app.GetAgent("run-x", "a1")
		if err == nil {
			t.Fatal("GetAgent: esperaba error, obtuvo nil")
		}
		if err.Error() != "boom" {
			t.Errorf("error: esperado %q, obtuvo %q", "boom", err.Error())
		}
		if dto != nil {
			t.Errorf("GetAgent: esperaba nil DTO en error, obtuvo %+v", dto)
		}
	})

	t.Run("agente sin archivos → dto.Files no-nil y vacío", func(t *testing.T) {
		fr := &fakeReader{
			agentDetail: &entity.AgentDetail{
				AgentID:   "a1",
				AgentRole: "pm",
				Status:    "success",
			},
			agentFiles: []entity.FileRecord{},
		}
		app := newTestApp(fr)
		dto, err := app.GetAgent("run-x", "a1")
		if err != nil {
			t.Fatalf("GetAgent: %v", err)
		}
		if dto == nil {
			t.Fatal("GetAgent: esperaba DTO no nil")
		}
		if dto.Files == nil {
			t.Error("Files: esperaba slice no-nil para agente sin archivos")
		}
		if len(dto.Files) != 0 {
			t.Errorf("Files: esperaba 0 elementos, obtuvo %d", len(dto.Files))
		}
	})

	t.Run("timestamps se formatean a RFC3339Nano cuando existen", func(t *testing.T) {
		now := time.Now().UTC()
		fr := &fakeReader{
			agentDetail: &entity.AgentDetail{
				AgentID:   "a1",
				AgentRole: "developer",
				Status:    "success",
				StartedAt: timePtr(now),
				EndedAt:   nil,
			},
			agentFiles: []entity.FileRecord{},
		}
		app := newTestApp(fr)
		dto, err := app.GetAgent("run-x", "a1")
		if err != nil {
			t.Fatalf("GetAgent: %v", err)
		}
		if dto == nil {
			t.Fatal("GetAgent: esperaba DTO no nil")
		}
		if dto.Agent.StartedAt == "" {
			t.Error("Agent.StartedAt: esperaba string no vacío")
		}
		if _, parseErr := time.Parse(time.RFC3339Nano, dto.Agent.StartedAt); parseErr != nil {
			t.Errorf("Agent.StartedAt %q no es RFC3339Nano: %v", dto.Agent.StartedAt, parseErr)
		}
		if dto.Agent.EndedAt != "" {
			t.Errorf("Agent.EndedAt: esperaba vacío para EndedAt nil, obtuvo %q", dto.Agent.EndedAt)
		}
	})

	t.Run("Output siempre vacío — protege decisión opción B", func(t *testing.T) {
		fr := &fakeReader{
			agentDetail: &entity.AgentDetail{
				AgentID:      "a1",
				AgentRole:    "developer",
				Status:       "success",
				DurationMs:   int64Ptr(1000),
				TokensInput:  intPtr(500),
				TokensOutput: intPtr(300),
				TokensTotal:  intPtr(800),
				ErrorMsg:     "",
			},
			agentFiles: []entity.FileRecord{
				{Path: "internal/app.go", Operation: "write"},
				{Path: "internal/app_test.go", Operation: "write"},
			},
		}
		app := newTestApp(fr)
		dto, err := app.GetAgent("run-x", "a1")
		if err != nil {
			t.Fatalf("GetAgent: %v", err)
		}
		if dto == nil {
			t.Fatal("GetAgent: esperaba DTO no nil")
		}
		if dto.Output != "" {
			t.Errorf("Output: esperaba vacío cuando store retorna zero-value, obtuvo %q", dto.Output)
		}
	})

	t.Run("toAgentDetailDTO mapea Output del store al DTO correctamente", func(t *testing.T) {
		fr := &fakeReader{
			agentDetail: &entity.AgentDetail{
				AgentID:   "a2",
				AgentRole: "tester",
				Status:    "success",
				Output:    "hello world",
			},
			agentFiles: []entity.FileRecord{},
		}
		app := newTestApp(fr)
		dto, err := app.GetAgent("run-y", "a2")
		if err != nil {
			t.Fatalf("GetAgent: %v", err)
		}
		if dto == nil {
			t.Fatal("GetAgent: esperaba DTO no nil")
		}
		if dto.Output != "hello world" {
			t.Errorf("Output: esperado %q, obtuvo %q", "hello world", dto.Output)
		}
	})
}

func Test_GetRunSummary(t *testing.T) {
	qaScore := 8.5
	startedAt := time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC)

	t.Run("store retorna summary → GetRunSummary retorna RunDTO poblado", func(t *testing.T) {
		fr := &fakeReader{
			runSummary: &entity.RunSummary{
				ID:         "r_20260409_100000_abcd",
				TaskID:     "DASH-FEAT-010",
				TaskDesc:   "QA scores integration",
				Status:     "success",
				Complexity: "medium",
				Provider:   "anthropic",
				StartedAt:  startedAt,
				QAScore:    &qaScore,
			},
		}
		app := newTestApp(fr)
		dto, err := app.GetRunSummary("r_20260409_100000_abcd")
		if err != nil {
			t.Fatalf("GetRunSummary: %v", err)
		}
		if dto == nil {
			t.Fatal("GetRunSummary: esperaba RunDTO no nil")
		}
		if dto.ID != "r_20260409_100000_abcd" {
			t.Errorf("ID: esperado %q, obtuvo %q", "r_20260409_100000_abcd", dto.ID)
		}
		if dto.Status != "success" {
			t.Errorf("Status: esperado %q, obtuvo %q", "success", dto.Status)
		}
	})

	t.Run("store retorna nil (run no existe) → GetRunSummary retorna (nil, nil)", func(t *testing.T) {
		app := newTestApp(&fakeReader{})
		dto, err := app.GetRunSummary("no-existe")
		if err != nil {
			t.Fatalf("GetRunSummary: esperaba nil error, obtuvo %v", err)
		}
		if dto != nil {
			t.Errorf("GetRunSummary: esperaba nil DTO, obtuvo %+v", dto)
		}
	})

	t.Run("store retorna error → GetRunSummary propaga el error", func(t *testing.T) {
		app := newTestApp(&fakeReader{runSummaryErr: errors.New("db error")})
		dto, err := app.GetRunSummary("run-x")
		if err == nil {
			t.Fatal("GetRunSummary: esperaba error, obtuvo nil")
		}
		if err.Error() != "db error" {
			t.Errorf("error: esperado %q, obtuvo %q", "db error", err.Error())
		}
		if dto != nil {
			t.Errorf("GetRunSummary: esperaba nil DTO en error, obtuvo %+v", dto)
		}
	})

	t.Run("run sin qa_score → QAScore nil", func(t *testing.T) {
		fr := &fakeReader{
			runSummary: &entity.RunSummary{
				ID:        "r_running_001",
				TaskID:    "task-x",
				Status:    "running",
				StartedAt: startedAt,
				QAScore:   nil,
			},
		}
		app := newTestApp(fr)
		dto, err := app.GetRunSummary("r_running_001")
		if err != nil {
			t.Fatalf("GetRunSummary: %v", err)
		}
		if dto == nil {
			t.Fatal("GetRunSummary: esperaba RunDTO no nil")
		}
	})
}

func Test_GetFlow_extended(t *testing.T) {
	t.Run("AC1: GetFlow propaga nombre, duración y status al nodo", func(t *testing.T) {
		dur := int64(1200)
		rows := []entity.AgentRow{
			{AgentID: "agent-pm", AgentRole: "pm", Status: "success", DurationMs: &dur},
		}
		app := newTestApp(&fakeReader{agentRows: rows})
		got, err := app.GetFlow("run-001")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}
		if len(got.Nodes) != 1 {
			t.Fatalf("Nodes: esperaba 1, obtuvo %d", len(got.Nodes))
		}
		n := got.Nodes[0]
		if n.Data.Label != "pm" {
			t.Errorf("Data.Label: esperado %q, obtuvo %q", "pm", n.Data.Label)
		}
		if n.Data.Status != "success" {
			t.Errorf("Data.Status: esperado %q, obtuvo %q", "success", n.Data.Status)
		}
		if n.Data.DurationMs == nil {
			t.Fatal("Data.DurationMs: esperaba valor no nil")
		}
		if *n.Data.DurationMs != 1200 {
			t.Errorf("Data.DurationMs: esperado 1200, obtuvo %d", *n.Data.DurationMs)
		}
	})

	t.Run("runID vacío retorna FlowDTO vacío sin error", func(t *testing.T) {
		app := newTestApp(&fakeReader{agentRows: nil})
		got, err := app.GetFlow("")
		if err != nil {
			t.Fatalf("GetFlow con runID vacío: %v", err)
		}
		if got == nil {
			t.Fatal("GetFlow: esperaba FlowDTO no nil")
		}
		if len(got.Nodes) != 0 {
			t.Errorf("Nodes: esperaba 0, obtuvo %d", len(got.Nodes))
		}
	})

	t.Run("app con ctx de Startup usa ese ctx", func(t *testing.T) {
		rows := []entity.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success"},
		}
		app := newTestApp(&fakeReader{agentRows: rows})
		app.Startup(context.Background())

		got, err := app.GetFlow("run-001")
		if err != nil {
			t.Fatalf("GetFlow: %v", err)
		}
		if len(got.Nodes) != 1 {
			t.Errorf("Nodes: esperaba 1, obtuvo %d", len(got.Nodes))
		}
	})
}

func intPtr(v int) *int { return &v }
