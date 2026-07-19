# MCP Server — Detección de Anti-Patrones

Detección pasiva: reportar `error` y `warning` siempre. Detección activa (`suggestion`): solo cuando el usuario pide "improve/refactor/optimize". Formato: `[file:line] [severity] [category] anti-pattern-name`.

| Patrón de Código | Anti-Patrón | Severidad | Categoría | Corrección |
|---|---|---|---|---|
| Server acepta o reenvía tokens no emitidos para él | token-passthrough | error | security | Validar audiencia; el server llama upstream como cliente OAuth propio con token distinto |
| `print()` / `console.log` / stdout en server stdio | stdout-pollution | error | protocol | Enviar todo logging a stderr; stdout solo mensajes MCP |
| Server HTTP local sin validar `Origin` o sin bind `127.0.0.1` | dns-rebinding-exposure | error | security | Validar `Origin` (403 si inválido) y bind a loopback |
| Session ID usado como mecanismo de autenticación | session-as-auth | error | security | Verificar autorización por request; session ID no autentica |
| `redirect_uri` con wildcard en flujo OAuth | wildcard-redirect | error | security | Validación exacta de `redirect_uri` |
| Secretos pasados por args de línea de comandos | secret-in-argv | error | security | Usar env vars / keychain |
| Errores de tool devueltos como error de protocolo JSON-RPC | protocol-error-misuse | warning | design | Devolver `isError: true` en el resultado para permitir auto-corrección del modelo |
| Una tool por cada endpoint REST (wrapper 1:1) | rest-mirroring | warning | design | Consolidar tools por workflow que el agente realmente ejecuta |
| Transport SSE independiente (`/sse` + `/messages`) en código nuevo | deprecated-transport | warning | design | Usar Streamable HTTP (o stdio) |
| Lógica de negocio acoplada a `Mcp-Session-Id` | session-coupling | warning | design | Diseño stateless; el RC 2026-07-28 elimina la sesión |
| Descripción de tool vacía o repite el nombre | empty-tool-description | warning | design | Describir qué hace y cuándo usarla, como a un empleado nuevo |
| Respuestas grandes sin paginar/truncar | unbounded-response | warning | performance | Paginación, filtrado, truncado con defaults; `response_format` concise/detailed |
| Schema JSON escrito a mano en Python/Go | manual-schema | warning | design | Derivar el schema de type hints/structs tipados |
| SDK no oficial o sin upper bound en versión mayor | unpinned-sdk | suggestion | maintainability | SDK oficial con `<2` / `mcp<2` / línea v1.x |
| Falta de annotations en tools destructivas | missing-annotations | suggestion | design | Declarar `destructiveHint` / `readOnlyHint` / `idempotentHint` / `openWorldHint` |
| Sin tests con cliente in-memory ni MCP Inspector | untested-server | suggestion | testing | Testear handler + contrato del schema + wiring end-to-end |
