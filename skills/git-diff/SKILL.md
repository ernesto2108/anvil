---
name: git-diff
description: Inspect and summarize repository changes using git diff. Use when user says "what changed", "show diff", "review changes", "summarize modifications", or before creating a commit or pull request.
---

Inspect only the changes made in the repository using git diff.

Why:
- review minimal changes
- avoid re-reading full files
- enable safe code review
- detect unintended edits
- reduce tokens and noise

Typical commands:
- `git diff` — current working changes
- `git diff --staged` — staged only
- `git diff path/to/file.go` — file specific
- `git diff main...feature-branch` — branch compare

Usage rules:
- ALWAYS prefer git diff before reading full files
- review only changed lines
- ignore whitespace-only changes
- summarize large diffs
- highlight: logic changes, new conditions, removed checks, schema changes, concurrency risks

Output format per file:

File: <path>
Change type: modified | added | deleted
Summary: <one line>

Diff:
<relevant lines only>

Best practices:
- small diffs (< 200 lines)
- one concern per change
- avoid unrelated refactors
- split large changes

Never:
- read entire repo when diff exists
- approve changes without diff
- hide changes
- auto-merge blindly
