# Context Navigator — <ProjectName>

last_full_scan: <YYYY-MM-DD>
last_updated: <YYYY-MM-DD>
coverage: bootstrap

## Índice

### Core
- [Workflows](Core/workflows.md) — ramas, ambientes, deploy, comandos operativos
- [Task Management](Core/task-management.md) — gestión de tareas, tickets, definition of done
- [Coding Standards](Core/coding-standards.md) — naming, linting, patrones de diseño detectados

### Technical domain
- [Proyecto](Technical%20domain/project.md) — stack, arquitectura, restricciones, SOLID
- [Dominio](Technical%20domain/domain.md) — entidades principales y bounded contexts
- [Glosario](Technical%20domain/glossary.md) — lenguaje del negocio ↔ lenguaje técnico
- [Contratos](Technical%20domain/contracts.md) — APIs, queues, eventos, reglas de negocio
- [Dependencias](Technical%20domain/dependencies.md) — grafo de dependencias entre dominios
- [Riesgos](Technical%20domain/risks.md) — deuda técnica, gotchas, restricciones
- [Service Map](service-map.yaml) — mapa de relaciones entre servicios del ecosistema

### Decisiones arquitectónicas
<!-- Solo cuando hay evidencia explícita -->
<!-- - [NNN — titulo](decisions/NNN-slug.md) -->

## Notas para agentes

- Leer `project.md` siempre — es el punto de entrada
- Cargar solo los dominios relevantes a la tarea
- Si `coverage: bootstrap`, el contexto fue generado automáticamente — puede tener gaps
- No modificar este archivo manualmente — actualizarlo vía skill `context-nav`
- `Technical domain/business-rules.md` es la fuente de verdad de reglas de negocio — no saltarlas bajo ninguna circunstancia
- Si `.project-context/` no existe, detenerse y pedir al humano que ejecute `context-init` antes de continuar — no implementar nada sin contexto
- Si el cambio toca más de un servicio → cargar `cross-service-dev` antes de implementar; el `service-map.yaml` en `.project-context/` es la fuente de verdad del ecosistema

### Workflow obligatorio para agentes developer

1. Leer el contexto relevante en `.project-context/` antes de tocar código — empezar por `project.md` y cargar dominios pertinentes a la tarea
2. Implementar el cambio en el dominio asignado, respetando convenciones de `Core/coding-standards.md` y reglas de `Technical domain/business-rules.md`
3. Correr la suite de tests del package/módulo afectado y confirmar que pasa
4. Correr lint/format del stack y resolver toda advertencia antes de continuar
5. Validar con el humano el resultado (diff, tests verdes, comportamiento esperado) y esperar confirmación
6. Invocar a `reporter` para registrar el cierre del run en `.project-context/`
