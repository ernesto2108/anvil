package transcript_test

import (
	"strings"
	"testing"

	"github.com/ernesto2108/anvil/internal/memory/transcript"
)

// ---------------------------------------------------------------------------
// Parse tests
// ---------------------------------------------------------------------------

func Test_Parse_EmptyReader(t *testing.T) {
	td, err := transcript.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if td.TurnCount != 0 {
		t.Errorf("TurnCount = %d, want 0", td.TurnCount)
	}
	if td.SessionID != "" {
		t.Errorf("SessionID = %q, want empty", td.SessionID)
	}
	if len(td.ToolsUsed) != 0 {
		t.Errorf("ToolsUsed = %v, want empty", td.ToolsUsed)
	}
}

func Test_Parse_SingleHumanTurn(t *testing.T) {
	line := `{"type":"human","uuid":"aaa","message":{"role":"user","content":"hi"}}`
	td, err := transcript.Parse(strings.NewReader(line))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if td.TurnCount != 1 {
		t.Errorf("TurnCount = %d, want 1", td.TurnCount)
	}
}

func Test_Parse_ExtractsSessionIDFromFirstUUID(t *testing.T) {
	line := `{"type":"system","uuid":"session-123","cwd":"/home/user/proj"}`
	td, err := transcript.Parse(strings.NewReader(line))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if td.SessionID != "session-123" {
		t.Errorf("SessionID = %q, want %q", td.SessionID, "session-123")
	}
}

func Test_Parse_ExtractsToolsUsed(t *testing.T) {
	line := `{"type":"assistant","uuid":"x","message":{"role":"assistant","content":[{"type":"tool_use","name":"Read","id":"t1"}]}}`
	td, err := transcript.Parse(strings.NewReader(line))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(td.ToolsUsed) != 1 || td.ToolsUsed[0] != "Read" {
		t.Errorf("ToolsUsed = %v, want [Read]", td.ToolsUsed)
	}
}

func Test_Parse_DedupesToolsUsed(t *testing.T) {
	lines := strings.Join([]string{
		`{"type":"assistant","uuid":"x1","message":{"role":"assistant","content":[{"type":"tool_use","name":"Read","id":"t1"}]}}`,
		`{"type":"assistant","uuid":"x2","message":{"role":"assistant","content":[{"type":"tool_use","name":"Read","id":"t2"}]}}`,
	}, "\n")
	td, err := transcript.Parse(strings.NewReader(lines))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(td.ToolsUsed) != 1 {
		t.Errorf("ToolsUsed = %v, want exactly one entry", td.ToolsUsed)
	}
}

func Test_Parse_DetectsErrors(t *testing.T) {
	line := `{"type":"tool","uuid":"x","message":{"role":"tool","content":[{"type":"tool_result","is_error":true}]}}`
	td, err := transcript.Parse(strings.NewReader(line))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !td.HasErrors {
		t.Errorf("HasErrors = false, want true")
	}
}

func Test_Parse_ContentAsString(t *testing.T) {
	line := `{"type":"human","uuid":"y","message":{"role":"user","content":"plain text content"}}`
	_, err := transcript.Parse(strings.NewReader(line))
	if err != nil {
		t.Fatalf("content-as-string caused error: %v", err)
	}
}

func Test_Parse_SkipsMalformedLines(t *testing.T) {
	lines := strings.Join([]string{
		`{"type":"human","uuid":"first","message":{"role":"user","content":"hello"}}`,
		`NOT JSON AT ALL`,
		`{"type":"human","uuid":"second","message":{"role":"user","content":"world"}}`,
	}, "\n")
	td, err := transcript.Parse(strings.NewReader(lines))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if td.TurnCount != 2 {
		t.Errorf("TurnCount = %d, want 2 (malformed line skipped)", td.TurnCount)
	}
}

func Test_ParseFile_FileNotFound(t *testing.T) {
	_, err := transcript.ParseFile("no/such/path/file.jsonl")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

// ---------------------------------------------------------------------------
// IsGeneric tests
// ---------------------------------------------------------------------------

func Test_IsGeneric_KnownNames(t *testing.T) {
	tests := []struct {
		name    string
		wantGen bool
	}{
		{name: "projects", wantGen: true},
		{name: "src", wantGen: true},
		{name: "home", wantGen: true},
		{name: "anvil", wantGen: false},
		{name: "myapp", wantGen: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transcript.IsGeneric(tt.name)
			if got != tt.wantGen {
				t.Errorf("IsGeneric(%q) = %v, want %v", tt.name, got, tt.wantGen)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// extractDecisions tests (tested via Parse output)
//
// extractDecisions is unexported; we exercise it through Parse by embedding
// assistant text turns containing decision keywords.
// ---------------------------------------------------------------------------

func assistantTextLine(uuid, text string) string {
	return `{"type":"assistant","uuid":"` + uuid + `","message":{"role":"assistant","content":[{"type":"text","text":"` + text + `"}]}}`
}

func Test_ExtractDecisions_EnglishKeywords(t *testing.T) {
	line := assistantTextLine("e1", "I decided to use table-driven tests for all parsers.")
	td, err := transcript.Parse(strings.NewReader(line))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(td.Decisions) == 0 {
		t.Error("Decisions is empty; expected at least one entry for 'decided' keyword")
	}
}

func Test_ExtractDecisions_SpanishKeywords(t *testing.T) {
	line := assistantTextLine("s1", "Decidí usar SQLite en memoria para los tests de integración.")
	td, err := transcript.Parse(strings.NewReader(line))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(td.Decisions) == 0 {
		t.Error("Decisions is empty; expected at least one entry for Spanish keyword 'decidí'")
	}
}

func Test_ExtractDecisions_MaxTen(t *testing.T) {
	var sb strings.Builder
	for i := 0; i < 15; i++ {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(assistantTextLine("m"+string(rune('a'+i)), "I decided to apply pattern number decision here."))
	}
	td, err := transcript.Parse(strings.NewReader(sb.String()))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(td.Decisions) > 10 {
		t.Errorf("Decisions has %d entries, want at most 10", len(td.Decisions))
	}
}

func Test_ExtractDecisions_ShortLinesSkipped(t *testing.T) {
	// "decided" is a keyword but the line is < 10 chars
	line := assistantTextLine("short1", "decided")
	td, err := transcript.Parse(strings.NewReader(line))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(td.Decisions) != 0 {
		t.Errorf("short line with keyword should be skipped; got Decisions = %v", td.Decisions)
	}
}
