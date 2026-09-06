---
name: test-api
description: Smoke testing de endpoints HTTP y flujos encadenados: escanea los cambios recientes del developer o prueba lo que el humano indique. Deriva casos desde el schema (campos, formatos, boundaries, reglas) con presupuesto por endpoint. Genera `tests/api-collection.json` (Postman Collection v2.1) y ambientes en `tests/environments/`, ejecutables con Newman CLI. Soporta flujos multi-paso, login automático, auth flexible y multi-servicio. Úsalo cuando termines de implementar endpoints, quieras probar APIs existentes, necesites una colección portable, o pidas "probar el flujo de X", "colección con ambientes", "environment de staging", "smoke cross-service", "login automático en la colección". No es para tests unitarios ni suites E2E permanentes.
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


# Smoke Test de API (Postman Collection)

Verificación rápida de endpoints HTTP y de flujos cortos recién implementados o modificados, ejecutada por el developer antes del handoff. No reemplaza la suite de tests permanente — confirma que lo que se acaba de escribir responde como se espera contra un servidor real.

## Filosofía

1. **El humano aporta valores, no estructura** — el developer leyó el código y conoce rutas, bodies, headers, esquema de auth y reglas. El humano solo conoce su entorno (credenciales reales, IDs que existen en su DB, URLs por ambiente). Pedirle únicamente eso.
2. **Cobertura derivada del schema, no fija** — un endpoint con happy path verde pero sin probar auth faltante, boundaries o cada validación no está verificado. Cada campo, formato, rango, enum y regla de negocio detectado se traduce en al menos un caso, dentro de un presupuesto y sin explosión combinatoria.
3. **La colección describe el sistema; el ambiente describe el entorno** — la colección nunca contiene URLs, tokens ni IDs reales: solo variables. Cambiar de local a staging debe ser cambiar de archivo de environment, nada más.

## Flujo de Trabajo

0. **Elegir modo de entrada** — preguntar al humano cómo arrancar:
   ```
   ¿Cómo quieres ejecutar el smoke test?
     [1] Escanear cambios recientes (git diff / archivos modificados en esta sesión)
     [2] Indicarme qué endpoints o flujo probar
   ```
   - **Si elige [1]** → continuar desde el Paso 1.
   - **Si elige [2]** → el humano indica qué probar, por ejemplo:
     - "el endpoint POST /api/orders"
     - "el flujo de login → crear proyecto → eliminar proyecto"
     - "todos los endpoints del archivo `handlers/users.go`"

     Con esa indicación, buscar en el repo los handlers/routes correspondientes y leer su código para extraer la MISMA información que el Paso 1 obtiene del diff. Si la indicación involucra un flujo de varios pasos, resolver el endpoint de cada paso y qué valor de la respuesta de un paso alimenta al siguiente. Si no se encuentra ningún endpoint que coincida → DETENER y reportar: "No encontré endpoints que coincidan con tu indicación; reformula con una ruta, un handler o un flujo concreto." Una vez extraída la información, continuar desde el Paso 2.

1. **Escanear los cambios** — leer el diff o los archivos modificados en esta sesión. Detectar por cada endpoint nuevo/modificado:
   - Ruta y método HTTP
   - Struct/schema de request (campos, tipos, obligatoriedad)
   - Struct/schema de response (y qué campos sirven como identificadores encadenables: `id`, `token`, `slug`)
   - Validaciones (campos requeridos, formatos, rangos)
   - Reglas de negocio del dominio (unicidad, límites, estados válidos)
   - **Esquema de auth REAL del código** — no asumir Bearer. Ver "Detección de auth y headers".
   - **Headers custom obligatorios** (tenant, api-version, correlation-id, firma) presentes en middlewares o handlers.
   - Si no se detecta ningún endpoint HTTP en los cambios → DETENER y reportar al humano: "No detecté endpoints HTTP en los cambios; no hay nada que smoke-testear."

1b. **Resolver servicios y URLs** — determinar a qué servicio pertenece cada endpoint:
   - Si existe `service-map.yaml` en el repo o el workspace → leerlo para resolver, por endpoint: servicio dueño, base path y esquema de auth de ese servicio.
   - Si no existe → preguntar al humano la lista de servicios involucrados y su variable de URL.
   - Asignar a cada servicio una variable propia de host: `{{AUTH_URL}}`, `{{ORDERS_URL}}`, `{{BILLING_URL}}`.
   - Si solo se detecta un servicio (monolito) → degradar a una sola variable `{{BASE_URL}}`.

