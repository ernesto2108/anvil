---
name: prd-template
description: PRD writing guide with discovery questionnaire, template, and acceptance criteria format. Used by the PM agent to create consistent, complete PRDs. Use when writing a PRD or when user says "create PRD", "write requirements", "new feature".
---

# PRD Guide

A PRD defines **what** to build and **why**. Never **how** — that belongs to the architect and developer.

## Discovery Questionnaire (Spanish)

**Agent mode:** Skip this section entirely. The orchestrator already gathered answers from the user and included them in the prompt. Go directly to the PRD Template section.

**Interactive mode:** Ask ONE topic at a time. Wait for the user's response before moving to the next topic. This is a conversation, not a form — let the user be specific and go deep on each area.

**Conversation flow:**
1. Ask the first topic (Problema)
2. Wait for the user's answer
3. If the answer is vague or incomplete, ask a follow-up to clarify
4. Only then move to the next topic
5. Skip topics the user already answered in previous messages
6. Stop when you have enough to write the PRD — not every topic is needed for every task

### Topic 1: Problema
- Que problema resuelve? Describilo sin mencionar soluciones
- (follow-up if needed) Como lo resuelven hoy? Que pasa si no lo hacemos?

### Topic 2: Usuario
- Quien es el usuario principal? En que contexto lo usa?
- (follow-up if needed) Hay otros usuarios afectados?

### Topic 3: Exito
- Como sabemos que funciono? Que metrica se mueve?
- (follow-up if needed) Cual es el baseline? Que NO debe empeorar?

### Topic 4: Alcance
- Cual es la version minima que entrega valor? (MVP)
- (follow-up if needed) Hay deadline? Que NO deberia incluir?

### Topic 5: Plataforma
- Para que plataforma es? Web, mobile, o ambos?
- (follow-up if mobile) iOS, Android, o ambos? Flutter o nativo?
- (follow-up if ambos) Se comparte el design system o son independientes?

### Topic 6: User Journeys
- Cual es el flujo principal del usuario? (paso a paso)
- (follow-up if needed) Que pasa si algo sale mal? Hay estados vacios o edge cases?

### Topic 7: Riesgos
- Que estamos asumiendo que no hemos validado?
- (follow-up if needed) Que puede salir mal en produccion? Que mitigacion hay?

### Topic 8: Dependencias
- Depende de otro equipo, API externa, o servicio compartido?

After gathering enough answers: confirm with a brief summary in Spanish, get approval before writing.

## PRD Template

Create at: `<docs>/03-tasks/<TASK-ID>/prd.md`

```markdown
# <TASK-ID>: <Titulo>

## Problema
Que problema existe, para quien, y por que ahora. Incluir datos de soporte (tickets, metricas, feedback).

## Objetivos y metricas de exito
- **Metrica principal:** <que se mueve> (baseline: X, objetivo: Y)
- **Como medir:** <herramienta, query, dashboard>
- **Countermetric:** <que NO debe empeorar>

## Journeys de usuario

### Camino feliz
1. Usuario hace X
2. Sistema responde con Y
3. Usuario ve Z

### Camino de error
1. Usuario hace X con input invalido
2. Sistema responde con mensaje de error claro
3. Usuario puede reintentar

## Alcance

### Plataforma
- **Platform:** web | mobile | both
- **Mobile stack:** Flutter | iOS native | Android native | N/A
- **Shared design system:** yes | no | N/A

### Incluido
- <capacidad 1> — P0 (obligatorio para lanzar)
- <capacidad 2> — P0
- <capacidad 3> — P1 (importante, pronto despues)
- <capacidad 4> — P2 (deseable, futuro)

### Fuera de alcance
- <que NO incluye esta tarea y por que>

## Requerimientos funcionales

| # | Requerimiento | Prioridad | Notas |
|---|---|---|---|
| 1 | <especifico, testeable> | P0 | |
| 2 | <especifico, testeable> | P0 | |
| 3 | <especifico, testeable> | P1 | |

## Requerimientos no funcionales
- **Performance:** <tiempo de respuesta, throughput esperado>
- **Seguridad:** <auth, sensibilidad de datos, compliance>
- **Accesibilidad:** <nivel WCAG si aplica>
- **Escalabilidad:** <carga esperada, crecimiento>

## Criterios de aceptacion

Usar formato Dado/Cuando/Entonces. Un comportamiento por escenario.

### Feature: <nombre>

**Escenario: camino feliz**
- Dado <precondicion>
- Cuando <accion del usuario>
- Entonces <resultado esperado>

**Escenario: caso de error**
- Dado <precondicion>
- Cuando <accion invalida>
- Entonces <comportamiento de error>

**Escenario: caso borde**
- Dado <precondicion inusual>
- Cuando <accion>
- Entonces <manejo esperado>

## Supuestos y riesgos

| Riesgo | Impacto | Mitigacion |
|---|---|---|
| <supuesto no validado> | <que se rompe si es incorrecto> | <como mitigar> |

## Dependencias
- <equipos externos, APIs, infra compartida>

## Preguntas abiertas
- [ ] <decision pendiente> — Responsable: <quien>, Deadline: <cuando>
- [ ] <otra pregunta abierta>

## Rollout
- <fases, feature flags, necesidades de migracion>
```

## Reglas

- **Criterios de aceptacion deben usar Dado/Cuando/Entonces** — nada vago como "deberia funcionar bien"
- **Incluir al menos 1 escenario de error y 1 caso borde** por feature
- **Sin detalles de implementacion** — sin schemas de DB, sin contratos de API, sin decisiones de arquitectura
- **Requerimientos funcionales deben tener prioridad** (P0/P1/P2)
- **Metricas de exito deben tener baseline** — "reducir X de 68% a 50%", no solo "reducir X"
- **Una pagina maximo** — si es muy grande, dividir en multiples tareas
- **Preguntas abiertas es obligatorio** — si todo esta decidido, escribir "Ninguna"

## Que NO va en un PRD

| Contenido | Donde pertenece |
|---|---|
| Schema de DB, contratos de API, arquitectura | Documento de diseno tecnico (architect) |
| Herramientas, lenguajes, frameworks especificos | Documento de diseno tecnico |
| Disenos de UI pixel-perfect | Figma / herramienta de diseno (designer) |
| Asignaciones de tareas, cronogramas detallados | Backlog / gestion de proyecto |
| Lenguaje vago ("rapido", "amigable") | Reescribir como criterio medible |
