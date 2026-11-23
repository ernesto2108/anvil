package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/fileutil"
	"github.com/ernesto2108/anvil/pkg/output"
)

func OpenCode(cfg *config.App, paths TargetPaths) {
	target := paths.OpenCode
	ts := AddTarget("OpenCode")

	deployOpenCodeAgents(cfg, target, ts)
	ts.Extras = append(ts.Extras, "Skills -> Claude Code path")
	deployOpenCodeCommands(cfg, target, ts)
}

func deployOpenCodeAgents(cfg *config.App, target string, ts *TargetStats) {
	agentDst := filepath.Join(target, config.CompAgents)
	deployAgents(cfg, agentDst, ts, adaptOpenCode)
}

// adaptOpenCode formats an agent file for the OpenCode target.
func adaptOpenCode(cfg *config.App, agent AgentData) string {
	desc := agent.Doc.Fields["description"]
	perm := agent.Perm

	permResolved := config.PermWrite
	if config.IsPerm(perm) {
		p := cfg.ResolvePermission(perm, config.TargetOpenCode)
		if p != "" {
			permResolved = p
		}
	}

	return fmt.Sprintf("---\ndescription: %s\nmanaged-by: anvil\nmode: subagent\npermission: %s\n---\n\n%s", desc, permResolved, agent.Doc.Body)
}

func deployOpenCodeCommands(cfg *config.App, target string, ts *TargetStats) {
	cmdSrc := filepath.Join(cfg.RepoDir, config.CompCommands)
	if !fileutil.IsDir(cmdSrc) {
		return
	}

	cmdDst := filepath.Join(target, config.CompCommands)
	fileutil.CleanPath(cmdDst)

	if err := fileutil.CopyDir(cmdSrc, cmdDst); err != nil {
		output.Error("copy commands: %s", err)
		return
	}

	entries, _ := os.ReadDir(cmdDst)
	ts.Commands.Deployed = len(entries)
}

func OpenCodeAgentsOnly(cfg *config.App, paths TargetPaths) {
	target := paths.OpenCode
	ts := AddTarget("OpenCode (agents only)")
	agentDst := filepath.Join(target, config.CompAgents)
	deployAgents(cfg, agentDst, ts, adaptOpenCode)
}
