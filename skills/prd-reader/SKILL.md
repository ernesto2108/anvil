---
name: prd-reader
description: Normaliza cualquier tipo de input de entrada (PRD informal, documento libre, path local, URL, texto libre) en un resumen estructurado que el architect presenta al usuario en el Paso 1 de su flujo. Extrae objetivo, stack, dominio, integraciones, restricciones, milestone y gaps sin inventar información. Usar al inicio del flujo del architect (Paso 1 — Resumir contexto).
---

# PRD Reader — normalización de inputs para el architect

> **Nota de contexto:** Esta skill se usa en el Paso 1 del flujo del architect para normalizar cualquier tipo de input en un resumen estructurado antes de presentarlo al usuario.

## Propósito

Tomar cualquier input de entrada que reciba el architect — un PRD formal, un documento libre, un path local a un archivo, una URL externa, o un brief informal en texto — y producir un **resumen estructurado** con los campos clave que el architect necesita para razonar.

El output de esta skill es lo que el architect presenta al usuario al final del Paso 1, antes de la pausa obligatoria de validación.

## Reglas de extracción

- **Si el input es una URL** → resumir el contenido leído antes de extraer campos. (El architect no ejecuta `WebFetch` directamente; si la URL aún no fue resumida, escalar al humano para que invoque al `explorer`.)
- **Si el input es un path local** → leer el archivo antes de extraer campos.
- **Si el input es texto libre** → extraer directamente.
- **Si hay múltiples inputs** (ej. PRD + ADR previo + brief inline) → combinar los hallazgos en un solo resumen, anotando la fuente cuando la información provenga de un input específico.
- **No inventar información que no esté en el input.** Si un campo no aparece, marcarlo como `no claro` y agregarlo a la lista de gaps.
- **No preguntar en esta etapa.** La skill solo extrae y marca gaps. Las preguntas para cubrir gaps van en el Paso 2 del flujo del architect.

## Campos a extraer (siempre los siete)

Para cada campo, registrar el valor encontrado o `no claro` si no aparece en el input.

| Campo | Descripción | Si falta |
|---|---|---|
| **Objetivo del feature** | Una línea que describa qué se quiere lograr. | `no claro` + gap |
| **Stack inferido** | Lenguajes, frameworks, DBs mencionados explícitamente o inferidos por contexto. | `no claro` + gap |
| **Dominio** | `backend` / `frontend` / `mobile` / `fullstack` / `no claro`. | `no claro` + gap |
| **Integraciones detectadas** | APIs externas, servicios internos, brokers, terceros mencionados. | `ninguna detectada` (no es necesariamente un gap) |
| **Restricciones conocidas** | NFRs (latencia, SLO, throughput), limitaciones técnicas, decisiones ya tomadas que no se deben pisar. | `ninguna detectada` |
| **Milestone o fecha objetivo** | Versión, sprint, fecha. | `no mencionado` |
| **Lo que NO quedó claro** | Lista explícita de gaps que el Paso 2 deberá resolver. | nunca vacío salvo que todos los campos anteriores estén completos |

## Formato de output

El architect imprime exactamente este bloque al cierre del Paso 1, antes de la pausa obligatoria:

```markdown
## Contexto capturado

**Objetivo:** <una línea o "no claro">
**Stack:** <tecnologías detectadas o "no claro">
**Dominio:** <backend / frontend / mobile / fullstack / no claro>
**Integraciones:** <lista o "ninguna detectada">
**Restricciones conocidas:** <lista o "ninguna detectada">
**Milestone:** <fecha/versión o "no mencionado">

**Gaps (no quedó claro):**
- <gap 1>
- <gap 2>
```

Si no hay gaps, escribir literal `Gaps (no quedó claro): ninguno`.

## Reglas duras

1. **Cero invenciones.** Si el input no menciona el stack, NO inferirlo de "lo más común" — marcarlo `no claro`.
2. **Cero preguntas en esta skill.** El output es siempre el bloque estructurado. Las preguntas vienen en el Paso 2 del architect.
3. **Resumen, no copia.** El objetivo y las restricciones se condensan a una línea cada una — no pegar párrafos del PRD.
4. **Fuente trazable cuando hay ambigüedad.** Si dos inputs se contradicen (ej. el PRD dice Postgres, el brief inline dice MySQL), anotarlo como gap en lugar de elegir.
5. **URLs no resumidas → escalar.** Si llega una URL sin contenido leído, el architect debe escalar al humano para invocar `explorer`; esta skill no fabrica el resumen.

## Checklist de validación

- [ ] Los siete campos están presentes en el output (aunque sea con `no claro` / `ninguna detectada` / `no mencionado`)
- [ ] La lista de gaps refleja literalmente los campos marcados `no claro`
- [ ] Ningún valor fue inventado más allá de lo que el input dice
- [ ] El objetivo cabe en una línea
- [ ] El output está listo para que el architect lo pegue antes de la pausa del Paso 1
