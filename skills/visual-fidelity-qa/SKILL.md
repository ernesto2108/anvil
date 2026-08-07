---
name: visual-fidelity-qa
description: >
  Compara una referencia de diseño (frame Pencil .pen, URL de Figma o
  screenshots) contra la implementación web/mobile usando Claude Vision.
  Produce JSON de issues clasificados por severidad y bloquea si hay críticos.
  Úsalo cuando termines de implementar UI desde un diseño aprobado, cuando una
  task tenga `Design reference`, o cuando el usuario diga "valida la fidelidad
  visual", "compara diseño contra implementación", "QA visual", "diff visual"
  o "verifica que coincida con el diseño".
user-invocable: false
---

# Visual Fidelity QA

> Verifica que la implementación UI coincide con la referencia de diseño aprobada. Usa Claude Vision (multimodal) para diff semántico — no pixel-a-pixel.

## Filosofía

1. **El diseño aprobado es la fuente de verdad** — si la implementación difiere, la implementación está mal hasta que el humano diga lo contrario.
2. **Diff semántico sobre pixel-perfect** — clasifica por impacto en UX (jerarquía, marca, flujo) no por diferencias de subpíxel.
3. **Puertas explícitas** — los issues críticos bloquean; los menores se marcan "por corregir"; el score es un umbral verificable, no una opinión.

## Inputs requeridos

Se necesita **al menos una** referencia de diseño utilizable de estos tres tipos, más el objetivo de la implementación:

| Campo | Requerido | Descripción |
|---|---|---|
| Referencia de diseño | sí (uno de los tres tipos) | `pen`, `figma` o `screenshots` — ver tabla abajo |
| `impl_url_or_component` | sí | URL del browser, ruta de pantalla o nombre del componente a evaluar |

| Tipo | Campos | Cómo se obtiene el screenshot de referencia |
|---|---|---|
| `pen` | `frame_id` + `pen_file` | `mcp__pencil__get_screenshot` con `frame_id` y `pen_file` |
| `figma` | URL de Figma | Screenshot vía MCP de Figma si está disponible; si no, pedir al humano un export de la pantalla |
| `screenshots` | path(s) a imagen(es) | Leer cada imagen con Read |

Solo DETENER si no existe **NINGUNA** referencia utilizable de los tres tipos. En ese caso, reportar al humano qué falta y detener.

## Flujo de trabajo

### Paso 1 — Obtener screenshot(s) de la referencia

Según el tipo de referencia disponible:

- **`pen`:** llamar a `mcp__pencil__get_screenshot` con `frame_id` y `pen_file`. Si falla → reportar: "No se pudo obtener screenshot del frame `{frame_id}`. Verifica que `frame_id` y `pen_file` sean correctos." y DETENER.
- **`figma`:** obtener screenshot vía MCP de Figma si está disponible. Si no lo está → pedir al humano un export PNG de la pantalla y leerlo con Read. Si no lo provee → DETENER.
- **`screenshots`:** leer cada path con Read.

Guardar como `design_screenshot` (o varios, si hay múltiples estados/variantes — ver Paso 4).

### Paso 2 — Obtener screenshot de la implementación (receta por plataforma)

Detectar la plataforma del proyecto y usar la receta correspondiente:

| Plataforma | Receta de captura |
|---|---|
| Web | Skill `verify` para navegar a `impl_url_or_component` y tomar screenshot |
| Flutter | Emulador/simulador corriendo → `flutter screenshot` o `adb exec-out screencap -p > impl.png` |
| iOS nativo | Simulador booteado → `xcrun simctl io booted screenshot impl.png` |
| Cualquiera con `.maestro/` | Flow de Maestro con `takeScreenshot` |

Guardar como `impl_screenshot`. **Fallback manual solo si todas las vías de la plataforma fallan:** pedir al humano un screenshot como `impl_screenshot.png` en la raíz del proyecto; si lo provee, leerlo con Read; si no → DETENER.

### Paso 3 — Análisis con Claude Vision

Enviar ambas imágenes a Claude con este prompt estructurado:

```
Primera imagen: diseño de referencia aprobado.
Segunda imagen: implementación actual en browser/app.

Analiza las diferencias visuales entre la referencia y la implementación.
Clasifica cada diferencia por severidad:

- crítica: elemento faltante, jerarquía visual incorrecta, flujo roto,
  color de marca incorrecto.
- menor: espaciado incorrecto (>4px), tipografía incorrecta, iconos
  incorrectos.
- cosmética: diferencias de subpíxel, antialiasing, sombras ligeramente
  distintas.

Responde ÚNICAMENTE en JSON con esta estructura:
{
  "score": <0-100, donde 100 = fidelidad perfecta>,
  "issues": [
    {
      "severity": "crítica|menor|cosmética",
      "element": "<nombre del elemento UI>",
      "description": "<qué difiere>",
      "suggestion": "<cómo corregirlo>"
    }
  ],
  "summary": "<resumen en 1-2 líneas>"
}
```

### Paso 4 — Multi-frame (si la referencia tiene varios estados/variantes)

Si la referencia incluye múltiples estados o variantes (dark/light, viewports o tamaños de pantalla, estados interactivos), iterar los Pasos 1-3 por cada par referencia↔implementación correspondiente. Consolidar un único reporte:

- El **score global = mínimo** de los scores individuales.
- La lista de issues concatena los issues de todos los pares, prefijando el `element` con la variante (ej. `[dark] Botón CTA`).

### Paso 5 — Gate y evaluación

- **Issues `críticos` presentes:** el reporte se marca **BLOQUEADO**. Esta skill solo reporta el bloqueo; qué hacer con él lo decide el agente host.
- **Issues `menores`:** marcarlos como "por corregir" en el reporte. No bloquean por sí mismos; el agente host decide si itera.
- **Cosméticos:** marcar como aceptables.
- **`score < 90`:** no es aprobable sin una justificación explícita escrita en el campo `summary` del reporte. Sin justificación, tratar como no aprobado.

### Paso 6 — Output final

Siempre producir el reporte en este formato para el handoff:

```
## Visual Fidelity QA — {fecha}

**Score:** {score}/100
**Referencia:** {tipo} — {frame_id/URL/paths}
**Variantes evaluadas:** {n} (si multi-frame)
**Estado:** APROBADO | BLOQUEADO

### Issues críticos ({n})
- [{element}] {description} → {suggestion}

### Issues menores ({n}) — por corregir
- [{element}] {description}

### Issues cosméticos ({n}) — aceptables
- [{element}] {description}
```

## Reglas

- No modificar código de producción dentro de esta skill — solo reportar.
- No evaluar lógica de negocio ni cobertura de tests.
- No rediseñar componentes ni regenerar el frame en Pencil.
- Sin ninguna referencia utilizable (`pen`, `figma` ni `screenshots`) no hay QA — detener y pedirla.
- `score < 90` no es aprobable sin justificación explícita en el reporte.

## Anti-patrones

| Anti-patrón | Corrección |
|---|---|
| Reportar "todo bien" sin haber obtenido ambos screenshots | Verificar que los Pasos 1 y 2 completaron antes del Paso 3 |
| Marcar diferencias de 1-2px como críticas | Usar la rúbrica: críticas = jerarquía/marca/flujo; subpíxel es cosmético |
| Aprobar con `score < 90` sin justificar | Escribir la justificación en `summary` o marcar no aprobado |
| Detener por falta de `.pen` habiendo screenshots o Figma | Aceptar cualquiera de los tres tipos de referencia |
| Modificar código aquí mismo | Esta skill solo reporta; la corrección la decide el agente host |
