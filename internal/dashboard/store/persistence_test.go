package store_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// ---------------------------------------------------------------------------
// DASH-TEST-002: Tests de persistencia — escritura, lectura y rotación
// ---------------------------------------------------------------------------

// --- ResolveRunBySession ---------------------------------------------------

func Test_ResolveRunBySession(t *testing.T) {
	t.Run("session existente retorna el run_id correcto", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		ev := mustNewEvent(t, runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
			TaskID:    "task-sess",
			Provider:  "anthropic",
			SessionID: "sess-abc-123",
		})
		if err := s.WriteEvent(ev); err != nil {
			t.Fatalf("WriteEvent: %v", err)
		}

		got, err := s.ResolveRunBySession("sess-abc-123")
		if err != nil {
			t.Fatalf("ResolveRunBySession: %v", err)
		}
		if got != runID {
			t.Errorf("esperado %q, obtuvo %q", runID, got)
		}
	})

	t.Run("session inexistente retorna string vacío sin error", func(t *testing.T) {
		s := newTestStore(t, 500)

		got, err := s.ResolveRunBySession("no-existe")
		if err != nil {
			t.Fatalf("ResolveRunBySession: %v", err)
		}
		if got != "" {
			t.Errorf("esperado string vacío, obtuvo %q", got)
		}
	})

	t.Run("run sin session_id no es encontrado por session vacía", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		// run.start sin SessionID → session_id queda NULL en la DB
		writeRunStart(t, s, runID)

		got, err := s.ResolveRunBySession("")
		if err != nil {
			t.Fatalf("ResolveRunBySession: %v", err)
		}
		if got != "" {
			t.Errorf("esperado string vacío para session vacía, obtuvo %q", got)
		}
	})

	t.Run("múltiples sessions retorna el run_id correcto para cada una", func(t *testing.T) {
		s := newTestStore(t, 500)

		runID1 := mustNewRunID(t)
		runID2 := mustNewRunID(t)

		ev1 := mustNewEvent(t, runID1, instrumentation.EventRunStart, instrumentation.RunStartPayload{
			TaskID:    "task-multi-sess",
			Provider:  "anthropic",
			SessionID: "session-aaa",
		})
		if err := s.WriteEvent(ev1); err != nil {
			t.Fatalf("WriteEvent run1: %v", err)
		}

		ev2 := mustNewEvent(t, runID2, instrumentation.EventRunStart, instrumentation.RunStartPayload{
			TaskID:    "task-multi-sess",
			Provider:  "anthropic",
			SessionID: "session-bbb",
		})
		if err := s.WriteEvent(ev2); err != nil {
			t.Fatalf("WriteEvent run2: %v", err)
		}

		got1, err := s.ResolveRunBySession("session-aaa")
		if err != nil {
			t.Fatalf("ResolveRunBySession session-aaa: %v", err)
		}
		if got1 != runID1 {
			t.Errorf("session-aaa: esperado %q, obtuvo %q", runID1, got1)
		}

		got2, err := s.ResolveRunBySession("session-bbb")
		if err != nil {
			t.Fatalf("ResolveRunBySession session-bbb: %v", err)
		}
		if got2 != runID2 {
			t.Errorf("session-bbb: esperado %q, obtuvo %q", runID2, got2)
		}
	})
}

// --- UpdateTaskDesc --------------------------------------------------------

