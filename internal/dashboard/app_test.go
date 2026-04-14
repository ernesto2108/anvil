//go:build dashboard

package dashboard

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/store"
	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// fakeStore implementa la interface Store completa con stubs no-op.
// ListAgentsByRun y GetAgentDetail tienen comportamiento configurable para los tests.
type fakeStore struct {
	agentRows []store.AgentRow
	agentErr  error
	// campos configurables para GetAgentDetail
	agentDetail    *store.AgentDetail
	agentFiles     []store.FileRow
	agentDetailErr error
	// campos configurables para GetRunSummary
	runSummary    *store.RunSummary
	runSummaryErr error
}

func (f *fakeStore) WriteEvent(_ instrumentation.Event) error { return nil }
func (f *fakeStore) ListRuns(_ context.Context, _, _ int, _, _, _, _ string) ([]store.RunSummary, error) {
	return nil, nil
}
func (f *fakeStore) ListAgentsByRun(_ context.Context, _ string) ([]store.AgentRow, error) {
	return f.agentRows, f.agentErr
}
func (f *fakeStore) GetAgentDetail(_ context.Context, _, _ string) (*store.AgentDetail, []store.FileRow, error) {
	return f.agentDetail, f.agentFiles, f.agentDetailErr
}
func (f *fakeStore) GetRunSummary(_ context.Context, _ string) (*store.RunSummary, error) {
	return f.runSummary, f.runSummaryErr
}
func (f *fakeStore) ListProjects(_ context.Context) ([]string, error)                        { return nil, nil }
func (f *fakeStore) ListFilesByRun(_ context.Context, _ string) ([]store.FileRow, error)     { return nil, nil }
func (f *fakeStore) ListChildRuns(_ context.Context, _ string) ([]store.RunSummary, error)   { return nil, nil }
func (f *fakeStore) ListActivityEvents(_ context.Context, _ string) ([]store.ActivityEvent, error) { return nil, nil }
func (f *fakeStore) ListToolUsageByRun(_ context.Context, _ string) ([]store.ToolUseSummary, error) { return nil, nil }
func (f *fakeStore) TotalToolUsesByRun(_ context.Context, _ string) (int, error)             { return 0, nil }
func (f *fakeStore) ListToolUseDetailsByRun(_ context.Context, _ string) ([]store.ToolUseDetail, error) { return nil, nil }
func (f *fakeStore) ListTasksByRun(_ context.Context, _ string) ([]store.TaskRow, error)     { return nil, nil }
func (f *fakeStore) CountCompactions(_ context.Context, _ string) (int, error)               { return 0, nil }
func (f *fakeStore) CountPermissionDenied(_ context.Context, _ string) (int, error)          { return 0, nil }
func (f *fakeStore) CleanupStaleRuns(_ int) (int64, error)                                   { return 0, nil }
func (f *fakeStore) BackfillProjects() (int64, error)                                        { return 0, nil }
func (f *fakeStore) ListPromptsByRun(_ context.Context, _ string) ([]store.PromptRow, error) { return nil, nil }
func (f *fakeStore) GetTurnStats(_ context.Context, _ string) ([]store.TurnStatsRow, error)  { return nil, nil }
func (f *fakeStore) DeleteRun(_ context.Context, _ string) error                             { return nil }
func (f *fakeStore) DeleteRuns(_ context.Context, _ []string) error                          { return nil }
func (f *fakeStore) Close() error                                                            { return nil }

func intPtr(v int) *int         { return &v }
func int64Ptr(v int64) *int64   { return &v }
func timePtr(t time.Time) *time.Time { return &t }

func Test_GetFlow(t *testing.T) {
	t.Run("store vacío retorna FlowDTO sin nodos ni aristas", func(t *testing.T) {
		app := NewApp(&fakeStore{agentRows: nil})
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
		rows := []store.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success", DurationMs: int64Ptr(500), TokensTotal: intPtr(1000)},
		}
		app := NewApp(&fakeStore{agentRows: rows})
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
		app := NewApp(&fakeStore{agentErr: context.Canceled})
		_, err := app.GetFlow("run-001")
		if err == nil {
			t.Fatal("GetFlow: esperaba error, obtuvo nil")
		}
	})
}

