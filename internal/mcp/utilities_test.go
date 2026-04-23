package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/state"
)

// newTestServerWithProviders constructs a Server whose cfg has two providers:
// "claude" (active) and "gemini", so switchProvider can be tested without real
// disk-backed YAML files.
func newTestServerWithProviders(t *testing.T) (*Server, string) {
	t.Helper()
	tmpDir := t.TempDir()

	cfg := &config.App{
		Name:    "anvil",
		RepoDir: tmpDir,
		Provider: config.ProviderConfig{
			Provider: "claude",
			Providers: map[string]config.TierMap{
				"claude": {},
				"gemini": {},
			},
		},
	}

	return &Server{cfg: cfg}, tmpDir
}

// ---------------------------------------------------------------------------
// countFilesChanged
// ---------------------------------------------------------------------------

func TestCountFilesChanged_empty(t *testing.T) {
	if got := countFilesChanged(""); got != 0 {
		t.Errorf("countFilesChanged(\"\") = %d, want 0", got)
	}
}

func TestCountFilesChanged_singleFile(t *testing.T) {
	diffOut := "1 file changed, 2 insertions(+)"
	if got := countFilesChanged(diffOut); got != 1 {
		t.Errorf("countFilesChanged = %d, want 1", got)
	}
}

func TestCountFilesChanged_multiFile(t *testing.T) {
	diffOut := "3 files changed, 12 insertions(+), 4 deletions(-)"
	if got := countFilesChanged(diffOut); got != 3 {
		t.Errorf("countFilesChanged = %d, want 3", got)
	}
}

// ---------------------------------------------------------------------------
// collectSkillNames
// ---------------------------------------------------------------------------

func TestCollectSkillNames_empty(t *testing.T) {
	names := collectSkillNames("/tmp/does-not-exist-xyzzy-12345")
	if len(names) != 0 {
		t.Errorf("expected empty slice, got %v", names)
	}
}

