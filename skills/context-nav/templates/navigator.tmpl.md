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

### Decisiones arquitectónicas
<!-- Solo cuando hay evidencia explícita -->
<!-- - [NNN — titulo](decisions/NNN-slug.md) -->

## Notas para agentes

- Leer `project.md` siempre — es el punto de entrada
- Cargar solo los dominios relevantes a la tarea
- Si `coverage: bootstrap`, el contexto fue generado automáticamente — puede tener gaps
- No modificar este archivo manualmente — actualizarlo vía skill `context-nav`
