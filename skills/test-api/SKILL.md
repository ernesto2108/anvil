---
name: test-api
description: Smoke testing de endpoints HTTP con dos modos (1) escanea cambios recientes del developer, o (2) el humano indica qué endpoints probar. Deriva una matriz sistemática de casos por campo y regla del schema (campo ausente, formato inválido, boundary values, tipo incorrecto, enum fuera de rango, reglas de negocio) con presupuesto por endpoint y sin explosión combinatoria. Genera `tests/api-collection.json` en formato Postman Collection v2.1 — importable en Postman, Insomnia, Thunder Client, Hoppscotch y ejecutable con Newman CLI. Úsalo cuando termines de implementar endpoints, quieras probar APIs existentes, o necesites una colección portable de pruebas manuales sin ligarte a ninguna herramienta. No es para tests unitarios ni archivos .hurl permanentes.
---

# Smoke Test de API (curl)

Verificación rápida de endpoints HTTP recién implementados o modificados, ejecutada por el developer antes del handoff. No reemplaza la suite de tests del `tester` — confirma que lo que se acaba de escribir responde como se espera contra un servidor real.

## Filosofía

1. **El humano aporta valores, no estructura** — el developer leyó el código y conoce rutas, bodies, headers y reglas. El humano solo conoce su entorno (tokens reales, IDs que existen en su DB, base URL). Pedirle únicamente eso.
2. **Cobertura derivada del schema, no fija** — un endpoint con happy path verde pero sin probar auth faltante, boundaries o cada validación no está verificado. Cada campo, formato, rango, enum y regla de negocio detectado se traduce en al menos un caso, dentro de un presupuesto y sin explosión combinatoria.
3. **Evidencia reproducible** — cada curl ejecutado queda documentado con su comando exacto y su respuesta real, para que el humano pueda re-correrlo o adjuntarlo al handoff.

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

     Con esa indicación, buscar en el repo los handlers/routes correspondientes y leer su código para extraer la MISMA información que el Paso 1 obtiene del diff (ruta y método, struct/schema de request y response, validaciones, reglas de negocio, requisitos de auth). Si la indicación involucra un flujo de varios pasos, resolver el endpoint de cada paso. Si no se encuentra ningún endpoint que coincida → DETENER y reportar: "No encontré endpoints que coincidan con tu indicación; reformula con una ruta, un handler o un flujo concreto." Una vez extraída la información, continuar desde el Paso 2 exactamente igual.

1. **Escanear los cambios** — leer el diff o los archivos modificados en esta sesión. Detectar por cada endpoint nuevo/modificado:
   - Ruta y método HTTP
   - Struct/schema de request (campos, tipos, obligatoriedad)
   - Struct/schema de response
   - Validaciones (campos requeridos, formatos, rangos)
   - Reglas de negocio del dominio (unicidad, límites, estados válidos)
   - Requisitos de auth (header esperado, scope/rol)
   - Si no se detecta ningún endpoint HTTP en los cambios → DETENER y reportar al humano: "No detecté endpoints HTTP en los cambios; no hay nada que smoke-testear."

2. **Derivar la matriz de casos y construir los curl templates** — por cada endpoint, derivar la matriz de casos a partir de la información extraída en el Paso 1 (campos, tipos, validaciones, rangos, enums, reglas de negocio), siguiendo la sección "Matriz derivada de casos" (abajo): los 6 escenarios base + un caso por cada eje aplicable, respetando la regla anti-explosión y el presupuesto del endpoint. Generar un curl por cada caso de la matriz derivada. El body y los headers van completos y estructurados; los valores que dependen del entorno van como placeholders en mayúsculas entre `<>`: `<BASE_URL>`, `<TOKEN>`, `<USER_ID>`, `<EXISTING_ID>`, `<NONEXISTENT_ID>`, etc.

