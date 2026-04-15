package mapper

import (
	"testing"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

func TestToRunDTO(t *testing.T) {
	now := time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)
	ended := now.Add(5 * time.Minute)
	dur := int64(300000)
	qa := 8.5

	r := entity.RunSummary{
		ID:            "run-1",
		TaskID:        "task-1",
		TaskDesc:      "test task",
		Status:        "success",
		Complexity:    "medium",
		Provider:      "anthropic",
		Project:       "anvil",
		StartedAt:     now,
		EndedAt:       &ended,
		DurationMs:    &dur,
		TotalTokens:   1000,
		FilesCount:    5,
		AgentsCount:   2,
		QAScore:       &qa,
		ParentRunID:   "",
		ChildrenCount: 0,
		Branch:        "main",
		ErrorReason:   "",
	}

	dto := ToRunDTO(r)

	if dto.ID != "run-1" {
		t.Errorf("ID: expected %q, got %q", "run-1", dto.ID)
	}
	if dto.StartedAt != now.Format(time.RFC3339Nano) {
		t.Errorf("StartedAt: expected %q, got %q", now.Format(time.RFC3339Nano), dto.StartedAt)
	}
	if dto.EndedAt != ended.Format(time.RFC3339Nano) {
		t.Errorf("EndedAt: expected %q, got %q", ended.Format(time.RFC3339Nano), dto.EndedAt)
	}
	if dto.DurationMs != 300000 {
		t.Errorf("DurationMs: expected 300000, got %d", dto.DurationMs)
	}
	if dto.Branch != "main" {
		t.Errorf("Branch: expected %q, got %q", "main", dto.Branch)
	}
}

func TestToRunDTO_NilOptionals(t *testing.T) {
	r := entity.RunSummary{
		ID:        "run-2",
		Status:    "running",
		StartedAt: time.Now(),
	}

	dto := ToRunDTO(r)

	if dto.EndedAt != "" {
		t.Errorf("EndedAt: expected empty, got %q", dto.EndedAt)
	}
	if dto.DurationMs != 0 {
		t.Errorf("DurationMs: expected 0, got %d", dto.DurationMs)
	}
}

func TestToRunDTOs(t *testing.T) {
	rows := []entity.RunSummary{
		{ID: "a", Status: "success", StartedAt: time.Now()},
		{ID: "b", Status: "failed", StartedAt: time.Now()},
	}
	dtos := ToRunDTOs(rows)
	if len(dtos) != 2 {
		t.Fatalf("expected 2 DTOs, got %d", len(dtos))
	}
	if dtos[0].ID != "a" || dtos[1].ID != "b" {
		t.Errorf("IDs: expected [a, b], got [%s, %s]", dtos[0].ID, dtos[1].ID)
	}
}
