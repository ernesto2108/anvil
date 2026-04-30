package memory

import (
	"strings"
	"testing"
)

func TestParseHandoff_FullHandoff(t *testing.T) {
	content := `# Handoff: MCP-001-F6 — Conversational Orchestration Persistence

## Input recibido
- Stack: Go
- Files: internal/mcp/orchestration.go (NEW)

## Estado actual

- [x] Handoff created
- [x] Implement handlers
- [x] Build + lint pass

## Decisiones tomadas

1. Pipeline YAML parsing usa gopkg.in/yaml.v3
2. agent_id format: a_<role>_<sequence>
3. files_touched arrives as []any, cast each element

## Handoff for tester

### Edge cases descubiertos

1. pipelineFile type collision — renamed local var to pipelinePath
2. nil DependsOn — normalized to []string{} before serialization
3. MAX(sequence) on empty table returns NULL

## Notas

- Migration files already existed, did not touch them
- Pre-existing lint issues (5) — bug in errcheck on rows.Close() in execution.go:159
- Found a regression: filescount counter overflows when > 1000 files
`

	draft, taskID, err := ParseHandoff(content, "/path/to/MCP-001-F6.md")
	if err != nil {
		t.Fatalf("ParseHandoff returned error: %v", err)
	}

	if taskID != "MCP-001-F6" {
		t.Errorf("taskID: got %q, want %q", taskID, "MCP-001-F6")
	}

	if !strings.Contains(draft.Summary, "MCP-001-F6") {
		t.Errorf("Summary missing TASK-ID: %q", draft.Summary)
	}
	if !strings.Contains(draft.Summary, "Conversational Orchestration") {
		t.Errorf("Summary missing title: %q", draft.Summary)
	}
	if !strings.Contains(draft.Summary, "Decisiones clave:") {
		t.Errorf("Summary missing decisions section: %q", draft.Summary)
	}

	if len(draft.Decisions) != 3 {
		t.Errorf("Decisions count: got %d, want 3 — %v", len(draft.Decisions), draft.Decisions)
	}
	if !strings.Contains(draft.Decisions[0], "yaml.v3") {
		t.Errorf("Decision[0] mismatch: %q", draft.Decisions[0])
	}

	if len(draft.EdgeCases) != 3 {
		t.Errorf("EdgeCases count: got %d, want 3 — %v", len(draft.EdgeCases), draft.EdgeCases)
	}

	if len(draft.Errors) < 2 {
		t.Errorf("Errors should pick up bug+regression, got %d: %v", len(draft.Errors), draft.Errors)
	}
}

func TestParseHandoff_TaskIDFromFilename(t *testing.T) {
	content := `# Some handoff without task id in title

## Decisiones tomadas

1. Decision A
`
	_, taskID, err := ParseHandoff(content, "/tmp/.handoff/BLT-123.md")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if taskID != "BLT-123" {
		t.Errorf("taskID: got %q, want BLT-123", taskID)
	}
}

func TestParseHandoff_EmptyContent(t *testing.T) {
	_, _, err := ParseHandoff("", "/tmp/x.md")
	if err == nil {
		t.Error("expected error for empty content, got nil")
	}
}

func TestParseHandoff_MissingSections(t *testing.T) {
	// Minimal handoff: only title, no decisions/edge cases/notes.
	content := `# Handoff: MCP-001 — Some title
`
	draft, taskID, err := ParseHandoff(content, "/tmp/MCP-001.md")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if taskID != "MCP-001" {
		t.Errorf("taskID: %q", taskID)
	}
	if len(draft.Decisions) != 0 {
		t.Errorf("Decisions should be empty, got %v", draft.Decisions)
	}
	if len(draft.EdgeCases) != 0 {
		t.Errorf("EdgeCases should be empty, got %v", draft.EdgeCases)
	}
	if !strings.Contains(draft.Summary, "MCP-001") {
		t.Errorf("Summary should contain MCP-001: %q", draft.Summary)
	}
}

