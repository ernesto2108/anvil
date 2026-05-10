# ADR 001 — Reglas inviolables del Líder: delegación de specs de agentes y aislamiento de sub-agentes

**Fecha:** 2026-05-10
**Run ID:** run-20260510T-rules1-2
**Estado:** Aceptado

## Contexto

Dos invariantes del sistema de IA estaban implícitas o aplicadas de forma inconsistente:

1. El Líder (`agents/leader.md`) no debe editar specs de agentes (`agents/*.md`, `skills/*/*.md`, `commands/*.md`, `pipelines/*.yaml`, hooks en `settings.json`, `CLAUDE.md` del proyecto). Toda edición de esos artefactos se delega al `agent-designer`. Esta regla aparecía en el spec del Líder pero no estaba reforzada como invariante explícita.
2. Los sub-agentes (architect, dba, designer, scanner, mkt-content, reviewer, pm, etc.) no hablan con el usuario. Solo conversan con el Líder. Varios specs tenían restos de "modo interactivo" o frases tipo "pregunta al usuario" que abrían un canal directo sub-agente ↔ usuario y rompían la responsabilidad única del Líder como punto de contacto.

## Decisión

- **Regla #1 (delegación obligatoria de specs de agentes):** reforzar en el bloque Rol del Líder y en la sección "Reglas inviolables" con clasificación explícita por path. Cualquier archivo bajo `agents/`, `skills/`, `commands/`, `pipelines/`, hooks en `settings.json`, o `CLAUDE.md` del proyecto → siempre delegar al `agent-designer`, sin importar el modo activo o lo "trivial" que parezca la edición.
- **Regla #2 (aislamiento de sub-agentes):** agregar como regla inviolable explícita en `agents/leader.md`. Eliminar "modo interactivo" de los sub-agentes. Toda frase tipo "pregunta al usuario", "confirmar con el usuario" en sub-agentes se reescribe como "escalar al Líder" (el Líder decide si consulta al usuario aplicando el Protocolo de debate).

## Alternativas consideradas

- **Dejar el "modo interactivo" como bifurcación de fallback en sub-agentes:** rechazado. Genera dos modos de operación distintos, duplica lógica de presentación de outputs, y borronea la responsabilidad del Líder como única interfaz con el usuario.
- **Aplicar las reglas solo en el Líder sin tocar los sub-agentes:** rechazado. Mientras los sub-agentes tengan instrucciones de hablar con el usuario, hay riesgo de violación cuando uno se invoque por accidente fuera del flujo del Líder.

## Consecuencias

- `agents/leader.md` queda como única fuente de las dos invariantes.
- Sub-agentes son consumibles solo a través del Líder; no son utilizables de forma standalone.
- `agent-designer` es el único autorizado a editar specs del sistema de IA — un solo punto de cambio para evoluciones futuras del meta-sistema.
- Cualquier nuevo sub-agente debe nacer ya sin canal directo al usuario.

## Archivos modificados

- `agents/leader.md` — refuerzo Regla #1 + nueva Regla inviolable #8
- `agents/architect.md` — 7 ediciones (líneas 214, 319, 332-339, 547, 555, 567-572, 574-579)
- `agents/dba.md` — 2 ediciones (líneas 66, 92)
- `agents/designer.md` — 3 ediciones (líneas 104-112 eliminadas, 129, 141)
- `agents/scanner.md` — 3 ediciones (líneas 22, 27, 28)
- `agents/mkt-content.md` — 2 ediciones (líneas 153, 220)
- `agents/reviewer.md` — 4 ediciones (líneas 22, 25, 29, 30)
- `agents/pm.md` — refuerzo en líneas 48-52

## Verificación

- Lectura cruzada de todos los `agents/*.md` afectados confirma ausencia de "modo interactivo" y de frases "pregunta al usuario" en sub-agentes.
- `agents/leader.md` declara explícitamente la clasificación de paths bajo Regla #1 y el aislamiento bajo Regla #8.
