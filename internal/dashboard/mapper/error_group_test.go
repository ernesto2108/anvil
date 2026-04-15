package mapper

import (
	"testing"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

func TestToErrorGroupDTO(t *testing.T) {
	now := time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)

	g := entity.ErrorGroup{
		ID:               "eg-abc",
		Fingerprint:      "abc123",
		Title:            "timeout after 30s",
		NormalizedMsg:    "timeout after *",
		ResolutionStatus: "new",
		FirstSeenAt:      now,
		LastSeenAt:       now,
		OccurrenceCount:  3,
		Notes:            "",
		CommitLink:       "",
	}

	dto := ToErrorGroupDTO(g)

	if dto.ID != "eg-abc" {
		t.Errorf("ID: expected %q, got %q", "eg-abc", dto.ID)
	}
	if dto.OccurrenceCount != 3 {
		t.Errorf("OccurrenceCount: expected 3, got %d", dto.OccurrenceCount)
	}
	if dto.FirstSeenAt != now.Format(time.RFC3339Nano) {
		t.Errorf("FirstSeenAt: expected %q, got %q", now.Format(time.RFC3339Nano), dto.FirstSeenAt)
	}
}

func TestToErrorGroupDetailDTO(t *testing.T) {
	now := time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)
	group := entity.ErrorGroup{ID: "eg-1", FirstSeenAt: now, LastSeenAt: now}
	runs := []entity.ErrorGroupRun{{RunID: "r1", ErrorMsg: "err"}}
	history := []entity.ErrorGroupHistory{{OldStatus: "new", NewStatus: "resolved"}}
	trend := []entity.TrendPoint{{Date: "2026-04-14", Count: 2}}

	dto := ToErrorGroupDetailDTO(group, runs, history, trend)

	if len(dto.Runs) != 1 {
		t.Errorf("Runs: expected 1, got %d", len(dto.Runs))
	}
	if len(dto.History) != 1 {
		t.Errorf("History: expected 1, got %d", len(dto.History))
	}
	if len(dto.Trend) != 1 {
		t.Errorf("Trend: expected 1, got %d", len(dto.Trend))
	}
}
