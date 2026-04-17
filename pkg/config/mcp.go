package config

import "os"

// MCPServerDef describes an MCP server declared in the mcp.servers section of anvil.yaml.
type MCPServerDef struct {
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"`
	Agents  []string          `yaml:"agents,omitempty"` // empty = available to all agents
}

// MCPSection is the top-level mcp: section in anvil.yaml.
type MCPSection struct {
	Servers map[string]MCPServerDef `yaml:"servers,omitempty"`
}

// BuildAgentMCPConfig returns the MCP servers the given agent role is allowed
// to use. A server with no Agents list is available to every agent.
// Env values are expanded via os.ExpandEnv so ${TOKEN} references are resolved
// at run time from the caller's environment.
func (a *App) BuildAgentMCPConfig(role string) map[string]MCPServerDef {
	if len(a.Manifest.MCP.Servers) == 0 {
		return nil
	}
	result := make(map[string]MCPServerDef)
	for name, def := range a.Manifest.MCP.Servers {
		if !isMCPAgentAllowed(role, def.Agents) {
			continue
		}
		expanded := MCPServerDef{
			Command: def.Command,
			Args:    def.Args,
		}
		if len(def.Env) > 0 {
			expanded.Env = make(map[string]string, len(def.Env))
			for k, v := range def.Env {
				expanded.Env[k] = os.ExpandEnv(v)
			}
		}
		result[name] = expanded
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// isMCPAgentAllowed returns true when the role is in the allowed list, or
// when the list is empty (no restriction).
func isMCPAgentAllowed(role string, agents []string) bool {
	if len(agents) == 0 {
		return true
	}
	for _, a := range agents {
		if a == role {
			return true
		}
	}
	return false
}
