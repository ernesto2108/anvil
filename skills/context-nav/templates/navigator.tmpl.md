# Context Navigator — <ProjectName>

last_full_scan: <YYYY-MM-DD>
last_updated: <YYYY-MM-DD>
coverage: bootstrap

## Índice

- [Proyecto](project.md) — stack, arquitectura, restricciones, SOLID
- [Operaciones](ops.md) — comandos para levantar, buildear, testear y operar
- [Patrones](patterns.md) — patrones de diseño inferidos con referencias
- [Contratos](contracts.md) — APIs, queues, eventos, servicios externos
- [Riesgos](risks.md) — gotchas, deuda técnica, restricciones conocidas

### Dominios activos

<!-- Un item por carpeta con lógica de negocio significativa -->
- [<domain>](domains/<domain>.md) — <una línea de responsabilidad>

### Decisiones arquitectónicas

<!-- Solo cuando hay evidencia explícita -->
<!-- - [NNN — titulo](decisions/NNN-slug.md) -->

## Notas para agentes

- Leer `project.md` siempre — es el punto de entrada
- Cargar solo los dominios relevantes a la tarea
- Si `coverage: bootstrap`, el contexto fue generado automáticamente — puede tener gaps
- No modificar este archivo manualmente — actualizarlo vía skill `context-nav`
