# Guia de pruebas: Orchestrator (ORCH-001)

Como probar el orquestador de pipelines paso a paso.

## Prerequisitos

- Go 1.22+
- Claude CLI instalado (`claude --version`)
- Estar en la raiz del proyecto: `cd ~/projects/anvil`

## 1. Compilar el binario

```bash
go build -o ./anvil ./cmd/anvil/
```

Si no compila, hay un problema. Revisar errores.

## 2. Probar con el pipeline de ejemplo (ejecucion real)

Este comando ejecuta los 4 nodos (pm, arch, dev, qa) usando `claude --print`.
Cada nodo invoca Claude de verdad, asi que tarda varios minutos por nodo.

```bash
./anvil run pipelines/example.yaml \
  --task "Crear una funcion Go que valide emails" \
  --auto-approve
```

**Que esperar:**
```
[anvil] Pipeline loaded: 4 nodes, order: [pm -> arch -> dev -> qa]
[anvil] Starting run r_20260412_...
[GATE] Node "arch" (role: architect) -- auto-approved
[GATE] Node "qa" (role: qa) -- auto-approved
[anvil] Run r_... finished: success (Xms)
   v pm -> success
   v arch -> success
   v dev -> success
   v qa -> success
```

**Sin `--auto-approve`** los gates piden input manual:
```bash
./anvil run pipelines/example.yaml \
  --task "Crear una funcion Go que valide emails"
```

Te va a preguntar:
```
[GATE] Node "arch" (role: architect) requires approval.
  Enter decision [approve/reject/skip]:
```

Escribe `approve`, `reject`, o `skip`.

## 3. Probar flags del comando run

### Concurrencia limitada
```bash
./anvil run pipelines/example.yaml \
  --task "test" --auto-approve --concurrency 1
```
Los nodos se ejecutan uno a uno (sin paralelismo).

### Modelo especifico
```bash
./anvil run pipelines/example.yaml \
  --task "test" --auto-approve --model claude-sonnet-4-20250514
```

## 4. Crear tu propio pipeline

Crea un archivo YAML en `pipelines/`. Ejemplo minimo:

```yaml
# pipelines/mi-test.yaml
nodes:
  - id: dev
    role: developer
    timeout: 5m
  - id: tester
    role: tester
    depends_on: [dev]
    timeout: 3m
```

Ejecutar:
```bash
./anvil run pipelines/mi-test.yaml \
  --task "Agregar endpoint GET /health que retorne 200" \
  --auto-approve
```

### Campos disponibles por nodo

| Campo | Tipo | Requerido | Ejemplo |
|-------|------|-----------|---------|
| `id` | string | si | `"dev"` |
| `role` | string | si | `"developer"` |
| `depends_on` | lista | no | `[pm, arch]` |
| `gate` | bool | no | `true` |
| `on_fail` | string | no | `retry`, `skip`, `abort` |
| `timeout` | string | no | `"5m"`, `"30s"`, `"1h"` |

### Roles validos

Solo estos roles son aceptados:

`pm`, `architect`, `designer`, `developer`, `tester`, `qa`, `devops`, `dba`, `security`

Un typo como `role: devloper` da error al cargar:
```
cli/pipeline: node "dev": orchestrator: unknown role "devloper"
```

## 5. Probar deteccion de errores

### Ciclo en dependencias
```yaml
# pipelines/test-cycle.yaml
nodes:
  - id: a
    role: pm
    depends_on: [c]
  - id: b
    role: developer
    depends_on: [a]
  - id: c
    role: qa
    depends_on: [b]
```

```bash
./anvil run pipelines/test-cycle.yaml --task "test"
```

Resultado esperado:
```
build DAG: orchestrator/dag: cycle detected among nodes
```

### Dependencia que no existe
```yaml
# pipelines/test-missing.yaml
nodes:
  - id: dev
    role: developer
    depends_on: [fantasma]
```

```bash
./anvil run pipelines/test-missing.yaml --task "test"
```

Resultado esperado:
```
build DAG: orchestrator/dag: node "dev" depends on unknown node "fantasma"
```

