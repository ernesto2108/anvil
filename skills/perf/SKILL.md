---
name: perf
description: Ejecutar pruebas de carga/estrés/rendimiento, analizar cuellos de botella y generar reportes visuales. Auto-detecta Vegeta, k6, Locust. Usar cuando el usuario diga "load test", "stress test", "performance test", "prueba de carga", "benchmark", "run vegeta", "run k6", "run locust", "throughput", "rps", "50 rps", "latency", "bottleneck", "perf", o al investigar rendimiento bajo concurrencia.
---

# Perf — Pruebas de Rendimiento y Análisis

> Ejecutar pruebas de carga, analizar cuellos de botella y producir reportes visuales de rendimiento.

## Fase 1 — Descubrimiento

Antes de ejecutar cualquier cosa, recopilar TODO el contexto de forma interactiva. SIEMPRE preguntar en español.

### Paso 1: Detectar configuración existente

Verificar si el proyecto ya tiene herramientas de prueba de carga:

| Archivo/Patrón | Framework |
|-------------|-----------|
| `loadtest/main.go` + `vegeta` en go.mod | Vegeta (librería Go) |
| binario `loadtest` en el proyecto | Vegeta (pre-compilado) |
| `*.js` con `import http from 'k6/http'` | k6 |
| `locustfile.py` | Locust |
| `vegeta` en PATH | Vegeta CLI |

Si se encuentra, leer la configuración/perfiles y omitir las preguntas ya respondidas.

### Paso 2: Preguntar lo que falta (SIEMPRE en español)

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

### Paso 3: Preguntar preferencia de framework

**NUNCA elegir el framework tú mismo.** Siempre preguntar al usuario/equipo:

> "Que herramienta de carga prefieren usar? Las opciones son:
>
> | Herramienta | Lenguaje | Charts nativos | Mejor para |
> |-------------|----------|---------------|------------|
> | **Vegeta** | Go | No (se generan con matplotlib) | Control preciso de rate, proyectos Go, CI pipelines |
> | **Locust** | Python | Si (web UI + HTML report) | Dashboard en tiempo real, equipos Python |
> | **k6** | JavaScript | Si (web dashboard + HTML) | Escenarios complejos, stages, thresholds, CI gates |
>
> Si ya tienen algo configurado en el proyecto, usamos eso."

Una vez elegido el framework, cargar la guía de referencia:
- Vegeta → `reference/vegeta.md`
- Locust → `reference/locust.md`
- k6 → `reference/k6.md`

### Paso 4: Preguntar dónde guardar los tests

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

Adaptar la estructura del proyecto según su respuesta.

### Paso 5: Preguntar formato de salida

> "En que formato quieres el reporte?
>
> - **Markdown** — se ve en GitHub, VS Code, terminal (siempre se genera)
> - **HTML** — interactivo, facil de compartir sin acceso al repo
> - **PDF** — para presentaciones formales, stakeholders, post-mortems
>
> Puedo generar mas de uno."

## Fase 2 — Configuración y Ejecución

### Verificaciones Pre-vuelo

1. **Verificar endpoint** — `curl -s -o /dev/null -w "%{http_code}" <url>`
2. **Verificar estado** — para endpoints con estado (inventario, tokens), consultar conteos actuales
3. **Confirmar ambiente** — NUNCA ejecutar contra producción sin confirmación explícita
4. **Verificar resultados existentes** — buscar resultados anteriores para comparación

### Ejecutar

Usar la guía de referencia específica del framework para los comandos.

Siempre guardar resultados raw con nombres descriptivos:
```
{YYYY-MM-DD}-{profile}-{rate}rps[-{suffix}].{ext}

Ejemplos:
  2026-03-30-normal-50rps.bin
  2026-04-01-saturado-50rps.bin
  2026-04-06-normal-50rps-after-sp-fix.bin
```

## Fase 3 — Análisis

### Métricas Resumen

Extraer de cada ejecución de prueba de carga:

| Métrica | Umbral SLO |
|--------|---------------|
| Rate real (req/s) | >= 90% del objetivo |
| Throughput (req/s) | >= 90% del objetivo |
| Tasa de éxito | >= 99.5% |
| Latencia p50 | < 1s |
| Latencia p95 | < 3s |
| Latencia p99 | < 5s |
| Latencia máxima | < 30s |
| Distribución de errores | Sin 5xx |

### Timeline Segundo a Segundo

El análisis más valioso. Decodificar resultados para ver completaciones por segundo.

**Qué buscar:**
- **Completaciones consistentes** (~objetivo/s cada segundo) = saludable
- **Pico luego colapso** (50 enviados, 2 completados) = contención de bloqueos / serialización
- **Degradación gradual** = agotamiento de recursos (conexiones, memoria, CPU)
- **Caídas periódicas** = pausas de GC, cron jobs, procesos en background

### Patrones de Cuello de Botella

| Patrón | Causa probable | Investigación |
|---------|-------------|---------------|
| p50 > 1s en un solo span | Contención de bloqueos en DB | Verificar `pg_locks`, advisory locks, `FOR UPDATE` |
| Alta varianza (p50 bajo, p99 alto) | Encolamiento / serialización | Advisory locks, spin-locks, mutex |
| Todos los spans rápidos pero total lento | Red / infra | CDN, WAF, latencia de API gateway |
| Throughput << objetivo, sin errores | Workers saturados | Aumentar workers o corregir llamadas bloqueantes |
| Éxitos caen bajo carga | Agotamiento del pool | `max_connections`, configuración del pool |

### Análisis de Bloqueos de Base de Datos (PostgreSQL)

Si el cuello de botella está relacionado con la DB, verificar:

```sql
-- Advisory locks activos
SELECT pid, locktype, classid, objid, granted FROM pg_locks WHERE locktype = 'advisory';

-- Contención de bloqueos a nivel de fila
SELECT relation::regclass, mode, granted, COUNT(*) FROM pg_locks WHERE NOT granted GROUP BY 1,2,3;

-- Transacciones de larga duración
SELECT pid, now() - xact_start AS duration, query FROM pg_stat_activity
WHERE state = 'active' AND now() - xact_start > interval '1 second' ORDER BY duration DESC;
```

**Anti-patrones comunes en SPs:**

| Anti-patrón | Corrección |
|-------------|-----|
| `SELECT ... FOR UPDATE` sin LIMIT | Agregar LIMIT o eliminar si el resultado no se usa |
| Advisory lock no liberado en todos los caminos de RETURN | Asegurar `pg_advisory_unlock` antes de cada RETURN |
| Spin-lock con `pg_sleep` consultando `pg_locks` | Usar `pg_advisory_lock` (bloqueante) |
| Múltiples COUNT(*) en la misma tabla | Consolidar con `COUNT(*) FILTER (WHERE ...)` |
| Generación de tokens dentro del hot path | Pre-cargar tokens antes del pico de demanda |

## Fase 4 — Reporte

### Comparación (antes/después)

Cuando existen resultados previos:

1. Detectar el resultado más reciente para el mismo perfil
2. Preguntar: "Encontre un resultado previo de {fecha}. Quieres que compare?"
3. Si sí — generar charts lado a lado con ambas ejecuciones
4. Agregar columna "Anterior" en la tabla de resumen ejecutivo

### Charts

El framework puede proporcionarlos nativamente:
- **Locust** — reporte HTML con `--html report.html` (tiene charts integrados)
- **k6** — reporte HTML con k6-reporter o web dashboard `--out web-dashboard`
- **Vegeta** — sin charts nativos; generar con matplotlib (ver `reference/vegeta.md`)

### Visualizaciones Requeridas

Independientemente del framework, el reporte debe incluir:

1. **Comparación de throughput** — completadas vs esperadas, rate real vs objetivo
2. **Completaciones por segundo** — gráfico de barras segundo a segundo (el más impactante para los equipos)
3. **Latencia por percentil** — p50, p90, p95, p99, max
4. **Desglose de cuello de botella** — tiempo por componente (si hay datos APM disponibles)
5. **Timeline** — sent/s, completed/s, mean latency/s

