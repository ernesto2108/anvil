---
description: Analyze staged changes and write a conventional commit message
---

You are a Git commit message expert. Analyze staged changes and write a high-quality conventional commit message.

## Step 1: Check staged changes

Run `git diff --cached --stat` to see what files are staged. If nothing is staged, tell the user "No staged changes found. Stage files with `git add` first." and stop.

Then run `git diff --cached` to see the full diff of staged changes.

## Step 2: Analyze the changes

Carefully read the diff and determine:
- **What** changed (files, functions, components)
- **Why** it changed (bug fix, new feature, refactor, etc.)
- **Impact** (breaking changes, API changes, behavior changes)
- **Scope** (which module/component/area is affected)
- **Related issues** (look for issue references in branch name, diff comments, or TODO markers)

Run `git branch --show-current` to check the branch name for issue references (e.g., `feat/123-add-login`, `fix/PROJ-456`).

## Step 3: Select the commit type

Choose the most appropriate type based on the PRIMARY purpose of the change:

| Type | When to use |
|------|-------------|
| `feat` | New feature or capability for the user (triggers MINOR version bump) |
| `fix` | Bug fix (triggers PATCH version bump) |
| `docs` | Documentation only (README, comments, JSDoc, etc.) |
| `style` | Formatting, whitespace, semicolons — no logic change |
| `refactor` | Code restructuring — no feature add, no bug fix |
| `test` | Adding or updating tests only |
| `chore` | Maintenance tasks (deps, configs, tooling) |
| `perf` | Performance improvement with no behavior change |
| `ci` | CI/CD pipeline changes (GitHub Actions, Jenkins, etc.) |
| `build` | Build system or external dependency changes |

## Step 4: Write the commit message

Follow these rules strictly:

### Subject line format
```
<type>(<scope>): <description>
```

**Rules:**
1. Type is lowercase, from the table above
2. Scope is optional — a noun describing the affected area (e.g., `auth`, `parser`, `api`, `ui`)
3. Description starts with a lowercase letter
4. Use imperative mood ("add" not "added", "fix" not "fixes")
5. Do NOT end with a period
6. Total subject line MUST be 50 characters or fewer — this is a hard limit
7. If 50 chars is too tight, shorten the description — never exceed it

### Body (when needed)
- Separate from subject with ONE blank line
- Wrap each line at 72 characters
- Explain WHAT changed and WHY, not HOW (the diff shows how)
- Use bullet points for multiple items
- Include body for any non-trivial change

### Footer (when needed)
- Separate from body with ONE blank line
- Issue references: `Closes #123`, `Fixes #456`, `Refs PROJ-789`
- Breaking changes: `BREAKING CHANGE: <description of what breaks>`
- If adding `!` after type/scope, ALSO include BREAKING CHANGE footer with details
- Co-authors: `Co-authored-by: Name <email>`

### Anti-patterns — NEVER write these:
- "fix bug" / "fix issue" — describe WHICH bug
- "update code" / "update file" — describe WHAT was updated
- "changes" / "misc" / "stuff" — always be specific
- "WIP" — commits should be atomic and complete
- Past tense ("added", "fixed", "removed") — use imperative
- Ending subject with a period

## Step 5: Ask for ticket/issue reference

Before presenting the final message, use the AskUserQuestion tool to ask:

**Question:** "Does this commit belong to a ticket or issue?"
**Header:** "Ticket"
**Options:**
1. "No ticket" — No issue reference needed
2. "Yes, let me type it" — I'll provide the ticket ID (e.g., TECHADMIN-123, PROJ-456, #78)

If the user provides a ticket reference:
- Add it to the commit message footer as `Refs <TICKET-ID>` (e.g., `Refs TECHADMIN-123`)
- If the commit is a fix, use `Fixes <TICKET-ID>` instead
- If there was already an issue reference detected from the branch name, show both and let the user confirm which to keep

## Step 6: Present to user

Show the complete commit message in a code block. Format it exactly as it will appear in git.

Then ask: **"Commit with this message? (yes/edit/cancel)"**

- If the user says **yes**: run `git commit -m "$(cat <<'EOF'\n<full message here>\nEOF\n)"` using a heredoc for proper formatting
- If the user says **edit** or provides changes: revise the message and ask again
- If the user says **cancel**: stop without committing

## Examples of good commit messages

```
feat(auth): add OAuth2 login with Google provider
```

```
fix(parser): handle empty input without crashing

Previously, passing an empty string to parse() would throw an
unhandled TypeError. Now returns an empty result object.

Closes #342
```

```
refactor(api)!: rename user endpoints to follow REST conventions

Rename /getUser to GET /users/:id and /createUser to POST /users
to align with REST standards.

BREAKING CHANGE: all /getUser and /createUser endpoints have been
removed. Clients must migrate to /users/:id (GET) and /users (POST).

Refs PROJ-891
```

```
perf(db): add index on orders.customer_id for faster lookups

Query time for customer order history reduced from ~800ms to ~15ms
on production dataset (12M rows).
```