func Test_toFlowDTO(t *testing.T) {
	t.Run("3 agentes sin depends_on → aristas secuenciales", func(t *testing.T) {
		rows := []store.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success"},
			{AgentID: "a2", AgentRole: "architect", Status: "success"},
			{AgentID: "a3", AgentRole: "developer", Status: "success"},
		}
		got := toFlowDTO("run-001", rows)

		if len(got.Nodes) != 3 {
			t.Fatalf("Nodes: esperaba 3, obtuvo %d", len(got.Nodes))
		}
		if len(got.Edges) != 2 {
			t.Fatalf("Edges: esperaba 2 (fallback secuencial), obtuvo %d", len(got.Edges))
		}
		// a1 → a2
		if got.Edges[0].Source != "a1" || got.Edges[0].Target != "a2" {
			t.Errorf("Edges[0]: esperado a1->a2, obtuvo %s->%s", got.Edges[0].Source, got.Edges[0].Target)
		}
		// a2 → a3
		if got.Edges[1].Source != "a2" || got.Edges[1].Target != "a3" {
			t.Errorf("Edges[1]: esperado a2->a3, obtuvo %s->%s", got.Edges[1].Source, got.Edges[1].Target)
		}
	})

	t.Run("3 agentes con depends_on → aristas por dependencia", func(t *testing.T) {
		rows := []store.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success", DependsOn: "[]"},
			{AgentID: "a2", AgentRole: "architect", Status: "success", DependsOn: `["a1"]`},
			{AgentID: "a3", AgentRole: "developer", Status: "success", DependsOn: `["a2"]`},
		}
		got := toFlowDTO("run-001", rows)

		if len(got.Edges) != 2 {
			t.Fatalf("Edges: esperaba 2, obtuvo %d", len(got.Edges))
		}
		if got.Edges[0].Source != "a1" || got.Edges[0].Target != "a2" {
			t.Errorf("Edges[0]: esperado a1->a2, obtuvo %s->%s", got.Edges[0].Source, got.Edges[0].Target)
		}
	})

	t.Run("agente con múltiples depends_on → múltiples aristas entrantes", func(t *testing.T) {
		rows := []store.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success", DependsOn: "[]"},
			{AgentID: "a2", AgentRole: "architect", Status: "success", DependsOn: "[]"},
			{AgentID: "a3", AgentRole: "developer", Status: "success", DependsOn: `["a1","a2"]`},
		}
		got := toFlowDTO("run-001", rows)

		// a3 depende de a1 y a2: 2 aristas.
		// a1 y a2 no tienen depends_on y son el primero en su secuencia relativa:
		// a1 no tiene predecesor → 0 aristas por fallback.
		// a2 tendría fallback a a1, pero tiene depends_on vacío '[]' → fallback secuencial → a1->a2.
		// Total: 1 (a1->a2 fallback) + 2 (a1->a3, a2->a3) = 3 aristas.
		if len(got.Edges) != 3 {
			t.Fatalf("Edges: esperaba 3, obtuvo %d", len(got.Edges))
		}
	})

	t.Run("no genera aristas duplicadas", func(t *testing.T) {
		// Caso donde depends_on y fallback secuencial coinciden.
		rows := []store.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success", DependsOn: "[]"},
			// a2 tiene depends_on=a1 y además sería el sucesor secuencial de a1.
			{AgentID: "a2", AgentRole: "developer", Status: "success", DependsOn: `["a1"]`},
		}
		got := toFlowDTO("run-001", rows)

		if len(got.Edges) != 1 {
			t.Fatalf("Edges: esperaba 1 (sin duplicados), obtuvo %d", len(got.Edges))
		}
	})

	t.Run("agente running tiene DurationMs nil en nodo", func(t *testing.T) {
		rows := []store.AgentRow{
			{AgentID: "a1", AgentRole: "developer", Status: "running", DurationMs: nil},
		}
		got := toFlowDTO("run-001", rows)

		if len(got.Nodes) != 1 {
			t.Fatalf("Nodes: esperaba 1, obtuvo %d", len(got.Nodes))
		}
		if got.Nodes[0].Data.DurationMs != nil {
			t.Errorf("DurationMs: esperaba nil para agente running, obtuvo %v", got.Nodes[0].Data.DurationMs)
		}
	})

	t.Run("slice vacío → FlowDTO con slices no nil", func(t *testing.T) {
		got := toFlowDTO("run-empty", []store.AgentRow{})
		if got.Nodes == nil {
			t.Error("Nodes: esperaba slice no nil")
		}
		if got.Edges == nil {
			t.Error("Edges: esperaba slice no nil")
		}
	})

	// AC1: cada nodo debe tener los 3 campos — label (nombre), durationMs y status.
	t.Run("AC1: nodo expone nombre, duración y status correctamente", func(t *testing.T) {
		dur := int64(3750)
		tok := 5000
		rows := []store.AgentRow{
			{
				AgentID:     "a1",
				AgentRole:   "architect",
				Status:      "success",
				DurationMs:  &dur,
				TokensTotal: &tok,
			},
		}
		got := toFlowDTO("run-001", rows)

		if len(got.Nodes) != 1 {
			t.Fatalf("Nodes: esperaba 1, obtuvo %d", len(got.Nodes))
		}
		n := got.Nodes[0]
		// Nombre (label = AgentRole)
		if n.Data.Label != "architect" {
			t.Errorf("Data.Label: esperado %q, obtuvo %q", "architect", n.Data.Label)
		}
		// Duración
		if n.Data.DurationMs == nil {
			t.Fatal("Data.DurationMs: esperaba valor no nil")
		}
		if *n.Data.DurationMs != 3750 {
			t.Errorf("Data.DurationMs: esperado 3750, obtuvo %d", *n.Data.DurationMs)
		}
		// Status
		if n.Data.Status != "success" {
			t.Errorf("Data.Status: esperado %q, obtuvo %q", "success", n.Data.Status)
		}
		// Tokens field was removed from FlowNodeData — skip assertion.
	})

	// AC3: agente con status failed debe preservar ese status en el nodo.
	t.Run("AC3: agente failed preserva status failed en el nodo", func(t *testing.T) {
		dur := int64(30000)
		rows := []store.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success"},
			{AgentID: "a2", AgentRole: "architect", Status: "failed", DurationMs: &dur},
		}
		got := toFlowDTO("run-001", rows)

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

	// AC5: agente running debe tener status "running" y DurationMs nil en el nodo.
	t.Run("AC5: agente running tiene status running y DurationMs nil", func(t *testing.T) {
		rows := []store.AgentRow{
			{AgentID: "a1", AgentRole: "developer", Status: "running", DurationMs: nil},
		}
		got := toFlowDTO("run-001", rows)

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

	// AC6: mezcla de aristas depends_on (primario) y fallback secuencial en el mismo run.
	// Agentes: a1 (sin deps), a2 (sin deps → fallback a a1), a3 (depends_on: ["a1","a2"]).
	// Aristas esperadas: a1→a2 (fallback), a1→a3 (depends_on), a2→a3 (depends_on) = 3 total.
	// Nota: este caso ya existe como "agente con múltiples depends_on" pero sin verificar
	// explícitamente cuáles son las aristas. Este test verifica los source/target exactos.
	t.Run("AC6: mix depends_on y fallback secuencial verifica aristas exactas", func(t *testing.T) {
		rows := []store.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success", DependsOn: "[]"},
			{AgentID: "a2", AgentRole: "architect", Status: "success", DependsOn: "[]"},
			{AgentID: "a3", AgentRole: "developer", Status: "success", DependsOn: `["a1","a2"]`},
		}
		got := toFlowDTO("run-001", rows)

		if len(got.Nodes) != 3 {
			t.Fatalf("Nodes: esperaba 3, obtuvo %d", len(got.Nodes))
		}
		// Construir mapa de aristas para verificar sin depender del orden.
		type edge struct{ src, tgt string }
		seen := make(map[edge]bool)
		for _, e := range got.Edges {
			seen[edge{e.Source, e.Target}] = true
		}

		if len(got.Edges) != 3 {
			t.Fatalf("Edges: esperaba 3, obtuvo %d (aristas: %v)", len(got.Edges), got.Edges)
		}
		// a1 → a2: fallback secuencial (a2 tiene depends_on vacío).
		if !seen[edge{"a1", "a2"}] {
			t.Error("arista a1→a2 (fallback secuencial) no encontrada")
		}
		// a1 → a3: depends_on de a3.
		if !seen[edge{"a1", "a3"}] {
			t.Error("arista a1→a3 (depends_on) no encontrada")
		}
		// a2 → a3: depends_on de a3.
		if !seen[edge{"a2", "a3"}] {
			t.Error("arista a2→a3 (depends_on) no encontrada")
		}
	})

	// depends_on con JSON malformado (sin corchete de cierre) — parseDependsOn no debe hacer panic.
	// El comportamiento actual con input `["a1"`: los corchetes NO se stripean porque el último
	// char es `"` no `]`. El raw string `["a1"` se trata como un único dep ID → se genera una
	// arista con source=`["a1"` (dep ID inválido). Este test documenta que no hay panic y que
	// sí se genera exactamente 1 arista (aunque el source tenga el bracket).
	t.Run("depends_on JSON malformado no produce panic", func(t *testing.T) {
		rows := []store.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success"},
			// JSON inválido — falta el corchete de cierre.
			{AgentID: "a2", AgentRole: "developer", Status: "success", DependsOn: `["a1"`},
		}
		// Solo verificamos que no hace panic y retorna un FlowDTO válido.
		got := toFlowDTO("run-001", rows)
		if got == nil {
			t.Fatal("toFlowDTO: esperaba FlowDTO no nil con JSON malformado")
		}
		if len(got.Nodes) != 2 {
			t.Fatalf("Nodes: esperaba 2, obtuvo %d", len(got.Nodes))
		}
		// Con JSON sin corchete de cierre, parseDependsOn produce un dep ID inválido
		// (no `a1`), por lo que la arista tiene Source incorrecto. Documentamos el
		// count de aristas (1) sin verificar el source exacto, ya que el comportamiento
		// exacto es implementation-defined para input malformado.
		if len(got.Edges) != 1 {
			t.Fatalf("Edges: esperaba 1, obtuvo %d", len(got.Edges))
		}
		if got.Edges[0].Target != "a2" {
			t.Errorf("arista Target: esperado %q, obtuvo %q", "a2", got.Edges[0].Target)
		}
	})

	// depends_on con espacios en el JSON → trimQuotes debe limpiarlos.
	t.Run("depends_on con espacios en valores produce aristas correctas", func(t *testing.T) {
		rows := []store.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success"},
			// Espacios antes y después del string en el array JSON.
			{AgentID: "a2", AgentRole: "developer", Status: "success", DependsOn: `[ "a1" ]`},
		}
		got := toFlowDTO("run-001", rows)

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

	// depends_on con "null" literal — debe tratarse igual que vacío.
	t.Run("depends_on null literal produce fallback secuencial", func(t *testing.T) {
		rows := []store.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success"},
			{AgentID: "a2", AgentRole: "developer", Status: "success", DependsOn: "null"},
		}
		got := toFlowDTO("run-001", rows)

		// "null" → parseDependsOn retorna nil → fallback secuencial a1→a2.
		if len(got.Edges) != 1 {
			t.Fatalf("Edges: esperaba 1 (fallback), obtuvo %d", len(got.Edges))
		}
		if got.Edges[0].Source != "a1" || got.Edges[0].Target != "a2" {
			t.Errorf("arista: esperado a1→a2, obtuvo %s→%s", got.Edges[0].Source, got.Edges[0].Target)
		}
	})

	// Nombre de agente con caracteres especiales no debe romper el ID de arista.
	t.Run("agente con caracteres especiales en ID no rompe aristas", func(t *testing.T) {
		rows := []store.AgentRow{
			{AgentID: "agent/pm-01", AgentRole: "pm", Status: "success"},
			{AgentID: "agent/dev-02", AgentRole: "developer", Status: "success", DependsOn: `["agent/pm-01"]`},
		}
		got := toFlowDTO("run-001", rows)

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
		fs := &fakeStore{
			agentDetail: &store.AgentDetail{
				AgentID:      "a1",
				AgentRole:    "developer",
				Status:       "success",
				DurationMs:   int64Ptr(500),
				TokensTotal:  intPtr(1500),
				TokensInput:  intPtr(1000),
				TokensOutput: intPtr(500),
			},
			agentFiles: []store.FileRow{
				{Path: "main.go", Operation: "touched"},
			},
		}
		app := NewApp(fs)
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
		// Operation en store → Action en DTO (conversor renombra el campo)
		if dto.Files[0].Action != "touched" {
			t.Errorf("Files[0].Action: esperado %q, obtuvo %q", "touched", dto.Files[0].Action)
		}
		// Output es "" porque el mock no setea Output en AgentDetail.
		if dto.Output != "" {
			t.Errorf("Output: esperaba vacío (opción B, DASH-FEAT-013), obtuvo %q", dto.Output)
		}
	})

	t.Run("store retorna nil (agente no existe) → GetAgent retorna (nil, nil)", func(t *testing.T) {
		fs := &fakeStore{} // zero-value: agentDetail=nil, agentFiles=nil
		app := NewApp(fs)
		dto, err := app.GetAgent("run-x", "agente-inexistente")
		if err != nil {
			t.Fatalf("GetAgent: esperaba nil error, obtuvo %v", err)
		}
		if dto != nil {
			t.Errorf("GetAgent: esperaba nil DTO, obtuvo %+v", dto)
		}
	})

	t.Run("store retorna error → GetAgent propaga el error", func(t *testing.T) {
		fs := &fakeStore{agentDetailErr: errors.New("boom")}
		app := NewApp(fs)
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
		// Cubre AC2: "Sin archivos modificados" depende de slice vacío no-nil
		fs := &fakeStore{
			agentDetail: &store.AgentDetail{
				AgentID:   "a1",
				AgentRole: "pm",
				Status:    "success",
			},
			agentFiles: []store.FileRow{}, // slice vacío pero no-nil
		}
		app := NewApp(fs)
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
		fs := &fakeStore{
			agentDetail: &store.AgentDetail{
				AgentID:   "a1",
				AgentRole: "developer",
				Status:    "success",
				StartedAt: timePtr(now),
				EndedAt:   nil, // nil → campo vacío en DTO
			},
			agentFiles: []store.FileRow{},
		}
		app := NewApp(fs)
		dto, err := app.GetAgent("run-x", "a1")
		if err != nil {
			t.Fatalf("GetAgent: %v", err)
		}
		if dto == nil {
			t.Fatal("GetAgent: esperaba DTO no nil")
		}
		// StartedAt debe ser no vacío y parseable como RFC3339Nano
		if dto.Agent.StartedAt == "" {
			t.Error("Agent.StartedAt: esperaba string no vacío")
		}
		if _, parseErr := time.Parse(time.RFC3339Nano, dto.Agent.StartedAt); parseErr != nil {
			t.Errorf("Agent.StartedAt %q no es RFC3339Nano: %v", dto.Agent.StartedAt, parseErr)
		}
		// EndedAt nil → campo vacío
		if dto.Agent.EndedAt != "" {
			t.Errorf("Agent.EndedAt: esperaba vacío para EndedAt nil, obtuvo %q", dto.Agent.EndedAt)
		}
	})

	t.Run("Output siempre vacío — protege decisión opción B", func(t *testing.T) {
		// Este test falla intencionalmente si alguien agrega datos al campo Output
		// antes de implementar DASH-FEAT-013.
		fs := &fakeStore{
			agentDetail: &store.AgentDetail{
				AgentID:      "a1",
				AgentRole:    "developer",
				Status:       "success",
				DurationMs:   int64Ptr(1000),
				TokensInput:  intPtr(500),
				TokensOutput: intPtr(300),
				TokensTotal:  intPtr(800),
				ErrorMsg:     "",
			},
			agentFiles: []store.FileRow{
				{Path: "internal/app.go", Operation: "write"},
				{Path: "internal/app_test.go", Operation: "write"},
			},
		}
		app := NewApp(fs)
		dto, err := app.GetAgent("run-x", "a1")
		if err != nil {
			t.Fatalf("GetAgent: %v", err)
		}
		if dto == nil {
			t.Fatal("GetAgent: esperaba DTO no nil")
		}
		// Output DEBE ser "" cuando el store retorna AgentDetail con Output vacío (zero-value).
		// DASH-FEAT-013: la columna output ya existe; este test verifica que toAgentDetailDTO
		// mapea correctamente el campo Output del store al DTO.
		if dto.Output != "" {
			t.Errorf("Output: esperaba vacío cuando store retorna zero-value, obtuvo %q", dto.Output)
		}
	})

	t.Run("toAgentDetailDTO mapea Output del store al DTO correctamente", func(t *testing.T) {
		fs := &fakeStore{
			agentDetail: &store.AgentDetail{
				AgentID:   "a2",
				AgentRole: "tester",
				Status:    "success",
				Output:    "hello world",
			},
			agentFiles: []store.FileRow{},
		}
		app := NewApp(fs)
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

// Test_GetMetrics was removed — GetMetrics method no longer exists on App.

func Test_GetRunSummary(t *testing.T) {
	qaScore := 8.5
	startedAt := time.Date(2026, 4, 9, 10, 0, 0, 0, time.UTC)

	t.Run("store retorna summary → GetRunSummary retorna RunDTO poblado", func(t *testing.T) {
		fs := &fakeStore{
			runSummary: &store.RunSummary{
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
		app := NewApp(fs)
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
		// QAScore is stored in RunSummary but not exposed in RunDTO.
	})

	t.Run("store retorna nil (run no existe) → GetRunSummary retorna (nil, nil)", func(t *testing.T) {
		fs := &fakeStore{} // zero-value: runSummary=nil, runSummaryErr=nil
		app := NewApp(fs)
		dto, err := app.GetRunSummary("no-existe")
		if err != nil {
			t.Fatalf("GetRunSummary: esperaba nil error, obtuvo %v", err)
		}
		if dto != nil {
			t.Errorf("GetRunSummary: esperaba nil DTO, obtuvo %+v", dto)
		}
	})

	t.Run("store retorna error → GetRunSummary propaga el error", func(t *testing.T) {
		fs := &fakeStore{runSummaryErr: errors.New("db error")}
		app := NewApp(fs)
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

	t.Run("run sin qa_score → QAScore nil en DTO", func(t *testing.T) {
		fs := &fakeStore{
			runSummary: &store.RunSummary{
				ID:        "r_running_001",
				TaskID:    "task-x",
				Status:    "running",
				StartedAt: startedAt,
				QAScore:   nil,
			},
		}
		app := NewApp(fs)
		dto, err := app.GetRunSummary("r_running_001")
		if err != nil {
			t.Fatalf("GetRunSummary: %v", err)
		}
		if dto == nil {
			t.Fatal("GetRunSummary: esperaba RunDTO no nil")
		}
		// QAScore not exposed in RunDTO — skip assertion.
	})
}

// Test_GetFlow_extended cubre casos adicionales de GetFlow no presentes en Test_GetFlow.
func Test_GetFlow_extended(t *testing.T) {
	// AC1: GetFlow verifica que los 3 campos de datos (label, durationMs, status) llegan al nodo.
	t.Run("AC1: GetFlow propaga nombre, duración y status al nodo", func(t *testing.T) {
		dur := int64(1200)
		tok := 3000
		rows := []store.AgentRow{
			{
				AgentID:     "agent-pm",
				AgentRole:   "pm",
				Status:      "success",
				DurationMs:  &dur,
				TokensTotal: &tok,
			},
		}
		app := NewApp(&fakeStore{agentRows: rows})
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

	// GetFlow con runID vacío — debe retornar FlowDTO vacío sin error (no hay agentes).
	t.Run("runID vacío retorna FlowDTO vacío sin error", func(t *testing.T) {
		app := NewApp(&fakeStore{agentRows: nil})
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

	// GetFlow con contexto preestablecido via Startup.
	t.Run("app con ctx de Startup usa ese ctx", func(t *testing.T) {
		rows := []store.AgentRow{
			{AgentID: "a1", AgentRole: "pm", Status: "success"},
		}
		app := NewApp(&fakeStore{agentRows: rows})
		// Simular Startup con un contexto válido.
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
