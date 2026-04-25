---
name: summarize-changes
disable-model-invocation: true
description: Crea un resumen legible para humanos de qué cambió y por qué, escribiendo last-run.md al vault. Úsalo cuando el usuario diga "resume lo que hicimos", "escribe un reporte", "qué cambiamos", o al final de una sesión de trabajo para documentar el progreso.
---

Crea un resumen legible para humanos de qué cambió y por qué.

## Prerrequisito

Invocar `/git-diff` primero para recopilar el diff crudo y el resumen de cambios. Usar su salida como base para el reporte.

## Entradas

- Salida de `/git-diff` (archivos cambiados, estadísticas del diff)
- `{backlog_path}` (sprint-current.md o Linear — según sistema de docs)
- `{task_path}/prd.md` (o documento de Outline en Linear+Outline)
- `{context_path}` (context.md — resuelto por el orquestador)

## Acciones

- Usar la salida de git-diff para identificar los archivos cambiados
- Agrupar por feature
- Inferir la intención a partir de las tareas
- Explicar las razones

Salida: `{reports_path}/last-run.md` (sobreescribir si existe). El orquestador resuelve `reports_path` según el sistema de docs.
