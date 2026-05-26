package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ernesto2108/anvil/pkg/config"
)

// stopDecision is the JSON Claude Code expects from a Stop hook to ask the
// model to keep working. See https://code.claude.com/docs/en/hooks.
// Emitted on stdout with exit 0; an empty stdout means "allow stop".
type stopDecision struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// verifyStackByExt maps source-file extensions to a stack identifier used to
// pick lint/test patterns. Files outside this map are treated as no-code and
// do not require verification.
var verifyStackByExt = map[string]string{
	".go":   "go",
	".ts":   "ts",
	".tsx":  "ts",
	".js":   "ts",
	".jsx":  "ts",
	".dart": "dart",
	".rs":   "rs",
	".py":   "py",
}

// verifyLintPatterns lists Bash command substrings that, when paired with
// exit_code 0, count as a successful lint run for that stack.
var verifyLintPatterns = map[string][]string{
	"go":   {"golangci-lint run", "go vet"},
	"ts":   {"eslint", "tsc --noEmit", "biome lint", "biome check", "pnpm lint", "npm run lint", "yarn lint", "bun lint"},
	"dart": {"dart analyze", "flutter analyze"},
	"rs":   {"cargo clippy", "cargo check"},
	"py":   {"ruff check", "mypy", "pylint", "flake8"},
}

// verifyTestPatterns lists Bash command substrings that count as a successful
// test run for that stack.
var verifyTestPatterns = map[string][]string{
	"go":   {"go test"},
	"ts":   {"vitest", "jest", "pnpm test", "pnpm run test", "npm test", "yarn test", "bun test"},
	"dart": {"dart test", "flutter test"},
	"rs":   {"cargo test"},
	"py":   {"pytest", "python -m pytest"},
}

// bypassPhrases short-circuit the gate when the latest user prompt contains
// any of them (case-insensitive). Lets the user opt-out for exploration or
// when the verify will happen manually.
var bypassPhrases = []string{"skip verify", "sin verify"}

// cmdVerifyDirect is wired to the Stop hook alongside `anvil emit`. It checks
// whether the just-finished turn edited source files without running
// successful lint + test, and asks Claude to keep working when so.
//
// Always exits 0; stdout carries an optional JSON decision. Errors are logged
// to stderr and the gate fails open — a broken verify must never block the
// user.
func cmdVerifyDirect(_ *config.App) {
	// Orchestrated subprocess — exempt; the orchestrator owns gating.
	if os.Getenv("ANVIL_PARENT_RUN_ID") != "" {
		_, _ = io.Copy(io.Discard, io.LimitReader(os.Stdin, maxStdinBytes))
		return
	}

	raw, err := io.ReadAll(io.LimitReader(os.Stdin, maxStdinBytes))
	if err != nil || len(raw) == 0 {
		return
	}

	var env hookEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return
	}
	if env.SessionID == "" {
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dbPath := filepath.Join(home, ".anvil", "runs.db")

	db, err := openDashboardDB(dbPath, false)
	if err != nil {
		log.Printf("anvil/verify-direct: open db: %s", err)
		return
	}
	defer db.Close() //nolint:errcheck

	decision, err := evaluateVerify(db, env.SessionID)
	if err != nil {
		log.Printf("anvil/verify-direct: evaluate: %s", err)
		return
	}
	if decision == nil {
		return
	}
	if err := json.NewEncoder(os.Stdout).Encode(decision); err != nil {
		log.Printf("anvil/verify-direct: encode decision: %s", err)
	}
}