2. **Derivar la matriz de casos y construir los requests** — por cada endpoint suelto, derivar la matriz según "Matriz derivada de casos": base obligatoria de 4 escenarios + un caso por eje aplicable, respetando la regla anti-explosión y el presupuesto. Para los flujos, aplicar el alcance reducido de "Flujos encadenados". Body y headers van completos; los valores que dependen del entorno van como variables Postman en mayúsculas: `{{ORDERS_URL}}`, `{{TOKEN}}`, `{{USER_ID}}`, `{{NONEXISTENT_ID}}`.

3. **Pedir solo los valores al humano** — mostrar una lista concisa de las variables únicas detectadas y para qué sirve cada una, agrupadas por servicio. No mostrar los requests aún. Ejemplo:
   ```
   Necesito estos valores de tu entorno (ambiente local):
   - AUTH_URL: URL del servicio de auth (ej. http://localhost:8081)
   - ORDERS_URL: URL del servicio de órdenes (ej. http://localhost:8080)
   - AUTH_EMAIL / AUTH_PASSWORD: credenciales de un usuario real (el flujo de login obtiene el token solo)
   - USER_ID: ID de un usuario que exista en tu DB
   - NONEXISTENT_ID: un ID que NO exista (para el caso not-found)
   ```
   - Preguntar además: **¿qué ambientes necesitas además de `local`?** (dev, staging, qa…). Por defecto se genera solo `local`.
   - Si el humano indica que el servidor NO está corriendo → no ejecutar; generar igual los archivos y marcarlos como "pendiente de ejecución manual".

4. **Generar la colección y los ambientes**:
   - `tests/api-collection.json` (Postman Collection v2.1) con todos los endpoints, flujos y el login opcional. NO sustituir valores reales: todo va como variable.
   - `tests/environments/local.postman_environment.json` siempre, más un archivo por cada ambiente extra pedido.
   - Por cada ambiente, commitear `tests/environments/<amb>.postman_environment.example.json` con placeholders vacíos; el archivo real con valores queda gitignored.
   - Verificar que `.gitignore` contiene la línea de exclusión; si falta, agregarla (ver "Ambientes y secretos").

5. **Confirmar** — informar qué archivos se crearon/actualizaron, cuántos endpoints, escenarios y flujos se agregaron, qué ejes quedaron fuera por presupuesto, y mostrar el comando de ejecución:
   ```
   newman run tests/api-collection.json -e tests/environments/local.postman_environment.json
   ```

## Detección de auth y headers

Detectar del código el esquema real y modelarlo con el bloque `auth` nativo de Postman a nivel de colección (si todos los servicios comparten esquema) o de carpeta (si un servicio tiene el suyo).

| Esquema detectado en el código | Bloque `auth` Postman | Variable |
|---|---|---|
| `Authorization: Bearer <jwt>` | `"type": "bearer"` | `{{TOKEN}}` |
| API key en header (`X-API-Key`) | `"type": "apikey"` con `in: header` | `{{<SERVICIO>_API_KEY}}` |
| Basic auth | `"type": "basic"` | `{{BASIC_USER}}` / `{{BASIC_PASS}}` |
| Cookie de sesión | `"type": "noauth"` + header `Cookie` explícito | `{{SESSION_COOKIE}}` |

- Los casos **auth faltante** y **auth inválida** se modelan con override a nivel del request: `"auth": { "type": "noauth" }` para el faltante, y el mismo tipo con valor malformado para el inválido.
- Headers custom obligatorios detectados (tenant, api-version, correlation-id) van como header del request con valor variable (`{{TENANT_ID}}`), y el valor vive en el environment.

### Flujo de login opcional

Generar una carpeta `Auth` como **primer item de la colección** solo si el proyecto tiene un endpoint de login detectable o el humano lo indica. Hace POST al endpoint de auth con credenciales tomadas del environment y guarda el token como variable de colección — el humano nunca pega tokens a mano.

```json
{
  "name": "Auth",
  "item": [{
    "name": "login",
    "event": [{
      "listen": "test",
      "script": {
        "type": "text/javascript",
        "exec": [
          "pm.test('login 200', () => pm.response.to.have.status(200));",
          "pm.collectionVariables.set('TOKEN', pm.response.json().access_token);"
        ]
      }
    }],
    "request": {
      "auth": { "type": "noauth" },
      "method": "POST",
      "header": [{ "key": "Content-Type", "value": "application/json" }],
      "body": {
        "mode": "raw",
        "raw": "{\"email\":\"{{AUTH_EMAIL}}\",\"password\":\"{{AUTH_PASSWORD}}\"}",
        "options": { "raw": { "language": "json" } }
      },
      "url": { "raw": "{{AUTH_URL}}/api/auth/login", "host": ["{{AUTH_URL}}"], "path": ["api","auth","login"] }
    }
  }]
}
```

