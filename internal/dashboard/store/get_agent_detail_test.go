package store_test

import (
	"context"
	"testing"

	"github.com/ernesto2108/anvil/internal/dashboard/store"
	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// writeAgentEndWithFiles escribe un evento agent.end con FilesTouched configurables.
func writeAgentEndWithFiles(t *testing.T, s *store.TestableStore, runID, agentID, status string, durationMs int64, tokens int, files []string) {
	t.Helper()
	ev := mustNewEvent(t, runID, instrumentation.EventAgentEnd, instrumentation.AgentEndPayload{
		AgentID:      agentID,
		Status:       status,
		DurationMs:   durationMs,
		TokensTotal:  tokens,
		FilesTouched: files,
		ExitCode:     0,
	})
	if err := s.WriteEvent(ev); err != nil {
		t.Fatalf("escribir agent.end con archivos para %q: %v", agentID, err)
	}
}

// writeAgentError escribe un evento agent.error para simular un agente fallido.
func writeAgentError(t *testing.T, s *store.TestableStore, runID, agentID, errMsg string, durationMs int64) {
	t.Helper()
	ev := mustNewEvent(t, runID, instrumentation.EventAgentError, instrumentation.AgentErrorPayload{
		AgentID:    agentID,
		Error:      errMsg,
		DurationMs: durationMs,
	})
	if err := s.WriteEvent(ev); err != nil {
		t.Fatalf("escribir agent.error para %q: %v", agentID, err)
	}
}

func Test_GetAgentDetail(t *testing.T) {
	t.Run("agente inexistente retorna nil sin error", func(t *testing.T) {
		s := newTestStore(t, 500)

		detail, files, err := s.GetAgentDetail(context.Background(), "run-no", "agent-no")
		if err != nil {
			t.Fatalf("GetAgentDetail: %v", err)
		}
		if detail != nil {
			t.Errorf("detail: esperaba nil, obtuvo %+v", detail)
		}
		if files != nil {
			t.Errorf("files: esperaba nil, obtuvo %v", files)
		}
	})

	t.Run("agente existe sin archivos retorna detalle + slice vacío no-nil", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		writeRunStart(t, s, runID)
		writeAgentStartWithDeps(t, s, runID, "agent-a", "developer", 1, []string{})
		writeAgentEndWithFiles(t, s, runID, "agent-a", "success", 800, 1200, []string{})

		detail, files, err := s.GetAgentDetail(context.Background(), runID, "agent-a")
		if err != nil {
			t.Fatalf("GetAgentDetail: %v", err)
		}
		if detail == nil {
			t.Fatal("detail: esperaba valor no nil")
		}
		if files == nil {
			t.Error("files: esperaba slice no-nil (no nil) aunque esté vacío")
		}
		if len(files) != 0 {
			t.Errorf("files: esperaba 0 elementos, obtuvo %d", len(files))
		}
		// Verificar campos del detalle
		if detail.AgentID != "agent-a" {
			t.Errorf("AgentID: esperado %q, obtuvo %q", "agent-a", detail.AgentID)
		}
		if detail.AgentRole != "developer" {
			t.Errorf("AgentRole: esperado %q, obtuvo %q", "developer", detail.AgentRole)
		}
		if detail.Status != "success" {
			t.Errorf("Status: esperado %q, obtuvo %q", "success", detail.Status)
		}
		if detail.DurationMs == nil {
			t.Error("DurationMs: esperaba valor no nil para agente completado")
		} else if *detail.DurationMs != 800 {
			t.Errorf("DurationMs: esperado 800, obtuvo %d", *detail.DurationMs)
		}
		if detail.TokensTotal == nil {
			t.Error("TokensTotal: esperaba valor no nil para agente completado")
		} else if *detail.TokensTotal != 1200 {
			t.Errorf("TokensTotal: esperado 1200, obtuvo %d", *detail.TokensTotal)
		}
	})

	t.Run("agente completado con archivos retorna lista de archivos ordenada", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		writeRunStart(t, s, runID)
		writeAgentStartWithDeps(t, s, runID, "agent-files", "developer", 1, []string{})
		writeAgentEndWithFiles(t, s, runID, "agent-files", "success", 1000, 2000,
			[]string{"a.go", "b.go", "c.go"},
		)

		detail, files, err := s.GetAgentDetail(context.Background(), runID, "agent-files")
		if err != nil {
			t.Fatalf("GetAgentDetail: %v", err)
		}
		if detail == nil {
			t.Fatal("detail: esperaba valor no nil")
		}
		if len(files) != 3 {
			t.Fatalf("files: esperaba 3 elementos, obtuvo %d", len(files))
		}
		// Verificar paths — deben venir ordenados por id ASC (orden de inserción)
		expected := []string{"a.go", "b.go", "c.go"}
		for i, want := range expected {
			if files[i].Path != want {
				t.Errorf("files[%d].Path: esperado %q, obtuvo %q", i, want, files[i].Path)
			}
		}
		// Verificar que la operación es "touched" (agent.end usa FilesTouched)
		for i, f := range files {
			if f.Operation != "touched" {
				t.Errorf("files[%d].Operation: esperado %q, obtuvo %q", i, "touched", f.Operation)
			}
		}
	})

	t.Run("agente running sin agent.end tiene nullable DurationMs y EndedAt nil", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		writeRunStart(t, s, runID)
		writeAgentStartWithDeps(t, s, runID, "agent-running", "pm", 1, []string{})
		// NO se escribe agent.end — el agente sigue "running"

		detail, files, err := s.GetAgentDetail(context.Background(), runID, "agent-running")
		if err != nil {
			t.Fatalf("GetAgentDetail: %v", err)
		}
		if detail == nil {
			t.Fatal("detail: esperaba valor no nil para agente running")
		}
		if detail.Status != "running" {
			t.Errorf("Status: esperado %q, obtuvo %q", "running", detail.Status)
		}
		// DurationMs debe ser nil para agente sin agent.end
		if detail.DurationMs != nil {
			t.Errorf("DurationMs: esperaba nil para agente running, obtuvo %d", *detail.DurationMs)
		}
		// EndedAt debe ser nil para agente sin agent.end
		if detail.EndedAt != nil {
			t.Errorf("EndedAt: esperaba nil para agente running, obtuvo %v", detail.EndedAt)
		}
		// Nota: tokens_input/output/total tienen DEFAULT 0 en la migración, así que
		// registros recién insertados con agent.start pueden tener *int = &0 (no nil).
		// Aceptamos cualquiera de los dos valores según el estado real de la DB.
		// La migración define DEFAULT 0, pero el scan usa sql.NullInt64 — si la
		// columna es 0, puede devolverse como *int(0). Esto es comportamiento correcto.
		if files == nil {
			t.Error("files: esperaba slice no-nil para agente running")
		}
		if len(files) != 0 {
			t.Errorf("files: esperaba 0 archivos para agente running, obtuvo %d", len(files))
		}
	})

	t.Run("agente fallido tiene ErrorMsg poblado y status failed", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		writeRunStart(t, s, runID)
		writeAgentStartWithDeps(t, s, runID, "agent-err", "architect", 1, []string{})
		writeAgentError(t, s, runID, "agent-err", "boom", 5000)

		detail, files, err := s.GetAgentDetail(context.Background(), runID, "agent-err")
		if err != nil {
			t.Fatalf("GetAgentDetail: %v", err)
		}
		if detail == nil {
			t.Fatal("detail: esperaba valor no nil para agente fallido")
		}
		if detail.Status != "failed" {
			t.Errorf("Status: esperado %q, obtuvo %q", "failed", detail.Status)
		}
		if detail.ErrorMsg != "boom" {
			t.Errorf("ErrorMsg: esperado %q, obtuvo %q", "boom", detail.ErrorMsg)
		}
		// Agente fallido no tiene archivos
		if len(files) != 0 {
			t.Errorf("files: esperaba 0 archivos para agente fallido sin FilesTouched, obtuvo %d", len(files))
		}
	})

	t.Run("agente con archivos filtra por runID + agentID — no mezcla runs distintos", func(t *testing.T) {
		s := newTestStore(t, 500)

		runID1 := mustNewRunID(t)
		runID2 := mustNewRunID(t)

		// Run 1 con agent-1 y sus archivos
		writeRunStart(t, s, runID1)
		writeAgentStartWithDeps(t, s, runID1, "agent-1", "developer", 1, []string{})
		writeAgentEndWithFiles(t, s, runID1, "agent-1", "success", 500, 1000,
			[]string{"run1-file.go"},
		)

		// Run 2 con agent-2 y sus archivos
		writeRunStart(t, s, runID2)
		writeAgentStartWithDeps(t, s, runID2, "agent-2", "developer", 1, []string{})
		writeAgentEndWithFiles(t, s, runID2, "agent-2", "success", 500, 1000,
			[]string{"run2-file.go"},
		)

		// GetAgentDetail de run1/agent-1 → sólo debe ver run1-file.go
		detail1, files1, err := s.GetAgentDetail(context.Background(), runID1, "agent-1")
		if err != nil {
			t.Fatalf("GetAgentDetail run1: %v", err)
		}
		if detail1 == nil {
			t.Fatal("detail: esperaba valor no nil para run1/agent-1")
		}
		if len(files1) != 1 {
			t.Fatalf("files para run1/agent-1: esperaba 1, obtuvo %d", len(files1))
		}
		if files1[0].Path != "run1-file.go" {
			t.Errorf("files[0].Path: esperado %q, obtuvo %q", "run1-file.go", files1[0].Path)
		}

		// GetAgentDetail de run2/agent-2 → sólo debe ver run2-file.go
		detail2, files2, err := s.GetAgentDetail(context.Background(), runID2, "agent-2")
		if err != nil {
			t.Fatalf("GetAgentDetail run2: %v", err)
		}
		if detail2 == nil {
			t.Fatal("detail: esperaba valor no nil para run2/agent-2")
		}
		if len(files2) != 1 {
			t.Fatalf("files para run2/agent-2: esperaba 1, obtuvo %d", len(files2))
		}
		if files2[0].Path != "run2-file.go" {
			t.Errorf("files[0].Path: esperado %q, obtuvo %q", "run2-file.go", files2[0].Path)
		}
	})

	t.Run("agente con archivos filtra por runID + agentID — no mezcla agentes del mismo run", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		writeRunStart(t, s, runID)

		// agent-a con sus archivos
		writeAgentStartWithDeps(t, s, runID, "agent-a", "developer", 1, []string{})
		writeAgentEndWithFiles(t, s, runID, "agent-a", "success", 500, 1000,
			[]string{"file-a.go"},
		)

		// agent-b con sus archivos en el mismo run
		writeAgentStartWithDeps(t, s, runID, "agent-b", "tester", 2, []string{"agent-a"})
		writeAgentEndWithFiles(t, s, runID, "agent-b", "success", 600, 1200,
			[]string{"file-b.go"},
		)

		// GetAgentDetail de agent-a → sólo debe ver file-a.go
		detailA, filesA, err := s.GetAgentDetail(context.Background(), runID, "agent-a")
		if err != nil {
			t.Fatalf("GetAgentDetail agent-a: %v", err)
		}
		if detailA == nil {
			t.Fatal("detail: esperaba valor no nil para agent-a")
		}
		if len(filesA) != 1 {
			t.Fatalf("files para agent-a: esperaba 1, obtuvo %d", len(filesA))
		}
		if filesA[0].Path != "file-a.go" {
			t.Errorf("files[0].Path: esperado %q, obtuvo %q", "file-a.go", filesA[0].Path)
		}
	})

	t.Run("agente no existe pero otros sí en el run — retorna (nil, nil, nil)", func(t *testing.T) {
		// Valida que el filtro AND agent_id = ? funciona, no sólo WHERE run_id = ?
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		writeRunStart(t, s, runID)
		writeAgentStartWithDeps(t, s, runID, "agent-a", "developer", 1, []string{})
		writeAgentEndWithFiles(t, s, runID, "agent-a", "success", 500, 1000,
			[]string{"some-file.go"},
		)

		// Buscar un agente que NO existe en este run
		detail, files, err := s.GetAgentDetail(context.Background(), runID, "agent-no-existe")
		if err != nil {
			t.Fatalf("GetAgentDetail: %v", err)
		}
		if detail != nil {
			t.Errorf("detail: esperaba nil para agente inexistente, obtuvo %+v", detail)
		}
		if files != nil {
			t.Errorf("files: esperaba nil para agente inexistente, obtuvo %v", files)
		}
	})
}
