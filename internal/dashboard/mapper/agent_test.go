package mapper

import (
	"testing"
	"time"

	"github.com/ernesto2108/anvil/internal/dashboard/entity"
)

func TestToAgentDetailDTO(t *testing.T) {
	now := time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)
	ended := now.Add(30 * time.Second)
	dur := int64(30000)

	d := entity.AgentDetail{
		AgentID:   "agent-1",
		AgentRole: "developer",
		Status:    "success",
		StartedAt: &now,
		EndedAt:   &ended,
		DurationMs: &dur,
		ErrorMsg:  "",
		Output:    "all done",
	}
	files := []entity.FileRecord{
		{Path: "/src/main.go", Operation: "write", AgentID: "agent-1"},
	}

	dto := ToAgentDetailDTO(d, files)

	if dto.Agent.ID != "agent-1" {
		t.Errorf("Agent.ID: expected %q, got %q", "agent-1", dto.Agent.ID)
	}
	if dto.Agent.Name != "developer" {
		t.Errorf("Agent.Name: expected %q, got %q", "developer", dto.Agent.Name)
	}
	if dto.Output != "all done" {
		t.Errorf("Output: expected %q, got %q", "all done", dto.Output)
	}
	if len(dto.Files) != 1 {
		t.Fatalf("Files: expected 1, got %d", len(dto.Files))
	}
	if dto.Files[0].Path != "/src/main.go" {
		t.Errorf("Files[0].Path: expected %q, got %q", "/src/main.go", dto.Files[0].Path)
	}
}

func TestToSessionAgentDTO(t *testing.T) {
	now := time.Date(2026, 4, 14, 10, 0, 0, 0, time.UTC)
	dur := int64(5000)

	row := entity.AgentRow{
		AgentID:    "agent-2",
		AgentRole:  "tester",
		Status:     "running",
		StartedAt:  &now,
		DurationMs: &dur,
	}

	dto := ToSessionAgentDTO(row, "test output")

	if dto.ID != "agent-2" {
		t.Errorf("ID: expected %q, got %q", "agent-2", dto.ID)
	}
	if dto.Role != "tester" {
		t.Errorf("Role: expected %q, got %q", "tester", dto.Role)
	}
	if dto.Output != "test output" {
		t.Errorf("Output: expected %q, got %q", "test output", dto.Output)
	}
	if dto.StartedAt == "" {
		t.Error("StartedAt: expected non-empty")
	}
	if dto.EndedAt != "" {
		t.Errorf("EndedAt: expected empty, got %q", dto.EndedAt)
	}
}
