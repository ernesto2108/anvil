# Contratos — Anvil

last_updated: 2026-05-08

## REST API

No hay servidor REST público. El único cliente HTTP saliente que Anvil usa es hacia APIs externas (ver Servicios externos abajo).

## MCP Server (JSON-RPC 2.0 sobre stdio)

### Transporte
`stdio` — el servidor MCP se lanza como subproceso desde Claude Code u otro cliente MCP. Protocolo: JSON-RPC 2.0.

### Entrada
- **Archivo:** `internal/mcp/server.go` — `Serve(ctx, cfg, db)`
- **Tools registradas:** construidas en `srv.buildRegistry()` → `internal/mcp/tools.go`
- **Categorías de tools:** orchestration, memory/search, context/vault, inventory, agent/skill management

### Tipos de request/response
```go
// internal/mcp/server.go
type request struct {
    JSONRPC string           `json:"jsonrpc"`
    ID      *json.RawMessage `json:"id,omitempty"`
    Method  string           `json:"method"`
    Params  json.RawMessage  `json:"params,omitempty"`
}
type response struct {
    JSONRPC string           `json:"jsonrpc"`
    ID      *json.RawMessage `json:"id,omitempty"`
    Result  any              `json:"result,omitempty"`
    Error   *rpcError        `json:"error,omitempty"`
}
```

### Variables de entorno para el servidor MCP
- `ANVIL_VAULT_PATH` — path al vault Obsidian (opcional); afecta herramientas de contexto en `internal/mcp/context.go`

## Servicios externos

### Ollama (embeddings + summarización local)
- **Cliente:** `internal/memory/ollama/client.go` (`Client`)
- **Endpoints:** `POST <baseURL>/api/embed`, `POST <baseURL>/api/chat`, `GET <baseURL>` (health)
- **Interfaz implementada:** `memory.Embedder`, `memory.Summarizer`
- **Default baseURL:** `http://localhost:11434`
- **Auth:** ninguna
- **Timeout:** no configurado explícitamente — usa http.DefaultClient
- **Usado en:** `internal/cli/run.go:206`, `internal/cli/digest_from_handoff.go:108`, `internal/mcp/context.go:63,176`

### Anthropic API / Claude Haiku (summarización cloud)
- **Cliente:** `internal/memory/haiku/client.go` (`Client`)
- **Endpoint:** `POST https://api.anthropic.com/v1/messages`
- **Interfaz implementada:** `memory.Summarizer`
- **Auth:** `x-api-key: $ANTHROPIC_API_KEY`, `anthropic-version: 2023-06-01`
- **Usado en:** `internal/cli/run.go:376` (fallback cuando `ANTHROPIC_API_KEY` está presente)

### Claude Code CLI (ejecución de agentes)
- **Mecanismo:** subprocess `claude --print --agent <role> --permission-mode acceptEdits -p <prompt>`
- **Implementado en:** `internal/runner/runner.go` (`ClaudeRunner.RunAgent`)
- **Env vars propagados:** `ANVIL_PARENT_RUN_ID`, `ANVIL_AGENT_ID`
- **Otros runners:** `internal/deploy/` — `claude.go`, `codex.go`, `gemini.go`, `opencode.go`, `cursor.go`

## Instrumentación — Event Schema

### Evento envelope
```go
// internal/instrumentation/events.go
type Event struct {
    SchemaVersion string          `json:"schema_version"` // "1"
    EventID       string          `json:"event_id"`
    RunID         string          `json:"run_id"`
    EventType     string          `json:"event_type"`
    Timestamp     time.Time       `json:"timestamp"`
    Payload       json.RawMessage `json:"payload"`
}
```

### Event types definidos
`run.start`, `run.end`, `agent.start`, `agent.end`, `agent.error`, `file.touched`, `qa.score`, `orchestrator.start`, `orchestrator.gate`, `tool.use`, `run.error`, `task.created`, `task.completed`, `context.compacted`, `permission.denied`, `user.prompt`

### Flujo de escritura
`Emitter` (chan interno) → `EventWriter` (interfaz) → SQLite via `internal/instrumentation/writer/`

## Contratos internos entre dominios

### memory.Embedder
- **Definida en:** `internal/memory/digest.go:46`
- **Implementada por:** `internal/memory/ollama/` (Client)
- **Consumida por:** `internal/cli/`, `internal/mcp/context.go`

### memory.Summarizer
- **Definida en:** `internal/memory/digest.go:51`
- **Implementada por:** `internal/memory/ollama/` (Summarizer), `internal/memory/haiku/` (Client)
- **Consumida por:** `internal/memory/capture/orchestrator.go`, `internal/cli/`

### orchestrator.AgentRunner
- **Definida en:** `internal/orchestrator/types.go:109`
- **Implementada por:** `internal/runner/` (ClaudeRunner)
- **Consumida por:** `internal/orchestrator/executor.go`

### orchestrator.EventSink
- **Definida en:** `internal/orchestrator/types.go:114`
- **Implementada por:** `internal/instrumentation/` (Emitter)
- **Consumida por:** `internal/orchestrator/executor.go`

### orchestrator.GateHandler
- **Definida en:** `internal/orchestrator/gate.go:19`
- **Implementada por:** `CLIGateHandler`, `AutoApproveHandler` (gate.go:32, gate.go:91)
- **Consumida por:** `internal/orchestrator/executor.go`

### pkg/output/rest.Client
- **Definida en:** `pkg/output/rest/client.go:15`
- **Consumida por:** módulos que necesitan POST de resultados a endpoints REST externos
