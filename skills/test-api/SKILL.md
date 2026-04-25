---
name: test-api
description: Convenciones para escribir tests de API con Hurl. Contract testing, chaining de requests, assertions de respuesta, y validación de esquemas. Usar cuando el tester necesite escribir tests de endpoints HTTP.
---

# API Test Conventions (Hurl)

Guía de convenciones para tests de API con Hurl. El tester carga este skill en PASO 0 cuando el handoff incluye tests de API.

## Qué es Hurl

CLI HTTP test runner (Rust/libcurl). Tests en archivos `.hurl` planos — legibles, diffeables, versionables. Sin runtime JS, sin GUI. Exit code 0 = todo pasa.

---

## Estructura de proyecto

```
tests/
  api/
    auth/
      login.hurl
      signup.hurl
    users/
      create-user.hurl
      get-user.hurl
    projects/
      crud-flow.hurl
    health.hurl
  vars/
    local.env
    staging.env
```

- Nombrar por escenario: `create-user.hurl`, no `post-users.hurl`
- Un flujo lógico por archivo (puede tener múltiples requests encadenados)
- Agrupar por dominio/recurso, no por método HTTP
- Máximo ~10 entries por archivo — dividir por escenario si crece

---

## Requests

```hurl
# GET con headers
GET {{base_url}}/api/users/{{user_id}}
Authorization: Bearer {{token}}
Accept: application/json

# POST con JSON
POST {{base_url}}/api/users
Content-Type: application/json
{
  "name": "Alice",
  "email": "alice@example.com"
}

# POST con form data
POST {{base_url}}/api/login
[FormParams]
username: admin
password: {{password}}

# PUT
PUT {{base_url}}/api/users/{{user_id}}
Content-Type: application/json
{
  "name": "Alice Updated"
}

# DELETE
DELETE {{base_url}}/api/users/{{user_id}}
Authorization: Bearer {{token}}
```

---

## Assertions

```hurl
HTTP 200
[Asserts]
status == 200
header "Content-Type" contains "application/json"
jsonpath "$.id" exists
jsonpath "$.name" == "Alice"
jsonpath "$.items" count == 3
jsonpath "$.email" matches /^[a-z]+@[a-z]+\.[a-z]+$/
jsonpath "$.age" >= 18
body contains "success"
duration < 1000
```

**Siempre** assertar el status code en cada response. `duration < N` para endpoints críticos en rendimiento.

---

## Chaining (captures)

```hurl
# Paso 1: Login, capturar token
POST {{base_url}}/api/auth/login
Content-Type: application/json
{
  "email": "{{admin_email}}",
  "password": "{{admin_password}}"
}

HTTP 200
[Captures]
token: jsonpath "$.access_token"
[Asserts]
jsonpath "$.access_token" isString

# Paso 2: Usar token capturado
GET {{base_url}}/api/profile
Authorization: Bearer {{token}}

HTTP 200
[Asserts]
jsonpath "$.email" == "{{admin_email}}"
```

Captures y asserts coexisten en el mismo bloque de response. Las variables persisten en todo el archivo.

---

## Contract testing

Assertar la forma del response explícitamente:

```hurl
[Asserts]
jsonpath "$.id" isInteger
jsonpath "$.name" isString
jsonpath "$.created_at" matches /^\d{4}-\d{2}-\d{2}T/
jsonpath "$.tags" isCollection
jsonpath "$.metadata" exists
```

Para validación completa contra OpenAPI, usar `--json` output + validador externo (Schemathesis para fuzzing automático desde specs OpenAPI).

---

## Variables y ambientes

```bash
# Variables por CLI
hurl --variable base_url=http://localhost:8080 --variable token=abc123 tests/

# Archivo de variables (key=value por línea)
hurl --variables-file vars/local.env --test tests/

# Secrets (redactados en logs/reports)
hurl --secret password=supersecret --test tests/
```

Overrides por entry con `[Options]`:

```hurl
GET {{base_url}}/api/slow-endpoint
[Options]
retry: 3
retry-interval: 2000
```

---

## CI

```bash
# Ejecutar todos los tests
hurl --test tests/api/

# JUnit para dashboards CI
hurl --test --report-junit results.xml tests/api/

# JSON report
hurl --test --report-json report/ tests/api/

# Ejecución paralela (Hurl 5.x)
hurl --test --jobs 4 tests/api/
```

Exit code 0 = pass, non-zero = failure.

---

## Ejemplo completo: CRUD flow

```hurl
# tests/api/projects/crud-flow.hurl
# Flow: login -> crear -> verificar -> actualizar -> eliminar

# --- Login ---
POST {{base_url}}/api/auth/login
Content-Type: application/json
{
  "email": "{{admin_email}}",
  "password": "{{admin_password}}"
}

HTTP 200
[Captures]
token: jsonpath "$.access_token"

# --- Crear ---
POST {{base_url}}/api/projects
Authorization: Bearer {{token}}
Content-Type: application/json
{
  "name": "Test Project",
  "description": "Created by Hurl"
}

HTTP 201
[Captures]
project_id: jsonpath "$.id"
[Asserts]
jsonpath "$.name" == "Test Project"
jsonpath "$.id" isInteger

# --- Verificar ---
GET {{base_url}}/api/projects/{{project_id}}
Authorization: Bearer {{token}}

HTTP 200
[Asserts]
jsonpath "$.id" == {{project_id}}
jsonpath "$.name" == "Test Project"
duration < 500

# --- Actualizar ---
PUT {{base_url}}/api/projects/{{project_id}}
Authorization: Bearer {{token}}
Content-Type: application/json
{
  "name": "Updated Project"
}

HTTP 200
[Asserts]
jsonpath "$.name" == "Updated Project"

# --- Eliminar ---
DELETE {{base_url}}/api/projects/{{project_id}}
Authorization: Bearer {{token}}

HTTP 204
```

Ejecutar: `hurl --test --variable base_url=http://localhost:8080 --variables-file vars/local.env tests/api/projects/crud-flow.hurl`

---

## Anti-patrones

| Prohibido | Correcto |
|---|---|
| URLs/tokens hardcodeados | `{{variables}}` siempre |
| Un request por archivo cuando forman un flujo | Encadenar requests relacionados (crear + verificar + eliminar) |
| Sin assertar status code | `HTTP <code>` en cada response |
| Archivos monolíticos (20+ entries) | Dividir por escenario, < 10 entries |
| Secrets en archivos `.hurl` | `--secret` o `--variables-file` (gitignored) |
| Assertar valores inestables (timestamps, UUIDs) | `exists`, `matches`, `isString` |
| Sin checks de `duration` en endpoints críticos | `duration < N` para performance |