## Flujos encadenados

Un flujo es una **carpeta** dentro de la colección (ej. `Flow: crear y eliminar proyecto`) con requests en orden de ejecución. Newman ejecuta los items en el orden del array, así que la colección sigue siendo portable a Postman, Insomnia y Hoppscotch.

Cada paso lleva un script `test` que (a) asserta el status y (b) extrae los valores que necesita el paso siguiente y los guarda como variable de colección:

```json
{
  "name": "2. POST /api/projects",
  "event": [{
    "listen": "test",
    "script": {
      "type": "text/javascript",
      "exec": [
        "pm.test('status 201', () => pm.response.to.have.status(201));",
        "pm.collectionVariables.set('PROJECT_ID', pm.response.json().id);"
      ]
    }
  }],
  "request": {
    "method": "POST",
    "header": [{ "key": "Content-Type", "value": "application/json" }],
    "body": {
      "mode": "raw",
      "raw": "{\"name\":\"Smoke Flow\"}",
      "options": { "raw": { "language": "json" } }
    },
    "url": { "raw": "{{ORDERS_URL}}/api/projects", "host": ["{{ORDERS_URL}}"], "path": ["api","projects"] }
  }
}
```

El paso siguiente consume `{{PROJECT_ID}}` en su URL o body.

### Alcance de los flujos

- **Smoke de flujos, no matriz completa:** el happy path encadenado completo, más opcionalmente 1–2 desvíos básicos (ej. "paso 2 sin el recurso creado en el paso 1", "paso 3 con el recurso ya eliminado").
- **Prohibido** aplicar la matriz completa de casos negativos a cada paso del flujo — esa matriz sigue aplicando solo a los endpoints sueltos.
- Nombrar los pasos con prefijo numérico (`1.`, `2.`, `3.`) para que el orden sea evidente al leer la colección.
- Cada paso limpia lo que creó cuando el flujo lo permita (último paso = DELETE del recurso), para que el flujo sea re-ejecutable.

### Flujos cross-service

Cada request lleva la variable de host de **su** servicio, así que un flujo cruza servicios sin mecanismo especial: login en `{{AUTH_URL}}` → crear orden en `{{ORDERS_URL}}` → consultar factura en `{{BILLING_URL}}`.

- Los flujos que tocan más de un servicio van en una carpeta `Cross-service flows` dentro de la colección del repo que inicia el flujo. Regla: **el flujo vive donde vive el paso 1**.
- Si todos los servicios comparten el token del servicio de auth, el flujo de login lo resuelve una vez para toda la colección.
- Si un servicio usa su propia API key, esa variable (`{{BILLING_API_KEY}}`) va en el environment y el bloque `auth` se declara a nivel de **la carpeta de ese servicio**, no de la colección.
- Caso API gateway: no requiere estructura distinta — el environment apunta todas las variables de host al mismo dominio con distinto path prefix.

## Ambientes y secretos

Los archivos de environment son Postman Environment v2 y viven en `tests/environments/`. La colección **solo referencia** variables; URLs, credenciales, IDs y headers custom viven aquí.

```json
{
  "name": "local",
  "values": [
    { "key": "AUTH_URL", "value": "http://localhost:8081", "type": "default", "enabled": true },
    { "key": "ORDERS_URL", "value": "http://localhost:8080", "type": "default", "enabled": true },
    { "key": "AUTH_EMAIL", "value": "", "type": "default", "enabled": true },
    { "key": "AUTH_PASSWORD", "value": "", "type": "secret", "enabled": true },
    { "key": "BILLING_API_KEY", "value": "", "type": "secret", "enabled": true },
    { "key": "USER_ID", "value": "", "type": "default", "enabled": true },
    { "key": "NONEXISTENT_ID", "value": "00000000-0000-0000-0000-000000000000", "type": "default", "enabled": true }
  ],
  "_postman_variable_scope": "environment"
}
```

Reglas de secretos, análogas al patrón `.env` / `.env.example`:

- Todo campo sensible (password, token, api key, cookie de sesión) lleva `"type": "secret"` y `value` vacío en el archivo `.example`.
- Se commitea `tests/environments/<amb>.postman_environment.example.json`; el archivo real con valores NO se commitea.
- Verificar `.gitignore` y agregar si falta:
  ```
  tests/environments/*.postman_environment.json
  !tests/environments/*.postman_environment.example.json
  ```