3. **Pedir solo los valores al humano** — mostrar una lista concisa de los placeholders únicos detectados y para qué sirve cada uno. No mostrar todos los curl aún. Ejemplo:
   ```
   Necesito estos valores de tu entorno para ejecutar el smoke test:
   - <BASE_URL>: URL local del servidor (ej. http://localhost:8080)
   - <TOKEN>: token Bearer válido de un usuario real
   - <USER_ID>: ID de un usuario que exista en tu DB
   - <NONEXISTENT_ID>: un ID que NO exista (para el caso not-found)
   ```
   - Si el humano indica que el servidor NO está corriendo → no ejecutar. Saltar al paso 5 documentando los templates como "pendiente de ejecución manual".

4. **Generar la Collection** — construir el JSON de `tests/api-collection.json` (formato Postman Collection v2.1) con todos los escenarios de todos los endpoints. NO sustituir los valores reales del humano: cada valor de entorno va como variable Postman (`{{BASE_URL}}`, `{{TOKEN}}`, `{{USER_ID}}`, `{{NONEXISTENT_ID}}`). Si el humano quiere ejecutar inmediatamente puede correr `newman run tests/api-collection.json --env-var BASE_URL=<valor> --env-var TOKEN=<valor>` — pero la ejecución es opcional; el archivo generado es el output principal.

5. **Confirmar** — informar al humano que `tests/api-collection.json` fue creado/actualizado, indicar cuántos endpoints y escenarios se agregaron, y mostrar el comando Newman por si quiere ejecutar en terminal.

## Matriz derivada de casos (por endpoint)

La matriz NO es una lista fija de 6 filas: se **deriva** de lo que el Paso 1 detectó en el schema del endpoint. La información sobre campos, tipos, validaciones, rangos, enums y reglas de negocio se traduce en casos concretos, uno por eje aplicable — no se colapsa en dos escenarios genéricos.

### Base obligatoria (siempre, si aplican)

| Escenario | Qué prueba | Status esperado típico |
|---|---|---|
| Happy path | Request válido + auth válida | 2xx |
| Auth inválida | Token expirado o malformado | 401 |
| Auth faltante | Sin header `Authorization` | 401 |
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
| Payload — campos extra | Campos desconocidos no declarados en el schema; documentar el comportamiento esperado (ignorado vs. rechazado) | depende (documentar) |
| Path / query params | Mismos ejes cuando apliquen: ID malformado (además del 404 base), param de paginación fuera de rango, filtro/enum inválido en query | 400 / 422 |

### Regla anti-explosión combinatoria

- **One-invalid-field-at-a-time:** cada caso inválido parte del happy path válido y altera UN solo campo o regla. **Prohibido** generar el producto cartesiano de valores inválidos (no combinar dos campos inválidos en el mismo caso).
- **Presupuesto orientativo por endpoint** según la complejidad del schema:

  | Complejidad | Campos / reglas | Casos objetivo |
  |---|---|---|
  | Simple | ≤3 campos, sin reglas de negocio | ~8–12 |
  | Medio | 4–8 campos | ~12–20 |
  | Complejo | >8 campos o reglas de negocio ricas | ~20–30 |

- **Priorización por riesgo** cuando se acerca el tope: primero reglas de negocio y boundaries, luego formatos y enums, y por último tipos incorrectos.
- **Nunca omitir en silencio:** si el tope del presupuesto se alcanza y quedan ejes sin cubrir, documentar explícitamente qué ejes y campos quedaron fuera y por qué.

## Ejemplo de curl templates (endpoint con auth)

```bash
# Happy path
curl -s -w '\n%{http_code}' -X POST '<BASE_URL>/api/projects' \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{"name":"Smoke Test","description":"valid"}'

# Validación fallida (name faltante)
curl -s -w '\n%{http_code}' -X POST '<BASE_URL>/api/projects' \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{"description":"missing name"}'

# Auth inválida
curl -s -w '\n%{http_code}' -X POST '<BASE_URL>/api/projects' \
  -H 'Authorization: Bearer malformed.token' \
  -H 'Content-Type: application/json' \
  -d '{"name":"x"}'

# Auth faltante
curl -s -w '\n%{http_code}' -X POST '<BASE_URL>/api/projects' \
  -H 'Content-Type: application/json' \
  -d '{"name":"x"}'

# Recurso no encontrado
curl -s -w '\n%{http_code}' -X GET '<BASE_URL>/api/projects/<NONEXISTENT_ID>' \
  -H 'Authorization: Bearer <TOKEN>'

# Boundary — name en el límite máximo de longitud (ej. 255 chars, esperado 2xx)
curl -s -w '\n%{http_code}' -X POST '<BASE_URL>/api/projects' \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{"name":"<255_CHARS>","description":"boundary max"}'

# Boundary — name un char por encima del máximo (256, esperado 400/422)
curl -s -w '\n%{http_code}' -X POST '<BASE_URL>/api/projects' \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{"name":"<256_CHARS>","description":"boundary over"}'

# Tipo incorrecto — name como número en lugar de string (esperado 400/422)
curl -s -w '\n%{http_code}' -X POST '<BASE_URL>/api/projects' \
  -H 'Authorization: Bearer <TOKEN>' \
  -H 'Content-Type: application/json' \
  -d '{"name":123,"description":"wrong type"}'
```

