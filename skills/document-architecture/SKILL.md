---
name: document-architecture
description: Documentar la arquitectura de un servicio frontend o backend. Auto-detecta el tipo de proyecto y ejecuta el pipeline correspondiente. Usar cuando el usuario diga "documenta arquitectura", "document service", "documenta frontend", "document architecture", o cuando orchestrate enruta aquí.
disable-model-invocation: true
---

# Document Architecture

Punto de entrada unificado para documentar cualquier proyecto. Detecta el tipo, carga la guía correspondiente y ejecuta el pipeline.

## Paso 0 — Input y Detección

Si se invoca sin argumentos, pregunta al usuario (en español):
1. **¿Que proyecto?** — nombre del repo
2. **¿Que tarea del backlog?** — ID. Si no hay, preguntar si crear una.

Resuelve `<docs>` desde `~/.claude/project-registry.md`.
Resuelve `<repo>` desde `~/projects/<project-name>`.

### Auto-detección

Ejecuta `ls <repo>/` y determina:

| El root contiene | Tipo |
|---|---|
| `package.json` + `src/` con `.tsx`/`.jsx` | **frontend** |
| `go.mod` + `internal/` (SIN `domain/`) | **service** (MVC) |
| `go.mod` + `domain/` + `usecase/` | **service** (Clean) |
| `go.mod` + `internal/` con `business/` | **service** (Hex) |
| Ambiguo | Preguntar: "¿Frontend o backend?" |

**Carga la guía correspondiente** desde `guides/frontend-pipeline.md` o `guides/service-pipeline.md`.

## Paso 1 — Verificar patrón de output

Verifica `<docs>/04-architecture/` para la estructura de archivos esperada. La guía especifica qué archivos generar.

## Paso 2 — Decidir seguridad

Pregúntate a ti mismo (NO al usuario): ¿este proyecto maneja auth, pagos, PII o datos sensibles?
- **Sí** → incluir seguridad en el pipeline
- **No** → omitir seguridad

## Paso 3 — Scanner (profundo, con conciencia de skeleton)

Lee el archivo skeleton especificado en la guía. Inyéctalo INLINE en el prompt del scanner. Lanza el agente `scanner` con `mode: deep` siguiendo las instrucciones del scanner de la guía.

- Modelo: **sonnet**
- **Objetivo: <25 tool calls**
- Después de completar: **Lee los archivos de contexto** con la herramienta Read.

## Paso 4 — Architect (2 agentes en paralelo)

Lanza DOS agentes architect con `mode: documentation` siguiendo la guía:
- **4a — Overview:** inyecta context-summary.md INLINE → produce `overview.md`
- **4b — Detail:** inyecta template + context-detail.md INLINE → produce `endpoints/*.md`

Nota: el modo documentation usa nombres `overview.md` / `endpoints/*.md` (no las vistas `architecture*.md` del modo de tarea). Esto es intencional — el modo documentation captura la arquitectura existente, no nuevas decisiones de diseño.

- Modelo: **sonnet**
- **Objetivo: 0 llamadas Read** para el agente detail (todo inline)

**Espera a que ambos terminen antes del Paso 5.** Lee overview.md para el resumen de seguridad.

## Paso 5 — Seguridad [condicional]

Omitir si el Paso 2 dijo que no. Lee `known-systemic-issues.md`, inyéctalo INLINE con context-risks.md + resumen de overview. Lanza el agente `security` siguiendo las instrucciones de seguridad de la guía.

- Modelo: **sonnet**
- **Objetivo: <10 tool calls**

## Paso 6 — Cerrar tarea

1. Actualiza el estado de la tarea a `done`
2. Actualiza board.md
3. Elimina duplicado del backlog si existe
4. Actualiza métricas del sprint

## Reglas

- **Todo el output en español** — títulos, descripciones, labels de Mermaid. Código/rutas en inglés.
- **Inyección de contexto OBLIGATORIA** — cada agente recibe SOLO su segmento INLINE.
- **Modelo: sonnet** para todos los agentes.
- **Mermaid: NUNCA usar `|` dentro de labels o mensajes.** Usar `/` en su lugar.
- **Presupuesto de tokens: <50 tool calls en total.**
