---
name: task-complete
description: Mark a task as done — updates task file status, moves card on Kanban board, cleans duplicates. Use when user says "task done", "complete task", "mark as done", or "/task-complete TASK-XXX".
disable-model-invocation: true
---

# Task Complete

Automates the 5 steps needed to close a task in the Obsidian vault.

## Usage

```
/task-complete TASK-006
/task-complete TASK-006 "Documentación completada + 2 bugs encontrados"
```

## Workflow

When invoked with a TASK-ID (and optional summary):

### Step 1 — Find the task file

Search for `<TASK-ID>.md` in `<docs>/03-tasks/` (check sprint folders first, then backlog).

### Step 2 — Update task status

In the task file frontmatter, set:
```yaml
status: done
```

### Step 3 — Update Kanban board

In `<docs>/02-backlog/board.md`:
1. **Remove** the task line from whatever column it's in (Backlog, To Do, In Progress, Blocked)
2. **Remove** any duplicate in Backlog if the task also exists in a sprint folder
3. **Add** to Done column:
   ```
   - [x] [[<sprint>/<TASK-ID>]] <summary or title> #<service> #<labels>
   ```

### Step 4 — Update sprint metrics (if applicable)

In `<docs>/02-backlog/sprint-current.md`, increment SP completed by the task's `story_points`.

### Step 5 — Confirm

Output a one-line confirmation:
```
✓ <TASK-ID> marcada como done (<story_points> SP)
```

## Rules

- Resolve `<docs>` from `~/.claude/project-registry.md`
- If task file not found → report error, do not create it
- If task is already `status: done` → skip, report "already done"
- Do NOT modify any file other than: the task .md, board.md, sprint-current.md
- Maximum 4 tool calls: 1 Read (task file) + 1 Edit (task status) + 1 Edit (board) + 1 Edit (sprint metrics)
