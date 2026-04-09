# Handoff Template

Use this template when creating a new handoff file at `.handoff/<TASK-ID or slug>.md`. Fill in the sections as you work — do not leave placeholder comments in the final file.

---

```markdown
# Handoff — <TASK-ID or slug>

## Estado actual
- [ ] Step 1: description
- [ ] Step 2: description
- [ ] Step 3: description

## Archivos modificados
<!-- path — what was done and why -->

## Decisiones tomadas
<!-- decision — reasoning -->

## Siguiente paso
<!-- Exactly what to do next to resume work -->

## Notas
<!-- Non-obvious context: workarounds, bugs found, blockers -->

## Handoff for tester
<!--
MANDATORY for the developer to fill BEFORE finishing. The tester reads THIS section
instead of re-reading production files, so it must be complete and precise.
If this section is empty or incomplete when the developer reports done, the
orchestrator will bounce the task back to the developer.
-->

### Archivos de producción tocados
<!-- one line per file: path — role (store query / handler / DTO / custom component / etc.) -->

### Public interfaces / contracts
<!--
Exact signatures of what was added or modified. Copy-paste from the code.
- New types/structs with all fields and tags
- New functions/methods with full signatures (params, return types, error behavior)
- New DTOs with JSON tags
-->

### Patrones aplicados
<!-- which patterns from the convention skill the developer followed -->

### Edge cases descubiertos
<!-- NULL handling, empty states, error paths, race conditions considered, unusual inputs -->

### Build tags / constraints
<!-- //go:build tags, embed.FS layout, Wails bindings, Rust cfg, any quirk that affects how tests must be written -->

### Casos de test sugeridos
<!-- bullet list — starting hint for the tester, NOT a mandatory list. Tester decides final coverage. -->

### Validación ya ejecutada
<!-- go build, go vet, npm run build, lint — exact commands and their exit status. Tester does NOT repeat build checks. -->

## Token usage
| Session | Tokens used | Tokens available | Tool calls | Files read | Files written |
|---------|------------|-----------------|------------|------------|---------------|
```
