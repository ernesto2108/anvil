//go:build dashboard

package store_test

import (
	"context"
	"testing"

	"github.com/ernesto2108/anvil/internal/dashboard/store"
)

// setRunStartedAt sobreescribe el started_at de un run en la DB para controlar
// la fecha exacta sin depender de time.Sleep.
func setRunStartedAt(t *testing.T, s *store.TestableStore, runID, ts string) {
	t.Helper()
	_, err := s.DB().Exec("UPDATE runs SET started_at = ? WHERE id = ?", ts, runID)
	if err != nil {
		t.Fatalf("setRunStartedAt(%q, %q): %v", runID, ts, err)
	}
}

// setupTresRuns crea tres runs con fechas distintas y retorna sus IDs en orden
// cronológico: antiguo (dia1), medio (dia2), reciente (dia3).
func setupTresRuns(t *testing.T) (s *store.TestableStore, runAntiguo, runMedio, runReciente string) {
	t.Helper()
	s = newTestStore(t, 500)

	runAntiguo = mustNewRunID(t)
	runMedio = mustNewRunID(t)
	runReciente = mustNewRunID(t)

	writeRunStart(t, s, runAntiguo)
	writeRunEnd(t, s, runAntiguo, "success", 100, 100)
	setRunStartedAt(t, s, runAntiguo, "2026-01-10T00:00:00Z")

	writeRunStart(t, s, runMedio)
	writeRunEnd(t, s, runMedio, "success", 200, 200)
	setRunStartedAt(t, s, runMedio, "2026-01-20T00:00:00Z")

	writeRunStart(t, s, runReciente)
	writeRunEnd(t, s, runReciente, "success", 300, 300)
	setRunStartedAt(t, s, runReciente, "2026-01-30T00:00:00Z")

	return
}

// idsDeRuns extrae los IDs de una lista de RunSummary en un mapa para búsquedas O(1).
func idsDeRuns(runs []store.RunSummary) map[string]bool {
	m := make(map[string]bool, len(runs))
	for _, r := range runs {
		m[r.ID] = true
	}
	return m
}

// ---------------------------------------------------------------------------
// Tests de filtro por status
// ---------------------------------------------------------------------------

