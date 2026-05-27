---
name: cross-service-dev
description: Entry point explícito para runs multi-repo/cross-service. Señala al humano que el cambio atraviesa varios microservicios y debe orquestarse con el pipeline cross-service completo.
allowed-tools:
---

# Cross-Service Dev — Entry Point Multi-Repo

Este command es un disparador delgado. **No orquesta agentes**: toda la orquestación la ejecuta el humano, cargando la skill `cross-service-dev` que define el pipeline (PM → Architect → Developer x N → DBA → Tester → QA → Reporter) y sus gates.

## Brief del usuario

$ARGUMENTS

## Señal al humano

Este run es **multi-repo / cross-service**. El humano debe:

1. Tratar la tarea como cross-service desde el Paso L0 (snapshot git de cada repo involucrado, Navigator por repo cuando exista).
2. Cargar la skill `cross-service-dev` y seguir su workflow extremo a extremo.
3. Resolver `service-map.yaml` desde el vault del proyecto; si no existe, pedirlo al usuario antes de continuar.
4. Aplicar los gates de la skill (veto del architect → STOP, QA < 7 → STOP) al cerrar.

Si el brief del usuario está vacío o ambiguo respecto a servicios involucrados, tipo de cambio o restricciones, el humano ejecuta el Paso 0 de clarificación antes de spawnear cualquier agente.
