//go:build dashboard

// Tests para DASH-FEAT-008 vista de métricas.
//
// Acceptance criteria cubiertos:
//
//	AC1 (≥5 ejecuciones → charts con datos reales): cinco_runs_sin_prev, treinta_mas_treinta
//	AC2 (1 ejecución → datos sin tendencia + mensaje): un_run, cuatro_runs
//	AC3 (últimos 30 runs, ejes legibles): treinta_mas_treinta, mezcla_status (verifica LIMIT 30)
package store_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/store"
	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// --- seed helpers ------------------------------------------------------------

// seedRun inserta un run completo (start + end) con el timestamp y valores dados.
// status debe ser "success" o "failed". Para "running", usa seedRunningRun.
func seedRun(t *testing.T, s *store.TestableStore, runID string, startedAt time.Time, status string, durationMs int64, totalTokens int) {
	t.Helper()

	startEv := mustNewEvent(t, runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
		TaskID:   "task-metrics",
		Provider: "anthropic",
	})
	startEv.Timestamp = startedAt
	if err := s.WriteEvent(startEv); err != nil {
		t.Fatalf("seedRun start %q: %v", runID, err)
	}

	endEv := mustNewEvent(t, runID, instrumentation.EventRunEnd, instrumentation.RunEndPayload{
		Status:      status,
		DurationMs:  durationMs,
		TotalTokens: totalTokens,
	})
	endEv.Timestamp = startedAt.Add(time.Duration(durationMs) * time.Millisecond)
	if err := s.WriteEvent(endEv); err != nil {
		t.Fatalf("seedRun end %q: %v", runID, err)
	}
}

// seedRunningRun inserta un run que queda en estado "running" (sin run.end).
func seedRunningRun(t *testing.T, s *store.TestableStore, runID string, startedAt time.Time) {
	t.Helper()

	startEv := mustNewEvent(t, runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
		TaskID:   "task-metrics",
		Provider: "anthropic",
	})
	startEv.Timestamp = startedAt
	if err := s.WriteEvent(startEv); err != nil {
		t.Fatalf("seedRunningRun %q: %v", runID, err)
	}
}

// seedAgentForRun inserta un agente para un run existente.
// Si durationMs == nil, el agente queda en estado "running" (sin agent.end),
// lo que significa que duration_ms queda NULL en la tabla agents.
func seedAgentForRun(t *testing.T, s *store.TestableStore, runID, agentID, role string, startedAt time.Time, durationMs *int64) {
	t.Helper()

	startEv := mustNewEvent(t, runID, instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
		AgentID:   agentID,
		AgentRole: role,
		Sequence:  1,
		DependsOn: []string{},
		Model:     "claude-sonnet-4-6",
	})
	startEv.Timestamp = startedAt
	if err := s.WriteEvent(startEv); err != nil {
		t.Fatalf("seedAgentForRun start %q/%q: %v", runID, agentID, err)
	}

	if durationMs == nil {
		return // agente queda en "running" — duration_ms permanece NULL
	}

	endEv := mustNewEvent(t, runID, instrumentation.EventAgentEnd, instrumentation.AgentEndPayload{
		AgentID:      agentID,
		Status:       "success",
		DurationMs:   *durationMs,
		TokensTotal:  100,
		FilesTouched: []string{},
		ExitCode:     0,
	})
	endEv.Timestamp = startedAt.Add(time.Duration(*durationMs) * time.Millisecond)
	if err := s.WriteEvent(endEv); err != nil {
		t.Fatalf("seedAgentForRun end %q/%q: %v", runID, agentID, err)
	}
}

// int64p retorna un puntero a int64 — útil para seeds de agentes.
func int64p(v int64) *int64 { return &v }

// metricsBaseTime es un tiempo de referencia fijo para seeds determinísticos.
var metricsBaseTime = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

// --- tests -------------------------------------------------------------------

