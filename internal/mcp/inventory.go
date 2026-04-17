package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ernesto2108/anvil/pkg/frontmatter"
	"gopkg.in/yaml.v3"
)

// listAgents reads agents/*.md and returns parsed frontmatter for each.
func (s *Server) listAgents(_ context.Context, _ map[string]any) (string, error) {
	files, err := filepath.Glob(filepath.Join(s.cfg.RepoDir, "agents", "*.md"))
	if err != nil {
		return "", fmt.Errorf("glob agents: %w", err)
	}

	type agentInfo struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Permission  string `json:"permission"`
		Model       string `json:"model"`
	}

	agents := make([]agentInfo, 0, len(files))
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		doc := frontmatter.Parse(string(data))
		agents = append(agents, agentInfo{
			Name:        doc.Fields["name"],
			Description: doc.Fields["description"],
			Permission:  doc.Fields["permission"],
			Model:       doc.Fields["model"],
		})
	}

	b, _ := json.MarshalIndent(agents, "", "  ")
	return string(b), nil
}

// listSkills reads skills/*/SKILL.md and returns name + description from each.
func (s *Server) listSkills(_ context.Context, _ map[string]any) (string, error) {
	pattern := filepath.Join(s.cfg.RepoDir, "skills", "*", "SKILL.md")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return "", fmt.Errorf("glob skills: %w", err)
	}

	type skillInfo struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	skills := make([]skillInfo, 0, len(files))
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		doc := frontmatter.Parse(string(data))
		name := doc.Fields["name"]
		if name == "" {
			// Derive from directory name.
			name = filepath.Base(filepath.Dir(f))
		}
		skills = append(skills, skillInfo{
			Name:        name,
			Description: doc.Fields["description"],
		})
	}

	b, _ := json.MarshalIndent(skills, "", "  ")
	return string(b), nil
}

// listPipelines reads pipelines/*.yaml and returns their node definitions.
func (s *Server) listPipelines(_ context.Context, _ map[string]any) (string, error) {
	files, err := filepath.Glob(filepath.Join(s.cfg.RepoDir, "pipelines", "*.yaml"))
	if err != nil {
		return "", fmt.Errorf("glob pipelines: %w", err)
	}

	type node struct {
		ID        string   `yaml:"id"`
		Role      string   `yaml:"role"`
		DependsOn []string `yaml:"depends_on"`
		Gate      bool     `yaml:"gate"`
	}
	type pipeline struct {
		Nodes []node `yaml:"nodes"`
	}

	type pipelineInfo struct {
		Name  string   `json:"name"`
		Nodes []string `json:"nodes"`
		Gates []string `json:"gates,omitempty"`
	}

	result := make([]pipelineInfo, 0, len(files))
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var p pipeline
		if err := yaml.Unmarshal(data, &p); err != nil {
			continue
		}
		name := strings.TrimSuffix(filepath.Base(f), ".yaml")
		roles := make([]string, 0, len(p.Nodes))
		var gates []string
		for _, n := range p.Nodes {
			roles = append(roles, n.Role)
			if n.Gate {
				gates = append(gates, n.Role)
			}
		}
		result = append(result, pipelineInfo{Name: name, Nodes: roles, Gates: gates})
	}

	b, _ := json.MarshalIndent(result, "", "  ")
	return string(b), nil
}

// getProjectInfo returns a summary of the current project: name, stack, provider, last run.
func (s *Server) getProjectInfo(_ context.Context, _ map[string]any) (string, error) {
	workDir, _ := os.Getwd()
	project := filepath.Base(workDir)

	type info struct {
		Project  string `json:"project"`
		Provider string `json:"provider"`
		RepoDir  string `json:"repo_dir"`
		VaultSet bool   `json:"vault_configured"`
	}

	out := info{
		Project:  project,
		Provider: s.cfg.Provider.Provider,
		RepoDir:  s.cfg.RepoDir,
		VaultSet: s.vaultDir != "",
	}

	// Count agents and skills.
	agentFiles, _ := filepath.Glob(filepath.Join(s.cfg.RepoDir, "agents", "*.md"))
	skillDirs, _ := filepath.Glob(filepath.Join(s.cfg.RepoDir, "skills", "*", "SKILL.md"))
	pipelines, _ := filepath.Glob(filepath.Join(s.cfg.RepoDir, "pipelines", "*.yaml"))

	type fullInfo struct {
		info
		AgentsCount   int `json:"agents_count"`
		SkillsCount   int `json:"skills_count"`
		PipelinesCount int `json:"pipelines_count"`
	}

	b, _ := json.MarshalIndent(fullInfo{
		info:          out,
		AgentsCount:   len(agentFiles),
		SkillsCount:   len(skillDirs),
		PipelinesCount: len(pipelines),
	}, "", "  ")
	return string(b), nil
}
