---
name: spec-writer
description: Provee el template accionable y las reglas de formato para producir `spec.md` a partir de Architecture Views + ADRs + requirements (modo normal) o de un brief técnico inline (modo liviano). Usar siempre al inicio de la invocación del agente `spec-writer`, antes de leer inputs o emitir el spec.
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

- **Siempre** al inicio de la invocación del agente `spec-writer`, antes del Paso 0 — Pre-flight.
- Antes de emitir cualquier `spec.md`, en cualquiera de los dos modos (`normal` o `liviano`).

## Guides por dominio

Cargar el guide correspondiente antes de materializar el `spec.md`. El guide contiene el template detallado con todas las secciones, ejemplos de filas y reglas específicas que el agente debe respetar.

| Guide | Cuándo cargar | Qué provee |
|---|---|---|
| `guides/spec.md` | **siempre** al inicio de la invocación — es el template canónico de `spec.md` | Estructura completa de secciones, tabla "ES / NO ES", template markdown del documento, reglas duras sobre ubicación justificada, utils a reutilizar, ACs testeables y tests por AC |

> El agente `spec-writer` define en su propio archivo (`agents/spec-writer.md`) las 12 secciones del modo normal y las 6 del modo liviano. Esta skill + `guides/spec.md` son la fuente de verdad del formato; cuando el agente y el guide se contradigan en una regla puntual, prevalece el agente (más reciente). Reportar la discrepancia al `agent-designer` para reconciliar.

## Reglas duras (extracto — el detalle vive en `guides/spec.md`)

1. **Cada archivo con acción `CREATE` debe tener "Ubicación: por qué aquí"** anclado en un archivo vecino existente o en el patrón del módulo. Sin esa columna llena → spec inválido.
2. **La sección "Utils a reutilizar" es obligatoria** si el spec propone cualquier helper, parser, formatter, validator o util nuevo. Justificar la ausencia de equivalentes existentes.
3. **Cada criterio de aceptación debe ser testeable tal cual** — formato `GIVEN / WHEN / THEN`, con marca `_Implementa: FR-N_` (modo normal) o `_Implementa: brief-N_` (modo liviano).
4. **Tabla "Tests por criterio de aceptación" sin filas vacías.** Una fila por AC declarado, sin excepción.
5. **Si una sección no aplica, mantener el header con `_No aplica._`** — no eliminar headers; el developer cuenta con el orden fijo.
6. **El spec NO duplica contratos** de las Architecture Views — los referencia.

## Checklist de validación (antes de cerrar el `spec.md`)

- [ ] Todas las secciones obligatorias del modo activo presentes y en orden
- [ ] Cada AC tiene su marca `_Implementa: ..._`
- [ ] Cada archivo `CREATE` del Mapa de implementación tiene justificación de ubicación
- [ ] "Utils a reutilizar" completa si el spec introduce helpers nuevos
- [ ] Tabla "Tests por criterio de aceptación" cubre todos los ACs declarados
- [ ] `## 2. No-objetivos` (modo normal) o `## 2. Alcance` (modo liviano) NO está vacía
- [ ] `## Design References` presente cuando la tarea toca UI
