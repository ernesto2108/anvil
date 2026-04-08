package deploy_test

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ernesto2108/anvil/internal/deploy"
)

func Test_IsManagedByAnvil(t *testing.T) {
	tmp := t.TempDir()

	managed := filepath.Join(tmp, "managed.md")
	os.WriteFile(managed, []byte("---\nmanaged-by: anvil\nmodel: high\n---\n\nBody"), 0o644)

	unmanaged := filepath.Join(tmp, "unmanaged.md")
	os.WriteFile(unmanaged, []byte("---\nmodel: high\n---\n\nBody"), 0o644)

	missing := filepath.Join(tmp, "missing.md")

	if !deploy.IsManagedByAnvil(managed) {
		t.Error("should detect managed-by: anvil")
	}
	if deploy.IsManagedByAnvil(unmanaged) {
		t.Error("should not detect managed-by on unmanaged file")
	}
	if deploy.IsManagedByAnvil(missing) {
		t.Error("should return false for missing file")
	}
}

func Test_CleanManagedFiles(t *testing.T) {
	tmp := t.TempDir()

	os.WriteFile(filepath.Join(tmp, "developer.md"), []byte("---\nmanaged-by: anvil\n---\n\nDev"), 0o644)
	os.WriteFile(filepath.Join(tmp, "custom.md"), []byte("---\nname: custom\n---\n\nCustom"), 0o644)

	removed := deploy.CleanManagedFiles(tmp)

	if len(removed) != 1 || removed[0] != "developer.md" {
		t.Errorf("removed = %v, want [developer.md]", removed)
	}

	if _, err := os.Stat(filepath.Join(tmp, "custom.md")); err != nil {
		t.Error("custom.md should not have been removed")
	}

	if _, err := os.Stat(filepath.Join(tmp, "developer.md")); !os.IsNotExist(err) {
		t.Error("developer.md should have been removed")
	}
}

func Test_ResolveCollisions_NoCollisions(t *testing.T) {
	tmp := t.TempDir()
	targetDir := filepath.Join(tmp, "agents")
	os.MkdirAll(targetDir, 0o755)

	srcDir := filepath.Join(tmp, "src")
	os.MkdirAll(srcDir, 0o755)
	os.WriteFile(filepath.Join(srcDir, "developer.md"), []byte("---\nmodel: high\n---\n"), 0o644)

	reader := bufio.NewReader(strings.NewReader(""))

	result := deploy.ResolveCollisions(
		[]string{filepath.Join(srcDir, "developer.md")},
		targetDir,
		reader,
	)

	if len(result.Safe) != 1 {
		t.Errorf("expected 1 safe, got %d", len(result.Safe))
	}
	if len(result.Resolutions) != 0 {
		t.Errorf("expected 0 resolutions, got %d", len(result.Resolutions))
	}
}

func Test_ResolveCollisions_KeepAll(t *testing.T) {
	tmp := t.TempDir()
	targetDir := filepath.Join(tmp, "agents")
	os.MkdirAll(targetDir, 0o755)
	os.WriteFile(filepath.Join(targetDir, "developer.md"), []byte("---\nname: dev\n---\n\nUser"), 0o644)

	srcDir := filepath.Join(tmp, "src")
	os.MkdirAll(srcDir, 0o755)
	os.WriteFile(filepath.Join(srcDir, "developer.md"), []byte("---\nmodel: high\n---\n\nAnvil"), 0o644)
	os.WriteFile(filepath.Join(srcDir, "tester.md"), []byte("---\nmodel: med\n---\n\nAnvil"), 0o644)

	// Simulate: "a" (all at once), "k" (keep)
	reader := bufio.NewReader(strings.NewReader("a\nk\n"))

	result := deploy.ResolveCollisions(
		[]string{filepath.Join(srcDir, "developer.md"), filepath.Join(srcDir, "tester.md")},
		targetDir,
		reader,
	)

	if len(result.Safe) != 1 || result.Safe[0] != "tester.md" {
		t.Errorf("safe = %v, want [tester.md]", result.Safe)
	}
	if len(result.Resolutions) != 1 {
		t.Fatalf("expected 1 resolution, got %d", len(result.Resolutions))
	}

	skip, _, _ := deploy.ApplyResolutions(result.Resolutions)
	if len(skip) != 1 || skip[0] != "developer.md" {
		t.Errorf("skip = %v, want [developer.md]", skip)
	}
}

func Test_ResolveCollisions_ReplaceAll(t *testing.T) {
	tmp := t.TempDir()
	targetDir := filepath.Join(tmp, "agents")
	os.MkdirAll(targetDir, 0o755)
	os.WriteFile(filepath.Join(targetDir, "developer.md"), []byte("---\nname: dev\n---\n\nUser"), 0o644)

	srcDir := filepath.Join(tmp, "src")
	os.MkdirAll(srcDir, 0o755)
	os.WriteFile(filepath.Join(srcDir, "developer.md"), []byte("---\nmodel: high\n---\n\nAnvil"), 0o644)

	reader := bufio.NewReader(strings.NewReader("a\nr\n"))

	result := deploy.ResolveCollisions(
		[]string{filepath.Join(srcDir, "developer.md")},
		targetDir,
		reader,
	)

	_, replace, _ := deploy.ApplyResolutions(result.Resolutions)
	if len(replace) != 1 || replace[0] != "developer.md" {
		t.Errorf("replace = %v, want [developer.md]", replace)
	}
}

func Test_ShouldDeployFile(t *testing.T) {
	skip := []string{"developer.md", "pm.md"}

	if deploy.ShouldDeployFile("developer.md", skip) {
		t.Error("should not deploy skipped file")
	}
	if !deploy.ShouldDeployFile("tester.md", skip) {
		t.Error("should deploy non-skipped file")
	}
}

func Test_GetDeployName(t *testing.T) {
	renames := map[string]string{
		"developer.md": "developer-anvil.md",
	}

	if got := deploy.GetDeployName("developer.md", renames); got != "developer-anvil.md" {
		t.Errorf("got %q, want %q", got, "developer-anvil.md")
	}
	if got := deploy.GetDeployName("tester.md", renames); got != "tester.md" {
		t.Errorf("got %q, want %q", got, "tester.md")
	}
}
