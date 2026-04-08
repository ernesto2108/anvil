package deploy

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/fileutil"
	"github.com/ernesto2108/anvil/pkg/frontmatter"
	"github.com/ernesto2108/anvil/pkg/output"
)

// stdinReader is the shared reader for interactive prompts during deploy.
var stdinReader = bufio.NewReader(os.Stdin)

// SetStdinReader replaces the stdin reader (used in tests).
func SetStdinReader(r *bufio.Reader) {
	stdinReader = r
}

func Claude(cfg *config.App, paths TargetPaths) {
	target := paths.Claude
	ts := AddTarget("Claude Code")

	deployClaudeAgents(cfg, target, ts)
	DeploySkillsSymlink(cfg, target, ts)
	DeployCommandsSymlink(cfg, target, ts)
	deployClaudeMD(cfg, target, ts)
}

func deployClaudeAgents(cfg *config.App, target string, ts *TargetStats) {
	files := FilteredAgentFiles(cfg)
	if len(files) == 0 {
		return
	}

	agentDst := filepath.Join(target, config.CompAgents)
	os.MkdirAll(agentDst, 0o755)

	// Resolve collisions interactively
	result := ResolveCollisions(files, agentDst, stdinReader)
	CleanManagedFiles(agentDst)
	skip, _, renames := ApplyResolutions(result.Resolutions)

	for _, f := range files {
		name := filepath.Base(f)
		if !ShouldDeployFile(name, skip) {
			ts.Agents.Preserved++
			continue
		}

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
				content = replaceKey(content, "permission", "tools")
			}
		}

		content = StampManagedBy(content)

		if len(content) > 0 && content[len(content)-1] != '\n' {
			content += "\n"
		}

		deployName := GetDeployName(name, renames)
		dstPath := filepath.Join(agentDst, deployName)
		os.WriteFile(dstPath, []byte(content), 0o644)
		ts.Agents.Deployed++
		ts.Agents.Names = append(ts.Agents.Names, deployName)
	}
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

func deployClaudeMD(cfg *config.App, target string, ts *TargetStats) {
	src := filepath.Join(cfg.RepoDir, config.FileClaudeMD)
	if !fileutil.Exists(src) {
		return
	}
	if err := fileutil.ForceSymlink(src, filepath.Join(target, config.FileClaudeMD)); err != nil {
		output.Error("symlink %s: %s", config.FileClaudeMD, err)
		return
	}
	ts.Extras = append(ts.Extras, config.FileClaudeMD+" -> symlink")
}

func ClaudeAgentsOnly(cfg *config.App, paths TargetPaths) {
	target := paths.Claude
	ts := AddTarget("Claude Code (agents only)")
	deployClaudeAgents(cfg, target, ts)
}