## Formato de Salida

El output es `tests/api-collection.json` en la raíz del proyecto, en formato **Postman Collection v2.1**:

- Crear el archivo si no existe; si ya existe, actualizarlo agregando los items nuevos sin borrar los existentes.
- El formato es importable en Postman, Insomnia, Thunder Client y Hoppscotch, y corrible con Newman CLI (`newman run tests/api-collection.json`).
- Los valores de entorno van como variables Postman (`{{BASE_URL}}`, `{{TOKEN}}`, `{{USER_ID}}`, `{{NONEXISTENT_ID}}`) — nunca sustituir los valores reales del humano en el JSON, siempre variables.
- Cada escenario es un `item` dentro de una carpeta nombrada por endpoint (ej. `POST /api/projects`).

Estructura mínima:

```json
{
  "info": {
    "name": "<nombre del proyecto>",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "POST /api/projects",
      "item": [
        {
          "name": "happy path",
          "request": {
            "method": "POST",
            "header": [
              { "key": "Authorization", "value": "Bearer {{TOKEN}}" },
              { "key": "Content-Type", "value": "application/json" }
            ],
            "body": {
              "mode": "raw",
              "raw": "{\"name\":\"Smoke Test\",\"description\":\"valid\"}",
              "options": { "raw": { "language": "json" } }
            },
            "url": { "raw": "{{BASE_URL}}/api/projects", "host": ["{{BASE_URL}}"], "path": ["api","projects"] }
          }
        }
      ]
    }
  ]
}
```

El humano puede configurar las variables en Postman/Insomnia como Environment, o pasarlas a Newman con `--env-var BASE_URL=http://localhost:8080`.

## Reglas

- Solo escanear los archivos modificados en la sesión actual; no auditar el repo completo.
- Nunca hardcodear tokens, IDs o URLs reales en los templates — siempre placeholders hasta sustituir en ejecución.
- Nunca persistir los valores reales del humano fuera del documento de resultados, y ahí redactados.
- Un curl por escenario, no encadenar flujos (eso es del `tester` con Hurl).
- No abortar la ejecución por un curl fallido; registrar y continuar.

## Lo que esta skill NO cubre

- Archivos `.hurl` permanentes y suite de contract testing versionada → trabajo del `tester`.
- Tests unitarios / de integración del código → skill `run-tests`.
- Endpoints de frontend o mobile → fuera de alcance; solo APIs backend.

## Checklist antes de documentar

- [ ] Se escanearon los cambios de la sesión y se listaron todos los endpoints afectados
- [ ] Cada endpoint tiene los escenarios base aplicables (o se documentó la omisión con razón)
- [ ] La matriz se derivó del schema: hay un caso por cada campo requerido, formato, boundary, tipo, enum y regla de negocio detectados (o se documentó qué ejes quedaron fuera por presupuesto)
- [ ] Se respetó la regla one-invalid-field-at-a-time (ningún caso combina dos inválidos) y el presupuesto por complejidad del endpoint
- [ ] Todos los placeholders se presentaron al humano antes de ejecutar
- [ ] Las variables de entorno usan `{{PLACEHOLDER}}` — ningún valor real hardcodeado en el JSON
- [ ] `tests/api-collection.json` creado o actualizado en el repo
