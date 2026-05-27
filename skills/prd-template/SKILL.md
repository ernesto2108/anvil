---
name: prd-template
description: Guía para redactar PRDs con cuestionario de descubrimiento, template y formato de criterios de aceptación. Usado por el agente PM para crear PRDs consistentes y completos. Usar al escribir un PRD o cuando el usuario diga "crear PRD", "escribir requerimientos", "nueva funcionalidad".
---

# Guía de PRD

Un PRD define **qué** construir y **por qué**. Nunca el **cómo** — eso corresponde al arquitecto y al desarrollador.

## Cuestionario de Descubrimiento (Español)

**Modo agente:** Omitir esta sección completamente. El orquestador ya recopiló las respuestas del usuario y las incluyó en el prompt. Ir directamente a la sección de Template de PRD.

**Modo interactivo:** Preguntar UN tema a la vez. Esperar la respuesta del usuario antes de pasar al siguiente tema. Es una conversación, no un formulario — dejar que el usuario sea específico y profundice en cada área.

**Flujo de conversación:**
1. Preguntar el primer tema (Problema)
2. Esperar la respuesta del usuario
3. Si la respuesta es vaga o incompleta, hacer una pregunta de seguimiento para aclarar
4. Solo entonces pasar al siguiente tema
5. Omitir temas que el usuario ya respondió en mensajes anteriores
6. Detenerse cuando se tenga suficiente para escribir el PRD — no todos los temas son necesarios para cada tarea

### Tema 1: Problema
- Que problema resuelve? Describilo sin mencionar soluciones
- (seguimiento si es necesario) Como lo resuelven hoy? Que pasa si no lo hacemos?

### Tema 2: Usuario
- Quien es el usuario principal? En que contexto lo usa?
- (seguimiento si es necesario) Hay otros usuarios afectados?

### Tema 3: Exito
- Como sabemos que funciono? Que metrica se mueve?
- (seguimiento si es necesario) Cual es el baseline? Que NO debe empeorar?

### Tema 4: Alcance
- Cual es la version minima que entrega valor? (MVP)
- (seguimiento si es necesario) Hay deadline? Que NO deberia incluir?

### Tema 5: Plataforma
- Para que plataforma es? Web, mobile, o ambos?
- (seguimiento si es mobile) iOS, Android, o ambos? Flutter o nativo?
- (seguimiento si es ambos) Se comparte el design system o son independientes?

### Tema 6: User Journeys
- Cual es el flujo principal del usuario? (paso a paso)
- (seguimiento si es necesario) Que pasa si algo sale mal? Hay estados vacios o edge cases?

### Tema 7: Riesgos
- Que estamos asumiendo que no hemos validado?
- (seguimiento si es necesario) Que puede salir mal en produccion? Que mitigacion hay?

### Tema 8: Dependencias
- Depende de otro equipo, API externa, o servicio compartido?

Después de recopilar suficientes respuestas: confirmar con un breve resumen en español, obtener aprobación antes de escribir.

## Template de PRD

Crear en: `{task_path}/prd.md` dentro de `.project-context/` o del repo (el orquestador provee `task_path` resuelto). El PRD es siempre un archivo local. Si el proyecto tiene `task_tool` configurado (campo de `.project-context/project.md`), al finalizar **indicar al humano**: "Vincula este PRD en {task_tool}". Nunca ejecutar acciones en la herramienta externa.

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

## Scope
- **Type:** new | visual-improvement | functional-improvement | both
- **Platform:** web | mobile | both
- **Stack:** backend | frontend | fullstack
- **Milestone:** <nombre del milestone> (ej. MVP, v1.0, v2.0, Sprint Q2)
- **Existing assets:** [lista de archivos, componentes, pantallas que ya existen]
- **Design status:** none | exists-no-changes | exists-needs-update | new-needed

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
- **Tests de carga requeridos:** <sí/no — si sí: rps objetivo, p99 target, herramienta preferida (k6/Vegeta/Locust)>
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

- **Los criterios de aceptación deben usar Dado/Cuando/Entonces** — nada vago como "debería funcionar bien"
- **Incluir al menos 1 escenario de error y 1 caso borde** por feature
- **Sin detalles de implementación** — sin schemas de DB, sin contratos de API, sin decisiones de arquitectura
- **Los requerimientos funcionales deben tener prioridad** (P0/P1/P2)
- **Las métricas de éxito deben tener baseline** — "reducir X de 68% a 50%", no solo "reducir X"
- **Máximo una página** — si es muy grande, dividir en múltiples tareas
- **Las preguntas abiertas son obligatorias** — si todo está decidido, escribir "Ninguna"

## Formato de salida para descomposición de tareas

Cuando el PM descompone un PRD en tareas (vía `/backlog-management`), las tareas se escriben siempre como archivos locales en `.project-context/` o en el repo, con frontmatter YAML simple. Ver la skill `/backlog-management` para el formato exacto.

Si el proyecto tiene `task_tool` configurado (campo de `.project-context/project.md`), al finalizar **indicar al humano** qué tareas crear en esa herramienta — nunca ejecutar acciones en la herramienta externa.

## Qué NO va en un PRD

| Contenido | Donde pertenece |
|---|---|
| Schema de DB, contratos de API, arquitectura | Documento de diseño técnico (architect) |
| Herramientas, lenguajes, frameworks específicos | Documento de diseño técnico |
| Diseños de UI pixel-perfect | Figma / herramienta de diseño (designer) |
| Asignaciones de tareas, cronogramas detallados | Backlog / gestión de proyecto |
| Lenguaje vago ("rápido", "amigable") | Reescribir como criterio medible |