func Test_UpdateTaskDesc(t *testing.T) {
	t.Run("actualiza task_desc cuando está vacío", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		ev := mustNewEvent(t, runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
			TaskID:          "task-upd",
			TaskDescription: "",
			Provider:        "anthropic",
		})
		if err := s.WriteEvent(ev); err != nil {
			t.Fatalf("WriteEvent: %v", err)
		}

		if err := s.UpdateTaskDesc(runID, "nueva descripción"); err != nil {
			t.Fatalf("UpdateTaskDesc: %v", err)
		}

		var desc string
		if err := s.DB().QueryRow("SELECT task_desc FROM runs WHERE id = ?", runID).Scan(&desc); err != nil {
			t.Fatalf("consultar task_desc: %v", err)
		}
		if desc != "nueva descripción" {
			t.Errorf("task_desc: esperado %q, obtuvo %q", "nueva descripción", desc)
		}
	})

	t.Run("no sobreescribe task_desc existente", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		ev := mustNewEvent(t, runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
			TaskID:          "task-upd",
			TaskDescription: "descripción original",
			Provider:        "anthropic",
		})
		if err := s.WriteEvent(ev); err != nil {
			t.Fatalf("WriteEvent: %v", err)
		}

		// Intentar sobreescribir — no debe cambiar.
		if err := s.UpdateTaskDesc(runID, "intento de sobreescritura"); err != nil {
			t.Fatalf("UpdateTaskDesc: %v", err)
		}

		var desc string
		if err := s.DB().QueryRow("SELECT task_desc FROM runs WHERE id = ?", runID).Scan(&desc); err != nil {
			t.Fatalf("consultar task_desc: %v", err)
		}
		if desc != "descripción original" {
			t.Errorf("task_desc: esperado %q (no sobreescrito), obtuvo %q", "descripción original", desc)
		}
	})

	t.Run("run inexistente no retorna error", func(t *testing.T) {
		s := newTestStore(t, 500)

		// UpdateTaskDesc usa UPDATE sin verificar rows affected — no falla por run inexistente.
		if err := s.UpdateTaskDesc("run-no-existe", "algo"); err != nil {
			t.Fatalf("UpdateTaskDesc para run inexistente: %v", err)
		}
	})
}

// --- ComputeRunTotals ------------------------------------------------------

