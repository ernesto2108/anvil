# Dominio: <name>

last_updated: <YYYY-MM-DD>

## Responsabilidad

<Qué hace este dominio — 1-2 líneas. Una sola responsabilidad.>

## Archivos clave

```
<path>/
├── <archivo>.go   — <rol>
├── <archivo>.go   — <rol>
└── <subdir>/      — <descripción>
```

## Flujo principal

<Secuencia de pasos del flujo más importante>

```
<Trigger> → <paso1> → <paso2> → <paso3> → <Resultado>
```

## Patrones usados

- **<Patrón>:** `<archivo>` — <qué varía o qué encapsula>
- **<Patrón>:** `<archivo>` — <descripción>

## Interfaces públicas

```go
// o TypeScript, Python según stack
type <Name> interface {
    <Method>(ctx context.Context, ...) (<ReturnType>, error)
}
```

## Dependencias de este dominio

- `<package interno>` — <para qué>
- `<servicio externo>` — <para qué>

## Quién depende de este dominio

- `<package>` — <cómo lo usa>

## Decisiones tomadas

<!-- Solo decisiones con evidencia — handoff, SPEC, o comentario explícito -->
- D1: <decisión> — ver `decisions/<NNN>-<slug>.md`

## Gotchas

<!-- Comportamientos sorprendentes, edge cases conocidos, restricciones operativas -->
- <gotcha> — <contexto o workaround>

## Deuda técnica

<!-- Solo si hay evidencia real — archivos > 300 líneas, TODOs, FIXMEs -->
- <item> — `<archivo>:<línea>`
