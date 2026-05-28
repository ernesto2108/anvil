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

### Tema 2: Objetivo
- Que vas a lograr especificamente? Como lo medirias en una frase?
- (seguimiento si es necesario) Es distinto del problema: el problema es el dolor, el objetivo es el resultado concreto y medible que persigues.

### Tema 3: Usuario
- Quien es el usuario principal? En que contexto lo usa?
- (seguimiento si es necesario) Hay otros usuarios afectados?

### Tema 4: Exito
- Como sabemos que funciono? Que metrica se mueve?
- (seguimiento si es necesario) Cual es el baseline? Que NO debe empeorar?

### Tema 5: Alcance
- Cual es la version minima que entrega valor? (MVP)
- (seguimiento si es necesario) Hay deadline? Que NO deberia incluir?

### Tema 6: Plataforma
- Para que plataforma es? Web, mobile, o ambos?
- (seguimiento si es mobile) iOS, Android, o ambos? Flutter o nativo?
- (seguimiento si es ambos) Se comparte el design system o son independientes?

### Tema 7: Requerimientos Funcionales
El objetivo es identificar los Requerimientos Funcionales (RF) como unidades discretas, cada una con sus escenarios Dado/Cuando/Entonces.
- Cuales son las capacidades concretas que el sistema debe permitir? Nombra cada una como un RF discreto (ej. "RF: crear workflow", "RF: editar perfil").
- Para cada RF: cual es el camino feliz? (Dado <precondición>, Cuando <acción>, Entonces <resultado>)
- (seguimiento por cada RF) Que pasa si algo sale mal — un error de validación o del sistema? Hay un caso borde o estado vacío?

### Tema 8: Requerimientos No Funcionales
- Performance: hay un tiempo de respuesta o throughput esperado?
- Seguridad: hay datos sensibles, requisitos de auth, o compliance?
- Accesibilidad: aplica algún nivel WCAG?
- Escalabilidad: cuál es la carga o el volumen esperado?

### Tema 9: Riesgos
- Que estamos asumiendo que no hemos validado?
- (seguimiento si es necesario) Que puede salir mal en produccion? Que mitigacion hay?

### Tema 10: Dependencias
- Depende de otro equipo, API externa, o servicio compartido?

Después de recopilar suficientes respuestas: confirmar con un breve resumen en español, obtener aprobación antes de escribir.

## Template de PRD

Crear en: `{task_path}/prd.md` dentro de `.project-context/` o del repo (el orquestador provee `task_path` resuelto). El PRD es siempre un archivo local. Si el proyecto tiene `task_tool` configurado (campo de `.project-context/project.md`), al finalizar **indicar al humano**: "Vincula este PRD en {task_tool}". Nunca ejecutar acciones en la herramienta externa.

```markdown
# <TASK-ID>: <Titulo>

## 1. Contexto y Problema
Qué problema existe, para quién, y por qué ahora.
Incluir datos: usuarios afectados, impacto en negocio, señales observadas.

## 2. Objetivo
Una frase clara y medible de qué vas a lograr.

## 3. Métricas de Éxito
- **Primaria:** <métrica> (baseline: X → objetivo: Y)
- **Secundaria:** <qué vigilar para no romper otras cosas>
- **Cómo medir:** <herramienta, query, dashboard>

## 4. Usuarios y Casos de Uso
Quién va a usar esto y en qué escenarios.
- **Usuario principal:** <descripción + contexto de uso>
- **Otros afectados:** <si aplica>
- **Jobs-to-be-done:** <qué tarea real resuelve>

## 5. Scope
### Incluido
- <capacidad 1> — P0
- <capacidad 2> — P1

### Fuera de alcance
- <qué NO incluye esta tarea y por qué>

## 6. Requerimientos Funcionales

### RF-01: <nombre del requerimiento>
<descripción del comportamiento esperado>

**Escenario: camino feliz**
- Dado <precondición>
- Cuando <acción del usuario>
- Entonces <resultado esperado>

**Escenario: error**
- Dado <precondición>
- Cuando <acción inválida>
- Entonces <comportamiento de error>

**Escenario: caso borde**
- Dado <precondición inusual>
- Cuando <acción>
- Entonces <manejo esperado>

---

### RF-02: <nombre del requerimiento>
...

## 7. Requerimientos No Funcionales
- **Performance:** <tiempo de respuesta, throughput>
- **Seguridad:** <auth, datos sensibles, compliance>
- **Accesibilidad:** <nivel WCAG si aplica>
- **Escalabilidad:** <carga esperada>

## 8. Dependencias y Riesgos

| Item | Tipo | Impacto | Mitigación |
|------|------|---------|------------|
| <equipo / sistema / tercero> | Dependencia | <qué se bloquea> | <acción> |
| <supuesto no validado> | Riesgo | <qué se rompe> | <acción> |

## 9. Milestones y Timeline
- **MVP:** <fecha> — <qué incluye>
- **v1.0:** <fecha> — <qué agrega>

## 10. Preguntas Abiertas
- [ ] <decisión pendiente> — Responsable: <quien>, Deadline: <cuando>
```

## Reglas

- **Cada RF lleva sus escenarios Dado/Cuando/Entonces** integrados en la sección 6 — nada vago como "debería funcionar bien"
- **Incluir al menos 1 escenario de error y 1 caso borde** por requerimiento funcional
- **Sin detalles de implementación** — sin schemas de DB, sin contratos de API, sin decisiones de arquitectura
- **Las capacidades en Scope deben tener prioridad** (P0/P1/P2)
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
