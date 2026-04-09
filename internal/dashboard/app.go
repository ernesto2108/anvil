//go:build dashboard

package dashboard

import "context"

// App holds Wails runtime context and the read-only store used
// by bindings exposed to the frontend.
type App struct {
	ctx   context.Context
	store Store
}

// NewApp wires the store dependency. Store is required: pass nil only
// in tests that do not exercise bindings.
func NewApp(s Store) *App {
	return &App{store: s}
}

// Startup is invoked by Wails once the native window is ready.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// --- Bindings (read-only stubs — real implementation lands in DASH-FEAT-005+) ---

// GetRuns returns a paginated list of runs matching the given query.
func (a *App) GetRuns(_ RunsQuery) ([]RunDTO, error) {
	return []RunDTO{}, nil
}

// GetRun returns the detail for a single run identified by its ID.
func (a *App) GetRun(_ string) (*RunDetailDTO, error) {
	return nil, nil
}

// GetAgents returns all agents belonging to the run identified by runID.
func (a *App) GetAgents(_ string) ([]AgentDTO, error) {
	return []AgentDTO{}, nil
}

// GetAgent returns the detail for a single agent within a run.
func (a *App) GetAgent(_, _ string) (*AgentDetailDTO, error) {
	return nil, nil
}

// GetFlow returns the execution flow graph for a run.
func (a *App) GetFlow(_ string) (*FlowDTO, error) {
	return nil, nil
}

// GetMetrics returns aggregated metrics for the given time range.
func (a *App) GetMetrics(_ MetricsQuery) (*MetricsDTO, error) {
	return nil, nil
}

// GetFiles returns files produced or consumed by an agent within a run.
func (a *App) GetFiles(_, _ string) ([]FileDTO, error) {
	return []FileDTO{}, nil
}
