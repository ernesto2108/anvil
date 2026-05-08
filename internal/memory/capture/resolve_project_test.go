package capture_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ernesto2108/anvil/internal/memory/capture"
	"github.com/ernesto2108/anvil/internal/memory/transcript"
)

func Test_ResolveProject_OverrideWins(t *testing.T) {
	td := transcript.TranscriptDigest{Cwd: "/some/projects/anvil"}
	got, err := capture.ResolveProject(context.Background(), td, "myproject", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "myproject" {
		t.Errorf("ResolveProject with override = %q, want %q", got, "myproject")
	}
}

func Test_ResolveProject_CwdWithGit(t *testing.T) {
	dir := t.TempDir()
	// Create a .git directory so hasGit returns true.
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatalf("create .git dir: %v", err)
	}

	projectName := filepath.Base(dir)
	td := transcript.TranscriptDigest{Cwd: dir}

	got, err := capture.ResolveProject(context.Background(), td, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != projectName {
		t.Errorf("ResolveProject cwd-with-git = %q, want %q", got, projectName)
	}
}

func Test_ResolveProject_GenericCwdFallback(t *testing.T) {
	// "projects" is generic — strategy 2 should skip it.
	// We use a temp dir whose base name is generic (simulate by using "src").
	// t.TempDir() returns something like /tmp/TestXxx123; rename last component
	// is tricky, so instead create a subdir called "src" without .git.
	parent := t.TempDir()
	genericDir := filepath.Join(parent, "src")
	if err := os.Mkdir(genericDir, 0o755); err != nil {
		t.Fatalf("create generic dir: %v", err)
	}

	td := transcript.TranscriptDigest{Cwd: genericDir}
	// Strategy 2: cwd base is "src" (generic) → skipped even with .git check;
	// no .git anyway, so strategy 2 never fires.
	// Strategy 3: no DB/sessionID → skipped.
	// Strategy 4 (fallback): returns Base(cwd) = "src".
	got, err := capture.ResolveProject(context.Background(), td, "", nil)
	// The fallback returns "src" without erroring — generic names are only
	// filtered at strategy 2. The test verifies we don't crash/error.
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Error("ResolveProject returned empty project name from generic cwd fallback")
	}
}
