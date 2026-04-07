package deploy

import (
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/fileutil"
	"github.com/ernesto2108/anvil/pkg/frontmatter"
	"github.com/ernesto2108/anvil/pkg/output"
)

func Claude(cfg *config.App, paths TargetPaths) {
	target := paths.Claude
	output.Info("%s -> %s", output.Bold("Claude Code"), target)

	deployClaudeAgents(cfg, target)
	DeploySkillsSymlink(cfg.RepoDir, target)
	DeployCommandsSymlink(cfg.RepoDir, target)
	deployClaudeMD(cfg, target)
}

func deployClaudeAgents(cfg *config.App, target string) {
	files := AgentFiles(cfg.RepoDir)
	if len(files) == 0 {
		return
	}

	agentDst := filepath.Join(target, config.CompAgents)
	fileutil.CleanPath(agentDst)
	os.MkdirAll(agentDst, 0o755)

	count := 0
	for _, f := range files {
		name := filepath.Base(f)
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}

		content := string(data)
		tier := frontmatter.Get(content, "model")
		perm := frontmatter.Get(content, "permission")

		resolved := tier
		if config.IsTier(tier) {
			model, err := cfg.ResolveTier(tier, config.TargetClaude)
			if err == nil {
				resolved = model
				content = frontmatter.ReplaceField(content, "model", tier, resolved)
			}
		}

		if config.IsPerm(perm) {
			tools := cfg.ResolvePermission(perm, config.TargetClaude)
			if tools != "" {
				content = frontmatter.ReplaceField(content, "permission", perm, tools)
				// Rename the key from permission to tools
				content = replaceKey(content, "permission", "tools")
			}
		}

		if len(content) > 0 && content[len(content)-1] != '\n' {
			content += "\n"
		}
		dstPath := filepath.Join(agentDst, name)
		os.WriteFile(dstPath, []byte(content), 0o644)

		output.Info("  agents/%s  %s->%s  %s->tools", name, tier, output.Green(resolved), perm)
		count++
	}
	output.Info("  %d agents (provider: %s)", count, output.Green(config.TargetClaude))
}

func replaceKey(content, oldKey, newKey string) string {
	// Replace first occurrence of "oldKey:" with "newKey:" in frontmatter
	old := oldKey + ":"
	new := newKey + ":"
	replaced := false
	lines := splitLines(content)
	for i, line := range lines {
		if !replaced && len(line) > len(old) && line[:len(old)] == old {
			lines[i] = new + line[len(old):]
			replaced = true
		}
	}
	return joinLines(lines)
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	result := lines[0]
	for _, l := range lines[1:] {
		result += "\n" + l
	}
	return result
}

func deployClaudeMD(cfg *config.App, target string) {
	src := filepath.Join(cfg.RepoDir, config.FileClaudeMD)
	if !fileutil.Exists(src) {
		return
	}
	if err := fileutil.ForceSymlink(src, filepath.Join(target, config.FileClaudeMD)); err != nil {
		output.Error("symlink %s: %s", config.FileClaudeMD, err)
		return
	}
	output.Info("  %s -> symlink", config.FileClaudeMD)
}

func ClaudeAgentsOnly(cfg *config.App, paths TargetPaths) {
	target := paths.Claude
	output.Info("%s -> %s", output.Bold("Claude Code"), target)
	deployClaudeAgents(cfg, target)
}
