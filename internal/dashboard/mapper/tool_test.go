package mapper_test

import (
	"testing"

	"github.com/ernesto2108/anvil/internal/dashboard/entity"
	"github.com/ernesto2108/anvil/internal/dashboard/mapper"
)

func TestToToolUsageDTO(t *testing.T) {
	got := mapper.ToToolUsageDTO(entity.ToolUsage{ToolName: "Bash", Count: 5})
	if got.ToolName != "Bash" || got.Count != 5 {
		t.Errorf("got %+v", got)
	}
}

func TestToToolUsageDTOs(t *testing.T) {
	rows := []entity.ToolUsage{
		{ToolName: "Bash", Count: 3},
		{ToolName: "Read", Count: 2},
	}
	got := mapper.ToToolUsageDTOs(rows)
	if len(got) != 2 {
		t.Fatalf("len: got %d, want 2", len(got))
	}
	if got[0].ToolName != "Bash" || got[1].ToolName != "Read" {
		t.Errorf("tool names mismatch")
	}
}

func TestToTurnStatsDTO(t *testing.T) {
	got := mapper.ToTurnStatsDTO(entity.TurnStats{
		TurnNumber: 1, Prompt: "fix bug", Timestamp: "2026-04-14T10:00:00Z",
		EndTimestamp: "2026-04-14T10:05:00Z", FilesCount: 3, ToolUsesCount: 5, AgentsCount: 1,
	})
	if got.TurnNumber != 1 || got.Prompt != "fix bug" || got.FilesCount != 3 {
		t.Errorf("got %+v", got)
	}
}

func TestToTurnStatsDTOs(t *testing.T) {
	rows := []entity.TurnStats{
		{TurnNumber: 1, Prompt: "a"},
		{TurnNumber: 2, Prompt: "b"},
	}
	got := mapper.ToTurnStatsDTOs(rows)
	if len(got) != 2 {
		t.Fatalf("len: got %d, want 2", len(got))
	}
}
