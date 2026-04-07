package fileutil

import (
	"fmt"
	"os"
	"path/filepath"
)

func CleanPath(p string) error {
	fi, err := os.Lstat(p)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("stat %s: %w", p, err)
	}

	if fi.Mode()&os.ModeSymlink != 0 || !fi.IsDir() {
		return os.Remove(p)
	}
	return os.RemoveAll(p)
}

func ForceSymlink(src, dst string) error {
	if err := CleanPath(dst); err != nil {
		return fmt.Errorf("clean %s: %w", dst, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("mkdir %s: %w", filepath.Dir(dst), err)
	}
	return os.Symlink(src, dst)
}

func CopyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read %s: %w", src, err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("mkdir for %s: %w", dst, err)
	}
	return os.WriteFile(dst, data, 0o644)
}

func CopyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("rel path %s: %w", path, err)
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return CopyFile(path, target)
	})
}

func IsSymlink(p string) bool {
	fi, err := os.Lstat(p)
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeSymlink != 0
}

func Exists(p string) bool {
	_, err := os.Lstat(p)
	return err == nil
}

func IsDir(p string) bool {
	fi, err := os.Stat(p)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

func ExpandHome(p string) string {
	if len(p) >= 2 && p[:2] == "~/" {
		home, err := os.UserHomeDir()
		if err != nil {
			return p
		}
		return filepath.Join(home, p[2:])
	}
	return p
}
