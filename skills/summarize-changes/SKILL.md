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
- `<vault>/02-backlog/sprint-current.md`
- `<vault>/03-tasks/<TASK-ID>/prd.md`
- `<vault>/01-project/context.md`

## Acciones

- Usar la salida de git-diff para identificar los archivos cambiados
- Agrupar por feature
- Inferir la intención a partir de las tareas
- Explicar las razones

Salida: `<vault>/06-reports/last-run.md` (sobreescribir si existe)
