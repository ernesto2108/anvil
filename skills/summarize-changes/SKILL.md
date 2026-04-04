---
name: summarize-changes
disable-model-invocation: true
description: Create a human-readable summary of what changed and why, writing last-run.md to the vault. Use when user says "summarize what we did", "write a report", "what did we change", or at the end of a work session to document progress.
---

Create a human-readable summary of what changed and why.

## Prerequisite

Invoke `/git-diff` first to gather the raw diff and change summary. Use its output as the base for the report.

## Inputs

- Output from `/git-diff` (changed files, diff stats)
- `<vault>/02-backlog/sprint-current.md`
- `<vault>/03-tasks/<TASK-ID>/prd.md`
- `<vault>/01-project/context.md`

## Actions

- Use git-diff output to identify changed files
- Group by feature
- Infer intent from tasks
- Explain reasons

Output: `<vault>/06-reports/last-run.md` (overwrite if exists)
