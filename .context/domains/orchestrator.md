# Dominio: orchestrator

last_updated: 2026-05-08

## Responsabilidad

Ejecutar pipelines multi-agente definidos como DAGs (Directed Acyclic Graphs). Gestiona la topología de dependencias entre agentes, el control de concurrencia, los gates de aprobación humana, las políticas de fallo, y la emisión de eventos de instrumentación.

## Archivos clave

```
internal/orchestrator/
├── types.go       — entidades: Node, DAG, AgentResult, RunState, NodeStatus, FailPolicy, interfaces AgentRunner/EventSink/GateHandler
├── dag.go         — construcción y validación del DAG desde definición de pipeline
├── executor.go    — ejecución concurrente del DAG con semáforo, channels de resultado
├── gate.go        — GateHandler: CLIGateHandler (interactivo) + AutoApproveHandler (CI)
├── replanner.go   — lógica de replanificación tras fallos
```

## Flujo principal

```
Pipeline YAML → dag.Build() → DAG validado
→ executor.Run(dag, runner, sink)
→ topological sort → nodos en cola según InDegree
→ goroutines paralelas (semáforo: chan struct{})
→ AgentRunner.RunAgent(ctx, node, upstream) por nodo
→ Gate.Handle() si node.Gate=true → espera decisión humana
→ EventSink.Emit() en cada transición de estado
→ RunState final con Statuses, Results, Gates
```

## Patrones usados

- **Strategy — AgentRunner:** `types.go:109` — ejecutor intercambiable (ClaudeRunner, runners de deploy)
- **Strategy — GateHandler:** `gate.go:19` — `CLIGateHandler` vs `AutoApproveHandler`
- **Strategy — EventSink:** `types.go:114` — desacopla instrumentación del executor
- **Pipeline / DAG:** `dag.go` + `executor.go` — concurrencia con `chan workerResult` y semáforo `chan struct{}`
- **Observer implícito:** results chan + goroutine collector en `executor.go:265`

## Interfaces públicas

```go
// internal/orchestrator/types.go
type AgentRunner interface {
    RunAgent(ctx context.Context, node Node, upstream map[string]AgentResult) (AgentResult, error)
}

type EventSink interface {
    Emit(ev instrumentation.Event)
}

// internal/orchestrator/gate.go
type GateHandler interface {
    Handle(ctx context.Context, node Node, state RunState) (GateDecision, string, error)
}
```

## Roles válidos de agente

`pm`, `architect`, `designer`, `developer`, `tester`, `qa`, `devops`, `dba`, `security` — definidos en `types.go:ValidRoles`; roles fuera del set son rechazados en load time.

## Dependencias de este dominio

- `internal/instrumentation` — tipos de Event para EventSink
- `internal/runner/` — ClaudeRunner implementa AgentRunner (inyectado desde cli)

## Quién depende de este dominio

- `internal/cli/pipeline.go` — construye el DAG y lanza el executor
- `internal/cli/run.go` — integra runner, sink y gate para runs orquestados
- `internal/mcp/orchestration.go` — expone ejecución de pipelines vía MCP tools

## Gotchas

- `FailPolicyRetry` está definido pero la lógica de retry vive en `replanner.go` — si el replanner no está conectado, el retry no ocurre
- `node.Timeout` = 0 significa sin límite de tiempo — no hay timeout global por defecto

## Deuda técnica

- `internal/orchestrator/executor.go` (557 líneas) — candidato a refactor; mezcla lógica de scheduling, retry, gate handling y event emission