// evaluateVerify returns a non-nil decision when the latest turn for the
// session edited source files and did not run successful lint+test
// afterwards. nil means "allow stop".
func evaluateVerify(db *sql.DB, sessionID string) (*stopDecision, error) {
	var runID string
	err := db.QueryRow(`SELECT id FROM runs WHERE session_id = ? LIMIT 1`, sessionID).Scan(&runID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("resolve run: %w", err)
	}

	var promptText, promptTS string
	err = db.QueryRow(`
		SELECT prompt, timestamp FROM prompts
		WHERE run_id = ?
		ORDER BY sequence DESC LIMIT 1`, runID).Scan(&promptText, &promptTS)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("latest prompt: %w", err)
	}

	promptLower := strings.ToLower(promptText)
	for _, phrase := range bypassPhrases {
		if strings.Contains(promptLower, phrase) {
			return nil, nil
		}
	}

	edits, err := queryFileEditsAfter(db, runID, promptTS)
	if err != nil {
		return nil, err
	}
	if len(edits) == 0 {
		return nil, nil
	}

	// Reduce to one latest-edit timestamp per stack. Files in stacks not in
	// verifyStackByExt (markdown, yaml, json, etc.) are dropped here.
	latestEditByStack := map[string]string{}
	for _, ed := range edits {
		ext := strings.ToLower(filepath.Ext(ed.path))
		stack, ok := verifyStackByExt[ext]
		if !ok {
			continue
		}
		if prev, exists := latestEditByStack[stack]; !exists || ed.ts > prev {
			latestEditByStack[stack] = ed.ts
		}
	}
	if len(latestEditByStack) == 0 {
		return nil, nil
	}

	var gaps []string
	for _, stack := range sortedKeys(latestEditByStack) {
		lastEdit := latestEditByStack[stack]
		if !hasSuccessfulBashAfter(db, runID, lastEdit, verifyLintPatterns[stack]) {
			gaps = append(gaps, fmt.Sprintf("Lint %s no corrió con éxito desde el último edit", stack))
		}
		if !hasSuccessfulBashAfter(db, runID, lastEdit, verifyTestPatterns[stack]) {
			gaps = append(gaps, fmt.Sprintf("Tests %s no corrieron con éxito desde el último edit", stack))
		}
	}
	if len(gaps) == 0 {
		return nil, nil
	}

	reason := "Verify pendiente:\n• " + strings.Join(gaps, "\n• ") +
		"\n\nCorre lint + test antes de cerrar el turn, o agrega \"sin verify\" al próximo prompt si es intencional."
	return &stopDecision{Decision: "block", Reason: reason}, nil
}

type fileEdit struct {
	path string
	ts   string
}

// queryFileEditsAfter loads file.touched events from the audit log written
// after `afterTS` (lexicographic compare; safe because Anvil always writes
// UTC nanosecond ISO-8601 timestamps).
func queryFileEditsAfter(db *sql.DB, runID, afterTS string) ([]fileEdit, error) {
	rows, err := db.Query(`
		SELECT timestamp, payload FROM events
		WHERE run_id = ? AND event_type = 'file.touched' AND timestamp > ?
		ORDER BY timestamp ASC`, runID, afterTS)
	if err != nil {
		return nil, fmt.Errorf("query file.touched: %w", err)
	}
	defer rows.Close() //nolint:errcheck

	var out []fileEdit
	for rows.Next() {
		var ts, payload string
		if err := rows.Scan(&ts, &payload); err != nil {
			return nil, fmt.Errorf("scan file.touched: %w", err)
		}
		var ft struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal([]byte(payload), &ft); err != nil {
			continue
		}
		if ft.Path == "" {
			continue
		}
		out = append(out, fileEdit{path: ft.Path, ts: ts})
	}
	return out, rows.Err()
}

// hasSuccessfulBashAfter returns true when at least one Bash tool_uses row
// recorded after `afterTS` exited with code 0 and contains any of `patterns`
// in its command string.
func hasSuccessfulBashAfter(db *sql.DB, runID, afterTS string, patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}
	rows, err := db.Query(`
		SELECT tool_input FROM tool_uses
		WHERE run_id = ?
		  AND tool_name = 'Bash'
		  AND exit_code = 0
		  AND timestamp > ?`, runID, afterTS)
	if err != nil {
		log.Printf("anvil/verify-direct: query tool_uses: %s", err)
		return false
	}
	defer rows.Close() //nolint:errcheck

	for rows.Next() {
		var toolInput string
		if err := rows.Scan(&toolInput); err != nil {
			continue
		}
		var ti struct {
			Command string `json:"command"`
		}
		if err := json.Unmarshal([]byte(toolInput), &ti); err != nil {
			continue
		}
		for _, p := range patterns {
			if strings.Contains(ti.Command, p) {
				return true
			}
		}
	}
	if err := rows.Err(); err != nil {
		log.Printf("anvil/verify-direct: rows.Err: %s", err)
	}
	return false
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	// Simple insertion sort — n is at most ~5.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j-1] > out[j]; j-- {
			out[j-1], out[j] = out[j], out[j-1]
		}
	}
	return out
}
