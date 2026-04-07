---
name: perf
description: Run load/stress/performance tests, analyze bottlenecks, and generate visual reports. Auto-detects Vegeta, k6, Locust. Use when user says "load test", "stress test", "performance test", "prueba de carga", "benchmark", "run vegeta", "run k6", "run locust", "throughput", "rps", "50 rps", "latency", "bottleneck", "perf", or investigating performance under concurrency.
---

# Perf — Performance Testing & Analysis

> Execute load tests, analyze bottlenecks, and produce visual performance reports.

## Phase 1 — Discovery

Before running anything, gather ALL context interactively. ALWAYS ask in Spanish.

### Step 1: Detect existing setup

Check if the project already has load test tooling:

| File/Pattern | Framework |
|-------------|-----------|
| `loadtest/main.go` + `vegeta` in go.mod | Vegeta (Go library) |
| `loadtest` binary in project | Vegeta (pre-compiled) |
| `*.js` with `import http from 'k6/http'` | k6 |
| `locustfile.py` | Locust |
| `vegeta` in PATH | Vegeta CLI |

If found, read the config/profiles and skip questions already answered.

### Step 2: Ask what's missing (ALWAYS in Spanish)

| # | Pregunta | Ejemplo |
|---|----------|---------|
| 1 | Cual es el host o URL base? | `https://web-api-qa.example.com` |
| 2 | Cual es el endpoint a probar? | `/api/v1/bookings/generate-booking` |
| 3 | Que metodo HTTP usa? | POST, GET, PUT, DELETE |
| 4 | Necesita headers especiales? | `Authorization: Bearer ...`, custom headers |
| 5 | Cual es el body del request? (si es POST/PUT) | JSON payload o template |
| 6 | A cuantas requests por segundo quieres probar? | 50, 100, 500 |
| 7 | Cuanto tiempo debe durar la prueba? | 30s, 1m, 5m |
| 8 | En que ambiente se corre? | QA, staging, produccion (advertir si es produccion) |
| 9 | Necesita autenticacion o bypass? | API key, recaptcha bypass, JWT |
| 10 | Los datos varian entre requests? | IDs random, contadores, datos de un CSV |

### Step 3: Ask framework preference

**NEVER choose the framework yourself.** Always ask the user/team:

> "Que herramienta de carga prefieren usar? Las opciones son:
>
> | Herramienta | Lenguaje | Charts nativos | Mejor para |
> |-------------|----------|---------------|------------|
> | **Vegeta** | Go | No (se generan con matplotlib) | Control preciso de rate, proyectos Go, CI pipelines |
> | **Locust** | Python | Si (web UI + HTML report) | Dashboard en tiempo real, equipos Python |
> | **k6** | JavaScript | Si (web dashboard + HTML) | Escenarios complejos, stages, thresholds, CI gates |
>
> Si ya tienen algo configurado en el proyecto, usamos eso."

Once the framework is chosen, load the reference guide:
- Vegeta → `reference/vegeta.md`
- Locust → `reference/locust.md`
- k6 → `reference/k6.md`

### Step 4: Ask where to put the tests

> "Donde quieres guardar las pruebas de carga?
>
> **Opcion A — Dentro del repo del servicio** (`my-service/loadtest/`)
> - El script vive junto al codigo que prueba
> - Facil de mantener si cambia el endpoint
> - Resultados y reportes quedan en el mismo repo
>
> **Opcion B — Repo de performance dedicado** (`blt-performance-tests/`)
> - Un solo lugar para todos los servicios
> - Tools compartidos (chart generators, decoders)
> - No contaminas repos de servicios con binarios y PNGs
> - Mejor si testeas multiples servicios
>
> **Opcion C — Hibrido**
> - Script de carga dentro del repo del servicio (esta acoplado al contrato)
> - Resultados y reportes en un repo de performance centralizado
>
> Cual prefieren?"

Adapt the project structure based on their answer.

