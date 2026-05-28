# Bootstrap — Generar `.project-context/` desde cero

Usado por `context-init` en `mode: init` o `mode: deep` cuando `.project-context/` no existe o tiene `coverage: none`.

## Paso 1 — Leer estructura base

Ya hecho por scan-project pasos 1-3. Reutilizar esa información.

## Paso 2 — Inferir patrones de diseño

No buscar nombres. Buscar **firmas estructurales**. Un patrón puede llamarse de cualquier manera o no tener nombre.

### Go

**Factory**
```
grep -r "func New[A-Z]" --include="*.go" | grep -v "_test.go"
```
Señal positiva: la función retorna una interfaz o puntero a struct, tiene switch/if sobre string o tipo enum que construye distintos concretos.
```go
func NewSummarizer(kind string) Summarizer { switch kind { ... } }
```

**Builder**
```
grep -rn "func.*With[A-Z].*{" --include="*.go"
grep -rn "func.*Build\(\)" --include="*.go"
```
Señal positiva: métodos `With*` que retornan el mismo tipo receptor + un `Build()` final.

**Strategy**
```
grep -rn "^type [A-Z][a-zA-Z]* interface" --include="*.go"
```
Luego para cada interfaz con 1-3 métodos: buscar cuántas structs la implementan.
Señal positiva: ≥ 2 structs implementando la misma interfaz pequeña + un struct que la embebe como campo.
```go
type Summarizer interface { Summarize(ctx, text) (string, error) }
// claude.go, ollama.go, haiku.go → Strategy
```

**Repository**
```
grep -rn "func.*\(ctx context\.Context" --include="*.go" | grep -E "Get|List|Create|Update|Delete|Save|Find"
```
Señal positiva: struct con múltiples métodos CRUD, todos reciben `context.Context`, acceden a DB o storage.

**Singleton / shared state**
```
grep -rn "sync\.Once" --include="*.go"
grep -rn "^var [a-zA-Z]* \*" --include="*.go"
```
Señal positiva: `sync.Once` + variable de paquete + función `Get/Instance/Shared`.

**Middleware / Decorator**
```
grep -rn "func(.*Handler.*) .*Handler" --include="*.go"
grep -rn "func(.*Middleware" --include="*.go"
```
Señal positiva: función que recibe y retorna el mismo tipo funcional.

**Observer / Event**
```
grep -rn "chan " --include="*.go" | grep -v "_test.go"
grep -rn "Subscribe\|Publish\|Emit\|Notify\|Listen" --include="*.go"
```

**Pipeline / Chain of responsibility**
```
grep -rn "\.Next\(\)\|\.Handle\(\)\|\.Process\(\)" --include="*.go"
```
Señal positiva: structs encadenados que pasan datos al siguiente.

**Functional options**
```
grep -rn "type Option func\|type.*Option.*func" --include="*.go"
```

### TypeScript / React

**Factory**
```
grep -rn "function create[A-Z]\|const create[A-Z]" --include="*.ts" --include="*.tsx"
grep -rn "switch.*return new\|switch.*return {" --include="*.ts"
```

**Strategy**
```
grep -rn "interface [A-Z].*{" --include="*.ts" | head -30
```
Señal positiva: múltiples clases/objetos implementando la misma interfaz pequeña.

**Repository**
```
grep -rn "class [A-Z]*Repository\|class [A-Z]*Store\|class [A-Z]*Service" --include="*.ts"
```

**Observer / hooks reactivos**
```
grep -rn "useCallback\|useEffect\|EventEmitter\|\.on(\|\.emit(" --include="*.ts" --include="*.tsx"
```

**Builder (fluent API)**
```
grep -rn "\.set[A-Z]\|\.with[A-Z]\|\.add[A-Z]" --include="*.ts" | grep "return this"
```

**Singleton**
```
grep -rn "export const [a-z].*=.*new [A-Z]" --include="*.ts"
```

**HOC / Decorator (React)**
```
grep -rn "export.*function with[A-Z]\|export.*const with[A-Z]" --include="*.tsx" --include="*.ts"
```

### Python

**Factory**
```
grep -rn "def create_\|def make_\|def build_" --include="*.py"
grep -rn "cls_map\|registry\|handlers = {" --include="*.py"
```

