package cli

import "fmt"

func cmdHelp(appName string) {
	t := title(appName)
	fmt.Printf(`
  %s - Multi-tool GitOps for AI coding configuration

  Manages agents, skills, and commands across:
    Claude Code, OpenCode, Gemini CLI, Codex, and Cursor

  USAGE:
    %s <command> --task "description" [flags]

  PIPELINES (preset):
    quick   --task "desc"        Trivial change        dev → tester
    bug     --task "desc"        Bug fix + review       dev → tester → qa
    feat    --task "desc"        Standard feature       arch → dev → tester → qa
    design  --task "desc"        Feature with UI/UX     designer → arch → dev → tester → qa
    epic    --task "desc"        Full pipeline          pm → designer → arch → dba → dev → tester → qa → security
    db      --task "desc"        Database change        dba → dev → tester
    infra   --task "desc"        Infrastructure         devops → security

  PIPELINES (custom):
    run <pipeline.yaml> --task "desc"   Run a custom pipeline via DAG orchestrator

  FLAGS (all pipelines):
    --task, -t "desc"            Task description (required)
    --model, -m <name>           Override AI model
    --concurrency, -c <N>        Max parallel agents (default: 4)
    --auto-approve, -y           Skip human gates

  OTHER COMMANDS:
    digests              List run digests for current project
    digests show <id>    Show full body of one digest (the memory injected)
    export               Export digests to Markdown or JSON
    emit                 Ingest a Claude Code hook event (reads JSON from stdin)
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
    %s quick --task "Fix typo in README"
    %s bug --task "Fix null pointer in /api/users" -y
    %s feat --task "Add GET /users/:id endpoint"
    %s design --task "New onboarding flow with wizard"
    %s epic --task "Implement OAuth2 authentication system"
    %s db --task "Add indexes to orders table"
    %s infra --task "Add GitHub Actions CI pipeline"
    %s run pipeline.yaml --task "Custom workflow"

  TARGETS:
    claude    ~/.claude/          agents + skills + commands
    opencode  ~/.config/opencode/ agents (adapted) + commands
    gemini    ~/.gemini/          skills + commands (toml)
    codex     ~/.codex/           skills + AGENTS.md (generated)
    cursor    per-project         rules (from agents)

`, t, appName, appName, appName, appName, appName, appName, appName, appName, appName, appName)
}
