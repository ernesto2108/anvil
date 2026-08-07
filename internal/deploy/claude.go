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
	agentDst := filepath.Join(target, config.CompAgents)
	deployAgents(cfg, agentDst, ts, adaptClaude)
}

// adaptClaude formats an agent file for the Claude Code target.
func adaptClaude(cfg *config.App, agent AgentData) string {
	content := agent.Content
	tier := agent.Tier
	perm := agent.Perm

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
			permKey := agent.PermKey
			content = frontmatter.ReplaceField(content, permKey, perm, tools)
			content = frontmatter.RenameField(content, permKey, "tools")
		}
	}

	content = StampManagedBy(content)

	if len(content) > 0 && content[len(content)-1] != '\n' {
		content += "\n"
	}

	return content
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
