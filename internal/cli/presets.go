package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/internal/orchestrator"
	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/output"
)

// builtinPresets lists the command names that map to pipelines/<name>.yaml.
var builtinPresets = map[string]bool{
	"quick":  true,
	"bug":    true,
	"feat":   true,
	"design": true,
	"epic":   true,
	"db":     true,
	"infra":  true,
}

// resolvePipeline searches for <name>.yaml in ./pipelines/ and then <repoDir>/pipelines/.
func resolvePipeline(cfg *config.App, name string) ([]orchestrator.Node, error) {
	candidates := []string{
		filepath.Join("pipelines", name+".yaml"),
		filepath.Join(cfg.RepoDir, "pipelines", name+".yaml"),
	}
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return loadPipeline(path)
		}
	}
	return nil, fmt.Errorf("pipeline %q not found (searched %v)", name, candidates)
}

func cmdPreset(cfg *config.App, presetName string, args []string) {
	flags, _ := parseRunFlags(args)
	if flags.task == "" {
		output.Error("usage: anvil %s --task \"description\" [--model name] [-y]", presetName)
		os.Exit(1)
	}

	nodes, err := resolvePipeline(cfg, presetName)
	if err != nil {
		output.Error("%s", err)
		os.Exit(1)
	}

	output.Info("Pipeline: %s (%d agents)", output.Cyan(presetName), len(nodes))
	executeRun(cfg, nodes, flags)
}