func Test_GetMetrics(t *testing.T) {
	t.Run("db_vacia", func(t *testing.T) {
		s := newTestStore(t, 500)

		m, err := s.GetMetrics(context.Background())
		if err != nil {
			t.Fatalf("GetMetrics: %v", err)
		}

		if m.RunsCount != 0 {
			t.Errorf("RunsCount: esperado 0, obtuvo %d", m.RunsCount)
		}
		if m.HasEnoughData {
			t.Error("HasEnoughData: esperado false para DB vacía")
		}
		// Los slices deben ser no-nil aunque estén vacíos.
		if m.TokensPerRun == nil {
			t.Error("TokensPerRun: esperaba slice no-nil")
		}
		if m.AvgTimePerAgent == nil {
			t.Error("AvgTimePerAgent: esperaba slice no-nil")
		}
		if len(m.TokensPerRun) != 0 {
			t.Errorf("TokensPerRun len: esperado 0, obtuvo %d", len(m.TokensPerRun))
		}
		if len(m.AvgTimePerAgent) != 0 {
			t.Errorf("AvgTimePerAgent len: esperado 0, obtuvo %d", len(m.AvgTimePerAgent))
		}
		// Todos los deltas nil cuando no hay datos.
		if m.TotalTokensDeltaPct != nil {
			t.Error("TotalTokensDeltaPct: esperado nil para DB vacía")
		}
		if m.AvgDurationDeltaMs != nil {
			t.Error("AvgDurationDeltaMs: esperado nil para DB vacía")
		}
		if m.SuccessRateDelta != nil {
			t.Error("SuccessRateDelta: esperado nil para DB vacía")
		}
	})

	t.Run("un_run", func(t *testing.T) {
		s := newTestStore(t, 500)

		seedRun(t, s, "run-metrics-single", metricsBaseTime, "success", 5000, 1200)

		m, err := s.GetMetrics(context.Background())
		if err != nil {
			t.Fatalf("GetMetrics: %v", err)
		}

		if m.RunsCount != 1 {
			t.Errorf("RunsCount: esperado 1, obtuvo %d", m.RunsCount)
		}
		if m.HasEnoughData {
			t.Error("HasEnoughData: esperado false para 1 run")
		}
		if m.TotalTokens != 1200 {
			t.Errorf("TotalTokens: esperado 1200, obtuvo %d", m.TotalTokens)
		}
		if m.AvgDurationMs != 5000 {
			t.Errorf("AvgDurationMs: esperado 5000, obtuvo %d", m.AvgDurationMs)
		}
		if math.Abs(m.SuccessRate-1.0) > 1e-9 {
			t.Errorf("SuccessRate: esperado 1.0, obtuvo %f", m.SuccessRate)
		}
		// Sin prev → todos los deltas nil.
		if m.TotalTokensDeltaPct != nil {
			t.Error("TotalTokensDeltaPct: esperado nil (sin prev)")
		}
		if m.AvgDurationDeltaMs != nil {
			t.Error("AvgDurationDeltaMs: esperado nil (sin prev)")
		}
		if m.SuccessRateDelta != nil {
			t.Error("SuccessRateDelta: esperado nil (sin prev)")
		}
		// Slices no-nil.
		if m.TokensPerRun == nil {
			t.Error("TokensPerRun: esperaba slice no-nil")
		}
		if m.AvgTimePerAgent == nil {
			t.Error("AvgTimePerAgent: esperaba slice no-nil")
		}
	})

	t.Run("cuatro_runs", func(t *testing.T) {
		s := newTestStore(t, 500)

		for i := 0; i < 4; i++ {
			id := fmt.Sprintf("run-four-%c", 'a'+i)
			seedRun(t, s, id, metricsBaseTime.Add(time.Duration(i)*time.Hour), "success", 1000, 100*(i+1))
		}

		m, err := s.GetMetrics(context.Background())
		if err != nil {
			t.Fatalf("GetMetrics: %v", err)
		}

		if m.RunsCount != 4 {
			t.Errorf("RunsCount: esperado 4, obtuvo %d", m.RunsCount)
		}
		if m.HasEnoughData {
			t.Error("HasEnoughData: esperado false para 4 runs")
		}
		// Sin prev → deltas nil.
		if m.TotalTokensDeltaPct != nil {
			t.Error("TotalTokensDeltaPct: esperado nil (sin prev)")
		}
		// Slices no-nil.
		if m.TokensPerRun == nil {
			t.Error("TokensPerRun: esperaba slice no-nil")
		}
		if m.AvgTimePerAgent == nil {
			t.Error("AvgTimePerAgent: esperaba slice no-nil")
		}
		if len(m.TokensPerRun) != 4 {
			t.Errorf("TokensPerRun len: esperado 4, obtuvo %d", len(m.TokensPerRun))
		}
	})

	t.Run("cinco_runs_sin_prev", func(t *testing.T) {
		// AC1: ≥5 runs → HasEnoughData=true, charts con datos reales.
		s := newTestStore(t, 500)

		tokens := []int{100, 200, 300, 400, 500}
		for i, tok := range tokens {
			id := fmt.Sprintf("run-five-%c", 'a'+i)
			seedRun(t, s, id, metricsBaseTime.Add(time.Duration(i)*time.Hour), "success", 2000, tok)
		}

		m, err := s.GetMetrics(context.Background())
		if err != nil {
			t.Fatalf("GetMetrics: %v", err)
		}

		if m.RunsCount != 5 {
			t.Errorf("RunsCount: esperado 5, obtuvo %d", m.RunsCount)
		}
		if !m.HasEnoughData {
			t.Error("HasEnoughData: esperado true para 5 runs")
		}
		// Sin prev → todos los deltas nil.
		if m.TotalTokensDeltaPct != nil {
			t.Error("TotalTokensDeltaPct: esperado nil (sin prev)")
		}
		if m.AvgDurationDeltaMs != nil {
			t.Error("AvgDurationDeltaMs: esperado nil (sin prev)")
		}
		if m.SuccessRateDelta != nil {
			t.Error("SuccessRateDelta: esperado nil (sin prev)")
		}
		// TokensPerRun: 5 filas en orden cronológico ASC.
		if len(m.TokensPerRun) != 5 {
			t.Fatalf("TokensPerRun len: esperado 5, obtuvo %d", len(m.TokensPerRun))
		}
		for i := 1; i < len(m.TokensPerRun); i++ {
			if !m.TokensPerRun[i].StartedAt.After(m.TokensPerRun[i-1].StartedAt) {
				t.Errorf("TokensPerRun[%d].StartedAt no está en orden ASC: %v <= %v",
					i, m.TokensPerRun[i].StartedAt, m.TokensPerRun[i-1].StartedAt)
			}
		}
		// El más antiguo tiene tokens=100, el más reciente tokens=500.
		if m.TokensPerRun[0].Tokens != 100 {
			t.Errorf("TokensPerRun[0].Tokens: esperado 100 (run más antiguo), obtuvo %d", m.TokensPerRun[0].Tokens)
		}
		if m.TokensPerRun[4].Tokens != 500 {
			t.Errorf("TokensPerRun[4].Tokens: esperado 500 (run más reciente), obtuvo %d", m.TokensPerRun[4].Tokens)
		}
	})

	t.Run("treinta_mas_treinta", func(t *testing.T) {
		// AC1 y AC3: 30 runs actuales + 30 runs previos con valores redondos.
		// prev: 30 runs, duration=2000ms, tokens=100, todos success.
		//   → prevTotalTokens=3000, prevAvgDuration=2000, prevSuccessRate=1.0
		// cur: 30 runs, duration=3000ms, tokens=120, todos success.
		//   → curTotalTokens=3600, curAvgDuration=3000, curSuccessRate=1.0
		// Deltas esperados:
		//   TotalTokensDeltaPct = (3600-3000)/3000 = 0.2
		//   AvgDurationDeltaMs  = 3000-2000 = 1000
		//   SuccessRateDelta    = 1.0-1.0 = 0.0
		s := newTestStore(t, 500)

		prevBase := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 30; i++ {
			id := fmt.Sprintf("run-prev-%02d", i)
			seedRun(t, s, id, prevBase.Add(time.Duration(i)*time.Hour), "success", 2000, 100)
		}

		curBase := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 30; i++ {
			id := fmt.Sprintf("run-cur-%02d", i)
			seedRun(t, s, id, curBase.Add(time.Duration(i)*time.Hour), "success", 3000, 120)
		}

		m, err := s.GetMetrics(context.Background())
		if err != nil {
			t.Fatalf("GetMetrics: %v", err)
		}

		if m.RunsCount != 30 {
			t.Errorf("RunsCount: esperado 30, obtuvo %d", m.RunsCount)
		}
		if !m.HasEnoughData {
			t.Error("HasEnoughData: esperado true para 30 runs")
		}

		// TotalTokens = 30 * 120 = 3600
		if m.TotalTokens != 3600 {
			t.Errorf("TotalTokens: esperado 3600, obtuvo %d", m.TotalTokens)
		}

		// AvgDurationMs = 3000
		if m.AvgDurationMs != 3000 {
			t.Errorf("AvgDurationMs: esperado 3000, obtuvo %d", m.AvgDurationMs)
		}

		// SuccessRate = 1.0
		if math.Abs(m.SuccessRate-1.0) > 1e-9 {
			t.Errorf("SuccessRate: esperado 1.0, obtuvo %f", m.SuccessRate)
		}

		// TotalTokensDeltaPct = (3600-3000)/3000 = 0.2
		if m.TotalTokensDeltaPct == nil {
			t.Fatal("TotalTokensDeltaPct: esperado no-nil con 30 prev runs")
		}
		if math.Abs(*m.TotalTokensDeltaPct-0.2) > 1e-9 {
			t.Errorf("TotalTokensDeltaPct: esperado 0.2, obtuvo %f", *m.TotalTokensDeltaPct)
		}

		// AvgDurationDeltaMs = 3000 - 2000 = 1000
		if m.AvgDurationDeltaMs == nil {
			t.Fatal("AvgDurationDeltaMs: esperado no-nil con 30 prev runs")
		}
		if *m.AvgDurationDeltaMs != 1000 {
			t.Errorf("AvgDurationDeltaMs: esperado 1000, obtuvo %d", *m.AvgDurationDeltaMs)
		}

		// SuccessRateDelta = 0.0
		if m.SuccessRateDelta == nil {
			t.Fatal("SuccessRateDelta: esperado no-nil con 30 prev runs")
		}
		if math.Abs(*m.SuccessRateDelta-0.0) > 1e-9 {
			t.Errorf("SuccessRateDelta: esperado 0.0, obtuvo %f", *m.SuccessRateDelta)
		}

		// TokensPerRun: exactamente 30 filas en orden ASC.
		if len(m.TokensPerRun) != 30 {
			t.Fatalf("TokensPerRun len: esperado 30, obtuvo %d", len(m.TokensPerRun))
		}
		for i := 1; i < len(m.TokensPerRun); i++ {
			if !m.TokensPerRun[i].StartedAt.After(m.TokensPerRun[i-1].StartedAt) {
				t.Errorf("TokensPerRun[%d].StartedAt no está en orden ASC", i)
			}
		}
	})

	t.Run("prev_total_tokens_zero", func(t *testing.T) {
		// Guard división por cero: si prev.totalTokens==0, TotalTokensDeltaPct debe ser nil.
		// AvgDurationDeltaMs y SuccessRateDelta sí deben calcularse.
		s := newTestStore(t, 500)

		prevBase := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 30; i++ {
			id := fmt.Sprintf("run-prev-zero-%02d", i)
			seedRun(t, s, id, prevBase.Add(time.Duration(i)*time.Hour), "success", 1000, 0)
		}

		curBase := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 30; i++ {
			id := fmt.Sprintf("run-cur-zero-%02d", i)
			seedRun(t, s, id, curBase.Add(time.Duration(i)*time.Hour), "success", 2000, 500)
		}

		m, err := s.GetMetrics(context.Background())
		if err != nil {
			t.Fatalf("GetMetrics: %v", err)
		}

		// Guard: prevTotalTokens==0 → delta nil (evita división por cero).
		if m.TotalTokensDeltaPct != nil {
			t.Errorf("TotalTokensDeltaPct: esperado nil cuando prev.totalTokens==0, obtuvo %f", *m.TotalTokensDeltaPct)
		}

		// AvgDurationDeltaMs sí debe calcularse: 2000 - 1000 = 1000.
		if m.AvgDurationDeltaMs == nil {
			t.Fatal("AvgDurationDeltaMs: esperado no-nil")
		}
		if *m.AvgDurationDeltaMs != 1000 {
			t.Errorf("AvgDurationDeltaMs: esperado 1000, obtuvo %d", *m.AvgDurationDeltaMs)
		}

		// SuccessRateDelta sí debe calcularse (ambos períodos tienen rate=1.0 → delta=0.0).
		if m.SuccessRateDelta == nil {
			t.Fatal("SuccessRateDelta: esperado no-nil")
		}
		if math.Abs(*m.SuccessRateDelta-0.0) > 1e-9 {
			t.Errorf("SuccessRateDelta: esperado 0.0, obtuvo %f", *m.SuccessRateDelta)
		}
	})

	t.Run("mezcla_status", func(t *testing.T) {
		// AC3: runs "running" deben excluirse de todos los agregados.
		// 3 success + 1 failed + 2 running → RunsCount=4, SuccessRate=0.75.
		s := newTestStore(t, 500)

		base := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)

		seedRun(t, s, "run-mix-s1", base.Add(0*time.Hour), "success", 1000, 200)
		seedRun(t, s, "run-mix-s2", base.Add(1*time.Hour), "success", 1000, 200)
		seedRun(t, s, "run-mix-s3", base.Add(2*time.Hour), "success", 1000, 200)
		seedRun(t, s, "run-mix-f1", base.Add(3*time.Hour), "failed", 1000, 200)
		seedRunningRun(t, s, "run-mix-r1", base.Add(4*time.Hour))
		seedRunningRun(t, s, "run-mix-r2", base.Add(5*time.Hour))

		m, err := s.GetMetrics(context.Background())
		if err != nil {
			t.Fatalf("GetMetrics: %v", err)
		}

		// Solo los 4 runs terminados (success+failed) cuentan.
		if m.RunsCount != 4 {
			t.Errorf("RunsCount: esperado 4 (running excluidos), obtuvo %d", m.RunsCount)
		}

		// SuccessRate = 3 success / 4 terminados = 0.75.
		if math.Abs(m.SuccessRate-0.75) > 1e-9 {
			t.Errorf("SuccessRate: esperado 0.75, obtuvo %f", m.SuccessRate)
		}

		// TokensPerRun solo contiene los 4 terminados (no los running).
		if len(m.TokensPerRun) != 4 {
			t.Errorf("TokensPerRun len: esperado 4, obtuvo %d", len(m.TokensPerRun))
		}
	})

	t.Run("agents_duration_null", func(t *testing.T) {
		// Agentes sin agent.end tienen duration_ms IS NULL en la tabla agents.
		// Q_AVG_PER_AGENT los excluye con WHERE duration_ms IS NOT NULL.
		s := newTestStore(t, 500)

		runID := "run-agent-null-dur"
		seedRun(t, s, runID, metricsBaseTime, "success", 3000, 500)

		// Agente con duración conocida: 2000ms.
		seedAgentForRun(t, s, runID, "agent-known", "developer", metricsBaseTime, int64p(2000))
		// Agente sin agent.end → duration_ms permanece NULL.
		seedAgentForRun(t, s, runID, "agent-null", "pm", metricsBaseTime, nil)

		m, err := s.GetMetrics(context.Background())
		if err != nil {
			t.Fatalf("GetMetrics: %v", err)
		}

		if m.RunsCount != 1 {
			t.Errorf("RunsCount: esperado 1, obtuvo %d", m.RunsCount)
		}

		// Solo el agente con duración conocida debe aparecer en AvgTimePerAgent.
		if len(m.AvgTimePerAgent) != 1 {
			t.Fatalf("AvgTimePerAgent len: esperado 1 (agente null excluido), obtuvo %d", len(m.AvgTimePerAgent))
		}
		row := m.AvgTimePerAgent[0]
		if row.Role != "developer" {
			t.Errorf("AvgTimePerAgent[0].Role: esperado %q, obtuvo %q", "developer", row.Role)
		}
		if row.AvgDurationMs != 2000 {
			t.Errorf("AvgTimePerAgent[0].AvgDurationMs: esperado 2000, obtuvo %d", row.AvgDurationMs)
		}
		// RunsIncluded es COUNT(*) sobre agents del grupo con duration_ms IS NOT NULL.
		if row.RunsIncluded != 1 {
			t.Errorf("AvgTimePerAgent[0].RunsIncluded: esperado 1, obtuvo %d", row.RunsIncluded)
		}
	})

	t.Run("top_8_roles", func(t *testing.T) {
		// Con 9 roles distintos, AvgTimePerAgent debe truncar a top 8 por AvgDurationMs DESC.
		// El rol con menor duración (r1=100ms) debe quedar fuera.
		s := newTestStore(t, 500)

		runID := "run-top8-roles"
		seedRun(t, s, runID, metricsBaseTime, "success", 5000, 1000)

		// Roles con duraciones 100, 200, ..., 900ms.
		roles := []string{"r1", "r2", "r3", "r4", "r5", "r6", "r7", "r8", "r9"}
		for i, role := range roles {
			dur := int64((i + 1) * 100)
			agentID := "agent-" + role
			seedAgentForRun(t, s, runID, agentID, role, metricsBaseTime.Add(time.Duration(i)*time.Minute), int64p(dur))
		}

		m, err := s.GetMetrics(context.Background())
		if err != nil {
			t.Fatalf("GetMetrics: %v", err)
		}

		if len(m.AvgTimePerAgent) != 8 {
			t.Fatalf("AvgTimePerAgent len: esperado 8 (top 8 por duración), obtuvo %d", len(m.AvgTimePerAgent))
		}

		// El primero debe ser el rol con mayor duración (r9 = 900ms).
		if m.AvgTimePerAgent[0].Role != "r9" {
			t.Errorf("AvgTimePerAgent[0].Role: esperado %q (mayor duración), obtuvo %q", "r9", m.AvgTimePerAgent[0].Role)
		}
		if m.AvgTimePerAgent[0].AvgDurationMs != 900 {
			t.Errorf("AvgTimePerAgent[0].AvgDurationMs: esperado 900, obtuvo %d", m.AvgTimePerAgent[0].AvgDurationMs)
		}

		// r1 (100ms, el menor) debe quedar fuera del top 8.
		for _, row := range m.AvgTimePerAgent {
			if row.Role == "r1" {
				t.Error("AvgTimePerAgent: rol r1 (menor duración) no debería estar en el top 8")
			}
		}
	})

	t.Run("tokens_overflow", func(t *testing.T) {
		// 1 run con total_tokens cerca de int32 max (2_000_000_000).
		// Verifica que no hay overflow ni panic, y que el valor se preserva.
		s := newTestStore(t, 500)

		const bigTokens = 2_000_000_000
		seedRun(t, s, "run-overflow", metricsBaseTime, "success", 10000, bigTokens)

		m, err := s.GetMetrics(context.Background())
		if err != nil {
			t.Fatalf("GetMetrics: %v", err)
		}

		if m.TotalTokens != bigTokens {
			t.Errorf("TotalTokens: esperado %d, obtuvo %d", bigTokens, m.TotalTokens)
		}
		// Sin prev → delta nil, sin riesgo de NaN o panic.
		if m.TotalTokensDeltaPct != nil {
			t.Error("TotalTokensDeltaPct: esperado nil (sin prev)")
		}
	})

	t.Run("context_cancelado", func(t *testing.T) {
		s := newTestStore(t, 500)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancelar inmediatamente antes de llamar a GetMetrics

		_, err := s.GetMetrics(ctx)
		if err == nil {
			t.Fatal("GetMetrics: esperaba error con context cancelado, obtuvo nil")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("GetMetrics error: esperaba errors.Is(err, context.Canceled), obtuvo %v", err)
		}
		if !strings.Contains(err.Error(), "dashboard/store:") {
			t.Errorf("GetMetrics error: esperaba prefijo %q en el error, obtuvo %q", "dashboard/store:", err.Error())
		}
	})
}
