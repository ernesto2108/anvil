# Dominio: memory

last_updated: 2026-05-08

## Responsabilidad

Capturar, resumir, embeber y buscar los outputs de los runs de agentes AI. Es el sistema de memoria semántica de Anvil: genera `Digest`s estructurados (resumen, decisiones, edge cases, errores), los almacena con embeddings vectoriales, y permite búsqueda por similitud.

## Archivos clave

```
internal/memory/
├── digest.go              — entidades de dominio: Digest, Embedder, Summarizer interfaces
├── store.go               — persistencia de digests en SQLite + búsqueda vectorial
├── search.go              — búsqueda semántica sobre digests
├── similarity.go          — cálculo de similitud coseno
├── handoff_digest.go      — parsing de archivos .handoff/<TASK-ID>.md a Digest
├── encoding.go            — serialización de embeddings (float32 ↔ blob)
├── capture/
│   └── orchestrator.go    — orquestador de captura: observa transcripts, llama summarizer
├── transcript/
│   ├── parser.go          — parser de transcripts de Claude Code (JSONL)
│   ├── browse.go          — navegación de transcripts
│   └── transcript.go      — tipos de transcript
├── ollama/
│   ├── client.go          — Embedder + HTTP client para Ollama
│   ├── summarizer.go      — Summarizer via Ollama /api/chat
│   └── lifecycle.go       — health check de Ollama local
├── haiku/
│   └── client.go          — Summarizer via Anthropic API (Claude Haiku)
└── claude/
    └── client.go          — cliente alternativo (claude CLI directo)
```

## Flujo principal

```
Run termina → capture/orchestrator detecta transcript nuevo
→ transcript/parser parsea JSONL de Claude Code
→ Summarizer.Summarize() genera DigestDraft (summary, decisions, edge cases, errors)
→ Embedder.Embed() genera vector float32
→ store.Save() persiste Digest + embedding en SQLite
→ search.Search() permite recuperar digests similares en futuros runs
```

## Patrones usados

- **Strategy:** `Embedder` / `Summarizer` interfaces — `internal/memory/digest.go:46,51`
- **Strategy implementada por:** `ollama.Client` (Embedder + Summarizer), `haiku.Client` (Summarizer)
- **Singleton:** `sync.Once` en tests (`store_test.go:16`) para sqlite-vec

## Interfaces públicas

```go
// internal/memory/digest.go
type Embedder interface {
    Embed(ctx context.Context, text string) ([]float32, error)
}

type Summarizer interface {
    Summarize(ctx context.Context, agentOutputs []AgentOutput) (DigestDraft, error)
}
```

## Dependencias de este dominio

- `pkg/storage` — `OpenDB` para SQLite + sqlite-vec
- `internal/memory/ollama/` — Embedder y Summarizer locales
- `internal/memory/haiku/` — Summarizer cloud (Anthropic)

## Quién depende de este dominio

- `internal/cli/` — llama al store, summarizer y embedder en `run.go`, `digest_from_handoff.go`, `digests.go`
- `internal/mcp/context.go` — búsqueda de digests para contexto de herramientas MCP
- `internal/memory/capture/orchestrator.go` — pipeline de captura automática

## Gotchas

- El embedding es `nil` cuando no se pudo generar (Ollama no disponible) — el Digest se guarda sin vector y no aparece en búsquedas semánticas
- `handoff_digest.go` parsea el formato Markdown de archivos `.handoff/<TASK-ID>.md` — el formato es rígido, los cambios en el template del handoff rompen el parser
- Ollama debe estar corriendo localmente en `http://localhost:11434` para embeddings — sin él, la captura falla silenciosamente

## Deuda técnica

- `internal/memory/transcript/parser.go` tiene switch sobre `b.Type` (`parser.go:140`) — agregar nuevos tipos de bloque requiere editar el switch (OCP en riesgo)
