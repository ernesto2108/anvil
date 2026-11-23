# Committer handoff — Fase 1 → Fase 2

- TASK-ID: (sin TASK-ID — cambio del sistema de agentes)
- run_id: ad-hoc
- Commit hash: 71b11eba6300143fa847696639cb34167f567681
- Commit subject: chore(agents): elimina qa-fixer y deja que qa sugiera el developer de stack
- Rama destino: develop
- Remoto: git@github.com:ernesto2108/anvil.git
- Fecha Fase 1: 2026-09-20T20:01:51Z

## Mensaje del commit (verbatim)

chore(agents): elimina qa-fixer y deja que qa sugiera el developer de stack

Retira por completo el agente qa-fixer del sistema. Ahora, cuando qa
reporta un bug o bloqueador, indica el developer de stack sugerido
(developer-backend / developer-frontend / developer-mobile según la
ubicación del código) para que el humano lo invoque directamente —
sin correcciones quirúrgicas intermedias.

Cambios: agents/qa-fixer.md eliminado; agents/qa.md, arch-reviewer.md,
developer-frontend.md, developer-mobile.md y skills/committer-flow
actualizados para quitar toda referencia a qa-fixer; .context/domains/
ai-system.md sincronizado. Exports de Codex (.codex/agents/,
.agents/skills/) regenerados vía export-system.

Co-Authored-By: Claude Sonnet 5 <noreply@anthropic.com>

## Notas

Corrí sin TASK-ID — gate de handoff omitido. Sin handoff de developer — staging basado en git status puro.
