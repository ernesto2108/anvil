//go:build dashboard

package dashboard

import (
	"context"

	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

// DashboardReader provides read-only access to dashboard data.
// *query.Reader satisfies this interface.
type DashboardReader interface {
	ListRuns(ctx context.Context, limit, offset int, status, startDate, endDate, project string) ([]entity.RunSummary, error)
	GetRunSummary(ctx context.Context, runID string) (*entity.RunSummary, error)
	ListChildRuns(ctx context.Context, parentRunID string) ([]entity.RunSummary, error)
	ListProjects(ctx context.Context) ([]string, error)
	ListAgentsByRun(ctx context.Context, runID string) ([]entity.AgentRow, error)
	GetAgentDetail(ctx context.Context, runID, agentID string) (*entity.AgentDetail, []entity.FileRecord, error)
	ListFilesByRun(ctx context.Context, runID string) ([]entity.FileRecord, error)
	ListActivityEvents(ctx context.Context, runID string) ([]entity.ActivityEvent, error)
	ListToolUsageByRun(ctx context.Context, runID string) ([]entity.ToolUsage, error)
	TotalToolUsesByRun(ctx context.Context, runID string) (int, error)
	ListToolUseDetailsByRun(ctx context.Context, runID string) ([]entity.ToolUseDetail, error)
	ListTasksByRun(ctx context.Context, runID string) ([]entity.Task, error)
	CountCompactions(ctx context.Context, runID string) (int, error)
	ListPromptsByRun(ctx context.Context, runID string) ([]entity.Prompt, error)
	GetTurnStats(ctx context.Context, runID string) ([]entity.TurnStats, error)
	DeleteRun(ctx context.Context, runID string) error
	DeleteRuns(ctx context.Context, runIDs []string) error
	ListErrorGroups(ctx context.Context, status, search string) ([]entity.ErrorGroup, error)
	GetErrorGroup(ctx context.Context, groupID string) (*entity.ErrorGroup, error)
	ListErrorGroupRuns(ctx context.Context, groupID string) ([]entity.ErrorGroupRun, error)
	ListErrorGroupHistory(ctx context.Context, groupID string) ([]entity.ErrorGroupHistory, error)
	GetErrorTrend(ctx context.Context, groupID string) ([]entity.TrendPoint, error)
	GetErrorCounts(ctx context.Context) (entity.ErrorCounts, error)
}

// DashboardWriter provides write access to dashboard data.
// *writer.EventWriter satisfies this interface.
type DashboardWriter interface {
	CleanupStaleRuns(staleMinutes int) (int64, error)
	BackfillProjects() (int64, error)
	UpdateErrorResolution(ctx context.Context, groupID, status, notes, commitLink string) error
	Close() error
}
