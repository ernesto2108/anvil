package mapper_test

import (
	"testing"

	"github.com/ernesto2108/anvil/internal/dashboard/entity"
	"github.com/ernesto2108/anvil/internal/dashboard/mapper"
)

func TestToFileDTOs(t *testing.T) {
	rows := []entity.FileRecord{
		{Path: "/a.go", Operation: "create", Diff: "+new", AgentID: "a1"},
		{Path: "/b.go", Operation: "modify", AgentID: ""},
	}
	got := mapper.ToFileDTOs(rows)
	if len(got) != 2 {
		t.Fatalf("len: got %d, want 2", len(got))
	}
	if got[0].Path != "/a.go" || got[0].Action != "create" {
		t.Errorf("first file: %+v", got[0])
	}
	if got[1].AgentID != "" {
		t.Errorf("second file agentID: got %q, want empty", got[1].AgentID)
	}
}
