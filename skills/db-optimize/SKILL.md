---
name: db-optimize
description: Identify slow queries and suggest schema, index, or query optimizations. Use when user says "slow query", "optimize SQL", "add index", "query performance", "EXPLAIN", or investigating database bottlenecks.
---

# Database Optimization

> Analyze query performance and suggest schema, index, or query improvements.

## Prerequisite

Run `/db-schema-scan` first (or receive schema context inline from orchestrator) to understand current tables, indexes, and relationships.

## When to Use

- User reports slow queries or timeouts
- After adding a new query to a repository
- During QA review of data-heavy features
- Proactive audit of query patterns

## Workflow

### Step 1 — Identify Target Queries

Find queries to analyze:
- Read repository files (`queries/`, `*_psql.go`, `*_repository.go`)
- Look for: full table scans, missing WHERE clauses, JOINs without indexes, N+1 patterns
- If user pointed to a specific query, start there

### Step 2 — Analyze Each Query

For each query, evaluate:

| Check | What to Look For | Fix |
|-------|-----------------|-----|
| **Missing index on WHERE/JOIN** | Column in WHERE or JOIN ON has no index | Create index |
| **Full table scan** | SELECT without WHERE on large table | Add filtering or pagination |
| **N+1 queries** | Loop calling single-row query N times | Batch query with IN clause or JOIN |
| **SELECT *** | Fetching all columns when only 2-3 needed | List specific columns |
| **Missing LIMIT** | List queries without pagination | Add LIMIT + OFFSET or cursor |
| **Unindexed ORDER BY** | Sorting on non-indexed column | Add index or compound index |
| **Implicit type cast** | WHERE varchar_col = 123 (no quotes) | Fix type to avoid cast |
| **OR in WHERE** | Multiple OR conditions prevent index use | UNION or restructure |
| **COUNT(*)** on large table | Full scan for count | Approximate count or cached counter |
| **Tenant isolation** | Multi-tenant query without tenant_id filter | Add tenant_id to WHERE + compound index |

### Step 3 — Index Recommendations

When suggesting indexes:

```markdown
## Index Recommendations

| Table | Suggested Index | Columns | Type | Rationale |
|-------|----------------|---------|------|-----------|
| instances | idx_instances_tenant_status | (tenant_id, status) | btree | Query filters by tenant + status, currently full scan |
| events | idx_events_instance_created | (instance_id, created_at DESC) | btree | Timeline query sorts by date per instance |

### Index Trade-offs
- Each index slows INSERT/UPDATE by ~5-10%
- Only add indexes for queries that run frequently
- Compound indexes: put high-cardinality column first (unless tenant_id for isolation)
- Partial indexes for filtered subsets: `WHERE deleted_at IS NULL`
```

### Step 4 — Query Rewrite Suggestions

If the query itself can be improved:

```markdown
## Query Rewrites

### Before (N+1 pattern)
for each workflow_id:
  SELECT * FROM instances WHERE workflow_id = $1

### After (batch)
SELECT * FROM instances WHERE workflow_id = ANY($1::uuid[])

### Impact: N queries → 1 query
```

## PostgreSQL-Specific Optimizations

| Technique | When | How |
|-----------|------|-----|
| `EXPLAIN ANALYZE` | Verify index usage | Prefix query with EXPLAIN ANALYZE |
| `CREATE INDEX CONCURRENTLY` | Large tables in production | Avoids table lock |
| Partial index | Column has many NULLs you filter out | `CREATE INDEX ... WHERE deleted_at IS NULL` |
| Covering index | Avoid heap lookup | `INCLUDE (col)` in index |
| `GROUPING SETS` | Multiple aggregations in one pass | Replace multiple GROUP BY queries |
| Connection pool tuning | Pool exhaustion errors | `max_open_conns`, `max_idle_conns`, `conn_max_lifetime` |

## Output

```markdown
## Query Performance Report — <project>

### Queries Analyzed: <N>
### Issues Found: <N>

### Critical (fix now)
1. **[file:line]** — <description> → <fix>

### Recommended (improve performance)
1. **[file:line]** — <description> → <fix>

### Index Changes
| Action | SQL |
|--------|-----|
| Add | `CREATE INDEX CONCURRENTLY idx_x ON table (col1, col2)` |
| Drop (unused) | `DROP INDEX idx_y` |

### Estimated Impact
- Query X: ~500ms → ~5ms (index on WHERE columns)
- Query Y: N+1 eliminated, ~N*10ms → ~15ms
```

## Rules

- **Read-only** — suggest changes, don't execute them. DBA agent creates migrations
- **Measure before optimizing** — don't add indexes speculatively. Identify the actual slow query first
- **Compound indexes > multiple single indexes** — one (tenant_id, status, created_at) beats three separate indexes
- **Don't over-index** — every index costs write performance. Only index what's queried frequently
