---
name: orchestrate/qa-fix
description: QA fix loop re-invocation protocol, prompt template, security-fix mode, and when NOT to use this mode. Load when orchestrator is about to re-invoke developer after QA or security blocking findings.
---

# QA Fix Loop

**Load when:** QA or security agent returns blocking findings on a completed task and the orchestrator is about to re-invoke developer to apply fixes.

---

## Rule: re-invoke the developer in `qa-fix` mode

When QA returns BLOCKING findings, the orchestrator must re-invoke the developer to apply fixes. A FRESH developer invocation costs ~20-30k tokens (re-loading convention skills, PRD, design, context.md, production files, handoff). That is pure waste — the previous developer invocation already had all that context and wrote the handoff to prove it.

The orchestrator passes `Mode: qa-fix` in the developer prompt. In this mode the developer:
1. Reads `.handoff/<TASK-ID>.md` (the handoff from the first invocation) as PRIMARY context
2. Does **NOT** re-read PRD, design, context.md
3. Does **NOT** re-load the full convention skill — the orchestrator injects ONLY the specific rules that apply to the files being fixed (3-5 rules inline, not the whole dispatcher)
4. Reads **ONLY** the files listed in the QA findings, not the whole package or the whole codebase
5. Applies SURGICAL fixes — no refactors, no "while I'm here" cleanups
6. Re-runs validation only for the files touched (`go vet -tags <tag> ./internal/<pkg>`, `npm run build` only if frontend changed)
7. Updates `## Notas` in the handoff with a one-line entry per fix applied
8. Does NOT rewrite `## Handoff for tester` unless a fix changed a public interface signature

---

## qa-fix prompt template

```
Mode: qa-fix. TASK-ID: <TASK-ID>.

The developer already completed the initial implementation for this task. The handoff at `.handoff/<TASK-ID>.md` contains the full context: files touched, patterns applied, decisions made, validation run. THAT is your primary context.

QA review returned the following BLOCKING findings (must fix before this task can merge):

<inline QA findings — exact issues with file paths and line numbers when available, one finding per bullet>

Execution rules:
1. Read `.handoff/<TASK-ID>.md` first. Do NOT re-read PRD, design, or context.md.
2. Read ONLY the files mentioned in the QA findings above — not the whole package, not the whole codebase.
3. Apply MINIMAL fixes — address ONLY the findings. No extra refactors, no cleanups, no "while I'm here".
4. Re-run validation commands scoped to the files you touched:
   - Go: `go vet -tags <tag> ./internal/<pkg>` + relevant unit test package
   - Frontend: `npm run build` only if you touched .ts/.tsx
5. Update `## Notas` in the handoff — one line per fix applied.
6. Do NOT modify `## Handoff for tester` unless a fix changed a public interface signature.

Convention rules that apply to your fix (injected inline — do NOT load the full skill):
<inline ONLY the 3-5 specific rules from the convention skill that apply to the fix — e.g., "error wrapping: wrap SQL errors with fmt.Errorf('package: action: %w', err)" — NOT the whole dispatcher>

Forbidden:
- Loading the full convention skill
- Reading PRD / design / context.md
- Touching files outside the findings
- Refactoring code that works
```

---

## When NOT to use qa-fix mode

- **Findings are non-blocking** (rubric score still ≥7) → add them to the backlog as follow-up tasks, do NOT re-invoke the developer at all
- **Findings require architectural changes** (new patterns, new abstractions, moving files) → re-invoke the developer in NORMAL mode with a new plan, not qa-fix
- **Findings span > 5 files** → fixes are no longer surgical; use normal mode with a focused plan

---

## Security-fix mode

When the security agent returns blocking findings on a completed task, use `Mode: security-fix` with the same template (swap "QA" for "security" in the prompt). The context-savings logic is identical.

---

## Expected savings

On tasks like DASH-FEAT-006, where a fresh developer re-invocation for a11y fixes cost **22k tokens**, qa-fix mode should bring that down to **~5-8k**. Savings per QA cycle: **~15-17k tokens**. Over 5 tasks with QA cycles, that is a full invocation's worth of budget reclaimed.
