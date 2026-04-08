package cli

import (
	"fmt"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/gitutil"
	"github.com/ernesto2108/anvil/pkg/output"
)

func cmdInit(cfg *config.App, git *gitutil.Repo) {
	fmt.Println()
	output.Info("%s", output.Bold(title(cfg.Name)+" init"))
	fmt.Println()

	// Step 1: Show current config
	output.Info("Repo:     %s", cfg.RepoDir)
	output.Info("Branch:   %s", git.CurrentBranch())
	output.Info("Provider: %s", output.Green(cfg.ActiveProvider()))
	fmt.Println()

	// Step 2: Show enabled targets
	output.Info("Enabled targets:")
	for _, t := range cfg.AllTargets() {
		if cfg.TargetEnabled(t) {
			fmt.Printf("  %s %s\n", output.Green("●"), t)
		} else {
			fmt.Printf("  %s %s\n", output.Red("○"), t)
		}
	}
	fmt.Println()

	// Step 3: Launch browse
	output.Info("Launching browser to select agents, skills, and commands...")
	fmt.Println()

	registryBrowse(cfg, git)
}
