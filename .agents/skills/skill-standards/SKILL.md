---
name: skill-standards
description: Estándares y checklist para crear o modificar skills. Úsalo cuando el usuario diga "create a skill", "nueva skill", "plantilla de skill", "checklist de skill", "modificar skill existente", o cuando el agent-designer vaya a escribir o revisar un SKILL.md.
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


# Estándares de Creación de Skills

Estos son los estándares obligatorios para cada skill en este proyecto. Basados en el estándar abierto Agent Skills (agentskills.io), las mejores prácticas de Anthropic, y lecciones aprendidas de nuestra propia iteración.

## Filosofía

1. **Procedimiento sobre identidad** — una skill enseña cómo hacer, no quién eres. Si el contenido dice "Eres el especialista en X", es un agente disfrazado.
2. **Divulgación progresiva** — el SKILL.md contiene lo esencial (< 500 líneas); los detalles van en archivos de referencia que se cargan bajo demanda.
3. **Reutilización como criterio de existencia** — una skill justifica su existencia solo si la consumen 2+ agentes o si es un guardrail del sistema. Si la usa solo un agente, la lógica va en el spec de ese agente.

## Checklist Pre-Creación

Antes de escribir una nueva skill, verificar:

- [ ] Ninguna skill existente ya cubre este caso de uso (revisar el directorio de skills del proyecto)
- [ ] La skill tiene una única responsabilidad clara
- [ ] Sabes si debe ser auto-invocable, solo-usuario, o solo-sistema

## Estructura de SKILL.md

Cada skill DEBE tener este frontmatter:

```yaml
---
name: skill-name              # REQUERIDO. Minúsculas, guiones, coincide con el nombre del directorio
description: Qué hace. Úsalo cuando el usuario diga "keyword1", "keyword2", "keyword3", o [contexto]. # REQUERIDO. Máx 1024 chars
user-invocable: false          # Solo si es guardrail/comportamental (el sistema lo carga automáticamente)
disable-model-invocation: true # Solo si es pesado/manual (el usuario lo invoca con /nombre)
---
```

### Reglas para el Nombre
- Solo letras minúsculas, números, guiones
- Máx 64 caracteres
- Debe coincidir exactamente con el nombre del directorio padre
- Sin guiones al inicio/final o consecutivos
- Se prefiere la forma gerundio: `processing-pdfs`, no `pdf-processor`

### Reglas para la Descripción (Crítico — controla la activación)
- Formato: `Qué hace. Úsalo cuando [condiciones de activación].`
- Incluir 3-5 palabras clave/frases específicas que los usuarios dirían naturalmente
- Incluir extensiones de archivo cuando sea relevante (`.go`, `.tsx`, `.dart`)
- Ser "proactivo" — listar contextos explícitamente, errar por el lado de sobre-activarse
- Máx 1024 caracteres
- Tercera persona, redacción imperativa

### Control de Invocación

| Modo | Frontmatter | Cuándo usar |
|---|---|---|
| **Auto** (por defecto) | ninguno | Skills que el sistema debe cargar durante el trabajo normal (lint, convenciones, guardrails) |
| **Solo-usuario** | `disable-model-invocation: true` | Operaciones pesadas solo bajo petición explícita (orchestrate, scan, e2e) |
| **Solo-sistema** | `user-invocable: false` | Guardrails pasivos y guías de comportamiento (boundary-guardrails, entity-guardrails) |

## Estándares de Contenido del Cuerpo

### Secciones Requeridas (skills de convenciones/referencia)

```markdown
## Filosofía
- 3 principios fundamentales que guían las decisiones (no reglas, principios)

## Flujo de Trabajo
1. Pasos procedimentales numerados
2. Incluir puertas: "Si X — detener y [acción]"
3. Incluir puntos de confirmación del usuario: "Preguntar al usuario: [pregunta]"

## Reglas
- Reglas concretas y accionables agrupadas por preocupación

## Checklist Pre-Implementación
- [ ] Ítems verificables antes de comenzar el trabajo

## Detección de Anti-Patrones (para skills de convenciones)
- Tabla con: Patrón de Código | Anti-Patrón | Severidad | Categoría | Corrección
- Detección pasiva: error + warning siempre, suggestion solo en "improve/refactor"
- Formato de reporte: [file:line] [severity] [category] anti-pattern-name
```

### Secciones Requeridas (skills operacionales/de tarea)

```markdown
## Flujo de Trabajo
1. Paso de detección/configuración
2. Paso de ejecución
3. Puertas de decisión con ramificación clara
4. Paso de reporte con formato de salida definido

## Formato de Salida
- Plantilla exacta de la salida esperada
```

### Principios de Contenido

