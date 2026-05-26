---
name: context-nav
disable-model-invocation: true
description: Sistema de contexto vivo para proyectos. Lee `.project-context/` al iniciar sesión, escribe deltas después de cada implementación. Aplica en modo directo y en pipeline, con o sin agentes.
---

# Context Navigator

Sistema de conocimiento acumulativo que vive en `.project-context/` al lado de `.handoff/`. Funciona en todo proyecto — nuevo o con historia, modo directo o pipeline.

## Cuándo aplica

**Al inicio de cualquier sesión** — modo directo o pipeline:
1. Si `.project-context/NAVIGATOR.md` existe → leerlo, verificar staleness, inyectar contexto relevante
2. Si no existe → indicar al usuario que puede ejecutar `context-init mode: init` para bootstrap

**Después de cada implementación** — el reporter (pipeline) o el propio Claude (directo) escribe deltas a `.project-context/`.

## Estructura de archivos

```
.project-context/
├── NAVIGATOR.md           # Índice + metadatos de cobertura y staleness
├── project.md             # Stack, arquitectura, restricciones, SOLID
├── patterns.md            # Patrones de diseño inferidos + referencias a archivos
├── contracts.md           # APIs REST, queues, eventos, webhooks, servicios externos
├── business-rules.md      # Invariantes de negocio que cruzan dominios
├── dependencies.md        # Grafo de dependencias entre dominios
├── domains/               # Un .md por bounded context significativo
│   └── <domain>.md
├── decisions/             # ADRs-lite: decisiones con contexto y alternativas
│   └── NNN-<slug>.md
└── risks.md               # Gotchas, deuda técnica, restricciones operativas
```

## Reglas de escritura

- **Delta, nunca sobrescritura total** — usar Edit para modificar secciones específicas
- **Actualizar `last_updated` en NAVIGATOR.md** siempre que se toque cualquier archivo
- **No inventar** — solo registrar lo que existe en el código o fue decidido explícitamente
- **Referencias a archivos obligatorias** — todo patrón o contrato debe citar `path:line` o al menos `path`
- **Un dominio por bounded context** — no crear dominios para paquetes utilitarios genéricos

## Reglas de lectura

- El orquestador lee `project.md` + `patterns.md` + `contracts.md` siempre
- Lee solo los dominios que la tarea va a tocar (inferir desde archivos afectados)
- En modo directo, inyectar resumen de NAVIGATOR.md en la primera respuesta de sesión

## Staleness

Ver `staleness.md` para reglas completas. Resumen:
- `last_updated` en NAVIGATOR.md vs fecha del último commit en `internal/` o `src/`
- > 3 días de diff → contexto posiblemente desactualizado, mencionarlo
- > 7 días → recomendar re-scan
- Nunca bloquear trabajo por staleness — es una advertencia, no un gate

## Skills relacionadas

- `scan-project` — produce el bootstrap inicial de `.project-context/`
- `leader` (agent) — consume `.project-context/` en el Paso 0.3 antes del primer sub-agente
- `reporter` — escribe deltas a `.project-context/` al final del pipeline