### Step 5: Ask output format

> "En que formato quieres el reporte?
>
> - **Markdown** — se ve en GitHub, VS Code, terminal (siempre se genera)
> - **HTML** — interactivo, facil de compartir sin acceso al repo
> - **PDF** — para presentaciones formales, stakeholders, post-mortems
>
> Puedo generar mas de uno."

## Phase 2 — Setup & Execution

### Pre-flight Checks

1. **Verify endpoint** — `curl -s -o /dev/null -w "%{http_code}" <url>`
2. **Check state** — for stateful endpoints (inventory, tokens), query current counts
3. **Confirm environment** — NEVER run against production without explicit confirmation
4. **Check for existing results** — look for previous results for comparison

### Execute

Use the framework-specific reference guide for commands.

Always save raw results with descriptive names:
```
{YYYY-MM-DD}-{profile}-{rate}rps[-{suffix}].{ext}

Examples:
  2026-03-30-normal-50rps.bin
  2026-04-01-saturado-50rps.bin
  2026-04-06-normal-50rps-after-sp-fix.bin
```

## Phase 3 — Analysis

### Summary Metrics

Extract from every load test run:

| Metric | SLO Threshold |
|--------|---------------|
| Rate real (req/s) | >= 90% of target |
| Throughput (req/s) | >= 90% of target |
| Success rate | >= 99.5% |
| p50 latency | < 1s |
| p95 latency | < 3s |
| p99 latency | < 5s |
| Max latency | < 30s |
| Error distribution | No 5xx |

### Second-by-Second Timeline

The most valuable analysis. Decode results to see completions per second.

**What to look for:**
- **Consistent completions** (~target/s each second) = healthy
- **Spike then collapse** (50 sent, 2 completed) = lock contention / serialization
- **Gradual degradation** = resource exhaustion (connections, memory, CPU)
- **Periodic dips** = GC pauses, cron jobs, background processes

### Bottleneck Patterns

| Pattern | Likely Cause | Investigation |
|---------|-------------|---------------|
| p50 > 1s on single span | Lock contention in DB | Check `pg_locks`, advisory locks, `FOR UPDATE` |
| High variance (p50 low, p99 high) | Queuing / serialization | Advisory locks, spin-locks, mutex |
| All spans fast but total slow | Network / infra | CDN, WAF, API gateway latency |
| Throughput << target, no errors | Workers saturated | Increase workers or fix blocking calls |
| Success drops under load | Pool exhaustion | `max_connections`, pool settings |

### Database Lock Analysis (PostgreSQL)

If bottleneck is DB-related, check:

```sql
-- Active advisory locks
SELECT pid, locktype, classid, objid, granted FROM pg_locks WHERE locktype = 'advisory';

-- Row-level lock contention
SELECT relation::regclass, mode, granted, COUNT(*) FROM pg_locks WHERE NOT granted GROUP BY 1,2,3;

-- Long-running transactions
SELECT pid, now() - xact_start AS duration, query FROM pg_stat_activity
WHERE state = 'active' AND now() - xact_start > interval '1 second' ORDER BY duration DESC;
```

**Common SP anti-patterns:**

| Anti-pattern | Fix |
|-------------|-----|
| `SELECT ... FOR UPDATE` without LIMIT | Add LIMIT or remove if result unused |
| Advisory lock not released on all RETURN paths | Ensure `pg_advisory_unlock` before every RETURN |
| Spin-lock with `pg_sleep` polling `pg_locks` | Use `pg_advisory_lock` (blocking) |
| Multiple COUNT(*) on same table | Consolidate with `COUNT(*) FILTER (WHERE ...)` |
| Token generation inside hot path | Pre-load tokens before peak demand |

## Phase 4 — Report

### Comparison (before/after)

When previous results exist:

1. Detect the most recent result for the same profile
2. Ask: "Encontre un resultado previo de {fecha}. Quieres que compare?"
3. If yes — generate side-by-side charts with both runs
4. Add "Anterior" column in the executive summary table

