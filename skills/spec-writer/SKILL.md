---
name: spec-writer
description: Provee el template canónico y reglas de formato para producir spec.md. Úsalo cuando el agente spec-writer genera un spec, cuando se materializa un spec.md, o cuando se valida la estructura de un spec.md existente.
---

# Spec Writer — Template accionable de `spec.md`

## Filosofía

El `spec.md` es el contrato self-contained que el developer consume sin re-leer PRD, ADRs ni `requirements.md`. Esta skill provee el **template canónico** (qué secciones, en qué orden, con qué reglas) que el agente `spec-writer` debe materializar. La skill NO toma decisiones técnicas ni reemplaza las validaciones de cobertura definidas en el agente — solo asegura que el artefacto emitido tenga la estructura correcta.

| ES | NO ES |
|---|---|
| Template estructural + reglas de formato | Lógica de decisión arquitectónica |
| Distinción ES / NO ES de cada sección | Duplicado de contratos de las Architecture Views |
| Reglas duras sobre justificación de ubicación, utils a reutilizar, ACs testeables | Código, SQL o instrucciones de implementación |

## Cuándo cargar

- Al inicio del **Paso 4** del agente `spec-writer` — antes de leer ningún input confirmado. Cargarlo antes garantiza que el template y las condiciones de inclusión estén disponibles al momento de construir el spec.
- Antes de emitir cualquier `spec.md`.

## Guides por dominio

Cargar el guide correspondiente antes de materializar el `spec.md`. El guide contiene el template detallado con todas las secciones, ejemplos de filas y reglas específicas que el agente debe respetar.

| Guide | Cuándo cargar | Qué provee |
|---|---|---|
| `guides/spec.md` | **siempre** al inicio de la invocación — es el template canónico de `spec.md` | Estructura completa de secciones, tabla "ES / NO ES", template markdown del documento, reglas duras sobre ubicación justificada, utils a reutilizar, ACs testeables y tests por AC |

> El agente `spec-writer` define en su propio archivo (`agents/spec-writer.md`) las secciones del spec y las condiciones de inclusión de cada una. Esta skill + `guides/spec.md` son la fuente de verdad del formato; cuando el agente y el guide se contradigan en una regla puntual, prevalece el agente (más reciente). Reportar la discrepancia al `agent-designer` para reconciliar.

## Reglas duras (extracto — el detalle vive en `guides/spec.md`)

1. **Cada archivo con acción `CREATE` debe tener "Ubicación: por qué aquí"** anclado en un archivo vecino existente o en el patrón del módulo. Sin esa columna llena → spec inválido.
2. **Cada criterio de aceptación debe ser testeable tal cual** — formato `GIVEN / WHEN / THEN`, con marca `_Implementa: FR-N_` (si vino de `requirements.md`) o `_Implementa: brief-N_` (si vino del brief inline).
3. **Tabla "Tests por criterio de aceptación" sin filas vacías.** Una fila por AC declarado, sin excepción.
4. **Si una sección no aplica, mantener el header con `_No aplica._`** — no eliminar headers; el developer cuenta con el orden fijo.
5. **El spec NO duplica contratos** de las Architecture Views — los referencia.

## Checklist de validación (antes de cerrar el `spec.md`)

- [ ] Secciones aplicables al contexto disponible presentes y en orden
- [ ] Cada AC tiene su marca `_Implementa: FR-N_` (si vino de requirements) o `_Implementa: brief-N_` (si vino de brief inline)
- [ ] Cada archivo `CREATE` del Mapa de implementación tiene justificación de ubicación
- [ ] Tabla "Tests por criterio de aceptación" cubre todos los ACs declarados
- [ ] `## No-objetivos` NO está vacía
- [ ] `## Design References` presente cuando la tarea toca UI
