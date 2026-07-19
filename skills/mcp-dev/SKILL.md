---
name: mcp-dev
description: Convenciones para CONSTRUIR servidores MCP (Model Context Protocol) en TypeScript, Python o Go — diseño de tools, transports (stdio / Streamable HTTP), auth OAuth 2.1, structured output, testing con MCP Inspector y distribución. Úsalo cuando se escriba o revise el código de un servidor MCP, o el usuario mencione "MCP server", "Model Context Protocol", "@modelcontextprotocol/sdk", "FastMCP", "registerTool", "mcp.tool", "AddTool", "StreamableHTTP", "MCP Inspector", "server.json", ".mcpb", o construir tools/resources/prompts para un cliente MCP. NO es para configurar/consumir MCPs de infra en un repo — eso es la skill mcp-setup (genera .mcp.json).
---

# MCP Server Development

Convenciones para construir servidores MCP. Postura por defecto verificada a julio 2026: spec estable **2025-11-25**, con el RC **2026-07-28** (stateless) en camino. Construir contra 2025-11-25 (lo que soportan los SDKs estables) sin acoplarse a features ya deprecadas en el RC.

> **Deslinde con `mcp-setup`:** esta skill es para **construir** servidores MCP (escribir el código del server, sus tools, transport y distribución). La skill `mcp-setup` es para **consumir** MCPs de infra existentes en un repo (generar `.mcp.json`). No confundir dominios.

## Filosofía

- **Pocas tools de alto impacto, no un espejo 1:1 del API REST** — cada tool corresponde a un workflow real que el agente ejecuta (`schedule_event`, no `list_users`+`list_events`+`create_event`). Si un ingeniero no sabe cuál tool elegir, el modelo tampoco.
- **La descripción de la tool es el contrato con el modelo** — escríbela "como se lo explicarías a un empleado nuevo": qué hace, cuándo usarla, qué significa cada parámetro. Refinamientos pequeños producen mejoras medibles grandes.
- **El server no confía en nadie y nadie confía en el server** — todo input al server y todo output que devuelve son datos no confiables; valida audiencia de tokens, sanitiza resultados, nunca hagas token passthrough.
- **Diseño evaluation-driven** — prototipo → evaluar con tareas reales → iterar descripciones y schemas. No pulir a ciegas.

## Selección de SDK y versión (verificado 2026-07-18)

Usar siempre el **SDK oficial** del lenguaje y fijar upper bound en la línea estable:

| Lenguaje | Paquete oficial | Estable | Fijar |
|---|---|---|---|
| TypeScript | `@modelcontextprotocol/sdk` | v1.29.0 | `<2` (v2 beta = spec 2026-07-28) |
| Python | `mcp` (SDK oficial) | 1.28.1 (Py ≥3.10) | `mcp<2` (v2.0.0b1 pre-release) |
| Go | `github.com/modelcontextprotocol/go-sdk` | v1.4.0+ (spec 2025-11-25) | línea v1.x |

Alternativas válidas pero no default: `fastmcp` v3.x standalone (Python, más azúcar), `mark3labs/mcp-go` (Go, precursor del oficial). Para código nuevo, el SDK oficial es la convención.

Ver `sdk-reference.md` para snippets idiomáticos de server + tool/resource/prompt por lenguaje, testing in-memory y comandos de registro en Claude Code.

## Elección de transport

La spec 2025-11-25 define **dos** transports estándar. El viejo HTTP+SSE independiente está **deprecado** (sobrevive solo como mecanismo interno de Streamable HTTP).

| Escenario | Transport |
|---|---|
| Server local (filesystem, sockets locales); distribución npx/uvx/binario | **stdio** |
| Server remoto multi-cliente, SaaS, detrás de un dominio, con OAuth | **Streamable HTTP** (endpoint único POST+GET) |
| `/sse` + `/messages` a secas | Deprecado — solo backcompat |

**Regla de oro de stdio:** el server NUNCA escribe en stdout nada que no sea un mensaje MCP. Un `print()` / `console.log` rompe el protocolo. Todo logging va a **stderr** (permitido para cualquier nivel).

**Streamable HTTP:** header `MCP-Protocol-Version: 2025-11-25` en cada request tras negociar; validar `Origin` (403 si inválido, anti DNS-rebinding); en local bind solo a `127.0.0.1`. No construir lógica de negocio sobre la sesión (`Mcp-Session-Id`) — el RC 2026-07-28 la elimina.

## Diseño de tools

1. **Naming**: descriptivo y sin ambigüedad (`get_current_weather` > `weather`); namespacing con prefijo común cuando hay muchas tools (`asana_search`, `asana_projects_search`).
2. **Schemas de input**: parámetros inequívocos (`user_id`, no `user`); `enum` para valores cerrados; `description` en cada propiedad; `required` solo lo realmente obligatorio; defaults sensatos. JSON Schema 2020-12 es el dialecto por defecto. En Python/Go el schema se **deriva de los type hints/structs** — no lo escribas a mano.
3. **Structured output**: declara `outputSchema` y devuelve `structuredContent` cuando el resultado sea datos (en Python/Go se deriva del tipo de retorno).
4. **Errores**: los errores de ejecución y de validación de input van como **Tool Execution Error** (`isError: true` en el resultado), NO como error de protocolo JSON-RPC — así el modelo se auto-corrige. Mensajes accionables: qué estuvo mal y cómo corregirlo, nunca un traceback opaco.
5. **Eficiencia de tokens**: paginación, filtrado y truncado con defaults sensatos; parámetro `response_format: "concise" | "detailed"`; devolver identificadores útiles para llamadas encadenadas, no blobs JSON completos.
6. **Annotations**: declara `readOnlyHint` / `destructiveHint` / `idempotentHint` / `openWorldHint` para tools destructivas o de mundo abierto (son hints no confiables para el cliente, pero guían al host).