func Test_ComputeRunTotals(t *testing.T) {
	t.Run("calcula files_touched y agents_count correctamente", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		writeRunStart(t, s, runID)

		// 2 agentes
		writeAgentStart(t, s, runID, "agent-1")
		ev1 := mustNewEvent(t, runID, instrumentation.EventAgentEnd, instrumentation.AgentEndPayload{
			AgentID:      "agent-1",
			Status:       "success",
			DurationMs:   500,
			TokensTotal:  100,
			FilesTouched: []string{"/a.go", "/b.go"},
			ExitCode:     0,
		})
		if err := s.WriteEvent(ev1); err != nil {
			t.Fatalf("agent.end 1: %v", err)
		}

		writeAgentStart(t, s, runID, "agent-2")
		ev2 := mustNewEvent(t, runID, instrumentation.EventAgentEnd, instrumentation.AgentEndPayload{
			AgentID:      "agent-2",
			Status:       "success",
			DurationMs:   300,
			TokensTotal:  50,
			FilesTouched: []string{"/a.go", "/c.go"}, // /a.go duplicado
			ExitCode:     0,
		})
		if err := s.WriteEvent(ev2); err != nil {
			t.Fatalf("agent.end 2: %v", err)
		}

		if err := s.ComputeRunTotals(runID); err != nil {
			t.Fatalf("ComputeRunTotals: %v", err)
		}

		var filesTouched, agentsCount int
		err := s.DB().QueryRow(
			"SELECT files_touched, agents_count FROM runs WHERE id = ?", runID,
		).Scan(&filesTouched, &agentsCount)
		if err != nil {
			t.Fatalf("consultar totales: %v", err)
		}

		// files: COUNT(DISTINCT path) = 3 (/a.go, /b.go, /c.go)
		if filesTouched != 3 {
			t.Errorf("files_touched: esperado 3 (distinct paths), obtuvo %d", filesTouched)
		}
		// agents: COUNT(*) = 2
		if agentsCount != 2 {
			t.Errorf("agents_count: esperado 2, obtuvo %d", agentsCount)
		}
	})

	t.Run("run sin agentes ni archivos retorna ceros", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		writeRunStart(t, s, runID)

		if err := s.ComputeRunTotals(runID); err != nil {
			t.Fatalf("ComputeRunTotals: %v", err)
		}

		var filesTouched, agentsCount int
		err := s.DB().QueryRow(
			"SELECT files_touched, agents_count FROM runs WHERE id = ?", runID,
		).Scan(&filesTouched, &agentsCount)
		if err != nil {
			t.Fatalf("consultar totales: %v", err)
		}
		if filesTouched != 0 {
			t.Errorf("files_touched: esperado 0, obtuvo %d", filesTouched)
		}
		if agentsCount != 0 {
			t.Errorf("agents_count: esperado 0, obtuvo %d", agentsCount)
		}
	})

	t.Run("run inexistente no retorna error", func(t *testing.T) {
		s := newTestStore(t, 500)

		// ComputeRunTotals sobre run no existente debe ser un no-op sin error.
		if err := s.ComputeRunTotals("run-no-existe"); err != nil {
			t.Fatalf("ComputeRunTotals para run inexistente: %v", err)
		}
	})

	t.Run("llamada idempotente no corrompe los totales", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		writeRunStart(t, s, runID)

		writeAgentStart(t, s, runID, "agent-idem")
		ev := mustNewEvent(t, runID, instrumentation.EventAgentEnd, instrumentation.AgentEndPayload{
			AgentID:      "agent-idem",
			Status:       "success",
			DurationMs:   400,
			TokensTotal:  80,
			FilesTouched: []string{"/x.go", "/y.go"},
			ExitCode:     0,
		})
		if err := s.WriteEvent(ev); err != nil {
			t.Fatalf("agent.end: %v", err)
		}

		// Llamar dos veces consecutivas.
		if err := s.ComputeRunTotals(runID); err != nil {
			t.Fatalf("ComputeRunTotals primera vez: %v", err)
		}
		if err := s.ComputeRunTotals(runID); err != nil {
			t.Fatalf("ComputeRunTotals segunda vez: %v", err)
		}

		var filesTouched, agentsCount int
		err := s.DB().QueryRow(
			"SELECT files_touched, agents_count FROM runs WHERE id = ?", runID,
		).Scan(&filesTouched, &agentsCount)
		if err != nil {
			t.Fatalf("consultar totales: %v", err)
		}
		// Debe conservar los valores correctos, no doblarlos.
		if filesTouched != 2 {
			t.Errorf("files_touched: esperado 2 (idempotente), obtuvo %d", filesTouched)
		}
		if agentsCount != 1 {
			t.Errorf("agents_count: esperado 1 (idempotente), obtuvo %d", agentsCount)
		}
	})
}

// --- Rotación edge cases ---------------------------------------------------

