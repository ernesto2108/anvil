# Dominio: instrumentation

last_updated: 2026-05-08

## Responsabilidad

Capturar y persistir eventos de telemetría de cada run. Desacopla la emisión de eventos (productores) de la escritura a SQLite (consumidor) via un canal interno con goroutine dedicada.

## Archivos clave

```
internal/instrumentation/
├── events.go      — tipos: Event envelope, EventWriter/Emitter interfaces, todos los payload types, constantes de EventType
├── emitter.go     — implementación del Emitter: canal buffered + goroutine consumidora
├── noop.go        — NoopEmitter para tests y contextos donde no se necesita persistencia
└── writer/        — EventWriter: persistencia en SQLite
```

## Flujo principal

```
Productor (orchestrator, CLI hook) → Emitter.Emit(ev)
→ ch chan Event (buffered)
→ goroutine consumidora → EventWriter.WriteEvent(ev)
→ SQLite (tabla de eventos)
```

## Patrones usados

- **Observer / Channel:** `emitter.go:13` — `ch chan Event` + goroutine; productores no bloquean
- **Strategy — EventWriter:** `events.go:46` — desacopla el emitter de la capa de persistencia
- **Null Object — NoopEmitter:** `noop.go` — para tests sin efectos secundarios

## Interfaces públicas

```go
// internal/instrumentation/events.go
type EventWriter interface {
    WriteEvent(ev Event) error
}

type Emitter interface {
    Emit(ev Event)
    Start()
    Stop()
}
```

## Schema de versión

`SchemaVersion = "1"` — hardcoded en `events.go`; cambios de schema requieren nueva versión y migración de datos.

## Dependencias de este dominio

- `pkg/storage` — SQLite para el writer

## Quién depende de este dominio

- `internal/orchestrator` — EventSink consume Emitter.Emit()
- `internal/cli/` — crea y gestiona el Emitter durante el lifecycle del run
- `internal/mcp/` — emite eventos de herramientas MCP