**Strategy**
```
grep -rn "class [A-Z].*ABC\|class [A-Z].*Protocol" --include="*.py"
grep -rn "@abstractmethod" --include="*.py"
```

**Repository**
```
grep -rn "class [A-Z]*Repository\|class [A-Z]*Store\|class [A-Z]*DAO" --include="*.py"
```

**Observer**
```
grep -rn "\.subscribe\|\.publish\|\.emit\|signal\." --include="*.py"
```

**Decorator (Python nativo)**
```
grep -rn "^def [a-z].*:\n.*def wrapper" --include="*.py"
grep -rn "@[a-z_]*\ndef " --include="*.py"
```

## Paso 3 — Inferir contratos

### Endpoints REST (Go)
```
grep -rn "\.Get(\|\.Post(\|\.Put(\|\.Delete(\|\.Patch(" --include="*.go" | grep -v "_test.go"
grep -rn "router\.\|mux\.\|r\." --include="*.go" | grep -E '"/'
```

### Endpoints REST (TypeScript/Express/Next)
```
grep -rn "app\.get\|app\.post\|router\.get\|router\.post" --include="*.ts"
grep -rn "export.*GET\|export.*POST\|export.*PUT" --include="*.ts" --include="*.tsx"
```

### Message queues / eventos
```
grep -rn "Publish\|Subscribe\|Consume\|Produce\|SendMessage\|ReceiveMessage" --include="*.go" --include="*.ts" | grep -v "_test"
grep -rn "nats\.\|rabbitmq\.\|kafka\.\|redis\.X\|streams\." --include="*.go" --include="*.ts"
```

### Servicios externos
```
grep -rn "http\.NewRequest\|http\.Get\|http\.Post\|NewClient\|grpc\.Dial" --include="*.go"
grep -rn "axios\.\|fetch(\|new.*Client(" --include="*.ts"
```

### WebSockets
```
grep -rn "websocket\.\|ws\.New\|Upgrade\|gorilla/websocket" --include="*.go"
grep -rn "WebSocket\|io\.connect\|socket\.on" --include="*.ts"
```

## Paso 3.5 — Extraer comandos operativos

Leer los archivos de operación del proyecto y extraer los comandos reales. Este paso produce `Core/workflows.md`.

### Makefile

```bash
# Listar todos los targets con sus comentarios
grep -E "^[a-zA-Z_-]+:.*?##" Makefile | sed 's/:.*//' | head -40
# Si no tienen ## comments, listar targets sin descripción
grep -E "^[a-zA-Z_-]+:" Makefile | grep -v "^\." | sed 's/:.*//' | head -40
```

Leer el contenido de los targets más relevantes (build, test, dev, lint, migrate, run) para entender exactamente qué ejecutan. No asumir — leer el recipe del target.

```bash
# Ver recipe de un target específico
grep -A 5 "^build:" Makefile
grep -A 5 "^test:" Makefile
grep -A 5 "^dev:" Makefile
```

### docker-compose.yml / docker-compose.yaml

```bash
# Ver servicios definidos
grep -E "^  [a-zA-Z_-]+:" docker-compose.yml | sed 's/://'
# Ver puertos expuestos
grep -A 2 "ports:" docker-compose.yml
# Ver dependencias entre servicios
grep -A 3 "depends_on:" docker-compose.yml
```

Leer el archivo completo si tiene < 100 líneas. Para archivos más grandes, leer solo los servicios y sus configuraciones de ports/volumes/env.

### package.json (scripts)

```bash
# Extraer todos los scripts definidos
cat package.json | grep -A 50 '"scripts"' | grep -B 1 -A 1 '":"'
```

### scripts/ directory

```bash
ls scripts/ 2>/dev/null
# Leer cada script para entender qué hace (head -10 de cada uno)
```

### Variables de entorno

```bash
# Detectar qué variables se esperan
ls .env.example .env.sample .env.template 2>/dev/null
cat .env.example 2>/dev/null | head -40
# En Go: buscar os.Getenv y viper/env
grep -rn "os\.Getenv\|viper\.Get\|env\.Get" --include="*.go" | grep -v "_test" | head -20
# En Node: buscar process.env
grep -rn "process\.env\." --include="*.ts" --include="*.js" | grep -v "node_modules" | head -20
```