## Auth y seguridad

- **stdio → sin OAuth.** Las credenciales vienen del entorno (env vars, keychain), nunca de argumentos de línea de comandos visibles en `ps`.
- **Streamable HTTP → OAuth 2.1** si se implementa auth (es opcional en MCP). El server actúa como **Resource Server**; publica `/.well-known/oauth-protected-resource` (RFC 9728, MUST); exige PKCE S256 y Resource Indicators (RFC 8707); **valida la audiencia del token en cada request** y **NUNCA acepta ni reenvía tokens de terceros** (token passthrough prohibido).
- Session IDs **no son autenticación**; verificar autorización en cada request.

Ver `security-and-auth.md` para OAuth 2.1 detallado, confused deputy, SSRF en discovery, prompt injection vía resultados de tools y el checklist de seguridad completo.

## Flujo de trabajo

1. **Detectar contexto** — ¿server nuevo o existente? Lenguaje, transport objetivo (local stdio vs remoto HTTP), SDK ya presente. Detectar package manager desde lockfile si es TS. Declarar lo inferido en una línea.
2. **Cargar `sdk-reference.md`** para el lenguaje objetivo antes de escribir el server.
3. **Diseñar el set de tools** — consolidar por workflow, no espejar endpoints. Si el server necesitara >~15 tools, DETENER y proponer consolidación antes de implementar.
4. **Implementar** server + tools/resources/prompts con schemas derivados de tipos. Logging a stderr en stdio.
5. **Gate de transport HTTP:** si el server expone HTTP → validar `Origin`, bind `127.0.0.1` en local, y si hay auth cargar `security-and-auth.md` y seguir OAuth 2.1. Si el diseño requiere token passthrough → DETENER y reportar (es un anti-patrón de seguridad, no negociable).
6. **Testear** con el cliente in-memory del SDK (handler como función + contrato del schema + wiring end-to-end) y con MCP Inspector (`npx @modelcontextprotocol/inspector <comando>`), probando inputs inválidos y edge cases.
7. **Distribuir** según empaquetado (ver `sdk-reference.md`): npm (`npx`), PyPI (`uvx`), binario Go, imagen OCI, o bundle `.mcpb` para Claude Desktop; `server.json` para el MCP Registry.

## Checklist Pre-Implementación

- [ ] SDK oficial del lenguaje, con upper bound en la línea estable (`<2` / `mcp<2` / v1.x)
- [ ] Transport correcto para el escenario (stdio local vs Streamable HTTP remoto); SSE independiente NO
- [ ] En stdio: cero escrituras a stdout que no sean MCP; logging a stderr
- [ ] Tools consolidadas por workflow, no un wrapper 1:1 del API REST
- [ ] Cada tool con descripción accionable y schema con `description` por propiedad
- [ ] Errores de ejecución/validación como `isError: true`, no error de protocolo
- [ ] Structured output (`outputSchema`/tipo de retorno) cuando el resultado son datos
- [ ] HTTP: `Origin` validado, bind `127.0.0.1` en local, `MCP-Protocol-Version` presente
- [ ] Auth (si aplica): OAuth 2.1, audiencia validada, sin token passthrough
- [ ] Sin lógica de negocio acoplada a la sesión HTTP (compat con RC stateless)
- [ ] Probado con cliente in-memory + MCP Inspector, incluyendo inputs inválidos

## Detección de Anti-Patrones

Ver `anti-patterns.md` para la tabla completa con severidades y correcciones.

Señales que siempre deben detener el trabajo:
- Token passthrough (aceptar/reenviar tokens no emitidos para el server) → token-passthrough (error)
- `print()`/`console.log` a stdout en un server stdio → stdout-pollution (error)
- Server HTTP local sin validar `Origin` / sin bind a `127.0.0.1` → dns-rebinding-exposure (error)
- Errores de tool como error de protocolo JSON-RPC en vez de `isError: true` → protocol-error-misuse (warning)
- Una tool por cada endpoint REST → rest-mirroring (warning)
- Transport SSE independiente en código nuevo → deprecated-transport (warning)

## Archivos de Soporte

- `sdk-reference.md` — snippets de server + tool/resource/prompt en TS/Python/Go, versiones exactas, transports por SDK, testing in-memory, MCP Inspector, registro en Claude Code, empaquetado y distribución (MCP Registry, `.mcpb`)
- `security-and-auth.md` — OAuth 2.1 para servers remotos, RFC 9728/8707/PKCE, confused deputy, session hijacking, SSRF en discovery, prompt injection vía resultados de tools, checklist de seguridad
- `anti-patterns.md` — tabla de detección con severidades y correcciones
