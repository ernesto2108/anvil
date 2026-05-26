[Leer en espanol](manual.es.md)

# User Manual — Anvil

Practical guide for using Anvil day to day. From installation to orchestrating agents in a real project.

---

## 1. Installation

```bash
# Clone the repo
git clone https://github.com/ernesto2108/anvil.git ~/projects/anvil
cd ~/projects/anvil

# Choose which AI tools you use
anvil targets claude              # Claude Code only
anvil targets claude opencode     # Claude + OpenCode
anvil targets all                 # All tools

# Choose model provider
anvil provider claude             # Anthropic (Claude)
anvil provider gemini             # Google (Gemini)
anvil provider local              # Ollama/local

# Deploy
anvil deploy

# Verify
anvil status
```

After `deploy`, your AI tools have access to all agents and skills.

---

## 2. How to invoke skills (most common)

Skills are invoked with `/skill-name` in your AI chat. They are the primary mechanism for activating specialized knowledge.

### Daily examples

```
# Before writing Go code
/go-conventions

# Before writing React code
/react-conventions

# Creating a Dockerfile
/devops-conventions

# Running linters
/lint

# Running tests
/run-tests

# Creating a PRD for a new feature
/prd-template

# Reviewing a design in Pencil or Figma
/design-review

# Checking database schema
/db-schema-scan

# Generating architecture diagrams
/generate-diagram

# Auditing accessibility
/a11y-check

# Checking dependencies
/dependency-check
```

### How to know what skills exist

Ask your AI: **"what skills do I have available?"** — the system lists them automatically because they're registered.

---

## 3. How to use agents

Agents are specialized roles the AI can assume. They are invoked in two ways:

### Method 1: Guided (via `/orchestrate`)

Tell your AI what you want to do and use `/orchestrate`:

```
I want to add a notifications endpoint to the backend.
/orchestrate
```

The `/orchestrate` skill classifies task complexity and proposes agents in order. You confirm before anything runs:

| Complexity | What happens |
|------------|-------------|
| **Trivial** (1-2 files) | AI does it directly, no agents |
| **Medium** (3-8 files) | PM (if PRD missing) → Developer → Tester → QA |
| **Large** (8+ files) | PM → Architect → Developer → Tester → QA → Reporter |

### Method 2: Manual (invoke a specific agent)

You are the orchestrator — there is no automatic leader agent. Pick the agent that matches your task and invoke it directly:

```
"Use the developer-backend agent to implement the login"
"I need the dba agent to create the migration"
"Launch the security agent to audit the code"
```

### Which agent for what?

| You need to... | Agent | Example |
|---|---|---|
| Write requirements | **pm** | "Write the PRD for the invitations feature" |
| Design architecture | **architect** | "Design the notifications bounded context" |
| Design UI/UX | **designer** | "Design the settings screen in Pencil" |
| Write Go/backend code | **developer-backend** | "Implement the GET /users/:id endpoint" |
| Write React/Astro/TS code | **developer-frontend** | "Implement the Workflows screen" |
| Write Flutter/Dart code | **developer-mobile** | "Implement the notifications screen" |
| Write tests | **tester** | "Write tests for the auth service" |
| Create migrations | **dba** | "Create the migration to add the invitations table" |
| CI/CD and infra | **devops** | "Create the Dockerfile and CI workflow" |
| Review quality | **qa** | "Review the code from the last PR" |
| Audit security | **security** | "Audit the JWT token handling" |
| Initialize or refresh project context | **context-init** | "Initialize the project context" |
| Write docs | **tech-writer** | "Update the README with the new endpoints" |
| Generate report + update context | **reporter** | "Generate the session report" |

---

## 4. Typical workflows

### Trivial task (quick fix)

```
> "Fix the typo in routes.go line 42"
```

The AI does it directly. No agents or skills needed.

### Medium task (new frontend screen)

```
> "I need to implement the Workflows screen from the Pencil design"
> /orchestrate
```

The AI:
1. Classifies as medium (~5 pts)
2. Loads `/react-conventions`
3. Launches **developer** in implementation mode
4. Launches **tester** for tests
5. Runs `/lint` and `/run-tests`

### Large task (new bounded context)

```
> "I want to add team management: invite users, assign roles, list members"
> /orchestrate
```

You (as orchestrator):
1. Classify as large (~13 pts)
2. Invoke **pm** → generates PRD
3. Invoke **architect** → designs contracts and bounded context
4. Invoke **designer** → designs screens
5. Invoke **developer-backend** and/or **developer-frontend** / **developer-mobile** → implements the relevant stacks
6. Invoke **tester** → writes tests
7. Invoke **qa** → reviews quality (blocks if score < 7)
8. Invoke **reporter** → generates session report and updates `.project-context/`

### Infrastructure task

```
> "I need to dockerize the backend and create CI with GitHub Actions"
> /devops-conventions
```

The AI loads DevOps conventions and has access to:
- Docker best practices (multi-stage, non-root, layer caching)
- GitHub Actions templates (CI with lint+test+build, CD with Cloud Run)
- Terraform patterns
- AWS, GCP, Kubernetes, Argo CD guides

---

## 5. Per-stack conventions

When working on a specific stack, load its convention skill. This gives the AI rules, patterns, and anti-patterns for that language.

### Go

```
/go-conventions
```

Includes: error wrapping, entity validation, parameterized SQL, context everywhere, defer after error check, concurrency (worker pools, errgroup), testing (table-driven, mocks).

### React/TypeScript

```
/react-conventions
```