func Test_Rotate_EdgeCases(t *testing.T) {
	t.Run("count igual a maxRuns no elimina nada", func(t *testing.T) {
		s := newTestStore(t, 3)

		// Insertar exactamente 3 runs (== maxRuns)
		for i := 0; i < 3; i++ {
			runID := mustNewRunID(t)
			startEv := mustNewEvent(t, runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
				TaskID:   "task-edge",
				Provider: "anthropic",
			})
			startEv.Timestamp = time.Date(2024, 1, 1, 0, 0, i, 0, time.UTC)
			if err := s.WriteEvent(startEv); err != nil {
				t.Fatalf("run.start [%d]: %v", i, err)
			}

			endEv := mustNewEvent(t, runID, instrumentation.EventRunEnd, instrumentation.RunEndPayload{
				Status:     "success",
				DurationMs: 100,
			})
			endEv.Timestamp = time.Date(2024, 1, 1, 0, 0, i, 500000000, time.UTC)
			if err := s.WriteEvent(endEv); err != nil {
				t.Fatalf("run.end [%d]: %v", i, err)
			}
		}

		var count int
		if err := s.DB().QueryRow("SELECT COUNT(*) FROM runs").Scan(&count); err != nil {
			t.Fatalf("contar runs: %v", err)
		}
		if count != 3 {
			t.Errorf("runs: esperado 3 (sin rotación), obtuvo %d", count)
		}
	})

	t.Run("maxRuns 1 retiene solo el run más reciente", func(t *testing.T) {
		s := newTestStore(t, 1)

		var lastRunID string
		for i := 0; i < 3; i++ {
			runID := mustNewRunID(t)
			lastRunID = runID

			startEv := mustNewEvent(t, runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
				TaskID:   "task-edge",
				Provider: "anthropic",
			})
			startEv.Timestamp = time.Date(2024, 1, 1, 0, 0, i, 0, time.UTC)
			if err := s.WriteEvent(startEv); err != nil {
				t.Fatalf("run.start [%d]: %v", i, err)
			}

			endEv := mustNewEvent(t, runID, instrumentation.EventRunEnd, instrumentation.RunEndPayload{
				Status:     "success",
				DurationMs: 100,
			})
			endEv.Timestamp = time.Date(2024, 1, 1, 0, 0, i, 500000000, time.UTC)
			if err := s.WriteEvent(endEv); err != nil {
				t.Fatalf("run.end [%d]: %v", i, err)
			}
		}

		var count int
		if err := s.DB().QueryRow("SELECT COUNT(*) FROM runs").Scan(&count); err != nil {
			t.Fatalf("contar runs: %v", err)
		}
		if count != 1 {
			t.Errorf("runs: esperado 1, obtuvo %d", count)
		}

		// El que sobrevive debe ser el más reciente.
		var id string
		if err := s.DB().QueryRow("SELECT id FROM runs").Scan(&id); err != nil {
			t.Fatalf("consultar run sobreviviente: %v", err)
		}
		if id != lastRunID {
			t.Errorf("run sobreviviente: esperado %q (más reciente), obtuvo %q", lastRunID, id)
		}
	})

	t.Run("rotación no elimina runs running (solo se dispara en run.end)", func(t *testing.T) {
		s := newTestStore(t, 2)

		// Crear 3 runs: 2 completos + 1 running
		for i := 0; i < 2; i++ {
			runID := fmt.Sprintf("run-rot-complete-%d", i)
			startEv := mustNewEvent(t, runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
				TaskID:   "task-rot",
				Provider: "anthropic",
			})
			startEv.Timestamp = time.Date(2024, 1, 1, 0, 0, i, 0, time.UTC)
			if err := s.WriteEvent(startEv); err != nil {
				t.Fatalf("run.start [%d]: %v", i, err)
			}
			endEv := mustNewEvent(t, runID, instrumentation.EventRunEnd, instrumentation.RunEndPayload{
				Status:     "success",
				DurationMs: 100,
			})
			endEv.Timestamp = time.Date(2024, 1, 1, 0, 0, i, 500000000, time.UTC)
			if err := s.WriteEvent(endEv); err != nil {
				t.Fatalf("run.end [%d]: %v", i, err)
			}
		}

		// Tercer run: solo start (running) — no dispara rotación.
		startEv := mustNewEvent(t, "run-rot-running", instrumentation.EventRunStart, instrumentation.RunStartPayload{
			TaskID:   "task-rot",
			Provider: "anthropic",
		})
		startEv.Timestamp = time.Date(2024, 1, 1, 0, 0, 3, 0, time.UTC)
		if err := s.WriteEvent(startEv); err != nil {
			t.Fatalf("run.start running: %v", err)
		}

		// Deben existir 3 runs — rotación no se disparó porque el tercer run no tiene run.end.
		var count int
		if err := s.DB().QueryRow("SELECT COUNT(*) FROM runs").Scan(&count); err != nil {
			t.Fatalf("contar runs: %v", err)
		}
		if count != 3 {
			t.Errorf("runs: esperado 3 (rotación solo en run.end), obtuvo %d", count)
		}
	})
}

// --- run.start con session_id ----------------------------------------------

