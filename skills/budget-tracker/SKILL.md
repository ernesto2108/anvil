---
name: budget-tracker
description: Contabilidad de budget y política de retry del Líder — estructura `max_retries`/`max_cost`, gate antes de spawn, gate antes de retry, heurística de estimación de costo, y flujo de retry con captura de firma de error y escalamiento al WebSearch o al usuario. Cárgalo cuando el Líder esté por spawnear un sub-agente y necesite chequear si cabe en presupuesto, o cuando un sub-agente falle y deba decidir si reintentar, buscar la firma en web, o escalar. Reemplaza la sección `## Referencia — Budget y retry` del leader.md.
user-invocable: false
---

# budget-tracker

Lógica de presupuesto y política de retry del Líder. Previene runaways de costo y bucles infinitos de retry sin requerir consultas a la API de billing — toda la contabilidad es local al run.

## Cuándo se ejecuta

Se consulta en dos momentos del run:

1. **Antes de cada spawn de sub-agente** — chequear si el costo estimado del próximo spawn cabe en `max_cost`.
2. **Antes de cada retry** — chequear si quedan reintentos disponibles según `max_retries` y aplicar la política de firma de error.

## Estructura de datos

El Líder mantiene este objeto durante todo el run, persistido en `.context/runs/<run-id>/plan.md`:

```
budget {
  max_retries: int        // del usuario o default 2
  max_cost: float (USD)   // del usuario o default $0.50
  retries_used: int       // contador acumulado durante el run
  cost_accumulated: float // estimado local — modelo high ≈ 3× medium × tamaño de prompt
}
```

**Defaults** (si el usuario no especifica en las "Preguntas antes de arrancar" del `leader.md`):

- `max_retries = 2`
- `max_cost = $0.50`

## Reglas de gate

### Antes de cada sub-agente (spawn nuevo o retry)

Si `cost_accumulated + estimate > max_cost` → **escalar al usuario** (vía Protocolo de debate del `leader.md`, Paso 2) con formato:

> "El presupuesto del run está casi agotado (`$X` usados de `$max_cost`). El próximo spawn (`<agente>`) estima `$Y`. ¿Amplío el presupuesto, ajusto el scope, o cierro el run aquí?"

No spawnear hasta recibir respuesta.

### Antes de cada retry

Si `retries_used >= max_retries` → **escalar al usuario** con formato:

> "Agoté los reintentos (`<N>/<max_retries>`) en `<agente>`. Última firma de error: `<firma>`. ¿Amplío el presupuesto de retry, cambio de estrategia, o cierro el run?"

No reintentar hasta recibir respuesta.

### Sin consulta a API de billing

El estimado es **local** — su único propósito es prevenir runaway, no facturar con precisión. NO consultar la API de billing del provider en ningún caso.

## Heurística de estimación de costo

Estimar el costo del próximo sub-agente con esta heurística (suficiente para el gate, no para reportes financieros):

- Modelo `high` ≈ **3×** modelo `medium` (mismo tamaño de prompt)
- Modelo `medium` ≈ **1×** (base)
- Modelo `low` ≈ **0.1×** modelo `medium`
- El costo escala linealmente con el tamaño del prompt (tokens de entrada + tokens esperados de salida)

**Cálculo:**

```
estimate = base_cost(modelo) × (tokens_prompt + tokens_output_esperados) / 1000
```

Donde `base_cost(modelo)` se infiere de los precios por 1K tokens del provider activo. El Líder usa números redondos del orden correcto — no precisión decimal.

## Flujo de retry (cuando un sub-agente falla)

1. **Capturar firma de error** del sub-agente que falló:
   - **Categoría:** el tipo de error reportado (timeout, validation, internal, lint, test-fail, gate-rejected, etc.)
   - **Substring del mensaje normalizado:** primeras 80 chars relevantes del mensaje de error, normalizadas (sin paths absolutos, sin timestamps, sin run IDs)
   - Guardar la firma en `## Errores acumulados` del `plan.md` del run.

2. **Comparar con el intento anterior:**

   - **Firma distinta al intento anterior** → reintento normal:
     - Verificar gates antes del retry (sección de arriba)
     - Si pasan → incrementar `retries_used`, re-invocar al sub-agente con el output del error inline y un hint del gap
     - Si no pasan → escalar al usuario

   - **Firma igual al intento anterior** → no reintentar a ciegas:
     - Llamar `WebSearch` con la firma del error como query (delegar al `explorer`, no usar `WebSearch` directo — el Líder no tiene esa tool por #9 del `leader.md`)
     - Si el `explorer` devuelve una solución aplicable → aplicarla como intento N+1 (sigue contando contra `retries_used`)
     - Si no hay solución encontrada → escalar al usuario

3. **Cortar siempre cuando:**

   - `retries_used >= max_retries`, o
   - `cost_accumulated + estimate > max_cost`

   En cualquiera de estos dos casos → escalar al usuario, no insistir.

## Actualización del `cost_accumulated`

Después de cada sub-agente que termina (exitoso o fallido):

1. Sumar el costo real estimado del spawn al `cost_accumulated` (usando la heurística de arriba).
2. Persistir el nuevo valor en `plan.md` vía `mcp__anvil__save_leader_log`.

Aunque sea un estimado, mantener el running total visible permite al usuario evaluar el costo del run en el output final.

## Anti-patrones a evitar

- **Reintentar con la misma firma sin WebSearch** → bucle silencioso. Siempre romper el bucle con búsqueda externa o escalamiento.
- **Spawnear "uno más" cuando `cost_accumulated > max_cost`** → escalar primero, decidir después.
- **No persistir `cost_accumulated` en `plan.md`** → al continuar un run vía `mcp__anvil__load_orchestration`, el budget se pierde y el Líder podría reanudar con `retries_used=0` y `cost_accumulated=$0` de facto.
- **Estimar con la API de billing** → latencia alta y no agrega valor — el estimado local es suficiente para prevenir runaway.
- **Olvidar normalizar la firma de error** → "Error en /Users/ernesto/.../runs/abc-123/foo.go:42" cambia de run a run; sin normalizar, dos errores iguales se ven distintos y el gate de firma no dispara.

## Reglas

- `max_retries` y `max_cost` se preguntan al usuario en "Preguntas antes de arrancar" (sección del `leader.md`) o caen a default (2 / $0.50).
- La firma de error es `categoría + substring del mensaje normalizado` — esos son los dos campos que comparar entre intentos.
- WebSearch para resolver una firma repetida **se delega al `explorer`** (el Líder no tiene `WebSearch` ni `WebFetch` por Regla inviolable #9).
- El estimado de costo es local — nunca llamar a la API de billing del provider.
- Si el Líder escala por budget, NUNCA continuar sin respuesta del usuario, ni siquiera para "intentar una última vez".