### Charts

The framework may provide them natively:
- **Locust** — HTML report with `--html report.html` (has built-in charts)
- **k6** — HTML report with k6-reporter or web dashboard `--out web-dashboard`
- **Vegeta** — no native charts; generate with matplotlib (see `reference/vegeta.md`)

### Required Visualizations

Regardless of framework, the report must include:

1. **Throughput comparison** — completed vs expected, rate real vs target
2. **Completions per second** — second-by-second bar chart (most impactful for teams)
3. **Latency by percentile** — p50, p90, p95, p99, max
4. **Bottleneck breakdown** — time per component (if APM data available)
5. **Timeline** — sent/s, completed/s, mean latency/s

### Output Formats

| Formato | Cuando | Como |
|---------|--------|------|
| **Markdown** (siempre) | Dia a dia, GitHub, PRs | Generado automaticamente con `![](charts/...)` |
| **HTML** (si lo pide) | Compartir sin acceso al repo | Locust/k6 nativo. Vegeta: generar desde md |
| **PDF** (si lo pide) | Presentaciones, stakeholders | `pandoc report.md -o report.pdf --pdf-engine=wkhtmltopdf` |

If PDF requested, verify tools:
```bash
which pandoc && which wkhtmltopdf
# If missing: brew install pandoc wkhtmltopdf
```

### Report Structure

```markdown
# Informe de Performance — <endpoint>

## Resumen Ejecutivo
| Metrica | Actual | Anterior | Target |
(comparative table if previous data exists)

![Throughput](charts/chart-throughput.png)

## Pruebas Realizadas
(parameters per test)

![Latencies](charts/chart-latency.png)

## Analisis Segundo a Segundo
![Completions](charts/chart-completions.png)
![Timeline](charts/chart-timeline.png)

## Causa Raiz
(if bottleneck found)

![Breakdown](charts/chart-breakdown.png)

## Plan de Remediacion
(phased fixes with SQL if applicable)

## Anexo — Todos los Graficos
(all charts for quick reference)
```

**Report rules:**
- Always embed images with `![desc](relative-path.png)` — never just list filenames
- Relative paths so it renders in GitHub, VS Code, any markdown viewer
- Completions-per-second chart is the most impactful — always prominent
- If comparing runs, put them side by side in the same chart
- Report narrative always in Spanish

## Workflow Summary

1. **Discovery** — detect existing setup, ask missing info, ask framework preference, ask where to store, ask output format
2. **Setup** — load framework reference, create project structure per user's choice
3. **Pre-flight** — verify endpoint, check state, confirm environment
4. **Execute** — run load test, save results with descriptive name
5. **Analyze** — metrics, timeline, bottleneck identification
6. **Chart** — generate visuals (native or matplotlib)
7. **Report** — markdown with embedded images + additional formats if requested
8. **Compare** — if previous results exist, ask and chart both runs
9. **Export** — HTML or PDF if requested

## Rules

- **Communicate in Spanish** — all questions, status updates, and report narratives in Spanish. Technical terms (endpoint, headers, throughput, p50) stay in English
- **Never choose the framework** — always ask the user/team which tool they prefer. Present options with trade-offs
- **Never choose the project structure** — always ask: inside repo, dedicated repo, or hybrid
- **Always ask discovery questions** if no existing setup — never guess endpoint/headers
- **Never run against production** without explicit user confirmation
- **Always save raw results** — enables re-analysis and comparison
- **Charts are mandatory** — reports without visuals are incomplete
- **Name results with date** — `{YYYY-MM-DD}-{profile}-{rate}rps.{ext}`
- **Compare before/after** — if previous results exist, ask and chart both runs
- **Pre-check state** — for stateful endpoints, verify data before and after
- **Include SQL** — if DB bottlenecks found, include production-ready SQL with guards and rollback
