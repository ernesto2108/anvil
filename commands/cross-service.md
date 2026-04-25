---
name: cross-service
description: Orquestar un feature/cambio a través de múltiples repos de microservicios usando el pipeline completo de agentes
allowed-tools: Agent, Read, Glob, Grep, Bash, Edit, Write
---

# Desarrollo Cross-Service — Pipeline Multi-Repo

Cargar la skill `cross-service-dev` y seguir su workflow.

## Contexto a recopilar primero:

1. Leer `<vault>/04-architecture/service-map.yaml` del vault del proyecto
2. Si no existe, pedir al usuario que cree uno (referenciar el template en la skill service-map)

## Lo que dijo el usuario: $ARGUMENTS

Si no se proporcionaron argumentos, preguntar al usuario:
- ¿Qué cambio necesitás? (nuevo endpoint, modificar existente, eliminar, deprecar)
- ¿Qué servicios están involucrados?
- ¿Alguna restricción? (backwards compatibility, shared DB, deadline)

## Luego seguir las fases de la skill cross-service-dev:

1. **PM** — agente pm: clasificar operación, discovery, escribir prd.md
2. **Architect** — 1 agente, design.md consolidado
3. **Developer** — N agentes (1 por servicio, en paralelo cuando sea posible)
4. **DBA** — 0-1 agente (solo si se necesita migración)
5. **Tester** — N agentes (1 por servicio, en paralelo)
6. **QA** — 1 agente (diff combinado de todos los servicios)
7. **Documentar** — changelog, diagrama, actualizar service-map.yaml
8. **Reporter** — resumen final

Los gates aplican: veto del architect → STOP, QA < 7 → STOP.
