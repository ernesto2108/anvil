# Update — Actualización incremental de `.project-context/`

Usado por el reporter (pipeline) y directamente por Claude (modo directo) después de cada implementación.

## Principio

**Delta, no sobrescritura.** Modificar solo las secciones que cambiaron. Nunca reemplazar un archivo completo a menos que sea imposible hacer un delta coherente.

## Detectar qué actualizar

A partir del diff de la implementación, mapear archivos cambiados a secciones de `.project-context/`:

| Si se tocaron archivos en... | Actualizar... |
|---|---|
| `internal/<domain>/` | `Technical domain/domain.md` — sección afectada |
| Handlers HTTP / routes | `Technical domain/contracts.md` — sección REST API |
| Queue producers/consumers | `Technical domain/contracts.md` — sección Message Queues |
| Clientes externos (http, grpc) | `Technical domain/contracts.md` — sección Servicios externos |
| Nuevas interfaces o estructuras con patrón claro | `Core/coding-standards.md` — agregar o actualizar entrada |
| Decisión arquitectónica documentada en SPEC | `decisions/NNN-slug.md` — crear si no existe |
| Archivos > 300 líneas introducidos | `Technical domain/risks.md` — agregar nota de deuda potencial |
| Cambio en validaciones de dominio o reglas cross-entidad | `Technical domain/contracts.md` — sección de invariantes |
| Nuevo dominio, o cambio en relaciones entre dominios | `Technical domain/dependencies.md` — actualizar el grafo |
| `Makefile`, `docker-compose.*`, `package.json` scripts, `scripts/` | `Core/workflows.md` — actualizar el target o comando que cambió |
| `agents/*.md` | `Technical domain/domain.md` — sección Agentes (si el proyecto es Anvil) |
| `skills/*/SKILL.md` | `Technical domain/domain.md` — sección Skills |
| `pipelines/*.yaml` | `Technical domain/domain.md` — sección Pipelines |
| `commands/*.md` | `Technical domain/domain.md` — sección Commands |
| Cualquier cambio | `NAVIGATOR.md` — actualizar `last_updated` |

## Formato de actualización

### Agregar entrada nueva a `Core/coding-standards.md`

```markdown
## <NombreInferido> — <archivo principal>
- Archivo: `<path>:<line>`
- Qué hace: <una línea>
- Cuándo usar: <contexto>
- Anti-pattern: <qué evitar>
```

### Agregar endpoint a `Technical domain/contracts.md`

```markdown
### <METHOD> <path>
- Handler: `<path>:<func>`
- Auth: <tipo o "ninguna">
- Request: `<tipo>` (campos clave)
- Response: `<tipo>` (campos clave)
```

### Actualizar dominio

No reemplazar el archivo. Editar solo la sección que cambió:
- Si el flujo principal cambió → reescribir solo `## Flujo principal`
- Si se agregó un patrón → agregar línea a `## Patrones usados`
- Si apareció un gotcha → agregar a `## Gotchas`
- Si se tomó una decisión → agregar línea a `## Decisiones tomadas`

### Crear decision ADR-lite

Solo cuando hay evidencia explícita de una decisión arquitectónica:
- Documentada en el SPEC del architect
- Mencionada en el handoff con D1/D2/... prefix
- El usuario la mencionó explícitamente

Numeración: siguiente número disponible en `decisions/`. Formato: `NNN-slug-corto.md`.

## Qué NO actualizar

- No agregar patrones que ya estén documentados — verificar antes de escribir
- No crear decisiones que no estén respaldadas por evidencia
- No agregar contratos que ya existan — verificar la sección antes de agregar
- No actualizar `project.md` a menos que el stack o la arquitectura cambió fundamentalmente

## Flujo en modo directo

Cuando Claude implementa directamente (sin agentes), al finalizar la implementación:

1. Revisar qué archivos se modificaron en la sesión
2. Aplicar la tabla de mapeo de arriba
3. Escribir deltas — máximo 3-5 ediciones a `.project-context/`
4. Actualizar `last_updated` en NAVIGATOR.md

No pedir confirmación al usuario para los deltas — es una operación de mantenimiento transparente.

## Flujo en modo pipeline

El reporter recibe el diff inline del orquestador y aplica los mismos pasos. El orquestador debe incluir en el brief del reporter:
```
## Delta para .project-context/
Archivos cambiados: [lista]
Nuevos patrones detectados: [si aplica]
Nuevos contratos: [si aplica]
Decisiones documentadas en SPEC: [si aplica]
```
