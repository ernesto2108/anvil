[Leer en espanol](README.es.md)

# Anvil

AI agent orchestration system. Define agents, skills, and conventions once — deploy to Claude Code, OpenCode, Gemini CLI, Codex, and Cursor. Compatible with the [AGENTS.md](https://agents.md/) open standard.

## What is Anvil?

Anvil is a collection of **agents** (specialized AI roles), **skills** (domain knowledge and conventions), and a **CLI** that deploys them to your AI coding tools. Write your agents and skills in markdown, run `anvil deploy`, and every tool gets the same knowledge.

Anvil automatically generates an `AGENTS.md` file — the [open standard](https://agents.md/) maintained by the Linux Foundation and adopted by Codex, Cursor, Copilot, and others.

## User Manual

For a complete guide on daily usage — invoking skills, using agents, typical workflows, per-stack conventions, and tips — read the **[User Manual](docs/manual.md)**.

## Quick Start

```bash
# 1. Clone
git clone https://github.com/ernesto2108/anvil.git ~/projects/anvil
cd ~/projects/anvil

# 2. Build the CLI (requires Go 1.25+)
make build

# 3. Make it available globally (optional)
ln -sf ~/projects/anvil/anvil /usr/local/bin/anvil

# 4. Choose your targets (which tools to deploy to)
anvil targets claude opencode    # or: all

# 5. Choose your provider (model tier mapping)
anvil provider claude            # or: gemini, local

# 6. Deploy
anvil deploy

# 7. Verify
anvil status
```

> After step 3, you can run `anvil` from anywhere. If you skip it, use `./anvil` from the anvil directory.

## How It Works

```mermaid
flowchart TD
    User([User]) -->|describes task| Orch[Orchestrator]
    Orch -->|triage| Decision{Complexity?}

    Decision -->|Trivial 1-2 files| Direct[Direct execution]
    Decision -->|Medium 3-8 files| Pipeline1[Developer → Tester → QA]
    Decision -->|Large 8+ files| Pipeline2[Full pipeline]

    Pipeline2 --> PM[PM agent]
    PM -->|PRD| Arch[Architect agent]
    Arch -->|design.md| Des[Designer agent]
    Des -->|ui-spec.md| Dev[Developer agent]
    Dev -->|code| Test[Tester agent]
    Test -->|tests| QA[QA agent]
    QA -->|score >= 7?| Gate{Quality Gate}
    Gate -->|Pass| Rep[Reporter agent]
    Gate -->|Fail| Dev

    style Orch fill:#6366f1,color:#fff
    style Gate fill:#f59e0b,color:#000
    style Direct fill:#10b981,color:#fff
```

Each agent has strict boundaries:
- **developer** writes production code only
- **tester** writes test files only
- **dba** manages migrations only
- **devops** manages infra/CI only
- Agents never cross boundaries

## Project Structure

```
anvil/
├── cmd/anvil/             # CLI entry point (Go)
├── internal/              # Core packages (config, deploy, state, etc.)
├── anvil.yaml             # Deployment manifest (targets, components)
├── anvil.config.yaml      # Provider & model mapping
├── agents/                # 12 specialized agent definitions
├── skills/                # 38 domain skills and conventions
├── commands/              # User-invocable slash commands
├── docs/                  # Documentation (en + es)
├── examples/              # CLAUDE.md template for projects
└── vault-template/        # Obsidian vault template for documentation
```

## Agents

Each agent is a markdown file with YAML frontmatter defining its role, permissions, and model tier.

| Agent | Role | Permission | Tier |
|-------|------|------------|------|
| **pm** | Requirements, PRDs, backlog, sprint planning | write | high |
| **architect** | System design, API contracts, ADRs | write | high |
| **designer** | UX/UI design, design system, user flows | write | high |
| **developer** | Production code (Go, React, Flutter, Astro) | execute | medium |
| **tester** | Test files across all stacks | execute | medium |
| **dba** | Migrations, schema design, query optimization | execute | medium |
| **devops** | CI/CD, Docker, Terraform, K8s, cloud infra | execute | medium |
| **qa** | Code review, quality gate (blocks if score < 7) | execute | medium |
| **security** | SAST, SCA, secrets audit, auth review | execute | medium |
| **scanner** | Repository scanning, project context generation | execute | medium |
| **tech-writer** | Documentation, README, API docs, changelogs | write | medium |
| **reporter** | Session execution reports | execute | low |

### How agents work

- The orchestrator (you or `/orchestrate`) triages task complexity
- **Trivial** tasks: execute directly, no agents
- **Medium+**: agents run in sequence with gates between phases
- Each agent has strict boundaries — developer can't touch tests, tester can't touch production code

### Permissions

| Level | Available tools |
|-------|----------------|
| **read** | Glob, Grep, LS, Read |
| **write** | + Write, Edit |
| **execute** | + Bash |

### Model tiers

| Tier | Use case | Claude example | Gemini example |
|------|----------|----------------|----------------|
| **high** | Complex decisions (PM, Architect) | Opus | gemini-2.5-pro |
| **medium** | Implementation (Developer, Tester) | Sonnet | gemini-2.5-flash |
| **low** | Simple tasks (Reporter) | Haiku | gemini-2.5-flash-lite |

## Skills

Skills are loadable knowledge modules. Agents load them on-demand based on the task.

### Per-Stack Conventions

| Skill | Covers |
|-------|--------|
| `/go-conventions` | Error handling, validation, SQL, concurrency, testing, Kafka, RabbitMQ |
| `/react-conventions` | Hooks, state management, Tailwind v4, accessibility, testing, anti-patterns |
| `/flutter-conventions` | BLoC/Riverpod, widget composition, theming, testing |
| `/astro-conventions` | Islands, content collections, components, styling |
| `/devops-conventions` | Docker, GitHub Actions, Terraform, K8s, AWS, GCP, Argo CD/Workflows/Rollouts |

### Workflow Skills

| Skill | Purpose |
|-------|---------|
| `/orchestrate` | Triage complexity, select agents, manage gates |
| `/lint` | Auto-detect stack, run linters and formatters |
| `/run-tests` | Auto-detect stack, run tests with coverage |
| `/design-system` | Create design tokens, variables, components (Pencil/Figma) |
| `/design-review` | Quality audit of designs with scoring |
| `/design-to-code` | Translate designs to production code |
| `/prd-template` | PRD writing with discovery questionnaire |
| `/backlog-management` | Break PRDs into tickets, manage sprints |

### Guard Skills

| Skill | Purpose |
|-------|---------|
| `/architecture-boundary-guardrails` | Enforce bounded contexts, prevent cross-domain leaks |
| `/domain-entity-guardrails` | Strict typing, no pointers for optional fields |
| `/code-review-rubric` | Scoring criteria for QA reviews |

### Utility Skills

| Skill | Purpose |
|-------|---------|
| `/dependency-check` | Audit packages for vulnerabilities and licenses |
| `/bundle-analyzer` | Frontend bundle size impact analysis |
| `/db-schema-scan` | Read-only schema inspection via migrations |
| `/db-optimize` | Slow query identification, index suggestions |
| `/generate-diagram` | Mermaid.js diagrams (C4, ERD, sequence, flow) |
| `/git-diff` | Summarize repository changes |
| `/service-map` | Cross-service dependency awareness |
| `/a11y-check` | WCAG 2.1 accessibility audit |
| `/test-api` | API endpoint contract validation |
| `/ui-component-scan` | Scan component library for reuse |

## CLI Reference

```bash
# Deployment
anvil deploy                     # Deploy all components to active targets
anvil status                     # Show what's deployed where

# Targets (which AI tools to deploy to)
anvil targets                    # Show active targets
anvil targets claude opencode    # Set exact targets
anvil targets --add gemini       # Enable one target
anvil targets --rm cursor        # Disable one target
anvil targets all                # Enable all

# Provider (model mapping)
anvil provider                   # Show current provider
anvil provider gemini            # Switch to Gemini models
anvil provider local             # Switch to local/Ollama models

# Version pinning
anvil pin skills/go-conventions v1.2.0    # Pin to git tag
anvil unpin skills/go-conventions         # Follow HEAD again

# Maintenance
anvil diff                       # Show changes since last deploy
anvil uninstall                  # Remove from all targets
```

## Configuration

### `anvil.yaml` — Deployment manifest

```yaml
targets:
  claude:
    enabled: true
    path: ~/.claude
  opencode:
    enabled: true
    path: ~/.config/opencode
  gemini:
    enabled: true
    path: ~/.gemini
  codex:
    enabled: true
    path: ~/.codex
  cursor:
    enabled: false

components:
  agents:
    tag: "HEAD"
  skills:
    tag: "HEAD"
  commands:
    tag: "HEAD"
```

### `anvil.config.yaml` — Provider & model mapping

```yaml
provider: claude

providers:
  claude:
    high: opus
    medium: sonnet
    low: haiku
  gemini:
    high: gemini-2.5-pro
    medium: gemini-2.5-flash
    low: gemini-2.5-flash-lite
  local:
    high: qwen3:32b
    medium: qwen3:14b
    low: qwen3:8b
```

## Backup & Restore

Anvil automatically protects your existing files:

- **First deploy**: snapshots everything found in `~/.claude/`, `~/.codex/`, etc.
- **Each deploy**: timestamps backup if it detects manual changes
- **Uninstall**: restores original files from the snapshot

See [full section in the manual](docs/manual.md#8-backup--restore).

## AGENTS.md Compatibility

[AGENTS.md](https://agents.md/) is an open standard maintained by the Linux Foundation for configuring AI agents in software projects. Anvil generates this file automatically on every `anvil deploy`.

### Which tools read it?

| Tool | Reads AGENTS.md | Native file |
|---|---|---|
| **OpenAI Codex** | Yes (primary) | `~/.codex/AGENTS.md` |
| **Cursor** | Yes (repo root) | `.cursor/rules/*.mdc` |
| **GitHub Copilot** | Via `.github/copilot-instructions.md` | `.github/agents/*.agent.md` |
| **OpenCode** | Yes (primary) | — |
| **Claude Code** | No (uses `CLAUDE.md`) | `~/.claude/agents/*.md` |
| **Gemini CLI** | Discussion active | `GEMINI.md` |

### How it works in Anvil

1. Define agents in `agents/*.md` with frontmatter (role, permissions, tier)
2. On `anvil deploy`, Anvil:
   - Deploys native agents to each target (Claude, OpenCode, Gemini, etc.)
   - Generates compact `AGENTS.md` to `~/.codex/` for Codex
3. Any AGENTS.md-compatible tool can read the generated file

## Creating New Agents

Create `agents/{name}.md`:

```markdown
---
name: my-agent
description: One-line description the system uses to decide when to invoke this agent
permission: execute    # read | write | execute
model: medium          # high | medium | low
---

# Agent Spec — Role Title

## Role
What this agent does and what it does NOT do.

## Input
What the orchestrator provides.

## Rules
Specific constraints and permissions.

## Output
What it produces and where.
```

## Creating New Skills

Create `skills/{name}/SKILL.md`:

```markdown
---
name: my-skill
description: One-line description of what this skill teaches
---

# Skill Name

## When to Load
Conditions that trigger loading this skill.

## Content
The actual knowledge, conventions, and patterns.
```

For complex skills, use subdirectories with a routing table:

```
skills/my-conventions/
├── SKILL.md           # Dispatcher with routing table
├── rules/             # Quick reference files
├── guides/            # Detailed patterns
└── examples/          # Good + bad patterns
```

## Documentation Vault

Use `vault-template/` to bootstrap an Obsidian vault for any project:

```bash
cp -r vault-template/ ~/projects/my-project-knowledge-base/
```

Structure:
```
01-project/context.md         # Scanner output
02-backlog/sprint-current.md  # Sprint board
03-tasks/<ID>/                # PRD, design, QA per task
04-architecture/              # ADRs, bounded contexts
05-bugs/                      # Postmortems
06-reports/last-run.md        # Session reports
07-references/                # Templates, external links
08-design/                    # Design files (.pen, .fig)
```

## License

MIT
