package mcp

import "context"

// toolDef is the MCP tool schema sent in tools/list responses.
type toolDef struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// tool pairs a definition with its handler.
type tool struct {
	def     toolDef
	handler func(ctx context.Context, args map[string]any) (string, error)
}

// buildRegistry returns all tools registered for this server.
func (s *Server) buildRegistry() map[string]tool {
	reg := map[string]tool{}

	add := func(name, desc string, schema map[string]any, h func(context.Context, map[string]any) (string, error)) {
		reg[name] = tool{
			def: toolDef{
				Name:        name,
				Description: desc,
				InputSchema: schema,
			},
			handler: h,
		}
	}

	// ── Inventory ─────────────────────────────────────────────────────────────

	add("list_agents", "List all available Anvil agents with their role, permission, and model tier.",
		schema(), s.listAgents)

	add("list_skills", "List all available Anvil skills with their name and description.",
		schema(), s.listSkills)

	add("list_pipelines", "List all available Anvil pipeline presets with their node definitions.",
		schema(), s.listPipelines)

	add("get_project_info", "Return project name, stack, active provider, and last run summary.",
		schema(), s.getProjectInfo)

	// ── Execution ─────────────────────────────────────────────────────────────

	add("run_pipeline",
		"Launch an Anvil pipeline preset in the background. Returns immediately. Use list_runs to check status.",
		schema(
			prop("pipeline", "string", "Pipeline preset name (e.g. bug, feat, quick, epic, db, infra, design)"),
			prop("task", "string", "Task description to pass to the pipeline"),
			optProp("stack", "string", "Stack hint (go, react, python, typescript, rust, flutter)"),
		),
		s.runPipeline)

	add("get_run_status",
		"Get the current status of a run including its agents, duration, and files touched. Use run_id='last' for the most recent run.",
		schema(prop("run_id", "string", "Run ID (e.g. r_20260417_143052_a4b2) or 'last'")),
		s.getRunStatus)

	add("list_runs",
		"List recent pipeline runs, optionally filtered by project or status.",
		schema(
			optProp("project", "string", "Filter by project name (defaults to current directory name)"),
			optProp("limit", "number", "Maximum number of runs to return (default 10, max 50)"),
			optProp("status", "string", "Filter by status: running, success, failed"),
		),
		s.listRuns)

	add("get_agent_output",
		"Get the full output of a specific agent in a run. Use agent_role instead of agent_id for convenience.",
		schema(
			prop("run_id", "string", "Run ID or 'last' for the most recent run"),
			prop("agent_role", "string", "Agent role (e.g. developer, architect, tester, qa)"),
		),
		s.getAgentOutput)

	// ── Context ───────────────────────────────────────────────────────────────

	add("search_memories",
		"Search past run digests by semantic similarity. Returns the most relevant past decisions and edge cases.",
		schema(
			prop("query", "string", "Natural language query describing what you are looking for"),
			optProp("project", "string", "Project name to search in (defaults to current directory name)"),
			optProp("limit", "number", "Number of results to return (default 3, max 10)"),
		),
		s.searchMemories)

	add("get_conventions",
		"Load coding conventions for a given stack and topic. Returns relevant rules and patterns.",
		schema(
			prop("stack", "string", "Stack name: go, react, typescript, python, rust, flutter, astro, devops"),
			optProp("topic", "string", "Topic hint to narrow results (e.g. testing, database, concurrency, errors, patterns)"),
		),
		s.getConventions)

	add("get_backlog",
		"Read the current sprint backlog from the project vault. Requires ANVIL_VAULT_PATH env var.",
		schema(
			optProp("status", "string", "Filter by column: backlog, todo, in_progress, blocked, done (default: all)"),
		),
		s.getBacklog)

	add("get_task",
		"Read a task's PRD, architecture, and design documents from the project vault. Requires ANVIL_VAULT_PATH.",
		schema(prop("task_id", "string", "Task ID (e.g. MCP-001, DASH-FEAT-017, ORCH-001)")),
		s.getTask)

	add("get_recent_changes",
		"Get a summary of recent git commits and pipeline runs to understand what changed and when.",
		schema(
			optProp("days", "number", "Number of days to look back (default 1, max 7)"),
			optProp("project", "string", "Project name for run filtering (defaults to current directory name)"),
		),
		s.getRecentChanges)

	// ── Orchestration ─────────────────────────────────────────────────────────

	add("start_orchestration",
		"Create a conversational orchestration run. Returns a run_id to use with save_step and load_orchestration.",
		schema(
			prop("objective", "string", "Task objective for this orchestration"),
			prop("stack", "string", "Stack (go, react, python, typescript, rust, flutter)"),
			prop("complexity", "string", "Complexity level (small, medium, complex)"),
			optProp("pipeline", "string", "Pipeline preset name (e.g. feat, bug). If provided, pipeline nodes are snapshotted for tracking."),
		),
		s.startOrchestration)

	add("save_step",
		"Persist an agent step within a conversational orchestration run. Each step is saved with full output (no truncation).",
		schema(
			prop("run_id", "string", "Run ID from start_orchestration"),
			prop("role", "string", "Agent role (e.g. architect, developer, tester, qa)"),
			prop("status", "string", "Step status (success, failed, skipped)"),
			prop("output", "string", "Full agent output text"),
			optProp("files_touched", "array", "List of file paths touched by this step"),
			optProp("duration_ms", "number", "Step duration in milliseconds"),
		),
		s.saveStep)

	add("load_orchestration",
		"Load a conversational orchestration run with all its steps and pending roles. Use run_id='last' for the most recent conversational run.",
		schema(
			prop("run_id", "string", "Run ID or 'last' for most recent conversational run"),
			optProp("project", "string", "Project filter when using 'last' (defaults to cwd)"),
		),
		s.loadOrchestration)

	return reg
}

// schema builds a JSON Schema object for tool input. props are key-value pairs
// of property name → property schema map.
func schema(props ...schemaEntry) map[string]any {
	if len(props) == 0 {
		return map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		}
	}
	properties := make(map[string]any, len(props))
	var required []string
	for _, p := range props {
		properties[p.name] = p.def
		if p.req {
			required = append(required, p.name)
		}
	}
	s := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

type schemaEntry struct {
	name string
	def  map[string]any
	req  bool
}

func prop(name, typ, desc string) schemaEntry {
	return schemaEntry{name: name, def: map[string]any{"type": typ, "description": desc}, req: true}
}

func optProp(name, typ, desc string) schemaEntry {
	return schemaEntry{name: name, def: map[string]any{"type": typ, "description": desc}, req: false}
}

// strArg extracts a string argument, returning defVal if absent.
func strArg(args map[string]any, key, defVal string) string {
	if v, ok := args[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return defVal
}

// intArg extracts a numeric argument (JSON numbers decode as float64), returning defVal if absent or zero.
func intArg(args map[string]any, key string, defVal int) int {
	if v, ok := args[key]; ok {
		switch n := v.(type) {
		case float64:
			if n > 0 {
				return int(n)
			}
		case int:
			if n > 0 {
				return n
			}
		}
	}
	return defVal
}
