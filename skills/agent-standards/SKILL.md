---
name: agent-standards
description: Estándares y checklist para crear o modificar agentes. Úsalo cuando el usuario diga "nuevo agente", "crear agente", "modificar agente", "plantilla de agente", "checklist de agente", o cuando el agent-designer vaya a escribir o revisar un agents/*.md.
---

# Estándares de Creación de Agentes

Estos son los estándares obligatorios para cada agente en este proyecto.

## Filosofía

1. **Rol sobre procedimiento** — un agente define quién es y qué no es suyo; los pasos detallados van en skills que carga.
2. **Dominio exclusivo sin solapamiento** — cada agente posee un conjunto de archivos/responsabilidades que ningún otro toca. Sin solapamiento = sin ambigüedad de routing.
3. **Mínimo permiso** — usar el nivel más bajo de `permissionMode` que la tarea permita. Escalar solo cuando hay evidencia de necesidad. Los agentes no declaran modelo: cada herramienta usa el modelo de su sesión activa.

## Checklist Pre-Creación

Antes de escribir un nuevo agente, verificar:

- [ ] Ningún agente existente ya cubre este dominio (revisar descriptions de agents/*.md)
- [ ] El nuevo agente tiene dominio exclusivo claro (archivos/responsabilidades que solo él toca)
- [ ] Se justifica como agente y no como skill (ver criterios de decisión abajo)
- [ ] El `permissionMode` está justificado

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
skills:                       # opcional — skills que se cargan al invocar este agente
  - skill-name
---
```

> **`model` es un campo obsoleto: NO se declara.** El mecanismo de tiers (`low|medium|high`) fue eliminado del sistema. Cada herramienta ejecuta el agente con el modelo de su sesión activa. Si un agente todavía trae una línea `model:`, es un residuo y debe borrarse.

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

## Límites de alcance (opcional)
Restricciones de scope por tamaño de tarea: máximo de archivos a escribir, modos de operación (scoped/full), profundidad de lectura. Sin cifras de tokens ni de tool calls — los modelos no pueden contarlos.

## Salida / Output de cierre
Formato y límite de palabras del mensaje final.
```

## Detección de Anti-Patrones

| Anti-Patrón | Señal | Severidad | Corrección |
|---|---|---|---|
| Agente sin "Lo que NO hago" | No existe la sección en el body | warning | Agregar sección con referencias explícitas a otros agentes |
| Campo `model` en el frontmatter | Existe una línea `model:` en el frontmatter del agente | warning | Eliminar la línea `model:` — el campo es obsoleto |
| permissionMode:execute sin uso real de Bash | Solo hace lectura pero tiene execute | warning | Downgrade a write o read |
| Dos agentes con mismo dominio | Descriptions con keywords idénticas sin cláusula diferenciadora | error | Agregar cláusula "a diferencia de X, este agente hace Y" |
| Agente que contiene procedimientos de otra skill | Repite pasos que ya viven en una skill que carga | warning | Reemplazar por referencia a la skill correspondiente |
| Agente sin skills cargadas con flujos complejos | Flujo de 5+ pasos sin ninguna skill en frontmatter | info | Evaluar extracción de procedimientos a skills reutilizables |

## Checklist de Calidad

- [ ] `name` coincide con el filename sin `.md`
- [ ] `description` permite routing correcto: incluye cuándo invocarlo y qué NO cubre
- [ ] `permissionMode` es el mínimo necesario para la tarea
- [ ] El frontmatter no declara `model` (campo obsoleto)
- [ ] Sección "Lo que NO hago" presente con referencias a otros agentes
- [ ] El agente no repite procedimientos que ya viven en sus skills cargadas
- [ ] El dominio es exclusivo: ningún otro agente cubre las mismas responsabilidades
- [ ] Si tiene handoff: formato consistente con el patrón del proyecto
