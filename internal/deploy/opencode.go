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
	output.Info("%s -> %s", output.Bold("OpenCode"), target)

	deployOpenCodeAgents(cfg, target)
	output.Info("  skills -> using Claude Code path (default)")
	deployOpenCodeCommands(cfg, target)
}

func deployOpenCodeAgents(cfg *config.App, target string) {
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

		adapted := fmt.Sprintf("---\ndescription: %s\nmode: subagent\nmodel: %s\npermission: %s\n---\n\n%s", desc, resolved, permResolved, doc.Body)

		dstPath := filepath.Join(agentDst, name)
		os.WriteFile(dstPath, []byte(adapted), 0o644)
		count++
	}
	output.Info("  %d agents adapted", count)
}

func deployOpenCodeCommands(cfg *config.App, target string) {
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
	output.Info("  commands -> copied")
}

func OpenCodeAgentsOnly(cfg *config.App, paths TargetPaths) {
	target := paths.OpenCode
	files := AgentFiles(cfg.RepoDir)
	if len(files) == 0 {
		return
	}

	agentDst := filepath.Join(target, config.CompAgents)
	fileutil.CleanPath(agentDst)
	os.MkdirAll(agentDst, 0o755)

	for _, f := range files {
		name := filepath.Base(f)
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

		adapted := fmt.Sprintf("---\ndescription: %s\nmode: subagent\nmodel: %s\npermission: %s\n---\n\n%s", desc, resolved, permResolved, doc.Body)
		os.WriteFile(filepath.Join(agentDst, name), []byte(adapted), 0o644)
	}
	output.Info("%s -> agents updated", output.Bold("OpenCode"))
}