- `local` se genera siempre. Otros ambientes (`dev`, `staging`, `qa`) solo cuando el humano los pide, con las mismas claves y valores vacíos.

## Matriz derivada de casos (por endpoint suelto)

La matriz NO es una lista fija de filas: se **deriva** de lo que el Paso 1 detectó en el schema del endpoint.

### Base obligatoria (siempre, si aplican)

| Escenario | Qué prueba | Status esperado típico |
|---|---|---|
| Happy path | Request válido + auth válida | 2xx |
| Auth inválida | Credencial expirada o malformada (override de `auth` en el request) | 401 |
| Auth faltante | `"auth": { "type": "noauth" }` | 401 |
| Recurso no encontrado | ID que no existe | 404 |

Omitir un escenario base solo si no aplica al endpoint (ej. un endpoint público sin auth omite los dos de auth). Documentar la omisión con su razón.

### Ejes de generación (uno o más casos por cada detección del Paso 1)

Partiendo del body del happy path, generar casos variando **un solo campo/regla a la vez**:

| Eje | Regla de generación | Status esperado |
|---|---|---|
| Campo requerido ausente | Un caso POR CADA campo requerido, omitiendo ese campo (no un único caso global) | 400 / 422 |
| Formato inválido | Un caso por cada campo con formato (email, UUID, fecha, URL) con valor mal formado | 400 / 422 |
| Boundary value (rango/límite) | Por cada min/max de longitud o numérico: un caso justo EN el límite (2xx) y un caso justo FUERA (400/422) | 2xx y 400/422 |
| Tipo incorrecto | Por cada campo tipado: número donde va string, `null` en campo no-nullable | 400 / 422 |
| Enum / estado inválido | Por cada enum o conjunto de estados válidos: un valor fuera del conjunto | 400 / 422 |
| Regla de negocio | Por cada regla detectada: unicidad → duplicado; límite de dominio → excedido; transición de estado inválida → estado ilegal | 409 / 422 / depende |
| Payload — body vacío | Body `{}` | 400 / 422 |
| Payload — JSON malformado | Body con JSON sintácticamente inválido | 400 |
| Payload — campos extra | Campos desconocidos no declarados en el schema; documentar el comportamiento esperado | depende (documentar) |
| Path / query params | ID malformado (además del 404 base), paginación fuera de rango, filtro/enum inválido en query | 400 / 422 |
| Header custom obligatorio ausente | Por cada header custom detectado como obligatorio: request sin él | 400 / 412 / depende |

### Regla anti-explosión combinatoria

- **One-invalid-field-at-a-time:** cada caso inválido parte del happy path válido y altera UN solo campo o regla. **Prohibido** generar el producto cartesiano de valores inválidos.
- **Presupuesto orientativo por endpoint** según la complejidad del schema:

  | Complejidad | Campos / reglas | Casos objetivo |
  |---|---|---|
  | Simple | ≤3 campos, sin reglas de negocio | ~8–12 |
  | Medio | 4–8 campos | ~12–20 |
  | Complejo | >8 campos o reglas de negocio ricas | ~20–30 |

- **Presupuesto de flujos:** máximo ~5 flujos por colección y ~6 pasos por flujo. Si el humano pide más, DETENER y pedir priorización.
- **Priorización por riesgo** cuando se acerca el tope: primero reglas de negocio y boundaries, luego formatos y enums, y por último tipos incorrectos.
- **Nunca omitir en silencio:** si el tope se alcanza y quedan ejes sin cubrir, documentar explícitamente qué ejes y campos quedaron fuera y por qué.

## Formato de Salida

Archivos generados:

| Archivo | Cuándo | Commit |
|---|---|---|
| `tests/api-collection.json` | siempre | sí |
| `tests/environments/local.postman_environment.example.json` | siempre | sí |
| `tests/environments/local.postman_environment.json` | siempre | no (gitignored) |
| `tests/environments/<amb>.postman_environment[.example].json` | solo si el humano lo pide | example sí, real no |

Sobre la colección:

- Crear si no existe; si ya existe, agregar items nuevos sin borrar los existentes.
- Formato Postman Collection v2.1, importable en Postman, Insomnia, Thunder Client y Hoppscotch, corrible con Newman.
- Organización del array `item`: `Auth` (si aplica) → una carpeta por servicio con sus endpoints → `Cross-service flows`.
- Cada escenario de endpoint suelto es un `item` dentro de una carpeta nombrada por endpoint (ej. `POST /api/projects`).

