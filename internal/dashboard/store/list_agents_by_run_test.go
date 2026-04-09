package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/store"
	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// writeAgentEnd escribe un evento agent.end para el agentID dado.
func writeAgentEnd(t *testing.T, s *store.TestableStore, runID, agentID, status string, durationMs int64, tokens int) {
	t.Helper()
	ev := mustNewEvent(t, runID, instrumentation.EventAgentEnd, instrumentation.AgentEndPayload{
		AgentID:      agentID,
		Status:       status,
		DurationMs:   durationMs,
		TokensTotal:  tokens,
		FilesTouched: []string{},
		ExitCode:     0,
	})
	if err := s.WriteEvent(ev); err != nil {
		t.Fatalf("escribir agent.end: %v", err)
	}
}

// writeAgentStartWithDeps escribe un evento agent.start con dependencias explícitas.
func writeAgentStartWithDeps(t *testing.T, s *store.TestableStore, runID, agentID, role string, seq int, deps []string) {
	t.Helper()
	ev := mustNewEvent(t, runID, instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
		AgentID:   agentID,
		AgentRole: role,
		Sequence:  seq,
		DependsOn: deps,
		Model:     "claude-sonnet-4-6",
	})
	if err := s.WriteEvent(ev); err != nil {
		t.Fatalf("escribir agent.start para %q: %v", agentID, err)
	}
}

