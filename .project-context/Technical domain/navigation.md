# Technical Domain — Navegación

last_updated: 2026-08-13

## Documentos base

| Archivo | Propósito | Cuándo consultarlo |
|---|---|---|
| [project.md](project.md) | Stack, arquitectura, restricciones no negociables y SOLID | Siempre — es el punto de entrada |
| [domain.md](domain.md) | Entidades principales, relaciones y bounded contexts | Al implementar lógica de negocio, al diseñar nuevas entidades |
| [glossary.md](glossary.md) | Lenguaje del negocio ↔ lenguaje técnico | Cuando el humano menciona términos del negocio, antes de nombrar entidades |
| [contracts.md](contracts.md) | APIs, estados válidos, transiciones permitidas, reglas de negocio | Al implementar endpoints/tools, al cambiar estados de entidades |
| [dependencies.md](dependencies.md) | Dependencias entre dominios y servicios externos | Antes de modificar un dominio, al agregar integraciones |
| [risks.md](risks.md) | Deuda técnica, gotchas y restricciones operativas | Antes de tocar áreas críticas, al planificar refactors |

## Documentos adicionales

| Archivo | Propósito | Cuándo consultarlo |
|---|---|---|
| [../infra-services.md](../infra-services.md) | Servicios de infra detectados (SQLite) y disponibilidad de MCP | Antes de escribir tests de integración o lógica de datos |

## Quién actualiza estos documentos

- `project.md`, `domain.md`, `contracts.md`, `dependencies.md`, `risks.md` — reporter automático al cierre de cada tarea
- `glossary.md` — pre-populado por context-init; validado y completado por el equipo manualmente
- Documentos adicionales — según lo defina el equipo al agregarlos
