# MCP SDK Reference

Snippets idiomáticos, versiones exactas, transports, testing y distribución por lenguaje. Verificado 2026-07-18 contra fuentes oficiales; los ítems marcados ⚠️ no se re-verificaron snippet a snippet — confirmar contra el repo/docs del SDK antes de fijarlos.

## Índice
- [TypeScript](#typescript)
- [Python](#python)
- [Go](#go)
- [Testing](#testing)
- [MCP Inspector y Claude Code](#mcp-inspector-y-claude-code)
- [Distribución](#distribución)

## TypeScript

Paquete: `@modelcontextprotocol/sdk` — estable **v1.29.0** (fijar `<2`; v2 beta implementa spec 2026-07-28 y divide en `@modelcontextprotocol/server` + `/client`).

En **v1.x** los imports son:

```typescript
import { McpServer } from '@modelcontextprotocol/sdk/server/mcp.js';
import { StdioServerTransport } from '@modelcontextprotocol/sdk/server/stdio.js';
import { z } from 'zod';

const server = new McpServer({ name: 'greeting-server', version: '1.0.0' });

server.registerTool(
  'greet',
  {
    description: 'Greet someone by name. Use when the user asks for a greeting.',
    inputSchema: z.object({ name: z.string().describe('the person to greet') }),
  },
  async ({ name }) => ({
    content: [{ type: 'text', text: `Hello, ${name}!` }],
  })
);

const transport = new StdioServerTransport();
await server.connect(transport);
```

- `registerResource` y `registerPrompt` siguen la misma forma (nombre, config, handler).
- Structured output: declarar `outputSchema` en la config del tool y devolver `structuredContent`.
- HTTP: `StreamableHTTPServerTransport` (en v2, middlewares `@modelcontextprotocol/express` / `/hono`).
- ⚠️ La forma exacta de `registerResource`/`outputSchema` en v1 no se re-verificó snippet a snippet (docs devolvieron 404). Referencia canónica: https://github.com/modelcontextprotocol/typescript-sdk y https://ts.sdk.modelcontextprotocol.io/.

## Python

Paquete oficial: `mcp` — estable **1.28.1** (Python ≥3.10, soporta 3.10–3.14). Fijar `mcp<2` (v2.0.0b1 es pre-release con API `from mcp.server import MCPServer`). Alternativa: `fastmcp` v3.4.4 standalone.

```python
from mcp.server.fastmcp import FastMCP

mcp = FastMCP("Demo")

@mcp.tool()
def add(a: int, b: int) -> int:
    """Add two numbers."""
    return a + b

@mcp.resource("greeting://{name}")
def greeting(name: str) -> str:
    """Greet someone by name."""
    return f"Hello, {name}!"

@mcp.prompt()
def review_code(code: str) -> str:
    return f"Please review this code:\n\n{code}"

if __name__ == "__main__":
    mcp.run(transport="stdio")            # o transport="streamable-http" (monta en /mcp)
```

- Los **type hints SON el schema** — no escribir JSON Schema a mano. El retorno tipado genera structured output.
- `transport="sse"` existe por compatibilidad — no usarlo en código nuevo.

## Go

Paquete oficial: `github.com/modelcontextprotocol/go-sdk` (mantenido con Google). **v1.4.0+** soporta spec 2025-11-25; **v1.7.0+** apunta al RC 2026-07-28 manteniendo compat. Paquetes: `.../mcp` (núcleo), `.../auth`, `.../oauthex`.

```go
type GreetInput struct {
    Name string `json:"name" jsonschema:"the name of the person to greet"`
}

func SayHi(ctx context.Context, req *mcp.CallToolRequest, input GreetInput) (
    *mcp.CallToolResult, any, error) {
    return &mcp.CallToolResult{
        Content: []mcp.Content{&mcp.TextContent{Text: "Hi " + input.Name}},
    }, nil, nil
}

server := mcp.NewServer(&mcp.Implementation{Name: "greeter", Version: "v1.0.0"}, nil)
mcp.AddTool(server, &mcp.Tool{Name: "greet", Description: "say hi"}, SayHi)
server.Run(context.Background(), &mcp.StdioTransport{})
```

- El schema JSON se genera desde el struct tipado (tags `json` + `jsonschema`).
- ⚠️ La firma del handler (tres valores de retorno) y el nombre del handler HTTP (`mcp.NewStreamableHTTPHandler`) provienen de conocimiento del SDK v1 no re-verificado; confirmar en https://pkg.go.dev/github.com/modelcontextprotocol/go-sdk/mcp.

## Testing

Testear tres niveles: (1) el handler como función pura, (2) el contrato (schema generado, input inválido → `isError`), (3) wiring end-to-end con el cliente del SDK.

- **Python**: cliente in-memory sin subprocess:
  ```python
  async with Client(mcp) as client:
      result = await client.call_tool("add", {"a": 1, "b": 2})
  ```
- **TypeScript**: `InMemoryTransport` conecta un `Client` al `McpServer` en el mismo proceso (Vitest/Jest). ⚠️ Nombre `InMemoryTransport` de memoria; verificar en el repo.
- **Go**: `mcp.CommandTransport` (lanza el server como comando) o transportes in-memory; handlers tipados se testean como funciones normales.

## MCP Inspector y Claude Code

Herramienta oficial de debugging (UI web + modo CLI):

```bash
npx @modelcontextprotocol/inspector node path/to/server/index.js       # TS local
npx @modelcontextprotocol/inspector uv --directory path run package    # Python local
npx -y @modelcontextprotocol/inspector npx @modelcontextprotocol/server-filesystem ~/Desktop
```

Workflow: conectar → verificar negociación de capabilities → probar tabs Resources/Prompts/Tools con inputs custom → iterar (rebuild + reconnect) → probar edge cases (inputs inválidos, args faltantes, errores). En Python v1: `mcp dev server.py` (con `mcp[cli]`) lanza el server bajo el Inspector.

Registrar en Claude Code para probar localmente (el `--` separa el comando del server; opciones antes del nombre):

```bash
claude mcp add my-server -- node ./dist/index.js
claude mcp add --transport http my-remote https://example.com/mcp
claude mcp list          # estado de conexión
```

Scopes: `local` (default, privado), `user` (todas tus sesiones), `project` (escribe `.mcp.json` commiteable en la raíz).

## Distribución

| Stack | Formato | Ejecución en cliente |
|---|---|---|
| TypeScript | paquete npm con `bin` | `npx -y <paquete>` |
| Python | paquete PyPI | `uvx <paquete>` |
| Go | binario estático (releases GitHub, `go install`) | ruta al binario |
| Cualquiera | imagen Docker/OCI | `docker run -i --rm <imagen>` |
| Local one-click | bundle `.mcpb` | doble-click en Claude Desktop |

- **MCP Registry** (https://registry.modelcontextprotocol.io): catálogo oficial. Publicar con `server.json` (namespace verificado `io.github.<user>/<server>`, ubicación del artefacto, args, env vars) usando la CLI `mcp-publisher`. ⚠️ En preview / pre-GA a 2026-07-18 — verificar estado antes de fijarlo.
- **Bundles `.mcpb`** (antes DXT): zip con server + dependencias + `manifest.json`, instalación one-click en Claude Desktop. Repo: `modelcontextprotocol/mcpb`.
