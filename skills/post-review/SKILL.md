---
name: post-review
description: "Post-development review skill with stack-specific checklists for Go, React, React Native, Terraform, and PostgreSQL"
---

# Post-Review Skill — Dispatcher

## Purpose

Provide stack-specific review checklists for the Reviewer agent. This file routes to the correct checklist based on detected stack.

## Routing Table

| Stack detectado | Checklist a cargar |
|---|---|
| Go | `skills/post-review/checklists/go.md` |
| React (JS/TS) | `skills/post-review/checklists/react.md` |
| React Native | `skills/post-review/checklists/react-native.md` |
| Terraform | `skills/post-review/checklists/terraform.md` |
| PostgreSQL | `skills/post-review/checklists/postgres.md` |

## Multi-Stack Reviews

When a diff contains files from multiple stacks, load ALL relevant checklists. Apply each checklist only to its corresponding files.

## Lint Detection Table

| Stack | Config files | Linter | Install |
|---|---|---|---|
| Go | `.golangci.yml`, `.golangci.yaml` | golangci-lint | `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest` |
| React/TS | `.eslintrc.*`, `eslint.config.*`, `eslint` in package.json | ESLint | `npm install -D eslint` / `pnpm add -D eslint` |
| React Native | Same as React | ESLint | Same as React |
| Terraform | `.tflint.hcl` | TFLint | `brew install tflint` / `curl -s https://raw.githubusercontent.com/terraform-linters/tflint/master/install_linux.sh \| bash` |
| PostgreSQL | N/A | N/A | — |

## Supporting Files

| File | Purpose |
|---|---|
| `rubric.md` | Universal scoring criteria (1-10 scale) |
| `report-format.md` | Console output format specification |

## Usage

1. Reviewer agent detects stacks from file extensions
2. Verify lint configuration exists per stack (see Lint Detection Table)
3. Run linter if config exists; flag absence as CRITICO if not
4. Load matching checklists from the routing table
5. Load `rubric.md` and `report-format.md`
6. Execute review per checklist
7. Aggregate score (including lint findings) and print report
