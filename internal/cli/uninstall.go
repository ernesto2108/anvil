package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/internal/deploy"
	"github.com/ernesto2108/anvil/pkg/fileutil"
	"github.com/ernesto2108/anvil/pkg/output"
)

func cmdUninstall(cfg *config.App) {
	st := loadState(cfg)
	paths := deploy.ResolvePaths()

	fmt.Println()
	output.Warn("This will remove %s from ALL targets:", cfg.Name)
	output.Warn("  Claude:   %s/{agents,skills,commands,CLAUDE.md}", paths.Claude)
	output.Warn("  OpenCode: %s/{agents,commands}", paths.OpenCode)
	output.Warn("  Gemini:   %s/{skills,commands,GEMINI.md}", paths.Gemini)
	output.Warn("  Codex:    %s/{skills,AGENTS.md}", paths.Codex)
	fmt.Println()

	fmt.Print("Continue? [y/N] ")
	reader := bufio.NewReader(os.Stdin)
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)
	if confirm != "y" && confirm != "Y" {
		output.Info("Aborted.")
		return
	}
	fmt.Println()

	snap := st.SnapshotDir()

	// Claude
	output.Info("%s", output.Bold("Claude Code:"))
	for _, comp := range []string{config.CompAgents, config.CompSkills, config.CompCommands} {
		deploy.RestoreItem(filepath.Join(paths.Claude, comp), filepath.Join(snap, config.TargetClaude, comp))
	}
	claudeMD := filepath.Join(paths.Claude, config.FileClaudeMD)
	if fileutil.IsSymlink(claudeMD) {
		os.Remove(claudeMD)
		snapMD := filepath.Join(snap, config.TargetClaude, config.FileClaudeMD)
		if fileutil.Exists(snapMD) {
			fileutil.CopyFile(snapMD, claudeMD)
			output.Info("  %s %s", output.Green("restored"), config.FileClaudeMD)
		}
	}

	// OpenCode
	output.Info("%s", output.Bold("OpenCode:"))
	for _, comp := range []string{config.CompAgents, config.CompCommands} {
		deploy.RestoreItem(filepath.Join(paths.OpenCode, comp), filepath.Join(snap, config.TargetOpenCode, comp))
	}

	// Gemini
	output.Info("%s", output.Bold("Gemini CLI:"))
	for _, comp := range []string{config.CompSkills, config.CompCommands} {
		deploy.RestoreItem(filepath.Join(paths.Gemini, comp), filepath.Join(snap, config.TargetGemini, comp))
	}
	geminiMD := filepath.Join(paths.Gemini, config.FileGeminiMD)
	if fileutil.Exists(geminiMD) {
		os.Remove(geminiMD)
		snapGMD := filepath.Join(snap, config.TargetGemini, config.FileGeminiMD)
		if fileutil.Exists(snapGMD) {
			fileutil.CopyFile(snapGMD, geminiMD)
		}
	}

	// Codex
	output.Info("%s", output.Bold("Codex:"))
	deploy.RestoreItem(filepath.Join(paths.Codex, config.CompSkills), filepath.Join(snap, config.TargetCodex, config.CompSkills))
	agentsMD := filepath.Join(paths.Codex, config.FileAgentsMD)
	if fileutil.Exists(agentsMD) {
		os.Remove(agentsMD)
		snapAMD := filepath.Join(snap, config.TargetCodex, config.FileAgentsMD)
		if fileutil.Exists(snapAMD) {
			fileutil.CopyFile(snapAMD, agentsMD)
		}
	}
	output.Info("  removed %s", config.FileAgentsMD)

	// Clean repo AGENTS.md
	repoAgentsMD := filepath.Join(cfg.RepoDir, config.FileAgentsMD)
	if fileutil.Exists(repoAgentsMD) {
		os.Remove(repoAgentsMD)
	}

	st.Remove()
	fmt.Println()
	output.Info("%s uninstalled. Pre-existing files restored where snapshots existed.", title(cfg.Name))
	output.Info("Run %s to reinstall.", output.Yellow(cfg.Name+" deploy"))
	fmt.Println()
}
