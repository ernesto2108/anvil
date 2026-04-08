package deploy

import (
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/fileutil"
	"github.com/ernesto2108/anvil/pkg/output"
)

type TargetPaths struct {
	Claude   string
	OpenCode string
	Gemini   string
	Codex    string
}

func ResolvePaths() TargetPaths {
	home, _ := os.UserHomeDir()
	claudeHome := os.Getenv("CLAUDE_HOME")
	if claudeHome == "" {
		claudeHome = filepath.Join(home, ".claude")
	}

	return TargetPaths{
		Claude:   claudeHome,
		OpenCode: filepath.Join(home, ".config", "opencode"),
		Gemini:   filepath.Join(home, ".gemini"),
		Codex:    filepath.Join(home, ".codex"),
	}
}

func SnapshotItem(src, dest string) bool {
	fi, err := os.Lstat(src)
	if err != nil {
		return false
	}

	// Symlink: save a .symlink file with the target path
	if fi.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(src)
		if err != nil {
			return false
		}
		os.MkdirAll(filepath.Dir(dest), 0o755)
		if err := os.WriteFile(dest+".symlink", []byte(target), 0o644); err != nil {
			return false
		}
		output.Info("  %s %s -> %s", output.Cyan("saved"), src, target)
		return true
	}

	if fi.IsDir() {
		if err := fileutil.CopyDir(src, dest); err != nil {
			return false
		}
		output.Info("  %s %s", output.Cyan("saved"), src)
		return true
	}

	if err := fileutil.CopyFile(src, dest); err != nil {
		return false
	}
	output.Info("  %s %s", output.Cyan("saved"), src)
	return true
}

func RestoreItem(current, snapshot string) {
	fileutil.CleanPath(current)

	// Check for saved symlink first
	symlinkFile := snapshot + ".symlink"
	if fileutil.Exists(symlinkFile) {
		data, err := os.ReadFile(symlinkFile)
		if err == nil {
			target := string(data)
			os.Symlink(target, current)
			output.Info("  %s %s -> %s", output.Green("restored"), current, target)
			return
		}
	}

	if fileutil.Exists(snapshot) {
		fi, _ := os.Stat(snapshot)
		if fi != nil && fi.IsDir() {
			fileutil.CopyDir(snapshot, current)
		} else {
			fileutil.CopyFile(snapshot, current)
		}
		output.Info("  %s %s", output.Green("restored"), current)
		return
	}
	output.Info("  removed %s", current)
}

// DeploySkillsSymlink creates individual symlinks per skill directory.
// User-managed skills (not symlinked to anvil repo) are preserved.
func DeploySkillsSymlink(cfg *config.App, targetDir string, ts *TargetStats) {
	skillsSrc := filepath.Join(cfg.RepoDir, config.CompSkills)
	if !fileutil.IsDir(skillsSrc) {
		return
	}

	skillsDst := filepath.Join(targetDir, config.CompSkills)

	// If target is currently a single directory symlink, remove it (migration from old format)
	if fileutil.IsSymlink(skillsDst) {
		os.Remove(skillsDst)
	}
	os.MkdirAll(skillsDst, 0o755)

	cleanAnvilSymlinks(skillsDst, skillsSrc)

	include, exclude := cfg.ComponentFilter(config.CompSkills)

	entries, err := os.ReadDir(skillsSrc)
	if err != nil {
		return
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()

		if !config.IsComponentAllowed(name, include, exclude) {
			ts.Skills.Skipped++
			continue
		}

		src := filepath.Join(skillsSrc, name)
		dst := filepath.Join(skillsDst, name)

		if fileutil.IsDir(dst) && !fileutil.IsSymlink(dst) {
			ts.Skills.Preserved++
			continue
		}

		if err := fileutil.ForceSymlink(src, dst); err != nil {
			output.Error("  symlink skills/%s: %s", name, err)
			continue
		}
		ts.Skills.Deployed++
	}
}

// DeployCommandsSymlink creates individual symlinks per command file.
func DeployCommandsSymlink(cfg *config.App, targetDir string, ts *TargetStats) {
	cmdSrc := filepath.Join(cfg.RepoDir, config.CompCommands)
	if !fileutil.IsDir(cmdSrc) {
		return
	}

	cmdsDst := filepath.Join(targetDir, config.CompCommands)

	if fileutil.IsSymlink(cmdsDst) {
		os.Remove(cmdsDst)
	}
	os.MkdirAll(cmdsDst, 0o755)

	cleanAnvilSymlinks(cmdsDst, cmdSrc)

	include, exclude := cfg.ComponentFilter(config.CompCommands)

	entries, err := os.ReadDir(cmdSrc)
	if err != nil {
		return
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()

		if !config.IsComponentAllowed(name, include, exclude) {
			ts.Commands.Skipped++
			continue
		}

		src := filepath.Join(cmdSrc, name)
		dst := filepath.Join(cmdsDst, name)

		if fileutil.Exists(dst) && !fileutil.IsSymlink(dst) {
			ts.Commands.Preserved++
			continue
		}

		if err := fileutil.ForceSymlink(src, dst); err != nil {
			output.Error("  symlink commands/%s: %s", name, err)
			continue
		}
		ts.Commands.Deployed++
	}
}

// cleanAnvilSymlinks removes symlinks in dst that point into srcBase (anvil repo).
// This ensures stale anvil symlinks are cleaned before re-linking.
func cleanAnvilSymlinks(dst, srcBase string) {
	entries, err := os.ReadDir(dst)
	if err != nil {
		return
	}
	for _, e := range entries {
		path := filepath.Join(dst, e.Name())
		if !fileutil.IsSymlink(path) {
			continue
		}
		target, err := os.Readlink(path)
		if err != nil {
			continue
		}
		// Remove if it points into the anvil repo
		if len(target) >= len(srcBase) && target[:len(srcBase)] == srcBase {
			os.Remove(path)
		}
	}
}

func AgentFiles(repoDir string) []string {
	agentDir := filepath.Join(repoDir, config.CompAgents)
	entries, err := os.ReadDir(agentDir)
	if err != nil {
		return nil
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".md" {
			files = append(files, filepath.Join(agentDir, e.Name()))
		}
	}
	return files
}

// FilteredAgentFiles returns agent files filtered by include/exclude from config.
func FilteredAgentFiles(cfg *config.App) []string {
	all := AgentFiles(cfg.RepoDir)
	include, exclude := cfg.ComponentFilter(config.CompAgents)
	if len(include) == 0 && len(exclude) == 0 {
		return all
	}

	var filtered []string
	for _, f := range all {
		name := filepath.Base(f)
		if config.IsComponentAllowed(name, include, exclude) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func DeployedTargets(cfg *config.App) []string {
	var targets []string
	for _, t := range cfg.AllTargets() {
		if cfg.TargetEnabled(t) {
			targets = append(targets, t)
		}
	}
	return targets
}