### Dockerfile

```bash
# Ver el ENTRYPOINT y CMD — revelan cómo se lanza en producción
grep -E "^ENTRYPOINT|^CMD|^EXPOSE|^FROM" Dockerfile 2>/dev/null
```

## Paso 4 — Detectar bounded contexts

```bash
ls internal/    # Go — cada carpeta es un dominio candidato
ls src/         # React — carpetas de features
ls lib/         # Flutter
```

Un dominio es relevante para `.project-context/domains/` si tiene > 3 archivos significativos y lógica de negocio propia (no es util, config, o types puros).

## Paso 5 — Detectar SOLID

**SRP:** archivos con > 300 líneas son candidatos a violar SRP. Reportar los top 5.
```
find . -name "*.go" -not -path "*/vendor/*" | xargs wc -l | sort -rn | head -10
```

**OCP:** buscar switch/if sobre tipos donde agregar un nuevo tipo requeriría modificar el código existente.
```
grep -rn "switch.*\.(type)\|switch.*kind\|switch.*type" --include="*.go" | grep -v "_test"
```

**DIP:** buscar dependencias directas a structs concretos vs interfaces.
```
grep -rn "func New[A-Z].* \*[A-Z]" --include="*.go"  # retorna concreto — posible violación
grep -rn "func New[A-Z].* [A-Z][a-z]*er\b" --include="*.go"  # retorna interfaz — bien
```

## Paso 6 — Consultar MCP memory antes de escribir decisions/

Antes de escribir archivos, si `mcp__anvil__search_memories` está disponible:

```
mcp__anvil__search_memories(
  query = "decisiones arquitectónicas " + <nombre del proyecto>,
  limit = 5
)
```

Filtrar resultados con `score >= 0.5`. Por cada hit relevante:
- Extraer el campo `decisions` del digest
- Crear un archivo `decisions/NNN-slug.md` por cada decisión con evidencia clara
- Citar el run de origen: `Evidencia: digest run-<ID>, hace X días`

Si no hay hits o MCP no está disponible → dejar `decisions/` vacío (se llena on-demand).

**Esto no reemplaza el análisis de código** — complementa. El código dice qué existe, la memoria dice por qué se decidió así.

## Paso 7 — Escribir archivos

**Idioma obligatorio:** todo el contenido generado debe estar en español. Esto incluye encabezados, descripciones, notas, comentarios, listas, evidencias y cualquier texto narrativo. Los identificadores técnicos (nombres de archivos, funciones, paquetes, comandos, paths, snippets de código) se preservan tal como aparecen en el repo. Si un template trae encabezados en inglés, traducirlos antes de escribir.

Usar los templates en `templates/`. Escribir en este orden:
1. `Technical domain/project.md`
2. `Core/workflows.md` — comandos operativos reales extraídos en el Paso 3.5
3. `Core/coding-standards.md` — patrones detectados
4. `Technical domain/contracts.md` — APIs + invariantes de negocio
5. `Technical domain/dependencies.md`
6. `Technical domain/domain.md` — dominios con > 3 archivos significativos
7. `Technical domain/glossary.md` — pre-populado con entidades detectadas, marcadas como `⚠️ pendiente validación`
8. `Technical domain/risks.md` — incluir top-5 archivos > 300 líneas como deuda potencial
9. `decisions/NNN-slug.md` — solo los que tienen evidencia del Paso 6
10. `Core/navigation.md` — índice de Core
11. `Technical domain/navigation.md` — índice de Technical domain
12. `NAVIGATOR.md` — al final, con el índice general

Marcar `coverage: bootstrap` en NAVIGATOR.md.

## Lo que NO hacer en bootstrap

- No inferir patrones de nombres de archivos solos — leer al menos la firma de las funciones clave
- No crear un `domains/<name>.md` para paquetes de < 3 archivos
- No inventar decisiones — solo las que tienen evidencia en código o en MCP memory
- No sobreescribir si `.project-context/` ya existe con `coverage: full` — en ese caso solo actualizar secciones stale
