# Guía de Escaneo Profundo

Se usa cuando el escáner se invoca con `mode: deep` (pipelines de documentación).

## Detección de stack (PRIMERO)

Verifica los archivos marcadores para determinar el stack:

| Archivo | Stack | Receta |
|---|---|---|
| `go.mod` | Go | Consulta `docs/scanner-recipe-go-gin.md` si existe el import de gin |
| `package.json` + `src/` con `.tsx` | React | Sigue los patrones de frontend a continuación |
| `pubspec.yaml` | Flutter | Sigue los patrones de mobile a continuación |
| `Cargo.toml` | Rust | Sigue el enfoque genérico |
| `requirements.txt` / `pyproject.toml` | Python | Sigue el enfoque genérico |

Si existe una receta específica para el stack en `docs/`, cárgala. Define el orden de extracción, patrones de grep, qué omitir y el presupuesto de tokens.

Si no existe receta para el stack detectado, sigue el enfoque genérico.

## Enfoque genérico

Funciona para cualquier stack:

1. Realiza el escaneo estándar (estructura del proyecto, dependencias, tests, CI)
2. Traza los flujos de endpoint/ruta: cadena handler → lógica → capa de datos
3. **Prefiere Grep sobre Read:** extrae firmas de funciones, bindings, queries, llamadas externas. Solo usa Read en funciones completas cuando el contexto de grep no es suficiente
4. **Escribe TRES archivos segmentados** (no un monolito)
5. El objetivo: **ningún agente posterior necesita releer el código fuente** — todos los flujos ya están trazados

## Archivos de salida

### context-summary.md (~200 líneas máx)

```markdown
# {project-name} — Contexto tecnico
## 1. Identificacion (framework, runtime, module, responsabilidades)
## 2. Estructura de directorios (arbol, patron arquitectonico)
## 3. Dependencias principales (tabla: libreria, version, uso)
## 4. Endpoints/routes expuestos (tabla: metodo, ruta, handler, descripcion)
## 5. Dependencias externas (DB, cache, HTTP, gRPC, brokers — con config keys)
## 6. Configuracion (estructura resumida, variables de entorno)
## 7. Middleware/interceptors (cadena completa en orden)
## 8. Logica de negocio destacada (state machines, async patterns, rollbacks)
## 9. Schema de BD inferido (tablas principales con campos clave)
## 10. Notas tecnicas (deuda tecnica, issues conocidos)
```

Sé conciso: tablas sobre prosa. Omite los valores de configuración — solo estructura y claves.

### context-detail.md (~400 líneas máx)

El archivo de "detalle" se adapta al tipo de proyecto:
- **Backend (Go, Python, Rust):** flujos de endpoint (handler → service → repo)
- **Frontend (React):** flujos de módulo/feature (routes, components, API calls, state)
- **Mobile (Flutter):** flujos de pantalla (widgets, BLoC/providers, API calls)

```markdown
# {project-name} — Flujos detallados

### POST /endpoint-name (o nombre de Screen/Module)
- Entry: {file}:{line} → descripción
- Logic: {file}:{line} → qué hace
- Data: {file}:{line} → queries, API calls, cache
- External: {url/service} (timeout, protocol)
- Side effects: events, notifications
```

Mantén cada flujo entre 5 y 15 líneas. Resume los similares.

### context-risks.md (~100 líneas máx)

```markdown
# {project-name} — Areas de riesgo para auditoria
## Archivos riesgosos (tabla: archivo, linea, razon)
## Patrones de concurrencia (threads, async, shared state — con archivos)
## Input sin validar (datos que llegan directo a DB/services sin sanitizar)
## Dependencias externas sin proteccion (sin timeout, sin TLS, sin auth)
## Codigo sospechoso (error ignorado, panic sin recover, SQL concat, eval, dangerouslySetInnerHTML)
```

Solo hechos relevantes para seguridad. Sin arquitectura, sin lógica de negocio.

## Presupuesto de líneas

**CRÍTICO:** Los tres archivos combinados NO DEBEN superar las 700 líneas. Si hay muchos endpoints/pantallas, resume los patrones (ej., "los endpoints CRUD siguen todos la misma cadena handler → service → repo").

Esto permite al orquestador inyectar SOLO el archivo relevante en cada prompt de agente, manteniéndolos por debajo de 10k tokens.
