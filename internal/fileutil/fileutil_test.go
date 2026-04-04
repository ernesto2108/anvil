package fileutil_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ernesto2108/anvil/internal/fileutil"
)

func Test_CleanPath(t *testing.T) {
	t.Run("nonexistent path", func(t *testing.T) {
		err := fileutil.CleanPath("/tmp/forge-test-nonexistent-xyz")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("removes file", func(t *testing.T) {
		f := filepath.Join(t.TempDir(), "test.txt")
		os.WriteFile(f, []byte("hello"), 0o644)

		if err := fileutil.CleanPath(f); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fileutil.Exists(f) {
			t.Error("file still exists after CleanPath")
		}
	})

	t.Run("removes directory", func(t *testing.T) {
		d := filepath.Join(t.TempDir(), "subdir")
		os.MkdirAll(d, 0o755)
		os.WriteFile(filepath.Join(d, "inner.txt"), []byte("x"), 0o644)

		if err := fileutil.CleanPath(d); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fileutil.Exists(d) {
			t.Error("directory still exists after CleanPath")
		}
	})

	t.Run("removes symlink", func(t *testing.T) {
		tmp := t.TempDir()
		target := filepath.Join(tmp, "target")
		link := filepath.Join(tmp, "link")
		os.WriteFile(target, []byte("x"), 0o644)
		os.Symlink(target, link)

		if err := fileutil.CleanPath(link); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if fileutil.Exists(link) {
			t.Error("symlink still exists after CleanPath")
		}
		if !fileutil.Exists(target) {
			t.Error("target was removed — should only remove the symlink")
		}
	})
}

func Test_ForceSymlink(t *testing.T) {
	tmp := t.TempDir()
	target := filepath.Join(tmp, "target")
	link := filepath.Join(tmp, "sub", "link")
	os.WriteFile(target, []byte("content"), 0o644)

	if err := fileutil.ForceSymlink(target, link); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !fileutil.IsSymlink(link) {
		t.Error("expected symlink")
	}

	dest, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if dest != target {
		t.Errorf("symlink target = %q, want %q", dest, target)
	}
}

func Test_ForceSymlink_replaces_existing(t *testing.T) {
	tmp := t.TempDir()
	target1 := filepath.Join(tmp, "target1")
	target2 := filepath.Join(tmp, "target2")
	link := filepath.Join(tmp, "link")
	os.WriteFile(target1, []byte("a"), 0o644)
	os.WriteFile(target2, []byte("b"), 0o644)

	fileutil.ForceSymlink(target1, link)
	fileutil.ForceSymlink(target2, link)

	dest, _ := os.Readlink(link)
	if dest != target2 {
		t.Errorf("symlink target = %q, want %q", dest, target2)
	}
}

func Test_CopyFile(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src.txt")
	dst := filepath.Join(tmp, "nested", "dst.txt")
	os.WriteFile(src, []byte("hello world"), 0o644)

	if err := fileutil.CopyFile(src, dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(dst)
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("got %q, want %q", string(data), "hello world")
	}
}

func Test_CopyDir(t *testing.T) {
	tmp := t.TempDir()
	src := filepath.Join(tmp, "src")
	dst := filepath.Join(tmp, "dst")

	os.MkdirAll(filepath.Join(src, "sub"), 0o755)
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("a"), 0o644)
	os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("b"), 0o644)

	if err := fileutil.CopyDir(src, dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dst, "a.txt"))
	if string(data) != "a" {
		t.Errorf("a.txt = %q, want %q", string(data), "a")
	}

	data, _ = os.ReadFile(filepath.Join(dst, "sub", "b.txt"))
	if string(data) != "b" {
		t.Errorf("sub/b.txt = %q, want %q", string(data), "b")
	}
}

func Test_ExpandHome(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "tilde path", input: "~/projects", want: filepath.Join(home, "projects")},
		{name: "absolute path", input: "/usr/local", want: "/usr/local"},
		{name: "relative path", input: "relative", want: "relative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fileutil.ExpandHome(tt.input)
			if got != tt.want {
				t.Errorf("ExpandHome(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func Test_IsSymlink(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "file")
	link := filepath.Join(tmp, "link")
	os.WriteFile(file, []byte("x"), 0o644)
	os.Symlink(file, link)

	if fileutil.IsSymlink(file) {
		t.Error("regular file reported as symlink")
	}
	if !fileutil.IsSymlink(link) {
		t.Error("symlink not detected")
	}
	if fileutil.IsSymlink(filepath.Join(tmp, "nope")) {
		t.Error("nonexistent path reported as symlink")
	}
}
