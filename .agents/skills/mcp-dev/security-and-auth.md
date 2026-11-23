# MCP Server — Auth y Seguridad

Basado en el documento oficial **Security Best Practices** y la spec de authorization de MCP 2025-11-25. Verificado 2026-07-18.

## Auth por transport

### stdio → sin OAuth

La spec es explícita: las implementaciones stdio **SHOULD NOT** usar la spec de authorization. Las credenciales se obtienen **del entorno** (env vars, keychain, archivos de config). Preferir stdio precisamente porque limita el acceso al cliente que lanzó el proceso. Nunca aceptar secretos por argumentos de línea de comandos (visibles en `ps`).

Si un server local usa HTTP: bind a `127.0.0.1`, exigir un token, validar `Origin`, o usar unix domain sockets/IPC con permisos restringidos.

### Streamable HTTP → OAuth 2.1

La autorización es **opcional** en MCP, pero si se implementa sobre HTTP debe seguir la spec (OAuth 2.1 draft-13). El server MCP actúa como **Resource Server**; el authorization server puede ser externo (Auth0, Keycloak, IdP corporativo).

Requisitos clave:
- **RFC 9728 (Protected Resource Metadata) es MUST**: publicar `/.well-known/oauth-protected-resource` (o indicarlo en `WWW-Authenticate` del 401) con el campo `authorization_servers`.
- El AS debe ofrecer discovery vía **RFC 8414** o **OpenID Connect Discovery 1.0**.
- **Registro de clientes** (orden de preferencia): pre-registro → **Client ID Metadata Documents** (URL HTTPS como `client_id`, nuevo en 2025-11-25) → Dynamic Client Registration RFC 7591 (fallback).
- **PKCE S256 obligatorio**; si el AS no anuncia `code_challenge_methods_supported`, el cliente debe negarse a continuar.
- **RFC 8707 Resource Indicators obligatorio**: el cliente envía `resource=<URI canónica del server>` en authorization y token requests.
- **El server DEBE validar la audiencia del token** (emitido para él) y **NUNCA aceptar ni reenviar tokens de terceros**.
- Token en `Authorization: Bearer ...` en cada request; nunca en query string. `401` = token inválido/ausente; `403` = scopes insuficientes (con `WWW-Authenticate: ... error="insufficient_scope"` para step-up); `400` = request malformado.
- **Scope minimization**: pedir el mínimo inicial y elevar incrementalmente; evitar scopes ómnibus (`*`, `all`).

## Riesgos y mitigaciones (lado server)

1. **Token passthrough — PROHIBIDO.** El server nunca acepta tokens no emitidos para él ni los reenvía a APIs downstream (rompe rate limiting, auditoría, fronteras de confianza). Si el server llama a un API upstream, actúa como cliente OAuth propio con un token distinto.
2. **Confused deputy** (proxy hacia APIs de terceros con client ID estático): consentimiento por cliente antes de reenviar al AS de terceros, registro de `client_id` aprobados por usuario, validación exacta de `redirect_uri` (sin wildcards), `state` criptográfico de un solo uso, cookies `__Host-` + `Secure/HttpOnly/SameSite`.
3. **Session hijacking**: los session IDs **no son autenticación** ("MUST NOT use sessions for authentication"). IDs con CSPRNG, ligados al usuario (`<user_id>:<session_id>`), con expiración/rotación; verificar autorización en cada request.
4. **DNS rebinding** (servers HTTP locales): validar `Origin` (403 si inválido), bind a `127.0.0.1`.
5. **Compromiso de servers locales**: corren con los privilegios del cliente. Sandboxing/contenedores, mínimo privilegio, consentimiento explícito mostrando el comando completo antes de ejecutar.
6. **SSRF en discovery OAuth**: nunca fetchear URLs de metadata sin validar (bloquear rangos privados y metadata endpoints cloud `169.254.169.254`, exigir HTTPS, cuidado con redirects y TOCTOU de DNS).

## Prompt injection vía resultados de tools

Los resultados de tools (y las descripciones de tools de servers de terceros — "tool poisoning") entran al contexto del modelo y pueden contener instrucciones maliciosas.

- Tratar todo output de tools y contenido externo como **datos no confiables** — sanitizar/estructurar antes de devolver (structured output reduce superficie).
- No incluir en las respuestas contenido de terceros sin marcarlo/limitarlo; no incluir secretos ni datos que el usuario no autorizó.
- Validar y acotar inputs: paths canónicos dentro de un root para tools de filesystem, allowlists para comandos.
- El **cliente** requiere confirmación humana para tools destructivas (las annotations son hints no confiables; human-in-the-loop para escrituras).

> ⚠️ La guía específica de prompt injection/tool poisoning está distribuida entre el documento oficial, la página de roots/consentimiento y literatura de Anthropic; estas recomendaciones combinan esas fuentes con consenso de industria 2025-2026.

## Checklist de seguridad

- [ ] stdio: credenciales del entorno, nunca de argumentos de línea de comandos
- [ ] HTTP con auth: OAuth 2.1, server como Resource Server
- [ ] `/.well-known/oauth-protected-resource` publicado (RFC 9728)
- [ ] Audiencia del token validada en cada request
- [ ] Sin token passthrough (server llama upstream con token propio)
- [ ] PKCE S256 y Resource Indicators (RFC 8707) exigidos
- [ ] Confused deputy mitigado (consentimiento por cliente, `redirect_uri` exacto, `state`)
- [ ] Session IDs no usados como autenticación; autorización verificada por request
- [ ] `Origin` validado y bind `127.0.0.1` en servers HTTP locales
- [ ] Discovery OAuth protegido contra SSRF
- [ ] Output de tools sanitizado; sin secretos ni contenido de terceros sin marcar

## Fuentes

- https://modelcontextprotocol.io/specification/2025-11-25/basic/authorization
- https://modelcontextprotocol.io/specification/2025-11-25/basic/security_best_practices
