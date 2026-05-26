package cli

import (
	"fmt"
	"os"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/internal/deploy"
	"github.com/ernesto2108/anvil/pkg/output"
)

func cmdProvider(cfg *config.App, args []string) {
	if len(args) == 0 {
		current := cfg.ActiveProvider()
		fmt.Println()
		fmt.Printf("%s %s\n", output.Cyan("Current provider:"), output.Green(current))
		fmt.Println()
		fmt.Println(output.Cyan("Available providers:"))
		for _, p := range cfg.ListProviders() {
			if p == current {
				fmt.Printf("  %s (active)\n", output.Green(p))
			} else {
				fmt.Printf("  %s\n", p)
			}
		}
		fmt.Println()
		fmt.Printf("Usage: %s provider <name>\n", cfg.Name)
		return
	}

	newProvider := args[0]

	// Validate
	found := false
	for _, p := range cfg.ListProviders() {
		if p == newProvider {
			found = true
			break
		}
	}
	if !found {
		output.Error("Provider '%s' not found in %s.config.yaml", newProvider, cfg.Name)
		cmdProvider(cfg, nil)
		os.Exit(1)
	}

	if err := cfg.SetProvider(newProvider); err != nil {
		output.Error("set provider: %s", err)
		os.Exit(1)
	}

	output.Info("Provider switched to %s", output.Green(newProvider))
	fmt.Println()

	// Redeploy agents
	paths := deploy.ResolvePaths()
	output.Info("Redeploying agents...")

	if cfg.TargetEnabled(config.TargetClaude) {
		deploy.ClaudeAgentsOnly(cfg, paths)
	}
	fmt.Println()

	if cfg.TargetEnabled(config.TargetOpenCode) {
		deploy.OpenCodeAgentsOnly(cfg, paths)
	}

	st := loadState(cfg)
	st.Provider = newProvider
	if err := st.Save(); err != nil {
		output.Warn("save state: %s", err)
	}

	fmt.Println()
	output.Info("Done. All agents now use %s models.", output.Green(newProvider))
}
