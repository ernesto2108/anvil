//go:build dashboard

package dashboard

import (
	"context"

	"github.com/ernesto2108/anvil/internal/dashboard/store"
	"github.com/ernesto2108/anvil/internal/instrumentation"
)

// Store es la interfaz de lectura/escritura que usa el paquete dashboard.
// *store.SQLiteStore satisface esta interfaz.
type Store interface {
	// WriteEvent persiste un evento de instrumentación en el almacenamiento subyacente.
	WriteEvent(ev instrumentation.Event) error
	// ListRuns retorna runs paginados ordenados por started_at DESC.
	// status, startDate y endDate son filtros opcionales (cadena vacía = sin filtro).
	ListRuns(ctx context.Context, limit, offset int, status, startDate, endDate, project string) ([]store.RunSummary, error)
	// ListAgentsByRun retorna todos los agentes de un run ordenados por sequence ASC.
	// Retorna nil, nil si el run no existe o no tiene agentes.
	ListAgentsByRun(ctx context.Context, runID string) ([]store.AgentRow, error)
	// GetAgentDetail retorna el detalle completo de un agente (runID + agentID) junto con sus archivos.
	// Retorna (nil, nil, nil) si el agente no existe — NO es un error.
	GetAgentDetail(ctx context.Context, runID, agentID string) (*store.AgentDetail, []store.FileRow, error)
	// ListProjects returns distinct non-empty project names.
	ListProjects(ctx context.Context) ([]string, error)
	// ListFilesByRun returns all files touched in a run, ordered by id ASC.
	ListFilesByRun(ctx context.Context, runID string) ([]store.FileRow, error)
	// GetRunSummary retorna el RunSummary del run identificado por runID.
	// Retorna (nil, nil) si el run no existe — NO es un error.
	GetRunSummary(ctx context.Context, runID string) (*store.RunSummary, error)
	// ListChildRuns retorna los runs cuyo parent_run_id coincide con parentRunID.
	// Ordenados por started_at ASC (orden cronológico dentro del run padre).
	ListChildRuns(ctx context.Context, parentRunID string) ([]store.RunSummary, error)
	// CleanupStaleRuns marks runs stuck in 'running' as 'abandoned'
	// if inactive for more than staleMinutes. Returns count of cleaned runs.
	CleanupStaleRuns(staleMinutes int) (int64, error)
	// BackfillProjects derives project names for runs that don't have one.
	BackfillProjects() (int64, error)
	// Close libera los recursos del store.
	Close() error
}
