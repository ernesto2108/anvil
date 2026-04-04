---
name: skill-standards
description: Standards and checklist for creating new skills. Use when creating a new skill, reviewing skill quality, or user says "create a skill", "new skill", "skill template", "skill checklist". Ensures all skills follow the Agent Skills open standard and project conventions.
disable-model-invocation: true
---

# Skill Creation Standards

These are the mandatory standards for every skill in this project. Based on the Agent Skills open standard (agentskills.io), Anthropic best practices, and lessons learned from our own iteration.

## Pre-Creation Checklist

Before writing a new skill, verify:

- [ ] No existing skill already covers this use case (check `~/.claude/skills/`)
- [ ] The skill has a clear single responsibility
- [ ] You know whether it should be auto-invocable, user-only, or System-only

## SKILL.md Structure

Every skill MUST have this frontmatter:

```yaml
---
name: skill-name              # REQUIRED. Lowercase, hyphens, matches directory name
description: What it does. Use when user says "keyword1", "keyword2", "keyword3", or [context]. # REQUIRED. Max 1024 chars
user-invocable: false          # Only if guardrail/behavioral (system auto-loads)
disable-model-invocation: true # Only if heavy/manual (user invokes with /name)
---
```

### Name Rules
- Lowercase letters, numbers, hyphens only
- Max 64 characters
- Must match parent directory name exactly
- No leading/trailing/consecutive hyphens
- Gerund form preferred: `processing-pdfs`, not `pdf-processor`

### Description Rules (Critical — drives activation)
- Format: `What it does. Use when [trigger conditions].`
- Include 3-5 specific keywords/phrases users would naturally say
- Include file extensions when relevant (`.go`, `.tsx`, `.dart`)
- Be "pushy" — list contexts explicitly, err on side of over-triggering
- Max 1024 characters
- Third person, imperative phrasing

### Invocation Control

| Mode | Frontmatter | When to use |
|---|---|---|
| **Auto** (default) | none | Skills the system should load during normal work (lint, conventions, guardrails) |
| **User-only** | `disable-model-invocation: true` | Heavy operations only on explicit request (orchestrate, scan, e2e) |
| **System-only** | `user-invocable: false` | Passive guardrails and behavioral guides (boundary-guardrails, entity-guardrails) |

## Body Content Standards

### Required Sections (conventions/reference skills)

```markdown
## Philosophy
- 3 core principles that guide decisions (not rules, principles)

## Workflow
1. Numbered procedural steps
2. Include gates: "If X — stop and [action]"
3. Include user confirmation points: "Ask user: [question]"

## Rules
- Concrete, actionable rules grouped by concern

## Pre-Implementation Checklist
- [ ] Verifiable items before starting work

## Anti-Pattern Detection (for convention skills)
- Table with: Code Pattern | Anti-Pattern | Severity | Category | Fix
- Passive detection: error + warning always, suggestion only on "improve/refactor"
- Report format: [file:line] [severity] [category] anti-pattern-name
```

### Required Sections (operational/task skills)

```markdown
## Workflow
1. Detect/setup step
2. Execute step
3. Decision gates with clear branching
4. Report step with defined output format

## Output Format
- Exact template of expected output
```

### Content Principles

1. **Philosophy before rules** — state the principle, then the rule
2. **Procedures over declarations** — teach how to approach, not what to produce
3. **Explain the why** — "because X causes Y" beats "MUST do Z"
4. **Gates over instructions** — "If >200 lines, stop and split" beats "keep files small"
5. **Examples over explanations** — 50-token code example beats 150-token prose
6. **Progressive disclosure** — SKILL.md < 500 lines, details in reference files
7. **No project-specific references** — no hardcoded paths, project names, or domain terms

### Anti-Pattern Severity Levels

| Level | Meaning | When to report |
|---|---|---|
| `error` | Causes bugs, crashes, data loss | Always (passive detection) |
| `warning` | Performance, maintainability, design issues | Always (passive detection) |
| `suggestion` | Code style, minor improvements | Only on "improve/refactor/optimize" (active detection) |

## Directory Structure

```
skill-name/
├── SKILL.md                    # Required, < 500 lines
├── reference.md                # Optional, detailed docs (loaded on demand)
├── anti-patterns.md            # Optional, detection reference table
├── examples/
│   ├── good-patterns.md        # Optional, idiomatic examples
│   └── bad-patterns.md         # Optional, anti-patterns with corrections
├── evals/
│   ├── evals.json              # Optional, test cases
│   └── files/                  # Optional, example files for evals
└── scripts/
    └── helper.sh               # Optional, executable utilities
```

### Reference File Rules
- Keep references one level deep (no A.md → B.md → C.md chains)
- For files >300 lines, include a table of contents
- Reference clearly from SKILL.md: "See `reference.md` for the full guide"

## Evals (Recommended for convention skills)

Create `evals/evals.json` with:
- 4 trigger tests (2 should-trigger, 2 should-not-trigger)
- 3-5 quality tests with assertions and example files

```json
{
  "skill_name": "skill-name",
  "evals": [
    {
      "id": 1,
      "type": "trigger",
      "prompt": "Realistic user prompt",
      "should_trigger": true
    },
    {
      "id": 2,
      "type": "quality",
      "prompt": "Task prompt",
      "files": ["evals/files/example.ext"],
      "assertions": [
        "Specific, verifiable assertion about output"
      ]
    }
  ]
}
```

## Cross-Agent Compatibility

Skills live in `~/.claude/skills/` with a symlink at `~/.agents/skills/` for cross-agent discovery (Cursor, Codex, Gemini CLI, etc.). No extra work needed — the symlink handles it.

## Quality Checklist (run after creating a skill)

- [ ] `name` matches directory name
- [ ] `description` includes "Use when" with 3-5 activation keywords
- [ ] Invocation mode set correctly (auto / user-only / System-only)
- [ ] SKILL.md < 500 lines
- [ ] Philosophy section states principles, not just rules
- [ ] Workflow section has numbered steps with gates
- [ ] No project-specific references (paths, domain terms, project names)
- [ ] Anti-pattern table has severity levels (for convention skills)
- [ ] Pre-implementation checklist exists (for convention skills)
- [ ] Reference files are one level deep
- [ ] Evals created (for convention skills)