func TestCollectSkillNames_withDirs(t *testing.T) {
	tmpDir := t.TempDir()
	skillsDir := filepath.Join(tmpDir, config.CompSkills)
	if err := os.MkdirAll(skillsDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	for _, d := range []string{"alpha", "beta", "gamma"} {
		if err := os.Mkdir(filepath.Join(skillsDir, d), 0o755); err != nil {
			t.Fatalf("mkdir skill %s: %v", d, err)
		}
	}
	// A file (not a dir) — should NOT appear in results.
	if err := os.WriteFile(filepath.Join(skillsDir, "README.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	names := collectSkillNames(tmpDir)
	if len(names) != 3 {
		t.Errorf("expected 3 skill names, got %v", names)
	}
	found := make(map[string]bool)
	for _, n := range names {
		found[n] = true
	}
	for _, want := range []string{"alpha", "beta", "gamma"} {
		if !found[want] {
			t.Errorf("skill %q not in result %v", want, names)
		}
	}
}

// ---------------------------------------------------------------------------
// appendTargetChecks
// ---------------------------------------------------------------------------

func TestAppendTargetChecks_missingBase(t *testing.T) {
	var checks []healthCheck
	issues := appendTargetChecks("claude", "/tmp/no-such-base-dir-xyzzy", nil, &checks)

	if issues != 1 {
		t.Errorf("issues = %d, want 1", issues)
	}
	if len(checks) != 1 {
		t.Fatalf("len(checks) = %d, want 1", len(checks))
	}
	if checks[0].Status != "warn" {
		t.Errorf("checks[0].Status = %q, want 'warn'", checks[0].Status)
	}
}

func TestAppendTargetChecks_allPresent(t *testing.T) {
	tmpDir := t.TempDir()
	// Create a dir and a file that the items will point to.
	if err := os.Mkdir(filepath.Join(tmpDir, "agents"), 0o755); err != nil {
		t.Fatalf("mkdir agents: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "CLAUDE.md"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write CLAUDE.md: %v", err)
	}

	items := []doctorTargetCheck{
		{name: "agents", checkType: "dir_or_symlink"},
		{name: "CLAUDE.md", checkType: "file_or_symlink"},
	}

	var checks []healthCheck
	issues := appendTargetChecks("claude", tmpDir, items, &checks)

	if issues != 0 {
		t.Errorf("issues = %d, want 0", issues)
	}
	for _, c := range checks {
		if c.Status != "ok" {
			t.Errorf("check %q status = %q, want 'ok'", c.Name, c.Status)
		}
	}
}

func TestAppendTargetChecks_partialMissing(t *testing.T) {
	tmpDir := t.TempDir()
	// Only create one of two items.
	if err := os.Mkdir(filepath.Join(tmpDir, "agents"), 0o755); err != nil {
		t.Fatalf("mkdir agents: %v", err)
	}
	// "skills" dir intentionally NOT created.

	items := []doctorTargetCheck{
		{name: "agents", checkType: "dir_or_symlink"},
		{name: "skills", checkType: "dir_or_symlink"},
	}

	var checks []healthCheck
	issues := appendTargetChecks("claude", tmpDir, items, &checks)

	if issues != 1 {
		t.Errorf("issues = %d, want 1", issues)
	}
	if len(checks) != 2 {
		t.Fatalf("len(checks) = %d, want 2", len(checks))
	}

	// First check (agents) — ok; second (skills) — warn.
	if checks[0].Status != "ok" {
		t.Errorf("checks[0] (agents) status = %q, want 'ok'", checks[0].Status)
	}
	if checks[1].Status != "warn" {
		t.Errorf("checks[1] (skills) status = %q, want 'warn'", checks[1].Status)
	}
}

// ---------------------------------------------------------------------------
// switchProvider
// ---------------------------------------------------------------------------

func TestSwitchProvider_missingArg(t *testing.T) {
	db := setupTestDB(t)
	srv, _ := newTestServer(t, db)

	_, err := srv.switchProvider(context.Background(), map[string]any{})
	if err == nil {
		t.Fatal("expected error for missing provider arg, got nil")
	}
}

func TestSwitchProvider_unknownProvider(t *testing.T) {
	srv, _ := newTestServerWithProviders(t)

	out, err := srv.switchProvider(context.Background(), map[string]any{
		"provider": "openai",
	})
	if err != nil {
		t.Fatalf("unexpected hard error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, ok := result["error"]; !ok {
		t.Errorf("expected 'error' key in response, got: %s", out)
	}
	available, ok := result["available_providers"]
	if !ok {
		t.Fatalf("expected 'available_providers' key in response, got: %s", out)
	}
	avSlice, ok := available.([]any)
	if !ok || len(avSlice) == 0 {
		t.Errorf("available_providers should be a non-empty list, got: %v", available)
	}
}

// ---------------------------------------------------------------------------
// getDiff
// ---------------------------------------------------------------------------

func TestGetDiff_noDeployedSHA(t *testing.T) {
	srv, tmpDir := newTestServerWithProviders(t)
	// Write state with empty DeployedSHA (default after Load on fresh dir).
	stateDir := filepath.Join(tmpDir, ".anvil")
	if _, err := state.Load(stateDir); err != nil {
		t.Fatalf("init state: %v", err)
	}
	// DeployedSHA is "" by default in a fresh state.

	out, err := srv.getDiff(context.Background(), map[string]any{})
	if err != nil {
		t.Fatalf("unexpected hard error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, ok := result["error"]; !ok {
		t.Errorf("expected 'error' key for no-deployed-SHA, got: %s", out)
	}
}

func TestGetDiff_noChanges(t *testing.T) {
	// Use the actual repo root so gitutil.CurrentSHA() returns the real HEAD.
	gitDir := findGitRoot(t)
	// gitutil.CurrentSHA() returns --short (7 chars). Match that format.
	currentSHA := readHEADSHA(t, gitDir)
	if len(currentSHA) > 7 {
		currentSHA = currentSHA[:7]
	}

	srv := &Server{
		cfg: &config.App{
			Name:    "anvil",
			RepoDir: gitDir,
			Provider: config.ProviderConfig{
				Provider: "claude",
				Providers: map[string]config.TierMap{
					"claude": {},
				},
			},
		},
	}

	out, err := srv.getDiff(context.Background(), map[string]any{
		"base": currentSHA,
	})
	if err != nil {
		t.Fatalf("unexpected hard error: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	fc, ok := result["files_changed"]
	if !ok {
		t.Fatalf("missing 'files_changed' key, got: %s", out)
	}
	// JSON numbers unmarshal as float64.
	if fc.(float64) != 0 {
		t.Errorf("files_changed = %v, want 0", fc)
	}
	if ds, _ := result["diff_summary"].(string); ds != "no changes since last deploy" {
		t.Errorf("diff_summary = %q, want 'no changes since last deploy'", ds)
	}
}

// findGitRoot walks up from the current working directory to find the .git dir.
func findGitRoot(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	dir := cwd
	for {
		if fi, err := os.Stat(filepath.Join(dir, ".git")); err == nil && fi.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find .git root")
		}
		dir = parent
	}
}

// readHEADSHA resolves the current HEAD commit SHA from the git root.
func readHEADSHA(t *testing.T, gitRoot string) string {
	t.Helper()
	headData, err := os.ReadFile(filepath.Join(gitRoot, ".git", "HEAD"))
	if err != nil {
		t.Fatalf("read HEAD: %v", err)
	}
	head := string(headData)
	if len(head) > 0 && head[len(head)-1] == '\n' {
		head = head[:len(head)-1]
	}
	const refPrefix = "ref: "
	if len(head) > len(refPrefix) && head[:len(refPrefix)] == refPrefix {
		refPath := head[len(refPrefix):]
		shaData, err := os.ReadFile(filepath.Join(gitRoot, ".git", refPath))
		if err != nil {
			t.Fatalf("read ref %s: %v", refPath, err)
		}
		sha := string(shaData)
		if len(sha) > 0 && sha[len(sha)-1] == '\n' {
			sha = sha[:len(sha)-1]
		}
		return sha
	}
	return head // detached HEAD
}
