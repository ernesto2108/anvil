# Riesgos y Deuda Técnica — anvil

last_updated: 2026-08-13

## Gotchas operativos

### Build sin tag `fts5`
- **Dónde:** `Makefile` (targets `build`/`install`/`dashboard-build`), `pkg/storage/sqlite.go`
- **Descripción:** compilar sin `-tags fts5` produce un binario que corre pero sin FTS5/sqlite-vec — la capa de memoria/búsqueda queda degradada sin error explícito visible al usuario
- **Workaround:** usar siempre `make build`/`make install` en vez de `go build` directo sin flags

### Digest de terminal se omite sin proveedor LLM
- **Dónde:** `internal/cli/run.go:552`
- **Descripción:** si ni `ANTHROPIC_API_KEY` ni Ollama están disponibles, el resumen de terminal se salta silenciosamente (solo un `output.Warn`)
- **Workaround:** revisar logs de warning si se espera un digest y no aparece

## Deuda técnica

### Archivos candidatos a refactor

| Archivo | Líneas | Razón |
|---------|--------|-------|
| `internal/tui/browse_test.go` | 977 | Archivo de test más grande del repo — posible sobre-cobertura concentrada o falta de split |
| `internal/instrumentation/instrumentation_test.go` | 937 | Test grande — candidato a split por escenario |
| `internal/cli/registry.go` | 928 | Posible violación de SRP — mezcla fetch, parseo e instalación de registry en un solo archivo |
| `internal/mcp/orchestration_test.go` | 898 | Test grande |
| `internal/tui/browse.go` | 851 | Posible violación de SRP en la capa TUI |
| `internal/dashboard/query/query_test.go` | 830 | Test grande |
| `internal/cli/emit_translate.go` | 781 | Candidato a split (traducción de eventos + lógica de emisión) |
| `internal/cli/run.go` | 762 | Mezcla selección de proveedor LLM, ejecución de run y digest — candidato a split |
| `internal/mcp/context.go` | 632 | Mezcla parseo de handoff, digest y tools de contexto — candidato a split |
| `internal/deploy/integration_test.go` | 630 | Test grande |
| `internal/orchestrator/executor.go` | 557 | Núcleo del executor — revisar antes de agregar más responsabilidades |

### TODOs y FIXMEs con impacto

```bash
# Detectar con:
grep -rn "TODO\|FIXME\|HACK\|XXX" --include="*.go" | grep -v "_test"
```

No se detectaron TODOs/FIXMEs/HACKs con impacto en código de aplicación — solo 2 ocurrencias de la palabra "TODO", ambas como literal de string en `internal/mcp/context.go` (nombres de secciones de un documento de backlog, no marcadores de deuda técnica).

## Restricciones conocidas

- **Dashboard requiere macOS + CGO:** `make dashboard-build` usa `CGO_ENABLED=1` y `CGO_LDFLAGS="-framework UniformTypeIdentifiers"` — solo soportado en macOS por ahora (comentario explícito en `Makefile`).
- **Sin `.golangci.yml`:** no hay linter configurado más allá de `go vet` — gap de calidad, considerar agregar golangci-lint si el equipo lo requiere.

## Dependencias frágiles

- **CGO (sqlite3, sqlite-vec):** el driver `mattn/go-sqlite3` y `sqlite-vec-go-bindings/cgo` requieren CGO habilitado — builds sin CGO (cross-compile puro) probablemente fallan o degradan funcionalidad.
- **Ollama como fallback:** depende de que el usuario tenga Ollama corriendo localmente; sin health-check explícito confirmado más allá de lo mencionado en `run.go` (`if Ollama healthy`).

## Áreas sin tests

- `pkg/output/rest/client.go` — no se confirmó archivo de test asociado en el sondeo grep-first — gap, verificar en rescan `deep`.
- `internal/deploy/`, `internal/instrumentation/`, `internal/runner/` — cobertura no confirmada en detalle en este bootstrap (presupuesto de líneas del scan) — profundizar si una tarea los toca.