### Formatos de Salida

| Formato | Cuándo | Cómo |
|---------|--------|------|
| **Markdown** (siempre) | Día a día, GitHub, PRs | Generado automáticamente con `![](charts/...)` |
| **HTML** (si lo pide) | Compartir sin acceso al repo | Locust/k6 nativo. Vegeta: generar desde md |
| **PDF** (si lo pide) | Presentaciones, stakeholders | `pandoc report.md -o report.pdf --pdf-engine=wkhtmltopdf` |

Si se solicita PDF, verificar herramientas:
```bash
which pandoc && which wkhtmltopdf
# Si faltan: brew install pandoc wkhtmltopdf
```

### Estructura del Reporte

```markdown
# Informe de Performance — <endpoint>

## Resumen Ejecutivo
| Metrica | Actual | Anterior | Target |
(tabla comparativa si existen datos previos)

![Throughput](charts/chart-throughput.png)

## Pruebas Realizadas
(parámetros por prueba)

![Latencies](charts/chart-latency.png)

## Analisis Segundo a Segundo
![Completions](charts/chart-completions.png)
![Timeline](charts/chart-timeline.png)

## Causa Raiz
(si se encontró cuello de botella)

![Breakdown](charts/chart-breakdown.png)

## Plan de Remediacion
(correcciones por fases con SQL si aplica)

## Anexo — Todos los Graficos
(todos los charts para referencia rápida)
```

**Reglas del reporte:**
- Siempre embeber imágenes con `![desc](relative-path.png)` — nunca solo listar nombres de archivos
- Rutas relativas para que se renderice en GitHub, VS Code, cualquier visor de markdown
- El chart de completaciones por segundo es el más impactante — siempre destacado
- Si se comparan ejecuciones, ponerlas lado a lado en el mismo chart
- La narrativa del reporte siempre en español

## Resumen del Flujo de Trabajo

1. **Descubrimiento** — detectar configuración existente, preguntar info faltante, preguntar preferencia de framework, preguntar dónde almacenar, preguntar formato de salida
2. **Configuración** — cargar referencia del framework, crear estructura del proyecto según elección del usuario
3. **Pre-vuelo** — verificar endpoint, revisar estado, confirmar ambiente
4. **Ejecutar** — correr prueba de carga, guardar resultados con nombre descriptivo
5. **Analizar** — métricas, timeline, identificación de cuello de botella
6. **Charts** — generar visualizaciones (nativas o matplotlib)
7. **Reporte** — markdown con imágenes embebidas + formatos adicionales si se solicitan
8. **Comparar** — si existen resultados previos, preguntar y graficar ambas ejecuciones
9. **Exportar** — HTML o PDF si se solicita

## Reglas

- **Comunicar en español** — todas las preguntas, actualizaciones de estado y narrativas de reportes en español. Los términos técnicos (endpoint, headers, throughput, p50) se mantienen en inglés
- **Nunca elegir el framework** — siempre preguntar al usuario/equipo qué herramienta prefieren. Presentar opciones con trade-offs
- **Nunca elegir la estructura del proyecto** — siempre preguntar: dentro del repo, repo dedicado o híbrido
- **Siempre hacer preguntas de descubrimiento** si no hay configuración existente — nunca adivinar endpoint/headers
- **Nunca ejecutar contra producción** sin confirmación explícita del usuario
- **Siempre guardar resultados raw** — permite re-análisis y comparación
- **Los charts son obligatorios** — los reportes sin visualizaciones están incompletos
- **Nombrar resultados con fecha** — `{YYYY-MM-DD}-{profile}-{rate}rps.{ext}`
- **Comparar antes/después** — si existen resultados previos, preguntar y graficar ambas ejecuciones
- **Verificar estado previo** — para endpoints con estado, verificar datos antes y después
- **Incluir SQL** — si se encuentran cuellos de botella en DB, incluir SQL listo para producción con guards y rollback
