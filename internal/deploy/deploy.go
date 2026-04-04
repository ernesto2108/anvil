package deploy

import (
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/internal/config"
	"github.com/ernesto2108/anvil/internal/fileutil"
	"github.com/ernesto2108/anvil/internal/output"
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

func DeploySkillsSymlink(repoDir, targetDir string) {
	skillsSrc := filepath.Join(repoDir, "skills")
	if !fileutil.IsDir(skillsSrc) {
		return
	}
	if err := fileutil.ForceSymlink(skillsSrc, filepath.Join(targetDir, "skills")); err != nil {
		output.Error("symlink skills: %s", err)
		return
	}
	output.Info("  skills -> symlink")
}

func DeployCommandsSymlink(repoDir, targetDir string) {
	cmdSrc := filepath.Join(repoDir, "commands")
	if !fileutil.IsDir(cmdSrc) {
		return
	}
	if err := fileutil.ForceSymlink(cmdSrc, filepath.Join(targetDir, "commands")); err != nil {
		output.Error("symlink commands: %s", err)
		return
	}
	output.Info("  commands -> symlink")
}

func AgentFiles(repoDir string) []string {
	agentDir := filepath.Join(repoDir, "agents")
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

func DeployedTargets(cfg *config.App) []string {
	var targets []string
	for _, t := range cfg.AllTargets() {
		if cfg.TargetEnabled(t) {
			targets = append(targets, t)
		}
	}
	return targets
}