func Test_WriteEvent_RunStartWithSessionID(t *testing.T) {
	t.Run("session_id se persiste en la tabla runs", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		ev := mustNewEvent(t, runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
			TaskID:    "task-sess",
			Provider:  "anthropic",
			SessionID: "session-xyz-456",
		})
		if err := s.WriteEvent(ev); err != nil {
			t.Fatalf("WriteEvent: %v", err)
		}

		var sessionID string
		err := s.DB().QueryRow("SELECT session_id FROM runs WHERE id = ?", runID).Scan(&sessionID)
		if err != nil {
			t.Fatalf("consultar session_id: %v", err)
		}
		if sessionID != "session-xyz-456" {
			t.Errorf("session_id: esperado %q, obtuvo %q", "session-xyz-456", sessionID)
		}
	})

	t.Run("session_id vacío se persiste como NULL", func(t *testing.T) {
		s := newTestStore(t, 500)
		runID := mustNewRunID(t)

		ev := mustNewEvent(t, runID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
			TaskID:   "task-no-sess",
			Provider: "anthropic",
			// SessionID omitido → ""
		})
		if err := s.WriteEvent(ev); err != nil {
			t.Fatalf("WriteEvent: %v", err)
		}

		var sessionID *string
		err := s.DB().QueryRow("SELECT session_id FROM runs WHERE id = ?", runID).Scan(&sessionID)
		if err != nil {
			t.Fatalf("consultar session_id: %v", err)
		}
		if sessionID != nil {
			t.Errorf("session_id: esperado NULL, obtuvo %q", *sessionID)
		}
	})
}

// --- parent_run_id / relaciones padre-hijo -----------------------------------

func TestListRuns_ExcludesChildRuns(t *testing.T) {
	s := newTestStore(t, 500)
	ctx := context.Background()

	// Run padre (sin ParentRunID)
	parentID := mustNewRunID(t)
	evParent := mustNewEvent(t, parentID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
		TaskID:   "task-parent",
		Provider: "test",
	})
	if err := s.WriteEvent(evParent); err != nil {
		t.Fatalf("WriteEvent padre: %v", err)
	}

	// Run hijo (ParentRunID = parentID)
	childID := mustNewRunID(t)
	evChild := mustNewEvent(t, childID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
		TaskID:      "task-child",
		Provider:    "test",
		ParentRunID: parentID,
	})
	if err := s.WriteEvent(evChild); err != nil {
		t.Fatalf("WriteEvent hijo: %v", err)
	}

	// Otro run normal (sin ParentRunID)
	normalID := mustNewRunID(t)
	evNormal := mustNewEvent(t, normalID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
		TaskID:   "task-normal",
		Provider: "test",
	})
	if err := s.WriteEvent(evNormal); err != nil {
		t.Fatalf("WriteEvent normal: %v", err)
	}

	runs, err := s.ListRuns(ctx, 0, 0, "", "", "", "")
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}

	// Solo deben aparecer el padre y el normal — NO el hijo
	if len(runs) != 2 {
		t.Errorf("ListRuns: esperado 2 runs (padre + normal), obtuvo %d", len(runs))
	}

	// El run hijo NO debe aparecer en la lista
	for _, r := range runs {
		if r.ID == childID {
			t.Errorf("ListRuns: el run hijo %q no debería aparecer en la lista top-level", childID)
		}
	}

	// El padre debe tener ChildrenCount == 1
	var parentFound bool
	for _, r := range runs {
		if r.ID == parentID {
			parentFound = true
			if r.ChildrenCount != 1 {
				t.Errorf("padre ChildrenCount: esperado 1, obtuvo %d", r.ChildrenCount)
			}
		}
	}
	if !parentFound {
		t.Errorf("ListRuns: el run padre %q no aparece en la lista", parentID)
	}
}

