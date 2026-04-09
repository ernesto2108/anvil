package cli

import "fmt"

func cmdHelp(appName string) {
	t := title(appName)
	fmt.Printf(`
  %s - Multi-tool GitOps for AI coding configuration

  Manages agents, skills, and commands across:
    Claude Code, OpenCode, Gemini CLI, Codex, and Cursor

  USAGE:
    %s <command> [args]

  COMMANDS:
    dashboard            Open Anvil Dashboard (observability UI)
    init                 First-time setup — show config and launch browser
    browse               Interactive TUI to manage agents/skills/commands
    update               Pull latest + rebuild binary
    targets [tool...]    Show or set which tools are active
    provider [name]      Show or switch AI provider
    status               Show version, branch, targets, tags
    doctor               Diagnose deployment health
    pin <comp> <tag>     Pin a component to a specific version
    unpin <comp>         Unpin a component back to HEAD
    diff [component...]  Show changes since last deploy
    uninstall            Remove %s from all targets
    help                 Show this help

  EXAMPLES:
    %s init                                # First-time setup
    %s browse                              # Manage agents/skills/commands
    %s update                              # Pull + rebuild
    %s targets claude opencode             # Set active tools
    %s provider gemini                     # Switch AI provider
    %s status                              # What's deployed where?
    %s doctor                              # Check health

  TARGETS:
    claude    ~/.claude/          agents + skills + commands
    opencode  ~/.config/opencode/ agents (adapted) + commands
    gemini    ~/.gemini/          skills + commands (toml)
    codex     ~/.codex/           skills + AGENTS.md (generated)
    cursor    per-project         rules (from agents)

`, t, appName, appName, appName, appName, appName, appName, appName, appName, appName)
}
