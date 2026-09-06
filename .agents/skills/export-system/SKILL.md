---
name: export-system
description: Receta determinista para exportar un sistema de IA basado en Claude Code (agents/*.md, skills/*/SKILL.md, commands/*.md, CLAUDE.md) a los formatos nativos de OpenCode, Kimi Code, Kiro y Codex CLI. Traduce frontmatter de agentes campo a campo, copia skills como estándar Agent Skills, convierte commands al equivalente de cada herramienta y genera AGENTS.md desde CLAUDE.md. Úsalo cuando el usuario diga "exportar el sistema", "exportar agentes a opencode", "portar a kiro", "generar AGENTS.md", "usar mis agentes en codex", "convertir a kimi code", o cuando pida que el mismo sistema de agentes funcione fuera de Claude Code. Pausa con confirmación antes de escribir.
disable-model-invocation: true
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


# Export System — Portar el sistema de IA a otras herramientas de coding

> Receta determinista de conversión. Lee la fuente de verdad (`agents/`, `skills/`, `commands/`, `CLAUDE.md`), traduce cada artefacto al formato nativo de la herramienta destino y escribe los exports. Un único paso de confirmación antes de escribir. Sin scripts: la conversión se hace leyendo y escribiendo archivos.

## Filosofía

1. **Una sola fuente de verdad** — `agents/`, `skills/`, `commands/` y `CLAUDE.md` son el original. Todo lo que esta receta genera es derivado y descartable: si el export y la fuente divergen, gana la fuente y se regenera el export. Editar un export a mano crea una segunda verdad que se pierde en la siguiente ejecución.
2. **Traducción explícita, nunca creativa** — cada campo del frontmatter origen tiene un destino documentado en una tabla. Si un campo no tiene equivalente, se degrada de forma declarada (se mueve al cuerpo, o se omite y se reporta), nunca se inventa un campo que la herramienta destino no soporte.
3. **Pérdida declarada antes que pérdida silenciosa** — ninguna herramienta destino soporta el 100% del modelo de Claude Code. Lo que se pierde en cada conversión se lista en el reporte final, para que quien use el export sepa qué garantías ya no tiene.

## Parámetros de entrada

| Parámetro | Requerido | Valores | Default | Notas |
|---|---|---|---|---|
| `targets` | No | `opencode`, `kimi`, `kiro`, `codex`, `all` | preguntar | Si el usuario no lo dice, preguntar antes del Paso 1 |
| `scope` | No | `all`, `agents`, `skills`, `commands`, `instructions` | `all` | Qué familia de artefactos exportar |
| `agents` | No | lista de nombres | todos | Subconjunto de agentes a exportar |

Si `targets` no se puede inferir del prompt → preguntar en una línea: **¿A qué herramienta(s) exporto: opencode, kimi, kiro, codex o todas?** No asumir `all`.

## Flujo de trabajo

### Paso 1 — Inventariar la fuente

1. Listar `agents/*.md` (excluir cualquier archivo bajo `.handoff/`).
2. Listar `skills/*/SKILL.md`.
3. Listar `commands/**/*.md` (excluir `INDEX.md` y archivos que sean índices sin frontmatter).
4. Verificar si existe `CLAUDE.md` en la raíz del proyecto.

Leer el frontmatter de cada agente y de cada skill. **No leer el body completo de todos los agentes en este paso** — el body se lee en el Paso 4, agente por agente, al momento de escribir su export.

Si `skills/` es un symlink, resolverlo y trabajar con el contenido real; escribir siempre a través del path del repo, no del destino del symlink.

**Gate:** si no existe ni un solo artefacto exportable (0 agentes, 0 skills, 0 commands y sin `CLAUDE.md`) → DETENER y reportar que no hay nada que exportar.

### Paso 2 — Resolver el modelo de conversión por herramienta

Cargar el archivo de referencia de cada herramienta destino seleccionada. Solo los seleccionados:

| Target | Referencia |
|---|---|
| OpenCode | `references/opencode.md` |
| Kimi Code | `references/kimi.md` |
| Kiro | `references/kiro.md` |
| Codex CLI | `references/codex.md` |

Cada referencia contiene: rutas de destino, tabla de mapeo de frontmatter campo a campo, tratamiento de commands, tratamiento de instrucciones y la lista de pérdidas conocidas.

### Paso 3 — PAUSA: mostrar plan y pedir confirmación

Antes de escribir nada, mostrar el plan y **esperar confirmación explícita del usuario**:

```
Fuente: 24 agentes, 57 skills, 4 commands, CLAUDE.md (312 líneas)

Destinos y archivos a generar:
  opencode → .opencode/agents/*.md (24), .opencode/commands/*.md (4), AGENTS.md (1)
  codex    → .codex/agents/*.toml (24), .agents/skills/*/SKILL.md (61), AGENTS.md (1)

Total: 115 archivos.

⚠️ AGENTS.md ya existe → se sobrescribe (contenido derivado de CLAUDE.md).
⚠️ OpenCode dejará de leer CLAUDE.md una vez exista AGENTS.md.

¿Procedo?
```

Reglas del plan:

- Listar **conteos y rutas**, no el contenido de los archivos.
- Marcar con `⚠️` todo archivo existente que se vaya a sobrescribir.
- Marcar con `⚠️` toda pérdida de comportamiento relevante que la referencia declare.
- **Si el usuario no confirma → DETENER. No escribir nada.**

### Paso 4 — Escribir los exports

Escribir en este orden, herramienta por herramienta:

1. **Skills** — copia literal del `SKILL.md` a la ruta destino, con la carpeta contenedora conservando el nombre exacto de la skill. Antes de copiar, validar el `name` contra `^[a-z0-9]+(-[a-z0-9]+)*$` y que coincida con el nombre del directorio; si falla, omitir esa skill y anotarla en el reporte. Los campos extra de Claude Code (`allowed-tools`, `user-invocable`, `disable-model-invocation`) se copian tal cual: las otras herramientas los ignoran sin error.
2. **Agentes** — leer el body completo del agente, aplicar la tabla de mapeo de la referencia, aplicar la regla de skills del Paso 5 y escribir el archivo destino.
3. **Commands** — convertir según lo que indique la referencia de cada herramienta.
4. **Instrucciones** — generar `AGENTS.md` desde `CLAUDE.md` según el Paso 6.

### Paso 5 — Regla del campo `skills:` (aplica a las 4 herramientas)

Ninguna herramienta destino soporta el campo `skills:` del frontmatter. Ese campo se elimina del frontmatter exportado y su contenido se transporta al **inicio del body**, como bloque literal:

```markdown
## Skills requeridas

Antes de trabajar, carga estas skills y sigue sus instrucciones:
`nombre-skill-1`, `nombre-skill-2`, `nombre-skill-3`.
```

Si el agente ya tiene una instrucción de carga de skills en su body (frases del tipo "Carga la skill X al inicio"), el bloque se agrega igual: la redundancia es preferible a perder la referencia.

### Paso 6 — Generación de `AGENTS.md`

`AGENTS.md` se genera desde el `CLAUDE.md` del proyecto (nunca desde un `CLAUDE.md` global de usuario).

1. Copiar el contenido íntegro de `CLAUDE.md`.
2. Anteponer el bloque de encabezado del Paso 7.
3. Reescribir toda referencia a rutas de Claude Code (`.claude/…`) por la ruta equivalente del destino, según la referencia de la herramienta.
4. **Gate de tamaño (Codex):** si el `AGENTS.md` resultante supera ~32 KiB, DETENER antes de escribir y reportar el tamaño, indicando que Codex truncará el contenido combinado. Ofrecer al usuario dividir el contenido o reducirlo antes de continuar.

Un único `AGENTS.md` en la raíz sirve a las cuatro herramientas: no generar uno por herramienta.

### Paso 7 — Marcar los exports como derivados

Todo archivo generado lleva, inmediatamente después del frontmatter (o como primer comentario en los archivos TOML), esta marca:

```markdown
<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->
```

Además, recomendar al usuario en el reporte final que agregue los directorios de export a `.gitignore` si no quiere versionarlos. No modificar `.gitignore` como parte de esta receta.

### Paso 8 — Re-ejecución

Esta receta es idempotente por diseño: correrla de nuevo tras cambios en la fuente regenera los exports.

- Archivos generados previamente por esta receta (los que llevan la marca del Paso 7) → **sobrescribir sin preguntar**, ya declarados en el plan del Paso 3.
- Archivos en una ruta de destino **sin** la marca del Paso 7 → son manuales: preguntar antes de pisarlos, uno por uno.
- Artefactos que ya no existen en la fuente pero sí en el destino → listarlos como huérfanos en el reporte. **No borrarlos automáticamente**; ofrecer borrarlos.

## Reglas

- **Nunca escribir en `agents/`, `skills/`, `commands/` ni `CLAUDE.md`** durante una exportación. El flujo es estrictamente unidireccional: fuente → export.
- **Nunca emitir `override: true`** en ningún agente exportado (reemplazaría el system prompt principal de la herramienta destino).
- **Nunca inventar valores de `model`.** Los tiers `low|medium|high` no son identificadores válidos en ninguna herramienta destino: usar el default documentado en cada referencia y anotar en el reporte que el modelo requiere revisión manual.
- **Nunca traducir el body de un agente.** El cuerpo markdown se transporta literal, salvo el bloque de skills del Paso 5 y la marca del Paso 7.
- **No exportar `.handoff/*.md`** ni ningún archivo de estado de ejecución.
- Los directorios generados (`.opencode/`, `.kimi-code/`, `.kiro/`, `.codex/`, `.agents/`, `AGENTS.md`) **no son artefactos protegidos del sistema**: son exports derivados y pueden regenerarse en cualquier momento.
- Tras escribir, verificar con un listado del directorio que los archivos existen en el path del repo. Un `Write` sin verificación posterior sobre un path con symlinks no es prueba de escritura.

## Checklist antes de reportar

- [ ] Cada skill exportada tiene `name` válido y coincidente con su directorio
- [ ] Ningún agente exportado conserva el campo `skills:` en su frontmatter
- [ ] Cada agente exportado con skills tiene el bloque "Skills requeridas" en su body
- [ ] Ningún archivo exportado contiene `override: true`
- [ ] Todo archivo generado lleva la marca de "GENERADO / NO EDITAR"
- [ ] `AGENTS.md` está por debajo del límite de 32 KiB si Codex está entre los destinos
- [ ] Ningún archivo bajo `agents/`, `skills/`, `commands/` ni `CLAUDE.md` fue modificado
- [ ] Los archivos escritos fueron verificados con un listado del directorio destino

## Formato de salida

```markdown
## Exportación completada

| Destino | Agentes | Skills | Commands | Instrucciones |
|---|---|---|---|---|
| opencode | 24 | 57 (leídos in-situ) | 4 | AGENTS.md |
| codex | 24 | 61 | → skills | AGENTS.md |

### Rutas generadas
- `.opencode/agents/` — 24 archivos
- `.codex/agents/` — 24 archivos TOML
- `.agents/skills/` — 61 directorios
- `AGENTS.md` — 312 líneas (14 KiB)

### Pérdidas declaradas
- `model: medium|high` no se tradujo — revisar manualmente en cada herramienta
- Codex no soporta allowlist granular de tools — solo `sandbox_mode`
- Kiro: los commands se convirtieron a steering `inclusion: manual` (se invocan con `#nombre`)

### Omitidos
- `skills/Foo Bar/` — `name` inválido para el estándar Agent Skills

### Huérfanos detectados
- `.codex/agents/agente-viejo.toml` — ya no existe en la fuente (no borrado)

### Siguiente paso
Los exports son derivados. Tras cambiar `agents/`, `skills/` o `CLAUDE.md`, volver a ejecutar esta receta. No editar los archivos generados a mano.
```

## Detección de anti-patrones

| Anti-patrón | Señal | Severidad | Corrección |
|---|---|---|---|
| Edición manual de un export | Archivo con la marca "GENERADO" modificado respecto a lo que produce la receta | error | Trasladar el cambio a la fuente y regenerar |
| Export como fuente de verdad | Un agente existe en `.opencode/agents/` pero no en `agents/` | error | Crear el agente en la fuente y regenerar |
| Tier de modelo copiado literal | `model: high` en un archivo exportado | error | Usar el identificador de modelo del formato destino o el default de la referencia |
| Campo `skills:` en un export | Frontmatter destino con `skills:` | error | Mover al body según el Paso 5 |
| `AGENTS.md` divergente de `CLAUDE.md` | Secciones presentes en uno y ausentes en el otro | warning | Regenerar `AGENTS.md` desde `CLAUDE.md` |
| Escritura sin confirmación | Archivos creados sin haber mostrado el plan del Paso 3 | error | Detener y mostrar el plan antes de cualquier escritura |
| `override: true` emitido | Presente en un agente exportado a Kimi Code | error | Eliminar el campo |

## Referencias

- `references/opencode.md` — rutas, frontmatter y permisos de OpenCode
- `references/kimi.md` — compatibilidad Claude Code, skills como slash commands
- `references/kiro.md` — steering files, vocabulario de tools propio
- `references/codex.md` — agentes en TOML, límite de 32 KiB del AGENTS.md