1. **Filosofía antes que reglas** — enunciar el principio, luego la regla
2. **Procedimientos sobre declaraciones** — enseñar cómo abordar, no qué producir
3. **Explicar el porqué** — "porque X causa Y" es mejor que "DEBE hacer Z"
4. **Puertas sobre instrucciones** — "Si >200 líneas, detener y dividir" es mejor que "mantener archivos pequeños"
5. **Ejemplos sobre explicaciones** — un ejemplo de código de 50 tokens es mejor que 150 tokens de prosa
6. **Divulgación progresiva** — SKILL.md < 500 líneas, detalles en archivos de referencia
7. **Sin referencias específicas del proyecto** — sin rutas hardcodeadas, nombres de proyecto, o términos de dominio

### Niveles de Severidad de Anti-Patrones

| Nivel | Significado | Cuándo reportar |
|---|---|---|
| `error` | Causa bugs, crashes, pérdida de datos | Siempre (detección pasiva) |
| `warning` | Problemas de rendimiento, mantenibilidad, diseño | Siempre (detección pasiva) |
| `suggestion` | Estilo de código, mejoras menores | Solo en "improve/refactor/optimize" (detección activa) |

## Estructura de Directorios

```
skill-name/
├── SKILL.md                    # Requerido, < 500 líneas
├── reference.md                # Opcional, docs detalladas (cargadas bajo demanda)
├── anti-patterns.md            # Opcional, tabla de referencia de detección
├── examples/
│   ├── good-patterns.md        # Opcional, ejemplos idiomáticos
│   └── bad-patterns.md        # Opcional, anti-patrones con correcciones
├── evals/
│   ├── evals.json              # Opcional, casos de prueba
│   └── files/                  # Opcional, archivos de ejemplo para evals
└── scripts/
    └── helper.sh               # Opcional, utilidades ejecutables
```

### Reglas para Archivos de Referencia
- Mantener referencias a un nivel de profundidad (sin cadenas A.md → B.md → C.md)
- Para archivos >300 líneas, incluir tabla de contenidos
- Referenciar claramente desde SKILL.md: "Ver `reference.md` para la guía completa"

## Evals (Recomendado para skills de convenciones)

Crear `evals/evals.json` con:
- 4 pruebas de activación (2 que deben activarse, 2 que no deben activarse)
- 3-5 pruebas de calidad con afirmaciones y archivos de ejemplo

```json
{
  "skill_name": "skill-name",
  "evals": [
    {
      "id": 1,
      "type": "trigger",
      "prompt": "Prompt realista del usuario",
      "should_trigger": true
    },
    {
      "id": 2,
      "type": "quality",
      "prompt": "Prompt de tarea",
      "files": ["evals/files/example.ext"],
      "assertions": [
        "Afirmación específica y verificable sobre la salida"
      ]
    }
  ]
}
```

## Compatibilidad Cross-Agent

Las skills viven en `~/.claude/skills/` con un symlink en `~/.agents/skills/` para descubrimiento cross-agent (Cursor, Codex, Gemini CLI, etc.). No se necesita trabajo extra — el symlink lo maneja.

## Detección de Anti-Patrones

| Anti-Patrón | Señal | Severidad | Corrección |
|---|---|---|---|
| Skill con routing logic | Contiene "derivar a", "invocar a", "escalar a" + nombre de agente | error | Mover lógica al agente que la consume |
| Skill con lenguaje de rol | Contiene "Eres el", "Tu rol es", "Actúas como" | error | Reescribir como instrucción procedimental |
| SKILL.md > 500 líneas | `wc -l SKILL.md` > 500 | warning | Extraer detalle a `reference.md` dentro del directorio |
| Descripción sin "Úsalo cuando" | No contiene la frase "Úsalo cuando" | warning | Agregar condiciones de activación con palabras clave |
| Flag contradictorio | `disable-model-invocation: true` en una skill que un agente carga en su flujo | error | Eliminar el flag o cambiar a `user-invocable: false` según el caso |
| Path hardcodeado | Rutas absolutas con `~`, `/home/`, `/Users/` en el contenido | warning | Usar referencias relativas o genéricas |
| Falta sección Filosofía (skills de convenciones) | No existe `## Filosofía` en el cuerpo | warning | Agregar 3 principios que guíen las decisiones |

## Checklist de Calidad (ejecutar después de crear una skill)

- [ ] `name` coincide con el nombre del directorio
- [ ] `description` incluye "Úsalo cuando" con 3-5 palabras clave de activación
- [ ] Modo de invocación configurado correctamente (auto / solo-usuario / solo-sistema)
- [ ] SKILL.md < 500 líneas
- [ ] La sección de Filosofía enuncia principios, no solo reglas
- [ ] La sección de Flujo de Trabajo tiene pasos numerados con puertas
- [ ] Sin referencias específicas del proyecto (rutas, términos de dominio, nombres de proyecto)
- [ ] La tabla de anti-patrones tiene niveles de severidad (para skills de convenciones)
- [ ] Existe checklist pre-implementación (para skills de convenciones)
- [ ] Los archivos de referencia están a un nivel de profundidad
- [ ] Evals creados (para skills de convenciones)
