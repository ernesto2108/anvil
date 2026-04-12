package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/ernesto2108/anvil/internal/orchestrator"
	"gopkg.in/yaml.v3"
)

type pipelineFile struct {
	Nodes []pipelineNode `yaml:"nodes"`
}

type pipelineNode struct {
	ID        string   `yaml:"id"`
	Role      string   `yaml:"role"`
	DependsOn []string `yaml:"depends_on"`
	Gate      bool     `yaml:"gate"`
	OnFail    string   `yaml:"on_fail"`
	Timeout   string   `yaml:"timeout"` // e.g. "5m", "30s"
}

func loadPipeline(path string) ([]orchestrator.Node, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cli/pipeline: read %s: %w", path, err)
	}
	var pf pipelineFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return nil, fmt.Errorf("cli/pipeline: parse %s: %w", path, err)
	}
	if len(pf.Nodes) == 0 {
		return nil, fmt.Errorf("cli/pipeline: no nodes defined in %s", path)
	}
	nodes := make([]orchestrator.Node, len(pf.Nodes))
	for i, pn := range pf.Nodes {
		if pn.ID == "" {
			return nil, fmt.Errorf("cli/pipeline: node %d missing id", i)
		}
		if pn.Role == "" {
			return nil, fmt.Errorf("cli/pipeline: node %q missing role", pn.ID)
		}
		if err := orchestrator.ValidateRole(pn.Role); err != nil {
			return nil, fmt.Errorf("cli/pipeline: node %q: %w", pn.ID, err)
		}
		var timeout time.Duration
		if pn.Timeout != "" {
			var parseErr error
			timeout, parseErr = time.ParseDuration(pn.Timeout)
			if parseErr != nil {
				return nil, fmt.Errorf("cli/pipeline: node %q invalid timeout %q: %w", pn.ID, pn.Timeout, parseErr)
			}
		}
		nodes[i] = orchestrator.Node{
			ID:        pn.ID,
			Role:      pn.Role,
			DependsOn: pn.DependsOn,
			Gate:      pn.Gate,
			OnFail:    orchestrator.FailPolicy(pn.OnFail),
			Timeout:   timeout,
		}
	}
	return nodes, nil
}
