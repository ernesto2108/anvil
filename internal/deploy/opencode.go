package deploy

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/fileutil"
	"github.com/ernesto2108/anvil/pkg/frontmatter"
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
	files := FilteredAgentFiles(cfg)
	if len(files) == 0 {
		return
	}

	agentDst := filepath.Join(target, config.CompAgents)
	os.MkdirAll(agentDst, 0o755)

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

		doc := frontmatter.Parse(string(data))
		desc := doc.Fields["description"]
		tier := doc.Fields["model"]
		perm := doc.Fields["permission"]

		resolved := tier
		if config.IsTier(tier) {
			model, err := cfg.ResolveTier(tier, cfg.ActiveProvider())
			if err == nil {
				resolved = model
			}
		}

		permResolved := config.PermWrite
		if config.IsPerm(perm) {
			p := cfg.ResolvePermission(perm, config.TargetOpenCode)
			if p != "" {
				permResolved = p
			}
		}

		deployName := GetDeployName(name, renames)
		adapted := fmt.Sprintf("---\ndescription: %s\nmanaged-by: anvil\nmode: subagent\nmodel: %s\npermission: %s\n---\n\n%s", desc, resolved, permResolved, doc.Body)

		os.WriteFile(filepath.Join(agentDst, deployName), []byte(adapted), 0o644)
		ts.Agents.Deployed++
	}
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

	files := FilteredAgentFiles(cfg)
	if len(files) == 0 {
		return
	}

	agentDst := filepath.Join(target, config.CompAgents)
	os.MkdirAll(agentDst, 0o755)

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

		doc := frontmatter.Parse(string(data))
		desc := doc.Fields["description"]
		tier := doc.Fields["model"]
		perm := doc.Fields["permission"]

		resolved := tier
		if config.IsTier(tier) {
			model, err := cfg.ResolveTier(tier, cfg.ActiveProvider())
			if err == nil {
				resolved = model
			}
		}

		permResolved := config.PermWrite
		if config.IsPerm(perm) {
			p := cfg.ResolvePermission(perm, config.TargetOpenCode)
			if p != "" {
				permResolved = p
			}
		}

		deployName := GetDeployName(name, renames)
		adapted := fmt.Sprintf("---\ndescription: %s\nmanaged-by: anvil\nmode: subagent\nmodel: %s\npermission: %s\n---\n\n%s", desc, resolved, permResolved, doc.Body)
		os.WriteFile(filepath.Join(agentDst, deployName), []byte(adapted), 0o644)
		ts.Agents.Deployed++
	}
}