func Test_ListAgentsByRun(t *testing.T) {
	t.Run("run inexistente retorna nil sin error", func(t *testing.T) {
		s := newTestStore(t, 500)

		got, err := s.ListAgentsByRun(context.Background(), "run-no-existe")
		if err != nil {
			t.Fatalf("ListAgentsByRun: %v", err)
		}
		if got != nil {
			t.Errorf("esperaba nil, obtuvo slice con %d elementos", len(got))
		}
	})

	t.Run("run con 1 agente completed retorna 1 fila con datos correctos", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		writeRunStart(t, s, runID)
		writeAgentStartWithDeps(t, s, runID, "agent-dev", "developer", 1, []string{})
		writeAgentEnd(t, s, runID, "agent-dev", "success", 1500, 8432)

		got, err := s.ListAgentsByRun(context.Background(), runID)
		if err != nil {
			t.Fatalf("ListAgentsByRun: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("esperaba 1 agente, obtuvo %d", len(got))
		}

		r := got[0]
		if r.AgentID != "agent-dev" {
			t.Errorf("AgentID: esperado %q, obtuvo %q", "agent-dev", r.AgentID)
		}
		if r.AgentRole != "developer" {
			t.Errorf("AgentRole: esperado %q, obtuvo %q", "developer", r.AgentRole)
		}
		if r.Status != "success" {
			t.Errorf("Status: esperado %q, obtuvo %q", "success", r.Status)
		}
		if r.DurationMs == nil {
			t.Fatal("DurationMs: esperaba valor no nil")
		}
		if *r.DurationMs != 1500 {
			t.Errorf("DurationMs: esperado 1500, obtuvo %d", *r.DurationMs)
		}
		if r.TokensTotal == nil {
			t.Fatal("TokensTotal: esperaba valor no nil")
		}
		if *r.TokensTotal != 8432 {
			t.Errorf("TokensTotal: esperado 8432, obtuvo %d", *r.TokensTotal)
		}
	})

	t.Run("run con 3 agentes en secuencia PM->arch->dev retorna 3 filas ordenadas", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		writeRunStart(t, s, runID)

		// Insertar en orden secuencial con dependencias PM -> arch -> dev.
		writeAgentStartWithDeps(t, s, runID, "agent-pm", "pm", 1, []string{})
		writeAgentEnd(t, s, runID, "agent-pm", "success", 500, 1000)

		writeAgentStartWithDeps(t, s, runID, "agent-arch", "architect", 2, []string{"agent-pm"})
		writeAgentEnd(t, s, runID, "agent-arch", "success", 700, 2000)

		writeAgentStartWithDeps(t, s, runID, "agent-dev", "developer", 3, []string{"agent-arch"})
		writeAgentEnd(t, s, runID, "agent-dev", "success", 1200, 5000)

		got, err := s.ListAgentsByRun(context.Background(), runID)
		if err != nil {
			t.Fatalf("ListAgentsByRun: %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("esperaba 3 agentes, obtuvo %d", len(got))
		}

		// Verificar orden correcto (sequence ASC).
		expectedOrder := []string{"agent-pm", "agent-arch", "agent-dev"}
		for i, id := range expectedOrder {
			if got[i].AgentID != id {
				t.Errorf("posición %d: esperado %q, obtuvo %q", i, id, got[i].AgentID)
			}
		}

		// Verificar depends_on del segundo agente contiene agent-pm.
		if got[1].DependsOn == "" {
			t.Error("agent-arch: DependsOn no debería estar vacío")
		}
	})

	t.Run("agente failed en medio preserva status, agentes posteriores pending", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		writeRunStart(t, s, runID)

		writeAgentStartWithDeps(t, s, runID, "agent-pm", "pm", 1, []string{})
		writeAgentEnd(t, s, runID, "agent-pm", "success", 500, 1000)

		// Agente fallido — usar agent.error para simular fallo.
		startEv := mustNewEvent(t, runID, instrumentation.EventAgentStart, instrumentation.AgentStartPayload{
			AgentID:   "agent-arch",
			AgentRole: "architect",
			Sequence:  2,
			DependsOn: []string{"agent-pm"},
			Model:     "claude-sonnet-4-6",
		})
		if err := s.WriteEvent(startEv); err != nil {
			t.Fatalf("escribir agent.start: %v", err)
		}
		errEv := mustNewEvent(t, runID, instrumentation.EventAgentError, instrumentation.AgentErrorPayload{
			AgentID:    "agent-arch",
			Error:      "timeout",
			DurationMs: 30000,
		})
		if err := s.WriteEvent(errEv); err != nil {
			t.Fatalf("escribir agent.error: %v", err)
		}

		// Agente pendiente (nunca llegó a correr).
		writeAgentStartWithDeps(t, s, runID, "agent-dev", "developer", 3, []string{"agent-arch"})

		got, err := s.ListAgentsByRun(context.Background(), runID)
		if err != nil {
			t.Fatalf("ListAgentsByRun: %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("esperaba 3 agentes, obtuvo %d", len(got))
		}

		if got[0].Status != "success" {
			t.Errorf("agent-pm Status: esperado %q, obtuvo %q", "success", got[0].Status)
		}
		if got[1].Status != "failed" {
			t.Errorf("agent-arch Status: esperado %q, obtuvo %q", "failed", got[1].Status)
		}
		if got[2].Status != "running" {
			// agent.start pone status=running; no llegó a agent.end.
			t.Errorf("agent-dev Status: esperado %q, obtuvo %q", "running", got[2].Status)
		}
	})

	t.Run("agente running tiene DurationMs nil", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		writeRunStart(t, s, runID)
		writeAgentStartWithDeps(t, s, runID, "agent-running", "developer", 1, []string{})
		// No se escribe agent.end — el agente sigue corriendo.

		got, err := s.ListAgentsByRun(context.Background(), runID)
		if err != nil {
			t.Fatalf("ListAgentsByRun: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("esperaba 1 agente, obtuvo %d", len(got))
		}

		if got[0].DurationMs != nil {
			t.Errorf("DurationMs: esperaba nil para agente en ejecución, obtuvo %v", got[0].DurationMs)
		}
		if got[0].Status != "running" {
			t.Errorf("Status: esperado %q, obtuvo %q", "running", got[0].Status)
		}
	})

	t.Run("contexto cancelado retorna error", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)
		writeRunStart(t, s, runID)
		writeAgentStartWithDeps(t, s, runID, "agent-x", "developer", 1, []string{})

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := s.ListAgentsByRun(ctx, runID)
		if err == nil {
			t.Fatal("esperaba error con contexto cancelado, obtuvo nil")
		}
	})

	t.Run("run de otro run no aparece", func(t *testing.T) {
		s := newTestStore(t, 500)

		runID1 := mustNewRunID(t)
		runID2 := mustNewRunID(t)

		writeRunStart(t, s, runID1)
		writeAgentStartWithDeps(t, s, runID1, "agent-r1", "developer", 1, []string{})

		writeRunStart(t, s, runID2)
		writeAgentStartWithDeps(t, s, runID2, "agent-r2", "developer", 1, []string{})
		time.Sleep(time.Millisecond) // ensure ordering

		got, err := s.ListAgentsByRun(context.Background(), runID1)
		if err != nil {
			t.Fatalf("ListAgentsByRun: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("esperaba 1 agente para runID1, obtuvo %d", len(got))
		}
		if got[0].AgentID != "agent-r1" {
			t.Errorf("AgentID: esperado %q, obtuvo %q", "agent-r1", got[0].AgentID)
		}
	})

	// AC7: agentes con sequence NULL se ordenan por id ASC (orden de inserción) como fallback.
	// Esto ocurre cuando el campo sequence no está seteado en la BD.
	// NOTA: 'id' en la query es el INTEGER PRIMARY KEY AUTOINCREMENT, no agent_id.
	// Por tanto el orden es de inserción, no lexicográfico por agent_id.
	t.Run("agentes con sequence NULL se ordenan por orden de inserción (id ASC)", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		writeRunStart(t, s, runID)

		// Insertar directamente con sequence NULL — primer registro obtiene id menor.
		// agent-first se inserta primero → id más bajo → aparece primero.
		_, err := s.DB().Exec(
			`INSERT INTO agents (run_id, agent_id, agent_role, sequence, depends_on, status)
			 VALUES (?, 'agent-first', 'tester', NULL, '[]', 'running'),
			        (?, 'agent-second', 'pm', NULL, '[]', 'running')`,
			runID, runID,
		)
		if err != nil {
			t.Fatalf("insertar agentes con sequence NULL: %v", err)
		}

		got, err := s.ListAgentsByRun(context.Background(), runID)
		if err != nil {
			t.Fatalf("ListAgentsByRun: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("esperaba 2 agentes, obtuvo %d", len(got))
		}

		// Con sequence NULL, el ORDER BY sequence ASC, id ASC ordena por id ASC (orden de inserción).
		if got[0].AgentID != "agent-first" {
			t.Errorf("posición 0: esperado %q, obtuvo %q (debe ordenar por id de inserción cuando sequence es NULL)", "agent-first", got[0].AgentID)
		}
		if got[1].AgentID != "agent-second" {
			t.Errorf("posición 1: esperado %q, obtuvo %q", "agent-second", got[1].AgentID)
		}

		// Verificar que Sequence es nil para ambos.
		for i, r := range got {
			if r.Sequence != nil {
				t.Errorf("posición %d: esperaba Sequence nil, obtuvo %d", i, *r.Sequence)
			}
		}
	})

	// Sequence duplicado: dos agentes con el mismo número de secuencia.
	// El desempate debe ser por id ASC (orden de inserción en AUTOINCREMENT).
	t.Run("sequence duplicado desempata por orden de inserción (id ASC)", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		writeRunStart(t, s, runID)

		// Ambos con sequence=1 — el store debería desempatar por id ASC (orden de inserción).
		// agent-inserted-first se inserta primero → id menor → aparece primero.
		_, err := s.DB().Exec(
			`INSERT INTO agents (run_id, agent_id, agent_role, sequence, depends_on, status)
			 VALUES (?, 'agent-inserted-first', 'tester', 1, '[]', 'running'),
			        (?, 'agent-inserted-second', 'developer', 1, '[]', 'running')`,
			runID, runID,
		)
		if err != nil {
			t.Fatalf("insertar agentes con sequence duplicado: %v", err)
		}

		got, err := s.ListAgentsByRun(context.Background(), runID)
		if err != nil {
			t.Fatalf("ListAgentsByRun: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("esperaba 2 agentes, obtuvo %d", len(got))
		}

		// El primero insertado obtiene id más bajo → aparece primero.
		if got[0].AgentID != "agent-inserted-first" {
			t.Errorf("posición 0: esperado %q, obtuvo %q (desempate por orden de inserción)", "agent-inserted-first", got[0].AgentID)
		}
		if got[1].AgentID != "agent-inserted-second" {
			t.Errorf("posición 1: esperado %q, obtuvo %q", "agent-inserted-second", got[1].AgentID)
		}
	})

	// Verificar que TokensTotal también se lee correctamente en el scan.
	t.Run("agente con tokens retorna TokensTotal correcto", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		writeRunStart(t, s, runID)
		writeAgentStartWithDeps(t, s, runID, "agent-tokens", "developer", 1, []string{})
		writeAgentEnd(t, s, runID, "agent-tokens", "success", 2000, 9999)

		got, err := s.ListAgentsByRun(context.Background(), runID)
		if err != nil {
			t.Fatalf("ListAgentsByRun: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("esperaba 1 agente, obtuvo %d", len(got))
		}
		if got[0].TokensTotal == nil {
			t.Fatal("TokensTotal: esperaba valor no nil")
		}
		if *got[0].TokensTotal != 9999 {
			t.Errorf("TokensTotal: esperado 9999, obtuvo %d", *got[0].TokensTotal)
		}
	})
}
