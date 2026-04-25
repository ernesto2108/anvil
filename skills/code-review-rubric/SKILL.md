---
name: code-review-rubric
description: Rúbrica de puntuación y formato de reporte para revisiones de código. Define criterios de evaluación, escala de puntuación y estructura de salida. Usado por el agente QA y cualquiera que revise calidad de código.
---

# Rúbrica de Revisión de Código

## Criterios de Evaluación

### Corrección
- Bugs de lógica, casos borde, seguridad ante nil/null
- Manejo de errores (errores envueltos, sin retornos desnudos)
- Cumplimiento de contrato (coincide con PRD/diseño)

### Rendimiento
- Asignaciones innecesarias, I/O bloqueante
- Queries N+1, algoritmos ineficientes
- Falta de paginación, queries sin límite

### Calidad de Código
- Claridad de nombres, legibilidad
- Complejidad ciclomática, duplicación
- Responsabilidad única, abstracciones mínimas

### Testing
- Tests unitarios presentes para lógica de negocio
- Rutas críticas cubiertas
- Casos borde y rutas de error testeadas
- Determinístico (sin flaky, sin dependencia de tiempo)

### Seguridad ante Concurrencia
- Race conditions, estado compartido
- Uso correcto de mutexes/channels/errgroups
- Propagación de contexto y cancelación

### Seguridad
- Validación de entrada en los límites
- SQL injection, XSS, command injection
- Verificaciones de auth/autorización presentes
- Secretos no hardcodeados

## Escala de Puntuación

| Puntuación | Significado | Acción |
|---|---|---|
| 9-10 | Excelente — listo para producción | Aprobar |
| 7-8 | Bueno — solo mejoras menores | Aprobar con sugerencias |
| 5-6 | Requiere trabajo — problemas significativos | Bloquear, crear tareas |
| 3-4 | Problemas mayores — repensar el enfoque | Bloquear, escalar al arquitecto |
| 1-2 | Crítico — riesgo de seguridad/datos | Bloquear de inmediato |

**Umbral: puntuación < 7 → BLOQUEAR y crear tareas en el backlog.**

## Formato del Reporte

Escribir en: `<docs>/03-tasks/<TASK-ID>/qa-review.md`

```markdown
# QA Review — <TASK-ID>

## Score: X/10

## Resumen
Un párrafo: qué se revisó, evaluación general.

## Fortalezas
- Qué se hizo bien (reconocer el buen trabajo)

## Problemas
| # | Severidad | Categoría | Archivo | Descripción |
|---|---|---|---|---|
| 1 | critical | correctness | internal/user/service.go:45 | Missing nil check on... |
| 2 | high | testing | — | No tests for error path in... |
| 3 | medium | quality | internal/order/handler.go:12 | Function too complex (cyclomatic 15) |

## Mejoras
- Sugerencias accionables (no vagas como "mejorar esto")

## Nivel de Riesgo
low / medium / high — basado en el radio de explosión de los problemas encontrados
```

## Formato de Tarea para el Backlog

Cuando se encuentran problemas, agregar a: `<docs>/02-backlog/sprint-current.md`

Cada tarea debe incluir:
- Título (imperativo: "Fix nil check in user service")
- Tipo (bug / tech-debt / test-gap)
- Descripción (qué está mal y por qué importa)
- Severidad (critical / high / medium / low)
- Corrección sugerida (concreta, no vaga)
- Archivos afectados (con números de línea)
