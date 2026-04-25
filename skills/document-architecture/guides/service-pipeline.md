# Pipeline de Servicio — Go Backend

## Detección de arquitectura

| Raíz contiene | Patrón | Esqueleto |
|---|---|---|
| `internal/` (SIN `domain/`) | MVC | `service-skeleton-mvc.md` |
| `domain/` + `usecase/` | Clean | `service-skeleton-clean.md` |
| `internal/` con `business/` | Hex | `service-skeleton-clean.md` (notar desviaciones) |

## Patrón de salida

```
{architecture_path}/<project>/
├── context-summary.md      # Resumen técnico
├── context-endpoints.md    # Flujos por endpoint
├── context-risks.md        # Áreas de riesgo con fragmentos de código
├── overview.md             # C4, componentes, ERD, diagramas de secuencia
├── security-audit.md       # Auditoría OWASP con puntaje
└── endpoints/
    └── <endpoint-name>.md  # Uno por endpoint con diagrama de secuencia
```

## Referencia de esqueleto

### MVC (`service-skeleton-mvc.md`)
- Estructura: `internal/{domain}/handlers|services|repositories`
- Middleware: MiddlewareError + MiddlewareTracking (separados)
- Config: JSON local -> SSM fallback
- SQL: queries crudas en `queries/*.go`

### Clean (`service-skeleton-clean.md`)
- Estructura: `domain/` + `usecase/` + `interface/` + `infrastructure/`
- Middleware: trackRequestAndResponse + authorizationMiddleware (fusionados)
- Config: YAML local -> SSM fallback
- SQL: squirrel + scany
- Multi-modo: HTTP, gRPC, SQS, Lambda

## Instrucciones para el Scanner

"El esqueleto describe patrones comunes en servicios de este tipo. NO re-explorar estos patrones. Enfocarse en lo que DIFIERE: dominios, endpoints, lógica de negocio, esquema, integraciones, librerías extra. En context-summary.md, referenciar el esqueleto y detallar solo las diferencias. En context-risks.md, incluir FRAGMENTOS DE CÓDIGO (5-10 líneas) para: (a) auth/autorización, (b) validación de inputs, (c) riesgos SQL, (d) exposición de datos, (e) CONCURRENCIA — Redis locks, errgroup, goroutines, os.Exit, (f) INTEGRACIÓN — config Kafka/SQS, interceptores gRPC, (g) MANEJO DE ERRORES — errores ignorados, rollbacks silenciados, fmt.Errorf con strings externas. Referenciar issues sistémicos heredados por ID."

Salida: `context-summary.md`, `context-endpoints.md`, `context-risks.md`

## Arquitecto — Overview

Inyectar: context-summary.md INLINE. NO inyectar endpoints ni risks.

Generar `overview.md` (+ opcional `state-machine.md`) con:
1. Descripcion del servicio
2. Diagrama de Contexto (C4)
3. Componentes internos
4. ERD
5. Sequence diagrams de flujos criticos
6. Dependencias externas (tabla)
7. Notas tecnicas

## Arquitecto — Detalle (endpoints)

Antes de lanzar, leer `<docs>/07-references/template-endpoint.md`.

Inyectar: template-endpoint.md + context-endpoints.md INLINE. NO inyectar summary ni risks.

Instrucciones: "Usar el template como referencia de formato EXACTA. NO leer archivos de ejemplo."

Salida: `endpoints/*.md` (uno por endpoint)

## Instrucciones de Seguridad (backend)

"Los issues sistémicos documentados en `known-systemic-issues.md` están PRE-VALIDADOS. Incluirlos con su prefijo de ID y anotar 'inherited from ecosystem'. NO gastar tool calls confirmándolos. Para riesgos específicos del servicio, context-risks.md tiene FRAGMENTOS DE CÓDIGO COMPLETOS cubriendo auth, input, SQL, exposición de datos, concurrencia, integraciones y manejo de errores. NO re-leer archivos en infrastructure/ ni usecase/. Solo usar Read para rastrear cadenas de llamadas entre archivos no cubiertas por los fragmentos."

Salida: `security-audit.md`, archivos de bugs en `{reports_path}/bugs/`, entradas en `{backlog_path}`

## Referencia de issues sistémicos

Archivo: `<docs>/07-references/known-systemic-issues.md`
- Issues sistémicos confirmados en servicios auditados
- Issues frecuentes encontrados en subconjunto de servicios
- Referencia de puntaje del audit del ecosistema

## Tabla de inyección rápida

| Agente | Inyectar INLINE | NO inyectar |
|---|---|---|
| Scanner | skeleton (resuelto por patrón) | — |
| Arquitecto (overview) | context-summary.md | endpoints, risks |
| Arquitecto (endpoints) | template-endpoint.md + context-endpoints.md | summary, risks |
| Seguridad | known-systemic-issues.md + context-risks.md + overview summary | full endpoints |
