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
	ListRuns(ctx context.Context, limit, offset int) ([]store.RunSummary, error)
	// ListAgentsByRun retorna todos los agentes de un run ordenados por sequence ASC.
	// Retorna nil, nil si el run no existe o no tiene agentes.
	ListAgentsByRun(ctx context.Context, runID string) ([]store.AgentRow, error)
	// GetAgentDetail retorna el detalle completo de un agente (runID + agentID) junto con sus archivos.
	// Retorna (nil, nil, nil) si el agente no existe — NO es un error.
	GetAgentDetail(ctx context.Context, runID, agentID string) (*store.AgentDetail, []store.FileRow, error)
	// GetMetrics calcula los agregados sobre los últimos 30 runs terminados.
	// Ver store.SQLiteStore.GetMetrics para la semántica completa.
	GetMetrics(ctx context.Context) (store.Metrics, error)
	// Close libera los recursos del store.
	Close() error
}
