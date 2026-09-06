---
name: context-nav
user-invocable: false
description: Sistema de contexto vivo para proyectos. Lee `.project-context/` al iniciar sesión, escribe deltas después de cada implementación. Aplica en modo directo y en pipeline, con o sin agentes. Úsalo cuando se necesite navegar o actualizar el contexto del proyecto en .project-context/, o cuando context-init lo cargue durante el bootstrap de sesión.
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


# Context Navigator

Sistema de conocimiento acumulativo que vive en `.project-context/` al lado de `.handoff/`. Funciona en todo proyecto — nuevo o con historia, modo directo o pipeline.

## Cuándo aplica

**Al inicio de cualquier sesión** — modo directo o pipeline:
1. Si `.project-context/NAVIGATOR.md` existe → leerlo, verificar staleness, inyectar contexto relevante
2. Si no existe → indicar al usuario que puede ejecutar `context-init mode: init` para bootstrap

**Después de cada implementación** — el reporter (pipeline) o el propio Claude (directo) escribe deltas a `.project-context/`.

## Gate de contexto al inicio (agentes que implementan)

Los agentes que implementan código (developers de stack) aplican este gate antes de escribir nada. Esta es la fuente única del procedimiento — los agentes solo cargan esta skill y lo ejecutan.

1. **Gate de existencia:** `.project-context/NAVIGATOR.md` debe existir. Si no existe, DETENER y responder al humano en una sola línea: **"No existe `.project-context/NAVIGATOR.md` — ejecuta el agente `context-init` primero y luego continúa."** No implementar nada hasta que exista el contexto.
2. **Carga proporcional al tamaño del cambio** (el agente decide, no pregunta; declara el nivel elegido — si su spec define un bloque de arranque con campo `contexto:`, esa es la declaración y no se repite en una línea aparte):
   - **Cambio acotado** (≤2 archivos, todos en rutas que ya existen — sin carpetas, paquetes ni dominios nuevos —, sin contratos nuevos, sin dependencias nuevas, sin decisiones de diseño): leer `NAVIGATOR.md` + `.project-context/Core/coding-standards.md` (siempre) + `.project-context/Core/patterns.md` si el área tocada lo amerita. Declarar: **"Contexto: ligero."**
   - **Cualquier otro caso** — incluida la creación de carpetas, paquetes o dominios nuevos: leer `NAVIGATOR.md`, `.project-context/Technical domain/project.md`, `.project-context/Core/coding-standards.md`, `.project-context/Core/patterns.md`, `.project-context/Technical domain/business-rules.md`, `.project-context/Technical domain/contracts.md` y `.project-context/Core/workflows.md`. Declarar: **"Contexto: completo."**
3. **Gate de placement (rutas nuevas):** antes de crear cualquier archivo en una ruta que no existe, validar el placement contra `Core/coding-standards.md §Estructura de carpetas` y la convención de paths de `Technical domain/project.md`. Declarar en una línea la ruta elegida y la regla que la respalda — formato: **"Placement: `internal/<dominio>/` según coding-standards §Estructura de carpetas."** Si ninguna regla documentada respalda la ruta → DETENER y proponer al humano la ruta con su justificación antes de crear el archivo.
4. Usar lo leído como contexto autoritativo durante todo el run. Si un archivo esperado no existe o está vacío, mencionar al humano cuál falta antes de continuar.

## Estructura de archivos

```
.project-context/
├── NAVIGATOR.md                          # INTOCABLE — índice general + metadatos
│
├── Core/
│   ├── navigation.md                     # Índice de Core
│   ├── workflows.md                      # Ramas, ambientes, deploy, comandos operativos
│   ├── task-management.md                # Gestión de tareas, tickets, DoD
│   ├── coding-standards.md               # Naming, linting, patrones detectados
│   └── patterns.md                       # Patrones de diseño inferidos del código
│
├── Technical domain/
│   ├── navigation.md                     # Índice de Technical domain
│   ├── project.md                        # Stack, arquitectura, restricciones, SOLID
│   ├── domain.md                         # Entidades y bounded contexts
│   ├── glossary.md                       # Lenguaje humano ↔ técnico
│   ├── contracts.md                      # APIs, queues, eventos, reglas de negocio
│   ├── business-rules.md                 # Invariantes de negocio + modelo de auth
│   ├── dependencies.md                   # Grafo de dependencias entre dominios
│   └── risks.md                          # Deuda técnica, gotchas, restricciones
│
├── decisions/                            # ADRs-lite
│   └── NNN-<slug>.md
└── runs/                                 # Reportes por run (opcional)
```

## Reglas de escritura

- **Delta, nunca sobrescritura total** — usar Edit para modificar secciones específicas
- **Actualizar `last_updated` en NAVIGATOR.md** siempre que se toque cualquier archivo — es SOLO una fecha `YYYY-MM-DD`: reemplazar el valor anterior, nunca concatenar, preservar el previo ni escribir texto narrativo del run
- **No inventar** — solo registrar lo que existe en el código o fue decidido explícitamente
- **Referencias a archivos obligatorias** — todo patrón o contrato debe citar `path:line` o al menos `path`
- **Un dominio por bounded context** — no crear dominios para paquetes utilitarios genéricos

## Reglas de lectura

- Los agentes que implementan (orquestador y developers de stack) aplican el **§Gate de contexto al inicio** — ese gate es la fuente única de qué archivos leer según el nivel (ligero/completo); no existe una lista "siempre" aparte
- Lee solo los dominios que la tarea va a tocar (inferir desde archivos afectados; si la tarea crea un dominio nuevo, no hay archivos de los que inferir → aplica el nivel completo y el gate de placement del §Gate)
- En modo directo, inyectar resumen de NAVIGATOR.md en la primera respuesta de sesión

## Staleness

Ver `staleness.md` para reglas completas. Resumen:
- `last_updated` en NAVIGATOR.md vs fecha del último commit en `internal/` o `src/`
- > 3 días de diff → contexto posiblemente desactualizado, mencionarlo
- > 7 días → recomendar re-scan
- Nunca bloquear trabajo por staleness — es una advertencia, no un gate

## Skills relacionadas

- `scan-project` — produce el bootstrap inicial de `.project-context/`
- `reporter` — escribe deltas a `.project-context/` al final del pipeline
