---
name: db-schema-scan
description: Read-only inspection of database schema via migration files and schema SQL. Use when user says "check the schema", "show tables", "inspect migrations", "what columns does X have", or needs to understand the database structure before writing queries.
---

# Database Schema Scan

> Read-only inspection of the current database schema. Prerequisite for DBA work and query optimization.

## When to Use

- Before any DBA migration work (understand current state)
- Before developer writes repository queries (verify table/column names)
- When user asks "what columns does X have", "show me the schema", "check the tables"
- As prerequisite for `/db-optimize`

## Workflow

### Step 1 — Find Migration Files

Search for migration files in common locations:

```
migrations/*.sql
db/migrations/*.sql
sql/migrations/*.sql
database/migrations/*.sql
**/migrate/*.sql
```

Also check for schema dumps: `schema.sql`, `init.sql`, `db/schema.sql`

### Step 2 — Parse Schema

Read migration files in order (by number/timestamp) and build a mental model of:

1. **Tables** — name, columns, types, constraints
2. **Relationships** — foreign keys, join tables
3. **Indexes** — name, columns, unique/partial
4. **Enums/Types** — custom types defined
5. **RLS Policies** — if multi-tenant
6. **Triggers** — if any exist

### Step 3 — Produce Schema Summary

Output a structured summary:

```markdown
## Schema Summary — <project>

### Tables (<count>)

#### <table_name>
| Column | Type | Nullable | Default | Constraint |
|--------|------|----------|---------|------------|
| id | UUID | NO | uuid_generate_v7() | PK |
| email | VARCHAR(255) | NO | — | UNIQUE |
| tenant_id | UUID | YES | — | FK → tenants(id) |
| created_at | TIMESTAMP | YES | CURRENT_TIMESTAMP | — |

**Indexes:** idx_users_email, idx_users_tenant_id
**RLS:** tenant_id = current_setting('app.tenant_id')

### Relationships
- users.tenant_id → tenants.id
- user_roles.user_id → users.id
- user_roles.role_id → roles.id

### Migration Count: <N> (latest: <filename>)

### Potential Issues
- [ ] Table X has no tenant_id (multi-tenant gap)
- [ ] Table Y has no index on frequently queried column Z
- [ ] Column A is VARCHAR(255) but only stores short codes
```

### Step 4 — Check for Schema/Code Mismatches (optional)

If the orchestrator asks, compare schema against repository query files:
- Columns referenced in queries that don't exist in schema
- Tables in schema that have no corresponding repository
- Type mismatches (schema says UUID, code scans as string)

## Rules

- **READ-ONLY** — never modify schema or migration files
- **Report, don't fix** — flag issues for the DBA agent to handle
- **Order matters** — read migrations in sequence to understand evolution
- **Check rollbacks** — note migrations that have `.up.sql` but no `.down.sql`
