# Template: architecture-db.md

Inspired by: Google "Data Storage" section + DBML spec format.

**Generate when:** database changes are involved.

## Template

```markdown
# Arquitectura de Base de Datos — <TASK-ID>

## Schema intent

<!-- DBML format — executable spec. DBA agent generates migrations from this. -->

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

<!-- Alternative: SQL DDL intent for simple changes -->

```sql
-- Intent: add status tracking to runs table
ALTER TABLE runs ADD COLUMN status TEXT NOT NULL DEFAULT 'pending';
ALTER TABLE runs ADD COLUMN finished_at DATETIME;
CREATE INDEX idx_runs_status ON runs(status);
```

## Estrategia de migración

- **Compatibilidad hacia atrás:** ...
- **Plan de rollback:** ...
- **Backfill de datos:** ...

## Índices recomendados

| Índice | Columnas | Justificación (qué query lo necesita) |
|---|---|---|

## Patrones de consulta

<!-- Expected query patterns and performance implications -->
- ...

## Diagrama ERD

```mermaid
erDiagram
  ...
```
```

## Rules

- DBML is the preferred format for new tables — machine-readable, generates migrations
- SQL DDL intent is acceptable for simple ALTER TABLE changes
- Every index must have a justification — which query needs it
- Migration strategy must address rollback — what happens if we need to revert
- ERD diagram shows relationships, not all columns — keep it readable
- Schema intent must match backend persistence types if both views exist
- NEVER propose new tables without confirming existing tables can't be extended (DB schema rule from architect agent)
