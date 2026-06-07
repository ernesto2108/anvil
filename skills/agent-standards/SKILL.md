---
name: agent-standards
description: Estándares y checklist para crear o modificar agentes. Úsalo cuando el usuario diga "nuevo agente", "crear agente", "modificar agente", "plantilla de agente", "checklist de agente", o cuando el agent-designer vaya a escribir o revisar un agents/*.md.
---

# Estándares de Creación de Agentes

Estos son los estándares obligatorios para cada agente en este proyecto.

## Filosofía

1. **Rol sobre procedimiento** — un agente define quién es y qué no es suyo; los pasos detallados van en skills que carga.
2. **Dominio exclusivo sin solapamiento** — cada agente posee un conjunto de archivos/responsabilidades que ningún otro toca. Sin solapamiento = sin ambigüedad de routing.
3. **Mínimo permiso y modelo** — usar el nivel más bajo de permissionMode y el tier más bajo de model que la tarea permita. Escalar solo cuando hay evidencia de necesidad.

## Checklist Pre-Creación

Antes de escribir un nuevo agente, verificar:

- [ ] Ningún agente existente ya cubre este dominio (revisar descriptions de agents/*.md)
- [ ] El nuevo agente tiene dominio exclusivo claro (archivos/responsabilidades que solo él toca)
- [ ] Se justifica como agente y no como skill (ver criterios de decisión abajo)
- [ ] El tier de modelo y permissionMode están justificados

## Criterios de Decisión: ¿Agente o Skill?

| Criterio | Agente | Skill |
|---|---|---|
| Tiene identidad y rol propio | Sí | No |
| Puede ejecutarse standalone | Sí | No (necesita un agente host) |
| Toma decisiones de routing/scope | Sí | No |
| Sabe a quién derivar | Sí | No |
| Es reutilizable por 2+ agentes | No | Sí |
| Contiene procedimientos/pasos | Secundario | Principal |
| Aparece en la tabla del CLAUDE.md | Sí | No |

**Test rápido:**
> "¿Necesita saber qué está fuera de su scope, a quién derivar, y actuar como especialista con criterio propio?" → Si SÍ: agente. Si NO: skill.

## Frontmatter Canónico

```yaml
---
name: <slug>                  # minúsculas, guiones, igual al filename sin .md
description: <texto>          # qué hace + cuándo invocarlo (routing del harness)
permissionMode: read | write | execute
model: low | medium | high
skills:                       # opcional — skills que se cargan al invocar este agente
  - skill-name
---
```

**Tiers de modelo:**
- `low` → rápido/barato: análisis puntual, formateo, reportes
- `medium` → balanceado: implementación estándar, tests
- `high` → capaz: diseño, arquitectura, decisiones complejas

**Niveles de permiso:**
- `read` → solo lectura (Glob, Grep, LS, Read)
- `write` → lectura + escritura (+ Edit, Write)
- `execute` → todo lo anterior + Bash

## Secciones Requeridas del Cuerpo

```markdown
## Rol
Una línea de identidad — qué hace este agente.

## Lo que NO hago
Lista explícita con el agente que sí maneja cada caso. Sin esta sección, el routing es ambiguo.

## Entradas requeridas / Inputs esperados
Tabla: Campo | Requerido | Fallback si falta

## Flujo de trabajo
Pasos numerados con gates explícitos ("Si X → DETENER y reportar").

## Auto-QA antes de entregar (opcional pero recomendado)
Checks que el agente corre sobre su propio output antes de cerrar.

## Presupuesto de tokens (opcional)
Objetivo | Máximo | Máx tool calls — por tamaño de tarea.

## Salida / Output de cierre
Formato y límite de palabras del mensaje final.
```

## Detección de Anti-Patrones

| Anti-Patrón | Señal | Severidad | Corrección |
|---|---|---|---|
| Agente sin "Lo que NO hago" | No existe la sección en el body | warning | Agregar sección con referencias explícitas a otros agentes |
| model:high para tarea mecánica | model: high en agente de formateo/reporte/lint | warning | Downgrade a medium o low |
| permissionMode:execute sin uso real de Bash | Solo hace lectura pero tiene execute | warning | Downgrade a write o read |
| Dos agentes con mismo dominio | Descriptions con keywords idénticas sin cláusula diferenciadora | error | Agregar cláusula "a diferencia de X, este agente hace Y" |
| Agente que contiene procedimientos de otra skill | Repite pasos que ya viven en una skill que carga | warning | Reemplazar por referencia a la skill correspondiente |
| Agente sin skills cargadas con flujos complejos | Flujo de 5+ pasos sin ninguna skill en frontmatter | info | Evaluar extracción de procedimientos a skills reutilizables |

## Checklist de Calidad

- [ ] `name` coincide con el filename sin `.md`
- [ ] `description` permite routing correcto: incluye cuándo invocarlo y qué NO cubre
- [ ] `permissionMode` es el mínimo necesario para la tarea
- [ ] `model` es el tier más bajo que produce calidad suficiente
- [ ] Sección "Lo que NO hago" presente con referencias a otros agentes
- [ ] El agente no repite procedimientos que ya viven en sus skills cargadas
- [ ] El dominio es exclusivo: ningún otro agente cubre las mismas responsabilidades
- [ ] Si tiene handoff: formato consistente con el patrón del proyecto
