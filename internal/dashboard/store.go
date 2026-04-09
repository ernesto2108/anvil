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
	// Close libera los recursos del store.
	Close() error
}
