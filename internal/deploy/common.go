package deploy

import (
	"os"
	"path/filepath"

	"github.com/ernesto2108/anvil/pkg/config"
	"github.com/ernesto2108/anvil/pkg/frontmatter"
	"github.com/ernesto2108/anvil/pkg/output"
)

// AgentData holds the parsed data for a single agent file, passed to the
// provider-specific formatting callback.
type AgentData struct {
	Name     string               // base filename (e.g. "developer.md")
	RawData  []byte               // raw file bytes
	Content  string               // string(RawData)
	Doc      frontmatter.Document // parsed frontmatter + body
	Tier     string               // frontmatter "model" value
	Perm     string               // resolved permission value (from "permissionMode" or "permission")
	PermKey  string               // actual frontmatter key used ("permissionMode" or "permission")
}

// PermField returns the permission field key and value from a parsed frontmatter document.
// Agents use "permissionMode" (preferred) or "permission" for the permission level.
func PermField(doc frontmatter.Document) (key, value string) {
	if v := doc.Fields["permissionMode"]; v != "" {
		return "permissionMode", v
	}
	if v := doc.Fields["permission"]; v != "" {
		return "permission", v
	}
	return "", ""
}

// AdaptFunc is the provider-specific callback that formats an agent file
// for a particular target. It receives the agent data and the config, and
// returns the final content string to write. If it returns an empty string,
// the file is skipped.
type AdaptFunc func(cfg *config.App, agent AgentData) string

// deployAgents is the shared pipeline for deploying agent files to any target.
// It handles filtering, collision resolution, reading, and parsing. The adapt
// callback handles target-specific formatting. The resulting content is written
// to agentDst with the resolved deploy name.
func deployAgents(cfg *config.App, agentDst string, ts *TargetStats, adapt AdaptFunc) {
	files := FilteredAgentFiles(cfg)
	if len(files) == 0 {
		return
	}

	if err := os.MkdirAll(agentDst, 0o755); err != nil {
		output.Error("create agents dir: %s", err)
		return
	}

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
			output.Error("read agent %s: %s", name, err)
			continue
		}

		content := string(data)
		doc := frontmatter.Parse(content)

		permKey, permVal := PermField(doc)
		agent := AgentData{
			Name:    name,
			RawData: data,
			Content: content,
			Doc:     doc,
			Tier:    doc.Fields["model"],
			Perm:    permVal,
			PermKey: permKey,
		}

		adapted := adapt(cfg, agent)
		if adapted == "" {
			continue
		}

		deployName := GetDeployName(name, renames)
		dstPath := filepath.Join(agentDst, deployName)
		// Remove any existing symlink to avoid writing through it to the source file
		if IsAnvilSymlink(dstPath) {
			os.Remove(dstPath)
		}
		if err := os.WriteFile(dstPath, []byte(adapted), 0o644); err != nil {
			output.Error("write agent %s: %s", deployName, err)
			continue
		}
		ts.Agents.Deployed++
		ts.Agents.Names = append(ts.Agents.Names, deployName)
	}
}