### Role invalido
```yaml
# pipelines/test-bad-role.yaml
nodes:
  - id: x
    role: hacker
```

```bash
./anvil run pipelines/test-bad-role.yaml --task "test"
```

Resultado esperado:
```
cli/pipeline: node "x": orchestrator: unknown role "hacker"
```

### Timeout invalido
```yaml
# pipelines/test-bad-timeout.yaml
nodes:
  - id: x
    role: developer
    timeout: "cinco minutos"
```

```bash
./anvil run pipelines/test-bad-timeout.yaml --task "test"
```

Resultado esperado:
```
cli/pipeline: node "x" invalid timeout "cinco minutos": ...
```

## 6. Probar gate rejection

```yaml
# pipelines/test-gate.yaml
nodes:
  - id: pm
    role: pm
  - id: arch
    role: architect
    depends_on: [pm]
    gate: true
  - id: dev
    role: developer
    depends_on: [arch]
```

```bash
./anvil run pipelines/test-gate.yaml --task "test"
```

Cuando pida aprobacion, escribe `reject`. El pipeline debe abortar sin ejecutar `dev`.

## 7. Probar retry on failure

El nodo `dev` en `example.yaml` tiene `on_fail: retry`. Si claude falla en ese nodo
(por ejemplo, por timeout de red), el orquestador lo reintenta una vez automaticamente.

Para simular esto puedes poner un timeout muy corto:
```yaml
# pipelines/test-timeout.yaml
nodes:
  - id: dev
    role: developer
    timeout: 1s
    on_fail: retry
```

```bash
./anvil run pipelines/test-timeout.yaml \
  --task "Escribe un ensayo de 5000 palabras" \
  --auto-approve
```

Claude no puede responder en 1 segundo, asi que:
1. Primera ejecucion: timeout -> fallo
2. Retry: timeout de nuevo -> pide decision al usuario (o falla con auto-approve)

## 8. Verificar eventos en el dashboard

Despues de cualquier run, los eventos se guardan en `~/.anvil/runs.db`.

Puedes verificar con sqlite3:
```bash
sqlite3 ~/.anvil/runs.db "SELECT event_type, json_extract(payload, '$.agent_id') as agent, json_extract(payload, '$.status') as status FROM events WHERE run_id = (SELECT run_id FROM events ORDER BY timestamp DESC LIMIT 1) ORDER BY timestamp"
```

Deberias ver eventos como:
```
orchestrator.start||
agent.start|pm|
agent.end|pm|success
agent.start|arch|
orchestrator.gate|arch|
agent.end|arch|success
agent.start|dev|
agent.end|dev|success
agent.start|qa|
orchestrator.gate|qa|
agent.end|qa|success
```

O abrir el dashboard:
```bash
./anvil dashboard
```

## 9. Correr los tests unitarios

```bash
go test -race -v ./internal/orchestrator/...
```

Todos deben pasar (14 tests, 33+ subtests).

Para correr todos los tests del proyecto:
```bash
go test -race ./internal/...
```

## Resumen de que probar

| Prueba | Comando | Resultado esperado |
|--------|---------|-------------------|
| Build | `go build -o ./anvil ./cmd/anvil/` | Sin errores |
| Pipeline real | `./anvil run pipelines/example.yaml --task "..." --auto-approve` | 4 nodos success |
| Gate manual | `./anvil run ... (sin --auto-approve)` | Pide input |
| Gate reject | Escribir `reject` en gate prompt | Pipeline aborta |
| Ciclo | Pipeline con dependencias circulares | Error de ciclo |
| Dep faltante | Pipeline con `depends_on` invalido | Error de dep |
| Role invalido | Pipeline con `role: hacker` | Error de role |
| Timeout | Pipeline con `timeout: 1s` | Nodo falla por timeout |
| Retry | Nodo con `on_fail: retry` que falla | Se reintenta 1 vez |
| Tests | `go test -race ./internal/orchestrator/...` | ALL PASS |
| Eventos DB | Query sqlite3 | Eventos agent.start/end presentes |
