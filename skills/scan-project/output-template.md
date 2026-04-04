# Output Template for project.md

Use this template when writing ``<vault>/01-project/context.md``. Include only sections for detected stacks. Delete unused stack sections.

---

```markdown
# <Project Name> — Project Context

## Product Objective (Canonical)
<short summary — what the product is and what problem it solves>

## Non-negotiable rules
- <rule from user>

## What the AI should optimize for
- <point from user>

## What NOT to suggest
- <point from user>

## Repository Snapshot (Technical Context)

### Stacks detected
- <Go 1.23 | React 18 + Vite | Flutter 3.x | etc.>

### Directory Tree (actual)
```text
<REAL TREE HERE — depth 3, excluding <vault>/, .git/, vendor/, node_modules/, dist/, build/>
```

<!-- === GO SECTION (include only if go.mod found) === -->
### Go
- **Version:** <from go.mod>
- **Module:** <module path>
- **Key dependencies:** <top-level only, e.g., chi, sqlx, pgx, slog>
- **Test files:** <count and example paths>
- **Linter config:** <.golangci.yml detected / not found>

<!-- === REACT/NODE SECTION (include only if package.json found) === -->
### React / Node
- **Framework:** <Next.js / Vite / CRA / none>
- **Key dependencies:** <react, typescript, tailwind, etc.>
- **Test runner:** <vitest / jest / other>
- **Test files:** <count and example paths>
- **Lint config:** <eslint config detected / not found>
- **TypeScript:** <tsconfig.json detected / not found>

<!-- === FLUTTER SECTION (include only if pubspec.yaml found) === -->
### Flutter
- **Dart version:** <from environment in pubspec.yaml>
- **Key dependencies:** <riverpod, bloc, dio, etc.>
- **Test files:** <count and example paths>
- **Analysis config:** <analysis_options.yaml detected / not found>
- **Localization:** <l10n.yaml detected / not found>

<!-- === OTHER STACKS (add as needed) === -->

### CI / Runtime detected
- <Dockerfile: yes/no>
- <docker-compose: yes/no>
- <GitHub Actions: list workflow files>

### Build tools
- <Makefile / taskfile / scripts detected>

### Config files
- <any other notable config>

## Notes for agents
- follow existing structure exactly
- do not introduce new architectural patterns
- place new code near similar existing files
```
