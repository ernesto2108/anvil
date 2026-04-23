package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ernesto2108/anvil/internal/deploy"
	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/fileutil"
	"github.com/ernesto2108/anvil/pkg/gitutil"
	"github.com/ernesto2108/anvil/pkg/state"
)

// healthCheck is a single doctor check result.
type healthCheck struct {
	Name   string `json:"name"`
	Status string `json:"status"` // "ok" | "warn"
	Detail string `json:"detail"`
}

// switchProvider changes the active provider in the config file and redeploys
// agents so model tier mappings take effect immediately.
// Input:  { provider: string }
// Output: { previous, current, models: { high, medium, low } }
func (s *Server) switchProvider(_ context.Context, args map[string]any) (string, error) {
	provider := strArg(args, "provider", "")
	if provider == "" {
		return "", fmt.Errorf("provider is required")
	}

	// Validate provider exists in config.
	available := s.cfg.ListProviders()
	found := false
	for _, p := range available {
		if p == provider {
			found = true
			break
		}
	}
	if !found {
		result := map[string]any{
			"error":               fmt.Sprintf("provider %q not found", provider),
			"available_providers": available,
		}
		out, _ := json.MarshalIndent(result, "", "  ")
		return string(out), nil
	}

	previous := s.cfg.ActiveProvider()

	if err := s.cfg.SetProvider(provider); err != nil {
		return "", fmt.Errorf("set provider: %w", err)
	}

	// Redeploy agents for enabled targets.
	paths := deploy.ResolvePaths()
	if s.cfg.TargetEnabled(config.TargetClaude) {
		deploy.ClaudeAgentsOnly(s.cfg, paths)
	}
	if s.cfg.TargetEnabled(config.TargetOpenCode) {
		deploy.OpenCodeAgentsOnly(s.cfg, paths)
	}

	// Persist provider to state file.
	stateDir := filepath.Join(s.cfg.RepoDir, "."+s.cfg.Name)
	if st, err := state.Load(stateDir); err == nil {
		st.Provider = provider
		_ = st.Save()
	}

	// Resolve tier model names for the new provider.
	high, _ := s.cfg.ResolveTier(config.TierHigh, provider)
	medium, _ := s.cfg.ResolveTier(config.TierMedium, provider)
	low, _ := s.cfg.ResolveTier(config.TierLow, provider)

	result := map[string]any{
		"previous": previous,
		"current":  provider,
		"models": map[string]any{
			"high":   high,
			"medium": medium,
			"low":    low,
		},
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

// getDiff returns the diff between the last deployed SHA (or an explicit base)
// and HEAD, formatted as a stat summary.
// Input:  { base?: string, stat_only?: bool }
// Output: { branch, deployed_sha, current_sha, files_changed, diff_summary }
func (s *Server) getDiff(_ context.Context, args map[string]any) (string, error) {
	stateDir := filepath.Join(s.cfg.RepoDir, "."+s.cfg.Name)
	st, err := state.Load(stateDir)
	if err != nil {
		return "", fmt.Errorf("load state: %w", err)
	}

	git := gitutil.New(s.cfg.RepoDir)
	branch := git.CurrentBranch()
	currentSHA := git.CurrentSHA()

	// Determine base SHA: explicit arg > deployed SHA > error.
	base := strArg(args, "base", "")
	if base == "" {
		if st.DeployedSHA == "" || st.DeployedSHA == state.StateNone {
			result := map[string]any{
				"error": "nothing deployed yet — no base SHA available; pass base explicitly",
			}
			out, _ := json.MarshalIndent(result, "", "  ")
			return string(out), nil
		}
		base = st.DeployedSHA
	}

	if base == currentSHA {
		result := map[string]any{
			"branch":        branch,
			"deployed_sha":  base,
			"current_sha":   currentSHA,
			"files_changed": 0,
			"diff_summary":  "no changes since last deploy",
		}
		out, _ := json.MarshalIndent(result, "", "  ")
		return string(out), nil
	}

	diffOut, diffErr := git.DiffStat(base)
	if diffErr != nil {
		return "", fmt.Errorf("git diff: %w", diffErr)
	}

	result := map[string]any{
		"branch":        branch,
		"deployed_sha":  base,
		"current_sha":   currentSHA,
		"files_changed": countFilesChanged(diffOut),
		"diff_summary":  diffOut,
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

// countFilesChanged extracts N from "N file(s) changed" in git diff --stat output.
func countFilesChanged(diffOut string) int {
	if diffOut == "" {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(diffOut), "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, "file") {
			var n int
			if _, scanErr := fmt.Sscanf(line, "%d", &n); scanErr == nil {
				return n
			}
		}
	}
	return 0
}

// deployAgents deploys agent files, skills, and/or commands to all enabled targets.
// Input:  { target?: "all"|"agents"|"skills" }  — default "all"
// Output: { target, deployed_agents: [...], deployed_skills: [...] }
func (s *Server) deployAgents(_ context.Context, args map[string]any) (string, error) {
	target := strArg(args, "target", "all")
	if target != "all" && target != "agents" && target != "skills" {
		return "", fmt.Errorf("target must be one of: all, agents, skills")
	}

	paths := deploy.ResolvePaths()
	deploy.StartSummary()

	switch target {
	case "agents":
		if s.cfg.TargetEnabled(config.TargetClaude) {
			deploy.ClaudeAgentsOnly(s.cfg, paths)
		}
		if s.cfg.TargetEnabled(config.TargetOpenCode) {
			deploy.OpenCodeAgentsOnly(s.cfg, paths)
		}
	case "skills":
		if s.cfg.TargetEnabled(config.TargetClaude) {
			ts := deploy.AddTarget(config.TargetClaude + " (skills)")
			deploy.DeploySkillsSymlink(s.cfg, paths.Claude, ts)
		}
		if s.cfg.TargetEnabled(config.TargetGemini) {
			ts := deploy.AddTarget(config.TargetGemini + " (skills)")
			deploy.DeploySkillsSymlink(s.cfg, paths.Gemini, ts)
		}
		if s.cfg.TargetEnabled(config.TargetCodex) {
			ts := deploy.AddTarget(config.TargetCodex + " (skills)")
			deploy.DeploySkillsSymlink(s.cfg, paths.Codex, ts)
		}
	default: // "all"
		if s.cfg.TargetEnabled(config.TargetClaude) {
			deploy.Claude(s.cfg, paths)
		}
		if s.cfg.TargetEnabled(config.TargetOpenCode) {
			deploy.OpenCode(s.cfg, paths)
		}
		if s.cfg.TargetEnabled(config.TargetGemini) {
			deploy.Gemini(s.cfg, paths)
		}
		if s.cfg.TargetEnabled(config.TargetCodex) {
			deploy.Codex(s.cfg, paths)
		}
	}

	// Collect names from repo for response.
	agentFiles := deploy.AgentFiles(s.cfg.RepoDir)
	agentNames := make([]string, 0, len(agentFiles))
	for _, f := range agentFiles {
		agentNames = append(agentNames, filepath.Base(f))
	}

	result := map[string]any{
		"target":          target,
		"deployed_agents": agentNames,
		"deployed_skills": collectSkillNames(s.cfg.RepoDir),
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

// collectSkillNames returns skill directory names from the repo skills/ dir.
func collectSkillNames(repoDir string) []string {
	skillsDir := filepath.Join(repoDir, config.CompSkills)
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return []string{}
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names
}

// runDoctor runs all 9 health checks and returns structured results.
// Input:  {} (no parameters)
// Output: { checks: [{ name, status, detail }], issues: N }
func (s *Server) runDoctor(_ context.Context, _ map[string]any) (string, error) {
	stateDir := filepath.Join(s.cfg.RepoDir, "."+s.cfg.Name)
	st, err := state.Load(stateDir)
	if err != nil {
		return "", fmt.Errorf("load state: %w", err)
	}

	git := gitutil.New(s.cfg.RepoDir)
	paths := deploy.ResolvePaths()

	var checks []healthCheck
	issues := 0

	addCheck := func(ok bool, name, detail string) {
		status := "ok"
		if !ok {
			status = "warn"
			issues++
		}
		checks = append(checks, healthCheck{Name: name, Status: status, Detail: detail})
	}

	// Check 1: Binary vs deployed SHA.
	currentSHA := git.CurrentSHA()
	switch {
	case st.DeployedSHA == "" || st.DeployedSHA == state.StateNone:
		addCheck(false, "deployed_sha", "nothing deployed yet — run 'anvil deploy'")
	case st.DeployedSHA != currentSHA:
		addCheck(false, "deployed_sha", fmt.Sprintf("binary (%s) differs from deployed (%s) — run 'anvil deploy'", currentSHA, st.DeployedSHA))
	default:
		addCheck(true, "deployed_sha", fmt.Sprintf("binary matches deployed SHA (%s)", currentSHA))
	}

	// Check 2: Uncommitted changes.
	if git.IsDirty() {
		addCheck(false, "working_tree", "uncommitted changes — deployed version may differ from working tree")
	} else {
		addCheck(true, "working_tree", "working tree clean")
	}

	// Check 3: Config files exist.
	manifestPath := filepath.Join(s.cfg.RepoDir, s.cfg.Name+".yaml")
	configPath := filepath.Join(s.cfg.RepoDir, s.cfg.Name+".config.yaml")
	manifestOK := fileutil.Exists(manifestPath)
	configOK := fileutil.Exists(configPath)
	if manifestOK && configOK {
		addCheck(true, "config_files", fmt.Sprintf("%s.yaml and %s.config.yaml present", s.cfg.Name, s.cfg.Name))
	} else {
		var missing []string
		if !manifestOK {
			missing = append(missing, s.cfg.Name+".yaml")
		}
		if !configOK {
			missing = append(missing, s.cfg.Name+".config.yaml")
		}
		addCheck(false, "config_files", "missing: "+strings.Join(missing, ", "))
	}

	// Check 4: Provider tier mappings.
	provider := s.cfg.ActiveProvider()
	if _, tierErr := s.cfg.ResolveTier(config.TierHigh, provider); tierErr != nil {
		addCheck(false, "provider_tiers", fmt.Sprintf("provider %q missing tier mappings", provider))
	} else {
		addCheck(true, "provider_tiers", fmt.Sprintf("provider %q has valid tier mappings", provider))
	}

	// Checks 5–8: Target health.
	type targetSpec struct {
		name     string
		basePath string
		items    []doctorTargetCheck
	}
	targets := []targetSpec{
		{
			name:     config.TargetClaude,
			basePath: paths.Claude,
			items: []doctorTargetCheck{
				{name: config.CompAgents, checkType: "dir_or_symlink"},
				{name: config.CompSkills, checkType: "dir_or_symlink"},
				{name: config.CompCommands, checkType: "dir_or_symlink"},
				{name: config.FileClaudeMD, checkType: "file_or_symlink"},
			},
		},
		{
			name:     config.TargetOpenCode,
			basePath: paths.OpenCode,
			items: []doctorTargetCheck{
				{name: config.CompAgents, checkType: "dir_or_symlink"},
				{name: config.CompCommands, checkType: "dir_or_symlink"},
			},
		},
		{
			name:     config.TargetGemini,
			basePath: paths.Gemini,
			items: []doctorTargetCheck{
				{name: config.CompSkills, checkType: "dir_or_symlink"},
				{name: config.CompCommands, checkType: "dir_or_symlink"},
				{name: config.FileGeminiMD, checkType: "file_or_symlink"},
			},
		},
		{
			name:     config.TargetCodex,
			basePath: paths.Codex,
			items: []doctorTargetCheck{
				{name: config.CompSkills, checkType: "dir_or_symlink"},
				{name: config.FileAgentsMD, checkType: "file_or_symlink"},
			},
		},
	}

	for _, ts := range targets {
		if !s.cfg.TargetEnabled(ts.name) {
			checks = append(checks, healthCheck{
				Name:   "target_" + ts.name,
				Status: "ok",
				Detail: "disabled (skipped)",
			})
			continue
		}
		issues += appendTargetChecks(ts.name, ts.basePath, ts.items, &checks)
	}

	// Check 9: Broken symlinks.
	brokenCount := brokenSymlinksMCP(paths, s.cfg)
	if brokenCount > 0 {
		issues += brokenCount
		checks = append(checks, healthCheck{
			Name:   "symlinks",
			Status: "warn",
			Detail: fmt.Sprintf("%d broken symlink(s) detected", brokenCount),
		})
	} else {
		checks = append(checks, healthCheck{Name: "symlinks", Status: "ok", Detail: "no broken symlinks"})
	}

	result := map[string]any{
		"checks": checks,
		"issues": issues,
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	return string(out), nil
}

// doctorTargetCheck specifies a single component to verify under a target directory.
type doctorTargetCheck struct {
	name      string
	checkType string // "dir_or_symlink" | "file_or_symlink"
}

// appendTargetChecks verifies each component in a target directory and appends
// results to checks. Returns number of issues found.
func appendTargetChecks(target, basePath string, items []doctorTargetCheck, checks *[]healthCheck) int {
	issues := 0
	if !fileutil.IsDir(basePath) {
		*checks = append(*checks, healthCheck{
			Name:   "target_" + target,
			Status: "warn",
			Detail: fmt.Sprintf("base directory missing: %s", basePath),
		})
		return 1
	}

	for _, c := range items {
		p := filepath.Join(basePath, c.name)
		ok := false
		switch c.checkType {
		case "dir_or_symlink":
			ok = fileutil.IsSymlink(p) || fileutil.IsDir(p)
		case "file_or_symlink":
			ok = fileutil.IsSymlink(p) || fileutil.Exists(p)
		}

		status := "ok"
		detail := fmt.Sprintf("%s/%s present", target, c.name)
		if !ok {
			status = "warn"
			detail = fmt.Sprintf("%s/%s missing", target, c.name)
			issues++
		}
		*checks = append(*checks, healthCheck{
			Name:   "target_" + target + "_" + c.name,
			Status: status,
			Detail: detail,
		})
	}
	return issues
}

// brokenSymlinksMCP counts broken symlinks across all enabled target directories.
func brokenSymlinksMCP(paths deploy.TargetPaths, cfg *config.App) int {
	broken := 0
	dirs := map[string]string{
		config.TargetClaude:   paths.Claude,
		config.TargetOpenCode: paths.OpenCode,
		config.TargetGemini:   paths.Gemini,
		config.TargetCodex:    paths.Codex,
	}
	for target, base := range dirs {
		if !cfg.TargetEnabled(target) {
			continue
		}
		entries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		for _, e := range entries {
			p := filepath.Join(base, e.Name())
			if fileutil.IsSymlink(p) {
				dest, readErr := os.Readlink(p)
				if readErr != nil || !fileutil.Exists(dest) {
					broken++
				}
			}
		}
	}
	return broken
}