func Test_ListRuns_FiltroStatus(t *testing.T) {
	t.Run("filtro por status success retorna solo runs completados exitosamente", func(t *testing.T) {
		s := newTestStore(t, 500)
		ctx := context.Background()

		runSuccess1 := mustNewRunID(t)
		runSuccess2 := mustNewRunID(t)
		runFailed := mustNewRunID(t)
		runRunning := mustNewRunID(t)

		writeRunStart(t, s, runSuccess1)
		writeRunEnd(t, s, runSuccess1, "success", 100, 500)

		writeRunStart(t, s, runSuccess2)
		writeRunEnd(t, s, runSuccess2, "success", 200, 600)

		writeRunStart(t, s, runFailed)
		writeRunEnd(t, s, runFailed, "failed", 150, 300)

		writeRunStart(t, s, runRunning)

		got, err := s.ListRuns(ctx, 0, 0, "success", "", "")
		if err != nil {
			t.Fatalf("ListRuns con status=success: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("esperaba 2 runs con status=success, obtuvo %d", len(got))
		}
		for _, r := range got {
			if r.Status != "success" {
				t.Errorf("run %q: esperaba status %q, obtuvo %q", r.ID, "success", r.Status)
			}
		}
	})

	t.Run("filtro por status failed excluye runs success y running", func(t *testing.T) {
		s := newTestStore(t, 500)
		ctx := context.Background()

		runFailed1 := mustNewRunID(t)
		runFailed2 := mustNewRunID(t)
		runSuccess := mustNewRunID(t)
		runRunning := mustNewRunID(t)

		writeRunStart(t, s, runFailed1)
		writeRunEnd(t, s, runFailed1, "failed", 100, 100)

		writeRunStart(t, s, runFailed2)
		writeRunEnd(t, s, runFailed2, "failed", 200, 200)

		writeRunStart(t, s, runSuccess)
		writeRunEnd(t, s, runSuccess, "success", 300, 300)

		writeRunStart(t, s, runRunning)

		got, err := s.ListRuns(ctx, 0, 0, "failed", "", "")
		if err != nil {
			t.Fatalf("ListRuns con status=failed: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("esperaba 2 runs con status=failed, obtuvo %d", len(got))
		}
		for _, r := range got {
			if r.Status != "failed" {
				t.Errorf("run %q: esperaba status %q, obtuvo %q", r.ID, "failed", r.Status)
			}
		}
	})

	t.Run("filtro por status running retorna solo runs en progreso", func(t *testing.T) {
		s := newTestStore(t, 500)
		ctx := context.Background()

		runRunning := mustNewRunID(t)
		runSuccess := mustNewRunID(t)

		writeRunStart(t, s, runRunning)
		writeRunStart(t, s, runSuccess)
		writeRunEnd(t, s, runSuccess, "success", 100, 100)

		got, err := s.ListRuns(ctx, 0, 0, "running", "", "")
		if err != nil {
			t.Fatalf("ListRuns con status=running: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("esperaba 1 run con status=running, obtuvo %d", len(got))
		}
		if got[0].ID != runRunning {
			t.Errorf("ID: esperado %q, obtuvo %q", runRunning, got[0].ID)
		}
	})

	t.Run("filtro por status vacío retorna todos los runs sin filtrar", func(t *testing.T) {
		s := newTestStore(t, 500)
		ctx := context.Background()

		for _, status := range []string{"success", "failed", "success"} {
			id := mustNewRunID(t)
			writeRunStart(t, s, id)
			writeRunEnd(t, s, id, status, 100, 100)
		}
		writeRunStart(t, s, mustNewRunID(t)) // running

		got, err := s.ListRuns(ctx, 0, 0, "", "", "")
		if err != nil {
			t.Fatalf("ListRuns sin filtros: %v", err)
		}
		if len(got) != 4 {
			t.Fatalf("esperaba 4 runs sin filtro, obtuvo %d", len(got))
		}
	})

	t.Run("filtro por status con 0 coincidencias retorna slice vacío sin error", func(t *testing.T) {
		s := newTestStore(t, 500)
		ctx := context.Background()

		id := mustNewRunID(t)
		writeRunStart(t, s, id)
		writeRunEnd(t, s, id, "success", 100, 100)

		got, err := s.ListRuns(ctx, 0, 0, "failed", "", "")
		if err != nil {
			t.Fatalf("ListRuns sin coincidencias: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("esperaba 0 runs, obtuvo %d", len(got))
		}
	})
}

// ---------------------------------------------------------------------------
// Tests de filtro por fecha (startDate / endDate)
// ---------------------------------------------------------------------------

func Test_ListRuns_FiltroFecha(t *testing.T) {
	t.Run("startDate filtra runs anteriores a esa fecha", func(t *testing.T) {
		s, _, _, runReciente := setupTresRuns(t)
		ctx := context.Background()

		// Solo runs desde 2026-01-30 inclusive.
		got, err := s.ListRuns(ctx, 0, 0, "", "2026-01-30T00:00:00Z", "")
		if err != nil {
			t.Fatalf("ListRuns con startDate: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("esperaba 1 run, obtuvo %d", len(got))
		}
		if got[0].ID != runReciente {
			t.Errorf("ID: esperado %q (runReciente), obtuvo %q", runReciente, got[0].ID)
		}
	})

	t.Run("endDate filtra runs posteriores a esa fecha", func(t *testing.T) {
		s, runAntiguo, _, _ := setupTresRuns(t)
		ctx := context.Background()

		// Solo runs hasta fin del día 2026-01-10.
		got, err := s.ListRuns(ctx, 0, 0, "", "", "2026-01-10T23:59:59.999999999Z")
		if err != nil {
			t.Fatalf("ListRuns con endDate: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("esperaba 1 run, obtuvo %d", len(got))
		}
		if got[0].ID != runAntiguo {
			t.Errorf("ID: esperado %q (runAntiguo), obtuvo %q", runAntiguo, got[0].ID)
		}
	})

	t.Run("endDate mismo día que un run lo incluye gracias a 23:59:59.999999999", func(t *testing.T) {
		s, _, runMedio, _ := setupTresRuns(t)
		ctx := context.Background()

		// startDate y endDate exactamente en el día de runMedio (2026-01-20).
		got, err := s.ListRuns(ctx, 0, 0, "", "2026-01-20T00:00:00Z", "2026-01-20T23:59:59.999999999Z")
		if err != nil {
			t.Fatalf("ListRuns endDate mismo día: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("esperaba 1 run (runMedio), obtuvo %d", len(got))
		}
		if got[0].ID != runMedio {
			t.Errorf("ID: esperado %q (runMedio), obtuvo %q", runMedio, got[0].ID)
		}
	})

	t.Run("startDate y endDate combinados retornan rango correcto", func(t *testing.T) {
		s, _, runMedio, runReciente := setupTresRuns(t)
		ctx := context.Background()

		// Rango: 2026-01-20 a fin de 2026-01-30 — excluye runAntiguo.
		got, err := s.ListRuns(ctx, 0, 0, "", "2026-01-20T00:00:00Z", "2026-01-30T23:59:59.999999999Z")
		if err != nil {
			t.Fatalf("ListRuns con rango de fechas: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("esperaba 2 runs en rango, obtuvo %d", len(got))
		}
		ids := idsDeRuns(got)
		if !ids[runMedio] {
			t.Errorf("esperaba runMedio (%q) en resultado, IDs obtenidos: %v", runMedio, ids)
		}
		if !ids[runReciente] {
			t.Errorf("esperaba runReciente (%q) en resultado, IDs obtenidos: %v", runReciente, ids)
		}
	})

	t.Run("startDate mayor al run más reciente retorna slice vacío", func(t *testing.T) {
		s, _, _, _ := setupTresRuns(t)
		ctx := context.Background()

		got, err := s.ListRuns(ctx, 0, 0, "", "2030-01-01T00:00:00Z", "")
		if err != nil {
			t.Fatalf("ListRuns con startDate futuro: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("esperaba 0 runs, obtuvo %d", len(got))
		}
	})

	t.Run("endDate anterior al run más antiguo retorna slice vacío", func(t *testing.T) {
		s, _, _, _ := setupTresRuns(t)
		ctx := context.Background()

		got, err := s.ListRuns(ctx, 0, 0, "", "", "2020-01-01T23:59:59.999999999Z")
		if err != nil {
			t.Fatalf("ListRuns con endDate pasado: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("esperaba 0 runs, obtuvo %d", len(got))
		}
	})
}

// ---------------------------------------------------------------------------
// Tests de filtros combinados (status + startDate + endDate)
// ---------------------------------------------------------------------------

func Test_ListRuns_FiltrosCombinados(t *testing.T) {
	t.Run("status + startDate + endDate retornan intersección correcta", func(t *testing.T) {
		s := newTestStore(t, 500)
		ctx := context.Background()

		const dia = "2026-03-15T00:00:00Z"
		const finDia = "2026-03-15T23:59:59.999999999Z"
		const fuera = "2026-04-01T00:00:00Z"

		// Run success dentro del rango — debe aparecer.
		runOK := mustNewRunID(t)
		writeRunStart(t, s, runOK)
		writeRunEnd(t, s, runOK, "success", 100, 100)
		setRunStartedAt(t, s, runOK, dia)

		// Run failed dentro del rango — excluido por status.
		runFailed := mustNewRunID(t)
		writeRunStart(t, s, runFailed)
		writeRunEnd(t, s, runFailed, "failed", 200, 200)
		setRunStartedAt(t, s, runFailed, dia)

		// Run success fuera del rango — excluido por fecha.
		runFuera := mustNewRunID(t)
		writeRunStart(t, s, runFuera)
		writeRunEnd(t, s, runFuera, "success", 300, 300)
		setRunStartedAt(t, s, runFuera, fuera)

		got, err := s.ListRuns(ctx, 0, 0, "success", dia, finDia)
		if err != nil {
			t.Fatalf("ListRuns con 3 filtros: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("esperaba 1 run con los 3 filtros, obtuvo %d", len(got))
		}
		if got[0].ID != runOK {
			t.Errorf("ID: esperado %q, obtuvo %q", runOK, got[0].ID)
		}
		if got[0].Status != "success" {
			t.Errorf("status: esperado %q, obtuvo %q", "success", got[0].Status)
		}
	})

	t.Run("status + fechas con 0 coincidencias retorna slice vacío sin error", func(t *testing.T) {
		s := newTestStore(t, 500)
		ctx := context.Background()

		id := mustNewRunID(t)
		writeRunStart(t, s, id)
		writeRunEnd(t, s, id, "success", 100, 100)
		setRunStartedAt(t, s, id, "2026-05-10T00:00:00Z")

		got, err := s.ListRuns(ctx, 0, 0, "failed", "2026-01-01T00:00:00Z", "2026-01-31T23:59:59.999999999Z")
		if err != nil {
			t.Fatalf("ListRuns con filtros sin coincidencias: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("esperaba 0 runs, obtuvo %d", len(got))
		}
	})

	t.Run("filtros vacíos no cambian el comportamiento base sin regresión", func(t *testing.T) {
		s, _, _, _ := setupTresRuns(t)
		ctx := context.Background()

		got, err := s.ListRuns(ctx, 0, 0, "", "", "")
		if err != nil {
			t.Fatalf("ListRuns sin filtros (no-regresión): %v", err)
		}
		if len(got) != 3 {
			t.Fatalf("esperaba 3 runs sin filtros, obtuvo %d", len(got))
		}
	})
}
