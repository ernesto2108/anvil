# Template: architecture-db.md

Inspirado en: sección "Data Storage" de Google + formato de spec DBML.

**Generar cuando:** hay cambios de base de datos involucrados.

## Template

```markdown
# Arquitectura de Base de Datos — <TASK-ID>

## Schema intent

<!-- Formato DBML — spec ejecutable. El agente DBA genera migraciones de esto. -->

```dbml
Table <table_name> {
  id text [pk, note: 'ULID']
  field1 text [not null]
  field2 integer [default: 0]
  created_at datetime [not null, default: `CURRENT_TIMESTAMP`]

  indexes {
    field1 [name: 'idx_<table>_field1']
    (field1, field2) [name: 'idx_<table>_compound']
  }
}

Ref: <table_name>.foreign_id > other_table.id
```

<!-- Alternativa: SQL DDL intent para cambios simples -->

```sql
-- Intent: agregar tracking de status a tabla runs
ALTER TABLE runs ADD COLUMN status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE runs ADD COLUMN finished_at DATETIME;
CREATE INDEX idx_runs_status ON runs(status);
```

## Estrategia de migración

- **Tipo de gestión:** [migraciones en repo / SQL manual / sync tool / otro]
- **Estado de la DB:** [nueva / existente en dev / existente en producción]
- **Compatibilidad hacia atrás:** ...
- **Orden de deploy:** [migración antes de código / código antes de migración / simultáneo]
- **Plan de rollback:** ...
- **Backfill de datos:** ...
- **Riesgos de producción:** [bloqueos de tabla, downtime estimado, datos afectados]

## Índices recomendados

| Índice | Columnas | Justificación (qué query lo necesita) |
|---|---|---|

## Patrones de consulta

<!-- Patrones de query esperados e implicaciones de rendimiento -->
- ...

## Diagrama ERD

```mermaid
erDiagram
  ...
```
```

## Patrones de acceso — incluir si aplica

<!-- Describir cómo fluyen los datos más allá de CRUD simple -->

### CQRS — incluir si hay separación read/write
- **Lado write:** qué tabla/aggregate recibe comandos
- **Lado read:** qué tabla/view sirve queries (puede ser diferente)
- **Sincronización:** cómo el lado read se actualiza (evento, trigger, polling)

### Event sourcing — incluir si el estado se reconstruye de eventos
- **Event store:** tabla donde se persisten eventos (`id`, `aggregate_id`, `event_type`, `payload`, `occurred_at`)
- **Snapshot:** frecuencia, tabla de snapshots
- **Proyecciones:** qué read models se construyen y cómo

### Outbox pattern — incluir si se publican eventos desde DB
- **Tabla outbox:** `id`, `topic`, `payload`, `published_at` (null = pendiente)
- **Poller / relay:** quién lee y publica los mensajes pendientes
- **Garantía:** at-least-once delivery desde DB hacia broker

## Reglas

- DBML es el formato preferido para tablas nuevas — legible por máquinas, genera migraciones
- SQL DDL intent es aceptable para cambios simples de ALTER TABLE
- Cada índice debe tener justificación — qué query lo necesita
- La estrategia de migración debe abordar rollback — qué pasa si necesitamos revertir
- El diagrama ERD muestra relaciones, no todas las columnas — mantenerlo legible
- El schema intent debe coincidir con los tipos de persistencia backend si ambas vistas existen
- NUNCA proponer tablas nuevas sin confirmar que las existentes no pueden extenderse
- Incluir sección "Patrones de acceso" cuando la feature usa eventos, proyecciones, o paths separados de read/write — tareas CRUD simples pueden omitirla