func TestParseHandoff_BulletStyleDecisions(t *testing.T) {
	// Some handoffs use `- ` instead of `1.` for decisions.
	content := `# Handoff: TASK-ABC

## Decisiones tomadas

- Use sync.Map instead of mutex+map
- TTL is 60s configurable via env

## Notas

- All good
`
	draft, _, err := ParseHandoff(content, "/tmp/TASK-ABC.md")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(draft.Decisions) != 2 {
		t.Errorf("Decisions count: got %d, want 2 — %v", len(draft.Decisions), draft.Decisions)
	}
}

func TestParseHandoff_NotesWithoutErrorsProducesEmpty(t *testing.T) {
	content := `# Handoff: TASK-1

## Notas

- Pre-existing pattern was kept consistent
- Used context.Background() because handlers receive _

## Decisiones tomadas

- Decision A
`
	draft, _, err := ParseHandoff(content, "/tmp/TASK-1.md")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(draft.Errors) != 0 {
		t.Errorf("Errors should be empty (no error keywords in notes), got %v", draft.Errors)
	}
}

func TestParseHandoff_CrossStackUsesFases(t *testing.T) {
	// Cross-stack handoffs use ## Fases instead of ## Estado actual.
	// The parser doesn't read either — it only cares about decisions/edges/notes.
	content := `# Handoff: CROSS-001 — Cross stack feature

## Fases

### Fase 1 — Backend (Go)
- [x] Implementar handler

### Fase 2 — Frontend (React)
- [x] Llamar endpoint

## Decisiones tomadas

1. JSON tags use camelCase
2. Error envelope: {"error": "..."}

## Handoff for tester

### Edge cases descubiertos

1. CORS preflight on PATCH
`
	draft, taskID, err := ParseHandoff(content, "/tmp/CROSS-001.md")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if taskID != "CROSS-001" {
		t.Errorf("taskID: %q", taskID)
	}
	if len(draft.Decisions) != 2 {
		t.Errorf("Decisions: got %d, want 2", len(draft.Decisions))
	}
	if len(draft.EdgeCases) != 1 {
		t.Errorf("EdgeCases: got %d, want 1", len(draft.EdgeCases))
	}
}

func TestExtractTaskID_FromTitle(t *testing.T) {
	cases := []struct {
		title string
		want  string
	}{
		{"# Handoff: MCP-001-F6 — something", "MCP-001-F6"},
		{"# Handoff: BLT-123 — some title", "BLT-123"},
		{"# Handoff — DASH-FEAT-007: implement X", "DASH-FEAT-007"},
		{"# Some title without ID", ""}, // no match → falls back to filename
	}
	for _, c := range cases {
		got := extractTaskID(c.title, "/tmp/no-id-here.md")
		if c.want == "" {
			// Should fall back to filename slug
			if got != "no-id-here" {
				t.Errorf("title %q: got %q, want filename fallback %q", c.title, got, "no-id-here")
			}
			continue
		}
		if got != c.want {
			t.Errorf("title %q: got %q, want %q", c.title, got, c.want)
		}
	}
}

func TestExtractListSection_StopsAtNextHeader(t *testing.T) {
	content := `## Decisiones tomadas

1. First
2. Second

## Otra sección

- Should not appear

### Otra subsección

- Tampoco
`
	items := extractListSection(content, "## Decisiones tomadas")
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2: %v", len(items), items)
	}
	if items[0] != "First" || items[1] != "Second" {
		t.Errorf("items mismatch: %v", items)
	}
}

func TestExtractListSection_HandlesContinuationLines(t *testing.T) {
	content := `## Decisiones tomadas

1. First decision
   that spans two lines
2. Second decision
`
	items := extractListSection(content, "## Decisiones tomadas")
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2: %v", len(items), items)
	}
	if !strings.Contains(items[0], "spans two lines") {
		t.Errorf("continuation not merged: %q", items[0])
	}
}
