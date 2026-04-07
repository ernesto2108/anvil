package cli

import "fmt"

func cmdHelp(appName string) {
	t := title(appName)
	fmt.Printf(`
  %s - Multi-tool GitOps for AI coding configuration

  Deploys agents, skills, and commands to:
    Claude Code, OpenCode, Gemini CLI, Codex, and Cursor

  USAGE:
    %s <command> [args]

  COMMANDS:
    deploy [version]     Deploy to all enabled targets (default: HEAD)
    self-update [ver]    Pull + rebuild + deploy in one step
    targets [tool...]    Show or set which tools are active
    provider [name]      Show or switch AI provider (redeploys agents)
    status               Show deployment state across all tools
    doctor               Diagnose deployment health
    rollback             Rollback to previous version
    pin <comp> <tag>     Pin a component to specific version (Claude)
    unpin <comp>         Unpin a component (Claude)
    uninstall            Remove %s from all targets
    tags                 List available versions
    diff [component...]  Show changes since last deploy
    help                 Show this help

  EXAMPLES:
    %s targets                             # Show active tools
    %s targets claude                      # Only use Claude Code
    %s targets claude opencode             # Claude + OpenCode
    %s targets all                         # Enable all tools
    %s deploy                              # Deploy to active tools
    %s provider gemini                     # Switch to Gemini models
    %s provider local                      # Switch to local/Ollama
    %s status                              # What's deployed where?
    %s self-update                         # Pull + build + deploy
    %s doctor                              # Check deployment health
    %s diff skills                         # Changes in skills only
    %s diff agents/developer               # Changes in one agent

  TARGETS:
    claude    ~/.claude/          agents + skills + commands
    opencode  ~/.config/opencode/ agents + commands
    gemini    ~/.gemini/          skills + commands (toml)
    codex     ~/.codex/           skills + AGENTS.md (auto-generated)
    cursor    per-project         rules from agents

`, t, appName, appName, appName, appName, appName, appName, appName, appName, appName, appName, appName, appName, appName, appName)
}