func TestListChildRuns(t *testing.T) {
	s := newTestStore(t, 500)
	ctx := context.Background()

	// Run padre
	parentID := mustNewRunID(t)
	evParent := mustNewEvent(t, parentID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
		TaskID:   "task-parent",
		Provider: "test",
	})
	if err := s.WriteEvent(evParent); err != nil {
		t.Fatalf("WriteEvent padre: %v", err)
	}

	// Hijo 1
	child1ID := mustNewRunID(t)
	evChild1 := mustNewEvent(t, child1ID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
		TaskID:      "task-child-1",
		Provider:    "test",
		ParentRunID: parentID,
	})
	if err := s.WriteEvent(evChild1); err != nil {
		t.Fatalf("WriteEvent hijo 1: %v", err)
	}

	// Pequeña pausa para garantizar started_at distinto
	time.Sleep(2 * time.Millisecond)

	// Hijo 2
	child2ID := mustNewRunID(t)
	evChild2 := mustNewEvent(t, child2ID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
		TaskID:      "task-child-2",
		Provider:    "test",
		ParentRunID: parentID,
	})
	if err := s.WriteEvent(evChild2); err != nil {
		t.Fatalf("WriteEvent hijo 2: %v", err)
	}

	children, err := s.ListChildRuns(ctx, parentID)
	if err != nil {
		t.Fatalf("ListChildRuns: %v", err)
	}

	// Debe retornar exactamente 2 hijos
	if len(children) != 2 {
		t.Fatalf("ListChildRuns: esperado 2 hijos, obtuvo %d", len(children))
	}

	// Verificar orden ASC por started_at: el primero debe ser child1
	if children[0].ID != child1ID {
		t.Errorf("orden ASC: primer hijo esperado %q, obtuvo %q", child1ID, children[0].ID)
	}
	if children[1].ID != child2ID {
		t.Errorf("orden ASC: segundo hijo esperado %q, obtuvo %q", child2ID, children[1].ID)
	}

	// Cada hijo debe tener ParentRunID == parentID
	for _, c := range children {
		if c.ParentRunID != parentID {
			t.Errorf("hijo %q: ParentRunID esperado %q, obtuvo %q", c.ID, parentID, c.ParentRunID)
		}
	}
}

func TestGetRunSummary_WithChildrenCount(t *testing.T) {
	s := newTestStore(t, 500)
	ctx := context.Background()

	// Run padre
	parentID := mustNewRunID(t)
	evParent := mustNewEvent(t, parentID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
		TaskID:   "task-parent",
		Provider: "test",
	})
	if err := s.WriteEvent(evParent); err != nil {
		t.Fatalf("WriteEvent padre: %v", err)
	}

	// 3 hijos
	var lastChildID string
	for i := 1; i <= 3; i++ {
		childID := mustNewRunID(t)
		lastChildID = childID
		ev := mustNewEvent(t, childID, instrumentation.EventRunStart, instrumentation.RunStartPayload{
			TaskID:      fmt.Sprintf("task-child-%d", i),
			Provider:    "test",
			ParentRunID: parentID,
		})
		if err := s.WriteEvent(ev); err != nil {
			t.Fatalf("WriteEvent hijo %d: %v", i, err)
		}
	}

	// GetRunSummary del padre debe tener ChildrenCount == 3
	summary, err := s.GetRunSummary(ctx, parentID)
	if err != nil {
		t.Fatalf("GetRunSummary padre: %v", err)
	}
	if summary == nil {
		t.Fatal("GetRunSummary padre: esperaba RunSummary no nil")
	}
	if summary.ChildrenCount != 3 {
		t.Errorf("padre ChildrenCount: esperado 3, obtuvo %d", summary.ChildrenCount)
	}

	// GetRunSummary de un hijo debe tener ChildrenCount == 0 y ParentRunID == parentID
	childSummary, err := s.GetRunSummary(ctx, lastChildID)
	if err != nil {
		t.Fatalf("GetRunSummary hijo: %v", err)
	}
	if childSummary == nil {
		t.Fatal("GetRunSummary hijo: esperaba RunSummary no nil")
	}
	if childSummary.ChildrenCount != 0 {
		t.Errorf("hijo ChildrenCount: esperado 0, obtuvo %d", childSummary.ChildrenCount)
	}
	if childSummary.ParentRunID != parentID {
		t.Errorf("hijo ParentRunID: esperado %q, obtuvo %q", parentID, childSummary.ParentRunID)
	}
}
