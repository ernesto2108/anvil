# Committer handoff — Fase 1 → Fase 2

- TASK-ID: (sin TASK-ID — ad-hoc)
- run_id: ad-hoc
- Commit hash: e7de67c
- Commit subject: fix(skills): shorten test-api description under Kiro limit
- Rama destino: develop (pusheado: fe6c012..e7de67c)
- Remoto: git@github.com:ernesto2108/anvil.git
- Fecha Fase 1: 2026-08-08T23:40:00Z

## Mensaje del commit (verbatim)

feat(skills): add task-flow for Linear + Outline

Two-phase task lifecycle skill for work repos: phase 1 drafts the
issue in Spanish and creates it in Linear behind a human confirmation
gate; phase 2 delegates commit/PR to committer-flow, moves the issue
to Done with the PR link, and writes a per-task doc in Outline.

Gate 0 resolves the docs backend from the project registry routing
table (single source of truth) instead of hardcoding repo prefixes,
asking the human only when the project has no match. Credentials are
read at runtime from the registry and never hardcoded in the skill.

Co-Authored-By: Claude Fable 5 <noreply@anthropic.com>
Claude-Session: https://claude.ai/code/session_01DzRvsa9HHMhEbfa8PmcSy5

## Notas

Corrí sin TASK-ID — gate de handoff omitido. run_id ausente — handoff propio en ad-hoc/. El usuario pidió commit y publicación en el mismo turno; `.project-context/` excluido del commit por decisión del usuario.
