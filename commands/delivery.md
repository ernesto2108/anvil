---
name: delivery
description: Inicia o retoma una entrega trazable con Linear, reporter, validaciones y PR estructurado.
tools: Agent, Read, Glob, Grep, Bash, Edit, Write
---

# Delivery

Cargar la skill `delivery-flow` y seguir el ciclo completo.

Entradas aceptadas:

```text
/delivery <plan|feat|fix|hotfix|refactor|chore> <descripción> [TASK-ID]
```

1. Resolver la configuración en `.project-context/` y reutilizar el estado de entrega si existe.
2. Si hay tracking requerido y no hay tarea, redactar el issue y pedir confirmación antes de crearlo.
3. Persistir el estado antes de implementar y pasar su path al ejecutor.
4. Antes del cierre, exigir reporter, evidencia de validación, PR estructurado y sincronización de Linear.
5. Aceptar `--no-tracking "motivo"` solo para trabajo explícitamente local/experimental; registrarlo en el estado.
