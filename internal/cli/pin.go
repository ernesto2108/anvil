package cli

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/internal/deploy"
	"github.com/ernesto2108/anvil/pkg/fileutil"
	"github.com/ernesto2108/anvil/pkg/gitutil"
	"github.com/ernesto2108/anvil/pkg/output"
)

func cmdPin(cfg *config.App, git *gitutil.Repo, args []string) {
	if len(args) < 2 {
		output.Error("Usage: %s pin <component> <tag>", cfg.Name)
		os.Exit(1)
	}

	component := args[0]
	version := args[1]
	st := loadState(cfg)
	paths := deploy.ResolvePaths()

	if !git.VersionExists(version) {
		output.Error("Version '%s' not found", version)
		os.Exit(1)
	}

	// Break parent symlink if pinning a nested component
	if strings.Contains(component, "/") {
		topLevel := strings.SplitN(component, "/", 2)[0]
		parentPath := filepath.Join(paths.Claude, topLevel)
		if fileutil.IsSymlink(parentPath) {
			output.Warn("Breaking symlink %s/", topLevel)
			link, err := os.Readlink(parentPath)
			if err != nil {
				output.Error("readlink %s: %s", parentPath, err)
				os.Exit(1)
			}
			if err := os.Remove(parentPath); err != nil {
				output.Error("remove symlink %s: %s", parentPath, err)
				os.Exit(1)
			}
			if err := fileutil.CopyDir(link, parentPath); err != nil {
				output.Error("copy dir %s: %s", link, err)
				os.Exit(1)
			}
		}
	}

	targetPath := filepath.Join(paths.Claude, component)
	_ = fileutil.CleanPath(targetPath)
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		output.Error("create dir: %s", err)
		os.Exit(1)
	}

	objType, err := git.CatFileType(version, component)
	if err != nil {
		output.Error("cat-file: %s", err)
		os.Exit(1)
	}

	if objType == "tree" {
		if err := os.MkdirAll(targetPath, 0o755); err != nil {
			output.Error("create dir %s: %s", targetPath, err)
			os.Exit(1)
		}
		if err := git.Archive(version, component, paths.Claude); err != nil {
			output.Error("archive: %s", err)
			os.Exit(1)
		}
	} else {
		content, err := git.ShowFile(version, component)
		if err != nil {
			output.Error("show file: %s", err)
			os.Exit(1)
		}
		if err := os.WriteFile(targetPath, []byte(content), 0o644); err != nil {
			output.Error("write file %s: %s", targetPath, err)
			os.Exit(1)
		}
	}

	output.Info("Pinned %s to %s", output.Cyan(component), output.Yellow(version))

	st.SetPin(component, version)
	if err := st.Save(); err != nil {
		output.Warn("save state: %s", err)
	}
}

func cmdUnpin(cfg *config.App, git *gitutil.Repo, args []string) {
	if len(args) < 1 {
		output.Error("Usage: %s unpin <component>", cfg.Name)
		os.Exit(1)
	}

	component := args[0]
	st := loadState(cfg)
	paths := deploy.ResolvePaths()

	source := filepath.Join(cfg.RepoDir, component)
	if !fileutil.Exists(source) {
		output.Error("Component '%s' not found in repo", component)
		os.Exit(1)
	}

	targetPath := filepath.Join(paths.Claude, component)
	_ = fileutil.CleanPath(targetPath)
	st.RemovePin(component)

	if strings.Contains(component, "/") {
		objType, err := git.CatFileType("HEAD", component)
		if err == nil {
			if objType == "tree" {
				if err := os.MkdirAll(targetPath, 0o755); err != nil {
					output.Error("create dir %s: %s", targetPath, err)
					os.Exit(1)
				}
				if err := git.Archive("HEAD", component, paths.Claude); err != nil {
					output.Error("archive: %s", err)
					os.Exit(1)
				}
			} else {
				content, err := git.ShowFile("HEAD", component)
				if err != nil {
					output.Error("show file: %s", err)
					os.Exit(1)
				}
				if err := os.WriteFile(targetPath, []byte(content), 0o644); err != nil {
					output.Error("write file %s: %s", targetPath, err)
					os.Exit(1)
				}
			}
		}

		// Restore parent symlink if no more pins in this top-level
		topLevel := strings.SplitN(component, "/", 2)[0]
		if st.PinCount(topLevel+"/") == 0 {
			parentPath := filepath.Join(paths.Claude, topLevel)
			if fileutil.IsDir(parentPath) && !fileutil.IsSymlink(parentPath) {
				if err := os.RemoveAll(parentPath); err != nil {
					output.Error("remove %s: %s", parentPath, err)
					os.Exit(1)
				}
				if err := os.Symlink(filepath.Join(cfg.RepoDir, topLevel), parentPath); err != nil {
					output.Error("symlink %s: %s", parentPath, err)
					os.Exit(1)
				}
				output.Info("Restored %s symlink", output.Cyan(topLevel+"/"))
			}
		}
	} else {
		if err := os.Symlink(source, targetPath); err != nil {
			output.Error("symlink %s: %s", targetPath, err)
			os.Exit(1)
		}
		output.Info("Unpinned %s", output.Cyan(component))
	}

	if err := st.Save(); err != nil {
		output.Warn("save state: %s", err)
	}
}
