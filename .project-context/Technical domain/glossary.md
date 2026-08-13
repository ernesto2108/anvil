# Glosario de Dominio — anvil

<!-- Mapa entre el lenguaje del negocio (cómo habla el equipo humano)
     y el lenguaje técnico (cómo vive en el código y la base de datos).
     Las filas marcadas con ⚠️ fueron pre-populadas automáticamente y requieren validación del equipo. -->

last_updated: 2026-08-13

## Entidades

| Término humano | Término técnico | Tabla / Struct / Tipo | Descripción |
|---|---|---|---|
| ⚠️ pendiente validación | Agent | `entity.Agent` / tabla `agents` | Un agente de IA invocado dentro de un run |
| ⚠️ pendiente validación | Run | `entity.Run` / tabla `runs` | Ejecución de un pipeline o sesión de agentes |
| ⚠️ pendiente validación | Event | `entity.Event` / tabla `events` | Evento emitido durante un run (capturado por `internal/memory/capture`) |
| ⚠️ pendiente validación | File | `entity.File` | Archivo tocado/afectado durante un run |
| ⚠️ pendiente validación | File edge | `entity.FileEdge` | Relación entre archivos (dependencia, referencia) detectada en un run |
| ⚠️ pendiente validación | Flow | `entity.Flow` | Secuencia/flujo de pasos dentro de un run |
| ⚠️ pendiente validación | Prompt | `entity.Prompt` | Prompt enviado a un agente/LLM |
| ⚠️ pendiente validación | Task | `entity.Task` | Unidad de trabajo dentro de un DAG de orquestación |
| ⚠️ pendiente validación | Tool | `entity.Tool` | Tool MCP invocada por un agente |
| ⚠️ pendiente validación | Error group | `entity.ErrorGroup` | Agrupación de errores relacionados detectados en un run |

## Acciones / Verbos

| Término humano | Término técnico | Método / Endpoint | Descripción |
|---|---|---|---|
| ⚠️ pendiente validación | Digest | `digestFromHandoff` (`internal/mcp/context.go`) | Resumen de un handoff/transcript escrito como memoria persistente |
| ⚠️ pendiente validación | Capture | `internal/memory/capture` | Captura en vivo de eventos de una sesión |
| ⚠️ pendiente validación | Emit | subcomando `emit` (`internal/cli`) | Emisión de eventos desde hooks de Claude Code hacia anvil |
| ⚠️ pendiente validación | Replan | `internal/orchestrator/replanner.go` | Reajuste del plan de ejecución ante un fallo en el DAG |

## Estados

| Término humano | Término técnico | Valor en código | Descripción |
|---|---|---|---|
| ⚠️ pendiente validación | Gate | `internal/orchestrator/gate.go` | Punto de aprobación/decisión entre fases de un pipeline |

## Términos ambiguos o conflictivos

<!-- Términos que significan cosas distintas según el contexto.
     Documentarlos explícitamente para evitar malentendidos. -->

| Término | Contexto A | Contexto B | Nota |
|---|---|---|---|
| "servicio" | En `internal/mcp` / `internal/orchestrator`, no implica un proceso de red separado | En arquitecturas de microservicios el término implica despliegue independiente | anvil es un monolito CLI — "servicio" en este repo nunca implica un proceso separado en runtime |

## Aliases conocidos

<!-- Variantes informales o históricas que el equipo sigue usando. -->

| Alias | Término canónico | Nota |
|---|---|---|
| `anvil-full` | build del dashboard (`go build -tags "dashboard production fts5"`) | Mismo binario `anvil`, con el módulo dashboard Wails embebido |
