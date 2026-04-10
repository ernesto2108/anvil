//go:build dashboard

package dashboard

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
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
	// campos configurables para GetMetrics
	metricsResult store.Metrics
	metricsErr    error
}

func (f *fakeStore) WriteEvent(_ instrumentation.Event) error              { return nil }
func (f *fakeStore) ListRuns(_ context.Context, _, _ int) ([]store.RunSummary, error) {
	return nil, nil
}
func (f *fakeStore) ListAgentsByRun(_ context.Context, _ string) ([]store.AgentRow, error) {
	return f.agentRows, f.agentErr
}
func (f *fakeStore) GetAgentDetail(_ context.Context, _, _ string) (*store.AgentDetail, []store.FileRow, error) {
	return f.agentDetail, f.agentFiles, f.agentDetailErr
}
func (f *fakeStore) GetMetrics(_ context.Context) (store.Metrics, error) {
	return f.metricsResult, f.metricsErr
}
func (f *fakeStore) Close() error { return nil }

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
		// Tokens
		if n.Data.Tokens == nil {
			t.Fatal("Data.Tokens: esperaba valor no nil")
		}
		if *n.Data.Tokens != 5000 {
			t.Errorf("Data.Tokens: esperado 5000, obtuvo %d", *n.Data.Tokens)
		}
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
		// CRÍTICO — opción B: Output siempre vacío hasta DASH-FEAT-013
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
		// Output DEBE ser "" — placeholder hasta DASH-FEAT-013.
		// Si este test falla, alguien modificó toAgentDetailDTO antes de que
		// exista la tabla/columna de output.
		if dto.Output != "" {
			t.Errorf("Output: esperaba vacío (placeholder DASH-FEAT-013), obtuvo %q", dto.Output)
		}
	})
}

func Test_GetMetrics(t *testing.T) {
	t.Run("store_retorna_vacio", func(t *testing.T) {
		// GetMetrics debe retornar *MetricsDTO no-nil con slices vacíos no-nil.
		// json.Marshal debe producir "tokensPerRun":[] no "tokensPerRun":null.
		fs := &fakeStore{
			metricsResult: store.Metrics{
				RunsCount:       0,
				HasEnoughData:   false,
				TokensPerRun:    []store.TokensPerRunRow{},
				AvgTimePerAgent: []store.AgentTimeRow{},
			},
		}
		app := NewApp(fs)
		dto, err := app.GetMetrics()
		if err != nil {
			t.Fatalf("GetMetrics: %v", err)
		}
		if dto == nil {
			t.Fatal("GetMetrics: esperaba *MetricsDTO no-nil, obtuvo nil")
		}
		if dto.RunsCount != 0 {
			t.Errorf("dto.RunsCount: esperado 0, obtuvo %d", dto.RunsCount)
		}
		if dto.HasEnoughData {
			t.Error("dto.HasEnoughData: esperado false")
		}
		// Slices no-nil para serialización correcta.
		if dto.TokensPerRun == nil {
			t.Error("dto.TokensPerRun: esperaba slice no-nil")
		}
		if dto.AvgTimePerAgent == nil {
			t.Error("dto.AvgTimePerAgent: esperaba slice no-nil")
		}
		// Serialización: "tokensPerRun":[] no "tokensPerRun":null.
		jsonBytes, marshalErr := json.Marshal(dto)
		if marshalErr != nil {
			t.Fatalf("json.Marshal: %v", marshalErr)
		}
		jsonStr := string(jsonBytes)
		if !strings.Contains(jsonStr, `"tokensPerRun":[]`) {
			t.Errorf("JSON: esperaba %q, obtuvo %q", `"tokensPerRun":[]`, jsonStr)
		}
		if !strings.Contains(jsonStr, `"avgTimePerAgent":[]`) {
			t.Errorf("JSON: esperaba %q, obtuvo %q", `"avgTimePerAgent":[]`, jsonStr)
		}
	})

	t.Run("store_retorna_error", func(t *testing.T) {
		// Cuando el store retorna error, GetMetrics debe retornar (nil, error).
		fs := &fakeStore{
			metricsErr: errors.New("boom"),
		}
		app := NewApp(fs)
		dto, err := app.GetMetrics()
		if err == nil {
			t.Fatal("GetMetrics: esperaba error, obtuvo nil")
		}
		if dto != nil {
			t.Errorf("GetMetrics: esperaba nil DTO en error, obtuvo %+v", dto)
		}
	})

	t.Run("deltas_nil_serializan_null", func(t *testing.T) {
		// Deltas nil en store.Metrics → campos *float64/*int64 nil en DTO →
		// json.Marshal produce "totalTokensDeltaPct":null no "totalTokensDeltaPct":0.
		fs := &fakeStore{
			metricsResult: store.Metrics{
				RunsCount:           3,
				HasEnoughData:       false,
				TotalTokens:         900,
				AvgDurationMs:       2000,
				SuccessRate:         1.0,
				TotalTokensDeltaPct: nil, // sin prev
				AvgDurationDeltaMs:  nil,
				SuccessRateDelta:    nil,
				TokensPerRun:        []store.TokensPerRunRow{},
				AvgTimePerAgent:     []store.AgentTimeRow{},
			},
		}
		app := NewApp(fs)
		dto, err := app.GetMetrics()
		if err != nil {
			t.Fatalf("GetMetrics: %v", err)
		}
		if dto == nil {
			t.Fatal("GetMetrics: esperaba *MetricsDTO no-nil")
		}

		// Verificar que los campos nil del store se mapean a nil en el DTO.
		if dto.TotalTokensDeltaPct != nil {
			t.Errorf("dto.TotalTokensDeltaPct: esperado nil, obtuvo %f", *dto.TotalTokensDeltaPct)
		}
		if dto.AvgDurationDeltaMs != nil {
			t.Errorf("dto.AvgDurationDeltaMs: esperado nil, obtuvo %d", *dto.AvgDurationDeltaMs)
		}
		if dto.SuccessRateDelta != nil {
			t.Errorf("dto.SuccessRateDelta: esperado nil, obtuvo %f", *dto.SuccessRateDelta)
		}

		// Serialización: los campos nil deben aparecer como "null" no como "0" ni ausentes.
		jsonBytes, marshalErr := json.Marshal(dto)
		if marshalErr != nil {
			t.Fatalf("json.Marshal: %v", marshalErr)
		}
		jsonStr := string(jsonBytes)
		if !strings.Contains(jsonStr, `"totalTokensDeltaPct":null`) {
			t.Errorf("JSON: esperaba %q, obtuvo %q", `"totalTokensDeltaPct":null`, jsonStr)
		}
		if !strings.Contains(jsonStr, `"avgDurationDeltaMs":null`) {
			t.Errorf("JSON: esperaba %q, obtuvo %q", `"avgDurationDeltaMs":null`, jsonStr)
		}
		if !strings.Contains(jsonStr, `"successRateDelta":null`) {
			t.Errorf("JSON: esperaba %q, obtuvo %q", `"successRateDelta":null`, jsonStr)
		}
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
