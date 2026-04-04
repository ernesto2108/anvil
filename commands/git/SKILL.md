---
description: Git commit message writing and review following Conventional Commits spec
---

# Git Commit Skills

Slash commands for writing and reviewing Git commit messages following industry best practices.

## Commands

### `/git:commit`
Analyzes staged changes (`git diff --cached`) and writes a conventional commit message. Asks for confirmation before committing.

**Usage:**
```
/git:commit
```

**What it does:**
1. Reads staged diff and branch name
2. Determines the appropriate commit type and scope
3. Writes a compliant commit message
4. Asks for confirmation before committing

### `/git:commit-review`
Reviews the last N commits (default 5) and scores each against the Conventional Commits spec.

**Usage:**
```
/git:commit-review        # reviews last 5 commits
/git:commit-review 10     # reviews last 10 commits
```

**What it does:**
1. Reads recent commit messages
2. Scores each on 12 criteria (structure, content, best practices)
3. Reports issues and suggests rewrites for failing commits
4. Provides a summary with common issues

**Score scale:**
- 90-100: Excellent
- 70-89: Acceptable
- 50-69: Needs improvement
- 0-49: Poor quality

### `/git:message`
Generates a commit message from a natural language description without committing.

**Usage:**
```
/git:message added login with Google OAuth
/git:message fixed crash on empty input in parser
/git:message renamed endpoints to follow REST conventions, breaks clients
```

**What it does:**
1. Parses natural language into type, scope, and description
2. Generates a properly formatted conventional commit message
3. Outputs the message — does NOT commit

## Standards enforced

Based on:
- [Conventional Commits v1.0.0](https://www.conventionalcommits.org/en/v1.0.0/)
- [How to Write a Git Commit Message](https://cbea.ms/git-commit/) (Chris Beams)
- [Angular Commit Convention](https://github.com/angular/angular/blob/main/CONTRIBUTING.md#commit)
- [Semantic Release](https://semantic-release.gitbook.io/semantic-release/)

### Commit types

| Type | Purpose | SemVer impact |
|------|---------|---------------|
| `feat` | New feature | MINOR |
| `fix` | Bug fix | PATCH |
| `docs` | Documentation | — |
| `style` | Formatting | — |
| `refactor` | Restructuring | — |
| `test` | Tests | — |
| `chore` | Maintenance | — |
| `perf` | Performance | — |
| `ci` | CI/CD | — |
| `build` | Build system | — |

### Rules

1. Subject line max 50 characters
2. Imperative mood ("add" not "added")
3. No period at end of subject
4. Blank line between subject and body
5. Body wraps at 72 characters
6. Body explains WHAT and WHY, not HOW
7. Breaking changes use `!` suffix AND `BREAKING CHANGE:` footer
8. Issue references in footer (`Closes #123`)
9. No vague messages ("fix bug", "update", "WIP")

## Scope

These commands focus exclusively on commit **message quality**. Code review is handled by CodeRabbit.
