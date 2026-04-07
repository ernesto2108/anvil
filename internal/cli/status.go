package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/internal/deploy"
	"github.com/ernesto2108/anvil/pkg/fileutil"
	"github.com/ernesto2108/anvil/pkg/gitutil"
	"github.com/ernesto2108/anvil/pkg/output"
)

func cmdStatus(cfg *config.App, git *gitutil.Repo) {
	st := loadState(cfg)
	paths := deploy.ResolvePaths()

	fmt.Println()
	fmt.Println(output.Bold(title(cfg.Name) + " Status"))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	fmt.Printf("  Repo:      %s\n", cfg.RepoDir)
	fmt.Printf("  Provider:  %s\n", output.Green(st.Provider))
	fmt.Printf("  Branch:    %s\n", git.CurrentBranch())
	fmt.Printf("  HEAD:      %s\n", git.CurrentSHA())
	fmt.Printf("  Tag:       %s\n", git.CurrentTag())
	fmt.Printf("  Deployed:  %s\n", output.Green(st.DeployedVersion))
	fmt.Printf("  Previous:  %s\n", st.PreviousVersion)
	fmt.Printf("  At:        %s\n", st.DeployedAt)
	fmt.Println()

	fmt.Println(output.Bold("Targets:"))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Claude
	claudeIcon := output.Red("○")
	if fileutil.IsDir(filepath.Join(paths.Claude, config.CompAgents)) || fileutil.IsSymlink(filepath.Join(paths.Claude, config.CompSkills)) {
		claudeIcon = output.Green("●")
	}
	fmt.Printf("  %s %-12s %s\n", claudeIcon, config.TargetClaude, paths.Claude)

	for _, comp := range []string{config.CompAgents, config.CompSkills, config.CompCommands, config.FileClaudeMD} {
		p := filepath.Join(paths.Claude, comp)
		if fileutil.IsSymlink(p) {
			fmt.Printf("    %s %s (symlink)\n", output.Green("●"), comp)
		} else if fileutil.Exists(p) {
			fmt.Printf("    %s %s (copy)\n", output.Yellow("◐"), comp)
		} else {
			fmt.Printf("    %s %s\n", output.Red("○"), comp)
		}
	}

	// OpenCode
	ocIcon := output.Red("○")
	if fileutil.IsDir(filepath.Join(paths.OpenCode, config.CompAgents)) {
		ocIcon = output.Green("●")
	}
	fmt.Printf("  %s %-12s %s\n", ocIcon, config.TargetOpenCode, paths.OpenCode)

	// Gemini
	gemIcon := output.Red("○")
	if fileutil.IsDir(filepath.Join(paths.Gemini, config.CompCommands)) || fileutil.IsSymlink(filepath.Join(paths.Gemini, config.CompSkills)) {
		gemIcon = output.Green("●")
	}
	fmt.Printf("  %s %-12s %s\n", gemIcon, config.TargetGemini, paths.Gemini)

	// Codex
	cdxIcon := output.Red("○")
	if fileutil.IsSymlink(filepath.Join(paths.Codex, config.CompSkills)) || fileutil.Exists(filepath.Join(paths.Codex, config.FileAgentsMD)) {
		cdxIcon = output.Green("●")
	}
	fmt.Printf("  %s %-12s %s\n", cdxIcon, config.TargetCodex, paths.Codex)

	fmt.Println()
	fmt.Printf("  %s deployed  %s copy/pinned  %s not deployed\n", output.Green("●"), output.Yellow("◐"), output.Red("○"))

	// Recent tags
	tagsOut, err := git.Tags()
	if err == nil && tagsOut != "" {
		fmt.Println()
		fmt.Println(output.Cyan("Recent tags:"))
		for i, t := range strings.Split(tagsOut, "\n") {
			if i >= 5 {
				break
			}
			fmt.Printf("  %s\n", t)
		}
	}
	fmt.Println()
}
