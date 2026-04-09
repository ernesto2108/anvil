package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/store"
	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// writeRunEnd escribe un evento run.end para el runID dado.
func writeRunEnd(t *testing.T, s *store.TestableStore, runID string, status string, durationMs int64, tokens int) {
	t.Helper()
	ev := mustNewEvent(t, runID, instrumentation.EventRunEnd, instrumentation.RunEndPayload{
		Status:          status,
		DurationMs:      durationMs,
		TotalTokens:     tokens,
		TotalFiles:      1,
		AgentsCompleted: 1,
		AgentsFailed:    0,
	})
	if err := s.WriteEvent(ev); err != nil {
		t.Fatalf("escribir run.end: %v", err)
	}
}

// writeQAScore escribe un evento qa.score para el runID dado.
func writeQAScore(t *testing.T, s *store.TestableStore, runID string, score float64) {
	t.Helper()
	ev := mustNewEvent(t, runID, instrumentation.EventQAScore, instrumentation.QAScorePayload{
		AgentID:  "agent-qa",
		Score:    int(score),
		MaxScore: 100,
	})
	if err := s.WriteEvent(ev); err != nil {
		t.Fatalf("escribir qa.score: %v", err)
	}
}

func Test_ListRuns(t *testing.T) {
	t.Run("store vacío retorna slice vacío", func(t *testing.T) {
		s := newTestStore(t, 500)

		got, err := s.ListRuns(context.Background(), 0, 0)
		if err != nil {
			t.Fatalf("ListRuns: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("esperaba 0 runs, obtuvo %d", len(got))
		}
	})

	t.Run("múltiples runs ordenados por started_at DESC", func(t *testing.T) {
		s := newTestStore(t, 500)

		runID1 := mustNewRunID(t)
		runID2 := mustNewRunID(t)
		runID3 := mustNewRunID(t)

		writeRunStart(t, s, runID1)
		time.Sleep(2 * time.Millisecond)
		writeRunStart(t, s, runID2)
		time.Sleep(2 * time.Millisecond)
		writeRunStart(t, s, runID3)

		writeRunEnd(t, s, runID1, "success", 100, 500)
		writeRunEnd(t, s, runID2, "success", 200, 600)
		writeRunEnd(t, s, runID3, "success", 300, 700)

		got, err := s.ListRuns(context.Background(), 0, 0)
		if err != nil {
			t.Fatalf("ListRuns: %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("esperaba 3 runs, obtuvo %d", len(got))
		}
		// El más reciente debe aparecer primero.
		if got[0].ID != runID3 {
			t.Errorf("primer run: esperado %q, obtuvo %q", runID3, got[0].ID)
		}
		if got[2].ID != runID1 {
			t.Errorf("último run: esperado %q, obtuvo %q", runID1, got[2].ID)
		}
	})

	t.Run("run en progreso tiene EndedAt y DurationMs nil", func(t *testing.T) {
		s := newTestStore(t, 500)

		runID := mustNewRunID(t)
		writeRunStart(t, s, runID)

		got, err := s.ListRuns(context.Background(), 0, 0)
		if err != nil {
			t.Fatalf("ListRuns: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("esperaba 1 run, obtuvo %d", len(got))
		}

		r := got[0]
		if r.Status != "running" {
			t.Errorf("status: esperado %q, obtuvo %q", "running", r.Status)
		}
		if r.EndedAt != nil {
			t.Errorf("EndedAt: esperaba nil para run en progreso, obtuvo %v", r.EndedAt)
		}
		if r.DurationMs != nil {
			t.Errorf("DurationMs: esperaba nil para run en progreso, obtuvo %v", r.DurationMs)
		}
	})

	t.Run("run con qa.score tiene QAScore no nil", func(t *testing.T) {
		s := newTestStore(t, 500)

		runID := mustNewRunID(t)
		writeRunStart(t, s, runID)
		writeRunEnd(t, s, runID, "success", 500, 1000)
		writeQAScore(t, s, runID, 87)

		got, err := s.ListRuns(context.Background(), 0, 0)
		if err != nil {
			t.Fatalf("ListRuns: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("esperaba 1 run, obtuvo %d", len(got))
		}
		if got[0].QAScore == nil {
			t.Fatal("QAScore: esperaba valor no nil")
		}
		if *got[0].QAScore != 87 {
			t.Errorf("QAScore: esperado 87, obtuvo %f", *got[0].QAScore)
		}
	})

	t.Run("limit y offset paginan correctamente", func(t *testing.T) {
		s := newTestStore(t, 500)

		for i := 0; i < 5; i++ {
			writeRunStart(t, s, mustNewRunID(t))
			time.Sleep(2 * time.Millisecond)
		}

		// Primera página: 2 runs más recientes.
		page1, err := s.ListRuns(context.Background(), 2, 0)
		if err != nil {
			t.Fatalf("ListRuns página 1: %v", err)
		}
		if len(page1) != 2 {
			t.Fatalf("página 1: esperado 2 runs, obtuvo %d", len(page1))
		}

		// Segunda página: siguientes 2.
		page2, err := s.ListRuns(context.Background(), 2, 2)
		if err != nil {
			t.Fatalf("ListRuns página 2: %v", err)
		}
		if len(page2) != 2 {
			t.Fatalf("página 2: esperado 2 runs, obtuvo %d", len(page2))
		}

		// Tercera página: el último.
		page3, err := s.ListRuns(context.Background(), 2, 4)
		if err != nil {
			t.Fatalf("ListRuns página 3: %v", err)
		}
		if len(page3) != 1 {
			t.Fatalf("página 3: esperado 1 run, obtuvo %d", len(page3))
		}

		// No debe haber IDs duplicados entre páginas.
		seen := map[string]bool{}
		for _, r := range append(append(page1, page2...), page3...) {
			if seen[r.ID] {
				t.Errorf("ID duplicado en paginación: %q", r.ID)
			}
			seen[r.ID] = true
		}
	})

	t.Run("contexto cancelado retorna error", func(t *testing.T) {
		s := newTestStore(t, 500)

		writeRunStart(t, s, mustNewRunID(t))

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancelar inmediatamente

		_, err := s.ListRuns(ctx, 0, 0)
		if err == nil {
			t.Fatal("esperaba error con contexto cancelado, obtuvo nil")
		}
	})
}
