package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/internal/deploy"
	"github.com/ernesto2108/anvil/pkg/fileutil"
	"github.com/ernesto2108/anvil/pkg/gitutil"
	"github.com/ernesto2108/anvil/pkg/output"
	"github.com/ernesto2108/anvil/pkg/state"
)

func cmdDoctor(cfg *config.App, git *gitutil.Repo) {
	st := loadState(cfg)
	paths := deploy.ResolvePaths()

	fmt.Println()
	fmt.Println(output.Bold(title(cfg.Name) + " Doctor"))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	issues := 0

	// Check 1: Binary vs deployed SHA
	currentSHA := git.CurrentSHA()
	if st.DeployedSHA == "" || st.DeployedSHA == state.StateNone {
		printCheck(false, "Nothing deployed yet — run '%s deploy'", cfg.Name)
		issues++
	} else if st.DeployedSHA != currentSHA {
		printCheck(false, "Binary (%s) differs from deployed (%s) — run '%s deploy'", currentSHA, st.DeployedSHA, cfg.Name)
		issues++
	} else {
		printCheck(true, "Binary matches deployed SHA (%s)", currentSHA)
	}

	// Check 2: Uncommitted changes
	if git.IsDirty() {
		printCheck(false, "Uncommitted changes in repo — deployed version may differ from working tree")
		issues++
	} else {
		printCheck(true, "Working tree clean")
	}

	// Check 3: Config files exist
	manifestPath := filepath.Join(cfg.RepoDir, cfg.Name+".yaml")
	configPath := filepath.Join(cfg.RepoDir, cfg.Name+".config.yaml")
	if fileutil.Exists(manifestPath) && fileutil.Exists(configPath) {
		printCheck(true, "%s.yaml and %s.config.yaml present", cfg.Name, cfg.Name)
	} else {
		if !fileutil.Exists(manifestPath) {
			printCheck(false, "%s.yaml missing", cfg.Name)
			issues++
		}
		if !fileutil.Exists(configPath) {
			printCheck(false, "%s.config.yaml missing", cfg.Name)
			issues++
		}
	}

	fmt.Println()
	fmt.Println(output.Bold("Targets:"))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Check 5: Claude target
	if cfg.TargetEnabled(config.TargetClaude) {
		issues += checkTarget(config.TargetClaude, paths.Claude, []targetCheck{
			{name: config.CompAgents, checkType: "dir_or_symlink"},
			{name: config.CompSkills, checkType: "dir_or_symlink"},
			{name: config.CompCommands, checkType: "dir_or_symlink"},
			{name: config.FileClaudeMD, checkType: "file_or_symlink"},
		})
	} else {
		printCheck(true, "%s — disabled (skipped)", config.TargetClaude)
	}

	// Check 6: OpenCode target
	if cfg.TargetEnabled(config.TargetOpenCode) {
		issues += checkTarget(config.TargetOpenCode, paths.OpenCode, []targetCheck{
			{name: config.CompAgents, checkType: "dir_or_symlink"},
			{name: config.CompCommands, checkType: "dir_or_symlink"},
		})
	} else {
		printCheck(true, "%s — disabled (skipped)", config.TargetOpenCode)
	}

	// Check 7: Gemini target
	if cfg.TargetEnabled(config.TargetGemini) {
		issues += checkTarget(config.TargetGemini, paths.Gemini, []targetCheck{
			{name: config.CompSkills, checkType: "dir_or_symlink"},
			{name: config.CompCommands, checkType: "dir_or_symlink"},
			{name: config.FileGeminiMD, checkType: "file_or_symlink"},
		})
	} else {
		printCheck(true, "%s — disabled (skipped)", config.TargetGemini)
	}

	// Check 8: Codex target
	if cfg.TargetEnabled(config.TargetCodex) {
		issues += checkTarget(config.TargetCodex, paths.Codex, []targetCheck{
			{name: config.CompSkills, checkType: "dir_or_symlink"},
			{name: config.FileAgentsMD, checkType: "file_or_symlink"},
		})
	} else {
		printCheck(true, "%s — disabled (skipped)", config.TargetCodex)
	}

	// Check 9: Broken symlinks
	fmt.Println()
	fmt.Println(output.Bold("Symlinks:"))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	brokenLinks := checkBrokenSymlinks(paths, cfg)
	if brokenLinks > 0 {
		issues += brokenLinks
	} else {
		printCheck(true, "No broken symlinks detected")
	}

	// Summary
	fmt.Println()
	if issues == 0 {
		output.Info("All checks passed")
	} else {
		output.Warn("%d issue(s) found", issues)
	}
	fmt.Println()
}

type targetCheck struct {
	name      string
	checkType string
}

func checkTarget(target, basePath string, checks []targetCheck) int {
	issues := 0
	if !fileutil.IsDir(basePath) {
		printCheck(false, "%s — base directory missing: %s", target, basePath)
		return 1
	}
	for _, c := range checks {
		p := filepath.Join(basePath, c.name)
		switch c.checkType {
		case "dir_or_symlink":
			if fileutil.IsSymlink(p) || fileutil.IsDir(p) {
				printCheck(true, "%s/%s", target, c.name)
			} else {
				printCheck(false, "%s/%s missing", target, c.name)
				issues++
			}
		case "file_or_symlink":
			if fileutil.IsSymlink(p) || fileutil.Exists(p) {
				printCheck(true, "%s/%s", target, c.name)
			} else {
				printCheck(false, "%s/%s missing", target, c.name)
				issues++
			}
		}
	}
	return issues
}

func checkBrokenSymlinks(paths deploy.TargetPaths, cfg *config.App) int {
	broken := 0
	dirs := map[string]string{
		config.TargetClaude:   paths.Claude,
		config.TargetOpenCode: paths.OpenCode,
		config.TargetGemini:   paths.Gemini,
		config.TargetCodex:    paths.Codex,
	}
	for target, base := range dirs {
		if !cfg.TargetEnabled(target) {
			continue
		}
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, e := range entries {
			p := filepath.Join(base, e.Name())
			if fileutil.IsSymlink(p) {
				dest, err := os.Readlink(p)
				if err != nil || !fileutil.Exists(dest) {
					printCheck(false, "%s/%s -> %s (broken)", target, e.Name(), dest)
					broken++
				}
			}
		}
	}
	return broken
}
