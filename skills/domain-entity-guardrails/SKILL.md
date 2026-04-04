---
name: domain-entity-guardrails
description: Enforce strict typing and explicit optionality in domain entities and value objects. Use when creating or modifying domain structs, reviewing domain layer code, or detecting pointer fields, `any` types, or `sql.Null*` in domain models.
user-invocable: false
---

Prevent fragile domain models by enforcing strict typing and explicit optionality.

Use when:
- creating or modifying domain entities or value objects
- mapping DB/HTTP DTOs into domain models
- reviewing domain layer code in a clean architecture project

## Detection

Before applying rules, locate the project's domain layer:
1. If `<vault>/01-project/context.md` exists, use it to find domain paths
2. Otherwise search for common patterns: `internal/**/domain/`, `src/domain/`, `pkg/domain/`, `lib/domain/`
3. Apply rules to any struct/type in the identified domain directories

## Rules

- avoid `any`/`interface{}` in domain entities
- avoid pointer fields in domain entities (`*string`, `*int`, `*time.Time`, etc.)
- do NOT introduce `Optional*` wrapper types in domain entities
- model optional semantics with concrete zero values (e.g., empty string, `0`, `time.Time{}`) and document that convention in code comments when needed
- represent JSON payloads with concrete types; prefer `json.RawMessage` at boundaries unless a strict schema exists
- prefer enums/value objects over free-form strings for constrained fields
- keep domain models persistence-agnostic (no `sql.Null*` in domain)

## Checklist Before Finishing

1. Search for `any` in changed domain files
2. Search for pointer fields in changed domain structs
3. Confirm optional semantics use zero-value conventions (no `Optional*` wrappers)
4. Run `gofmt` on edited files

## Output

- If a rule is violated: list file + field + proposed fix
- If all pass: report "domain guardrails OK"