Includes: custom hooks, state (TanStack Query, Zustand), Tailwind v4 syntax, accessibility, testing (Vitest + RTL), anti-patterns, functional components only.

### Flutter/Dart

```
/flutter-conventions
```

Includes: BLoC/Riverpod, widget composition, freezed, theming, testing.

### DevOps/Infra

```
/devops-conventions
```

Includes: Docker, GitHub Actions, Terraform, Kubernetes, AWS, GCP, Argo CD/Workflows/Rollouts, infra security.

---

## 6. Documentation vault

Each project can have an Obsidian vault for structured documentation. Initialize it:

```bash
cp -r ~/projects/anvil/vault-template/ ~/projects/my-project-knowledge-base/
```

### What goes where

| Folder | Content | Who writes |
|---|---|---|
| `01-project/` | `context.md` — technical snapshot | scanner |
| `02-backlog/` | `sprint-current.md` — sprint board | pm |
| `03-tasks/<ID>/` | PRD, design, QA review per task | pm, architect, qa |
| `04-architecture/` | ADRs, bounded contexts, diagrams | architect |
| `05-bugs/` | Critical bug postmortems | security, qa |
| `06-reports/` | `last-run.md` — last session report | reporter |
| `07-references/` | Templates, external links | manual |
| `08-design/` | Design files (.pen, .fig) | designer |

### Configure the project in Anvil

Add an entry to `~/.claude/project-registry.md`:

```markdown
| my-project | personal | ~/projects/my-project-knowledge-base/ |
```

Now all agents know where to read and write documentation for that project.

---

## 7. CLI management

### Update agents and skills

```bash
cd ~/projects/anvil
git pull
anvil deploy
```

### Check what's deployed

```bash
anvil status
```

### Change AI provider

```bash
anvil provider gemini    # Switch models to Gemini
anvil deploy             # Redeploy with new models
```

### Pin a skill to a version

```bash
anvil pin skills/go-conventions v1.2.0
anvil unpin skills/go-conventions    # Back to HEAD
```

### Uninstall

```bash
anvil uninstall    # Clean anvil and restore original files
```

---

## 8. Backup & Restore

Anvil automatically protects files you had before installation. No manual action needed.

### How it works

#### On first `anvil deploy`

Anvil scans all targets (Claude, OpenCode, Gemini, Codex) looking for existing files. If it finds anything, it saves an exact copy before touching anything:

```
.anvil/pre-install/
├── claude/
│   ├── agents/        # Your original ~/.claude/agents/
│   ├── skills/        # Your original ~/.claude/skills/
│   ├── commands/      # Your original commands
│   └── CLAUDE.md      # Your original CLAUDE.md
├── opencode/
│   ├── agents/
│   └── commands/
├── gemini/
│   ├── skills/
│   ├── commands/
│   └── GEMINI.md
└── codex/
    ├── skills/
    └── AGENTS.md
```

You'll see in the terminal:

```
[anvil] Saving pre-install snapshot of existing files...
  saved ~/.claude/agents
  saved ~/.claude/skills
  saved ~/.claude/CLAUDE.md
```

#### On each subsequent deploy

If there are directories that aren't symlinks (someone edited directly in `~/.claude/agents/`), Anvil moves them to a timestamped backup before overwriting:

```
[anvil] Backing up ~/.claude/agents → ~/.claude/agents.backup.20260403143022
```

#### On `anvil uninstall`

Anvil removes what it deployed and **restores original files** from the snapshot:

```
[anvil] Claude Code:
  restored ~/.claude/agents
  restored ~/.claude/skills
  restored ~/.claude/CLAUDE.md
[anvil] OpenCode:
  removed agents (no snapshot)

[anvil] Anvil uninstalled. Pre-existing files restored where snapshots existed.
```

### Common scenarios

| Situation | What happens |
|---|---|
| First time, nothing existed | Clean deploy, empty snapshot |
| First time, you had your own agents | Snapshot saves them, deploy overwrites, uninstall restores them |
| Already using Anvil, manually edited an agent | Next deploy makes timestamped backup of the edited version |
| Want to go back to pre-Anvil state | `anvil uninstall` restores everything |
| Lost the snapshot | Timestamped backups are in `~/.claude/*.backup.*` |

### Where are the backups

| Type | Location | Created when |
|---|---|---|
| Original snapshot | `.anvil/pre-install/` | First deploy |
| Timestamped backup | `~/.claude/*.backup.YYYYMMDDHHMMSS` | Each subsequent deploy (if manual changes detected) |

### Important

- The snapshot is created **once** — on the first deploy. If you deploy, uninstall, and deploy again, the second deploy creates a new snapshot.
- `anvil uninstall` deletes the entire `.anvil/` directory, including the snapshot. Copy it first if you want to keep it.
- Timestamped backups in `~/.claude/` are not auto-deleted. Clean them manually when no longer needed.

---

## 9. Tips

- **Don't load skills you don't need** — each skill consumes tokens. Only load what's relevant to your task.
- **Use `/orchestrate` when you don't know where to start** — the system classifies and picks agents for you.
- **Conventions stack** — if you load `/go-conventions` and then `/devops-conventions`, both apply.
- **Agents have strict boundaries** — developer doesn't touch tests, tester doesn't touch production, dba doesn't touch business logic. This is by design.
- **context-init saves tokens** — run `context-init` (mode: `init` or `regular`) at the start of a session so other agents have context without reading every file.
- **Documenting is part of done** — after tests pass, invoke `reporter` to update `.project-context/`. This is as mandatory as the tests themselves.
- **AGENTS.md is auto-generated** — don't edit it manually, it gets overwritten on every anvil deploy.
