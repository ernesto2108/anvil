package deploy_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ernesto2108/anvil/internal/deploy"
	"github.com/ernesto2108/anvil/pkg/fileutil"
)

func Test_SnapshotItem_file(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "source.txt")
	dst := filepath.Join(tmp, "snap", "source.txt")
	os.WriteFile(src, []byte("content"), 0o644)

	if !deploy.SnapshotItem(src, dst) {
		t.Fatal("SnapshotItem returned false")
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if string(data) != "content" {
		t.Errorf("snapshot content = %q, want %q", string(data), "content")
	}
}

func Test_SnapshotItem_directory(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "srcdir")
	dst := filepath.Join(tmp, "snap", "srcdir")
	os.MkdirAll(src, 0o755)
	os.WriteFile(filepath.Join(src, "file.txt"), []byte("hello"), 0o644)

	if !deploy.SnapshotItem(src, dst) {
		t.Fatal("SnapshotItem returned false")
	}

	data, _ := os.ReadFile(filepath.Join(dst, "file.txt"))
	if string(data) != "hello" {
		t.Errorf("got %q, want %q", string(data), "hello")
	}
}

func Test_SnapshotItem_symlink(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "real")
	src := filepath.Join(tmp, "link")
	dst := filepath.Join(tmp, "snap", "link")

	os.MkdirAll(target, 0o755)
	os.Symlink(target, src)

	if !deploy.SnapshotItem(src, dst) {
		t.Fatal("SnapshotItem returned false for symlink")
	}

	symlinkFile := dst + ".symlink"
	if !fileutil.Exists(symlinkFile) {
		t.Fatal(".symlink file not created")
	}

	data, _ := os.ReadFile(symlinkFile)
	if string(data) != target {
		t.Errorf("symlink target = %q, want %q", string(data), target)
	}
}

func Test_SnapshotItem_nonexistent(t *testing.T) {
	if deploy.SnapshotItem("/tmp/does-not-exist-xyz", "/tmp/snap") {
		t.Error("SnapshotItem should return false for nonexistent path")
	}
}

func Test_RestoreItem_from_copy(t *testing.T) {
	tmp := t.TempDir()
	current := filepath.Join(tmp, "current.txt")
	snapshot := filepath.Join(tmp, "snap.txt")

	os.WriteFile(current, []byte("old"), 0o644)
	os.WriteFile(snapshot, []byte("original"), 0o644)

	deploy.RestoreItem(current, snapshot)

	data, _ := os.ReadFile(current)
	if string(data) != "original" {
		t.Errorf("restored content = %q, want %q", string(data), "original")
	}
}

func Test_RestoreItem_from_symlink(t *testing.T) {
	tmp := t.TempDir()
	current := filepath.Join(tmp, "skills")
	snapshot := filepath.Join(tmp, "snap", "skills")

	os.WriteFile(current, []byte("overwritten"), 0o644)

	// Create .symlink file
	os.MkdirAll(filepath.Join(tmp, "snap"), 0o755)
	os.WriteFile(snapshot+".symlink", []byte("/original/skills"), 0o644)

	deploy.RestoreItem(current, snapshot)

	if !fileutil.IsSymlink(current) {
		t.Fatal("expected symlink after restore")
	}

	dest, _ := os.Readlink(current)
	if dest != "/original/skills" {
		t.Errorf("symlink target = %q, want %q", dest, "/original/skills")
	}
}

func Test_RestoreItem_no_snapshot(t *testing.T) {
	tmp := t.TempDir()
	current := filepath.Join(tmp, "file.txt")
	os.WriteFile(current, []byte("x"), 0o644)

	deploy.RestoreItem(current, filepath.Join(tmp, "nonexistent"))

	if fileutil.Exists(current) {
		t.Error("file should be removed when no snapshot exists")
	}
}

func Test_AgentFiles(t *testing.T) {
	tmp := t.TempDir()
	agentDir := filepath.Join(tmp, "agents")
	os.MkdirAll(agentDir, 0o755)

	os.WriteFile(filepath.Join(agentDir, "developer.md"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(agentDir, "tester.md"), []byte("y"), 0o644)
	os.WriteFile(filepath.Join(agentDir, "readme.txt"), []byte("z"), 0o644)

	files := deploy.AgentFiles(tmp)
	if len(files) != 2 {
		t.Errorf("got %d agent files, want 2", len(files))
	}
}