```json
{
  "info": {
    "name": "<nombre del proyecto>",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "auth": { "type": "bearer", "bearer": [{ "key": "token", "value": "{{TOKEN}}", "type": "string" }] },
  "variable": [{ "key": "TOKEN", "value": "" }],
  "item": [
    {
      "name": "orders-service",
      "item": [
        {
          "name": "POST /api/projects",
          "item": [
            {
              "name": "happy path",
              "event": [{ "listen": "test", "script": { "type": "text/javascript", "exec": [
                "pm.test('status 201', () => pm.response.to.have.status(201));"
              ]}}],
              "request": {
                "method": "POST",
                "header": [
                  { "key": "Content-Type", "value": "application/json" },
                  { "key": "X-Tenant-Id", "value": "{{TENANT_ID}}" }
                ],
                "body": {
                  "mode": "raw",
                  "raw": "{\"name\":\"Smoke Test\",\"description\":\"valid\"}",
                  "options": { "raw": { "language": "json" } }
                },
                "url": { "raw": "{{ORDERS_URL}}/api/projects", "host": ["{{ORDERS_URL}}"], "path": ["api","projects"] }
              }
            },
            {
              "name": "auth faltante",
              "request": {
                "auth": { "type": "noauth" },
                "method": "POST",
                "header": [{ "key": "Content-Type", "value": "application/json" }],
                "body": { "mode": "raw", "raw": "{\"name\":\"x\"}", "options": { "raw": { "language": "json" } } },
                "url": { "raw": "{{ORDERS_URL}}/api/projects", "host": ["{{ORDERS_URL}}"], "path": ["api","projects"] }
              }
            }
          ]
        }
      ]
    }
  ]
}
```

Ejecución:

```bash
newman run tests/api-collection.json -e tests/environments/local.postman_environment.json
newman run tests/api-collection.json -e tests/environments/staging.postman_environment.json
```

## Reglas

- Solo escanear los archivos modificados en la sesión actual; no auditar el repo completo (salvo el modo [2], acotado a lo que el humano indicó).
- Nunca hardcodear tokens, IDs, credenciales o URLs reales en la colección — siempre variables `{{VAR}}` resueltas por el environment.
- Nunca commitear un environment con valores reales; solo el `.example` con campos sensibles vacíos y `type: "secret"`.
- No asumir Bearer: usar el esquema de auth detectado en el código.
- Una variable de host por servicio; `{{BASE_URL}}` única solo cuando se detecta un único servicio.
- Los flujos son smoke: happy path encadenado + máximo 1–2 desvíos. No aplicar la matriz negativa completa a cada paso.
- Los scripts de test solo assertan status y extraen variables; no contienen lógica de negocio ni bucles.
- No abortar la generación por un request fallido en ejecución; registrar y continuar.

## Lo que esta skill NO cubre

- Suites E2E permanentes y versionadas (Hurl, Playwright, contract testing) → trabajo del `tester`; esta skill solo produce smoke desechable/portable.
- Tests unitarios / de integración del código → skill `run-tests`.
- Pruebas de carga, stress o performance con NFRs → fuera de alcance.
- Mantenimiento de `service-map.yaml` → esta skill solo lo lee.
- Endpoints de frontend o mobile → fuera de alcance; solo APIs backend.

## Checklist antes de documentar

- [ ] Se escanearon los cambios (o lo indicado por el humano) y se listaron todos los endpoints y flujos afectados
- [ ] Se resolvió el servicio dueño de cada endpoint (vía `service-map.yaml` o preguntando al humano) y cada uno usa su variable de host
- [ ] El esquema de auth se detectó del código, no se asumió; los casos auth faltante/inválida usan override en el request
- [ ] Si hay endpoint de login detectable, existe la carpeta `Auth` como primer item y guarda `{{TOKEN}}` vía script
- [ ] Cada endpoint suelto tiene los escenarios base aplicables (o se documentó la omisión con razón)
- [ ] La matriz se derivó del schema: un caso por campo requerido, formato, boundary, tipo, enum, header obligatorio y regla de negocio (o se documentó qué quedó fuera por presupuesto)
- [ ] Se respetó one-invalid-field-at-a-time y el presupuesto por complejidad
- [ ] Cada flujo es una carpeta con pasos numerados, scripts de assert + extracción, y alcance smoke (sin matriz negativa completa)
- [ ] Los flujos cross-service viven en la colección del repo que inicia el paso 1
- [ ] `tests/api-collection.json` creado o actualizado
- [ ] `tests/environments/local.postman_environment.json` + su `.example` generados; ambientes extra solo si el humano los pidió
- [ ] Campos sensibles con `type: "secret"` y valor vacío en el `.example`; `.gitignore` actualizado
- [ ] Ningún valor real hardcodeado en la colección
