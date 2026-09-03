package config

// Target tool identifiers used in anvil.yaml.
const (
	TargetClaude   = "claude"
	TargetOpenCode = "opencode"
	TargetGemini   = "gemini"
	TargetCodex    = "codex"
	TargetCursor   = "cursor"
)

// Component directory names inside the repo.
const (
	CompAgents   = "agents"
	CompSkills   = "skills"
	CompCommands = "commands"
)

// Generated/deployed markdown files.
const (
	FileClaudeMD = "CLAUDE.md"
	FileGeminiMD = "GEMINI.md"
	FileAgentsMD = "AGENTS.md"
)

// Permission levels in agent frontmatter.
const (
	PermRead    = "read"
	PermWrite   = "write"
	PermExecute = "execute"
)

// IsPerm returns true if the string is a valid permission level.
func IsPerm(s string) bool {
	return s == PermRead || s == PermWrite || s == PermExecute
}
