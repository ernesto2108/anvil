package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/output"
)

func cmdTargets(cfg *config.App, args []string) {
	allTargets := cfg.AllTargets()

	if len(args) == 0 {
		fmt.Println()
		fmt.Println(output.Bold("Deploy targets:"))
		fmt.Println()
		for _, t := range allTargets {
			if cfg.TargetEnabled(t) {
				fmt.Printf("  %s %s\n", output.Green("●"), t)
			} else {
				fmt.Printf("  %s %s\n", output.Red("○"), t)
			}
		}
		fmt.Println()
		fmt.Printf("Usage:\n")
		fmt.Printf("  %s targets <tool> [tool...]    set exact targets\n", cfg.Name)
		fmt.Printf("  %s targets --add <tool>        enable one target\n", cfg.Name)
		fmt.Printf("  %s targets --rm <tool>         disable one target\n", cfg.Name)
		fmt.Printf("  %s targets all                 enable all\n", cfg.Name)
		return
	}

	mode := "set"
	var tools []string

	switch args[0] {
	case "--add":
		mode = "add"
		tools = args[1:]
	case "--rm":
		mode = "rm"
		tools = args[1:]
	case "all":
		mode = "set"
		tools = allTargets
	default:
		mode = "set"
		tools = args
	}

	if len(tools) == 0 {
		output.Error("No targets specified")
		os.Exit(1)
	}

	// Validate
	valid := make(map[string]bool)
	for _, t := range allTargets {
		valid[t] = true
	}
	for _, t := range tools {
		if !valid[t] {
			output.Error("Unknown target: %s", t)
			output.Error("Valid targets: %s", strings.Join(allTargets, ", "))
			os.Exit(1)
		}
	}

	// Apply
	requested := make(map[string]bool)
	for _, t := range tools {
		requested[t] = true
	}

	switch mode {
	case "add":
		for _, t := range tools {
			cfg.SetTargetEnabled(t, true)
		}
	case "rm":
		for _, t := range tools {
			cfg.SetTargetEnabled(t, false)
		}
	case "set":
		for _, t := range allTargets {
			cfg.SetTargetEnabled(t, requested[t])
		}
	}

	output.Info("Targets updated:")
	for _, t := range allTargets {
		if cfg.TargetEnabled(t) {
			fmt.Printf("  %s %s\n", output.Green("●"), t)
		} else {
			fmt.Printf("  %s %s\n", output.Red("○"), t)
		}
	}
	fmt.Println()
	output.Info("Run %s to apply.", output.Yellow(cfg.Name+" deploy"))
}
