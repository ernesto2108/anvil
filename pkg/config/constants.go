package config

// Tier names used in agent frontmatter to map to provider-specific models.
const (
	TierHigh   = "high"
	TierMedium = "medium"
	TierLow    = "low"
)

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

// IsTier returns true if the string is a valid tier name.
func IsTier(s string) bool {
	return s == TierHigh || s == TierMedium || s == TierLow
}

// IsPerm returns true if the string is a valid permission level.
func IsPerm(s string) bool {
	return s == PermRead || s == PermWrite || s == PermExecute
}
