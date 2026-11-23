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
4. **Los datos exactos ganan sobre la impresión visual** — un screenshot de baja resolución da una primera impresión útil, pero no es ground truth. Esto aplica a **ambos lados** de la comparación: cuando el diseño expone datos estructurados (`design_nodes`) y/o la implementación expone datos estructurados (`impl_nodes`), esos datos son la fuente de verdad para colores, tamaños y tipografía; los screenshots solo orientan dónde mirar y sirven para layout/jerarquía/composición.

## Inputs requeridos

Se necesita **al menos una** referencia de diseño utilizable de estos tres tipos, más el objetivo de la implementación:

| Campo | Requerido | Descripción |
|---|---|---|
| Referencia de diseño | sí (uno de los tres tipos) | `pen`, `figma` o `screenshots` — ver tabla abajo |
| `impl_url_or_component` | sí | URL del browser, ruta de pantalla o nombre del componente a evaluar |

Cuando la implementación es **web**, además de `impl_url_or_component` esta skill extrae `impl_nodes` (colores, tamaños, tipografía vía `getComputedStyle`) siempre que sea posible — ver Paso 2. El agente host que corre esta skill necesita tener acceso a las tools `mcp__claude-in-chrome__*` (cargarlas via ToolSearch si están deferred) para ejecutar los Pasos 2 y 3 en plataforma web.

| Tipo | Campos | Cómo se obtiene el screenshot de referencia |
|---|---|---|
| `pen` | `frame_id` + `pen_file` | `mcp__pencil__get_screenshot` con `frame_id` y `pen_file` |
| `figma` | URL de Figma | Screenshot vía MCP de Figma si está disponible; si no, pedir al humano un export de la pantalla |
| `screenshots` | path(s) a imagen(es) | Leer cada imagen con Read |

Solo DETENER si no existe **NINGUNA** referencia utilizable de los tres tipos. En ese caso, reportar al humano qué falta y detener.

## Flujo de trabajo

### Paso 1 — Obtener screenshot(s) de la referencia

Según el tipo de referencia disponible:

- **`pen`:** llamar a `mcp__pencil__get_screenshot` con `frame_id` y `pen_file` para una primera impresión visual. Si falla → reportar: "No se pudo obtener screenshot del frame `{frame_id}`. Verifica que `frame_id` y `pen_file` sean correctos." y DETENER.

  **Además del screenshot, extraer siempre el árbol de nodos exacto** del mismo `frame_id` vía `mcp__pencil__batch_get`, con `resolveVariables:true` y `readDepth` suficiente para cubrir todos los elementos relevantes de la pantalla (6-8 en pantallas con jerarquía profunda). El screenshot de Pencil suele ser de baja resolución respecto al frame real (ej. ~300px para un frame de 1440px) — detalles como presencia de íconos, su nombre/tipo exacto, tamaño de variante de un componente (`sm` vs `md`), o grosor de borde son indistinguibles a esa resolución y solo el árbol de nodos los expone con precisión. Guardar este árbol como `design_nodes`.

  El análisis del Paso 3 debe basarse en **ambas fuentes**: `design_screenshot` para impresión visual general (layout, composición, primera pasada), y `design_nodes` para verificar valores puntuales (colores hex, tamaños en px, orden de children, contenido de texto literal, nombres de íconos). **Ante conflicto entre lo que "se ve" en el screenshot y lo que dicen los datos exactos del nodo, los datos del nodo ganan.**

  **Además, capturar el `width` declarado del frame raíz evaluado** (el ancho explícito del frame `.pen`, ej. `width:1440` desktop, `width:375`/`390` mobile, `width:768`/`834` tablet) y guardarlo como `design_viewport_width`. Este valor define el viewport objetivo con el que debe compararse la implementación en el Paso 2 — sin él, cualquier diff de layout capturado en un tamaño de ventana distinto es potencialmente un falso positivo por reflow responsive, no un bug real.

- **`figma`:** obtener screenshot vía MCP de Figma si está disponible. Si el MCP de Figma expone también propiedades de nodo (colores, texto, spacing), extraerlas igual que en `pen` y aplicar la misma regla de precedencia. Si no hay screenshot disponible → pedir al humano un export PNG de la pantalla y leerlo con Read. Si no lo provee → DETENER.
- **`screenshots`:** leer cada path con Read. Sin datos de nodo disponibles, el análisis depende enteramente de la imagen — advertir en el reporte final que esta referencia no permite verificar valores exactos (colores hex, copy literal, orden) y que el score debe tratarse con más cautela.

Guardar como `design_screenshot` (o varios, si hay múltiples estados/variantes — ver Paso 4), y `design_nodes` cuando aplique.

### Paso 2 — Obtener screenshot de la implementación (receta por plataforma)

Detectar la plataforma del proyecto y usar la receta correspondiente:

| Plataforma | Receta de captura |
|---|---|
| Web | `mcp__claude-in-chrome__navigate` a `impl_url_or_component`, luego `mcp__claude-in-chrome__computer` (o el tool de screenshot del set `claude-in-chrome`) para capturar |
| Flutter | Emulador/simulador corriendo → `flutter screenshot` o `adb exec-out screencap -p > impl.png` |
| iOS nativo | Simulador booteado → `xcrun simctl io booted screenshot impl.png` |
| Cualquiera con `.maestro/` | Flow de Maestro con `takeScreenshot` |

**Solo para plataforma web** — antes de capturar `impl_screenshot`, fijar el tamaño exacto de la ventana con `mcp__claude-in-chrome__resize_window`, usando `width` = `design_viewport_width` extraído en el Paso 1 (el `height` puede quedar libre, según contenido/scroll). Nunca capturar `impl_screenshot` con el tamaño de ventana que el browser tenía abierto por defecto — el ancho de la ventana debe estar pinneado al ancho exacto del frame de diseño correspondiente antes de cada captura. Si `design_viewport_width` no está disponible (referencia `figma` sin datos de nodo, o `screenshots`), intentar inferir el ancho desde la imagen de referencia; si tampoco es posible, dejar constancia explícita en el reporte final de que la comparación de viewport no pudo alinearse.

Guardar como `impl_screenshot`. **Fallback manual solo si todas las vías de la plataforma fallan:** pedir al humano un screenshot como `impl_screenshot.png` en la raíz del proyecto; si lo provee, leerlo con Read; si no → DETENER.

**Solo para plataforma web** — inmediatamente después de capturar `impl_screenshot`, extraer también `impl_nodes`: usar `mcp__claude-in-chrome__read_page` o `mcp__claude-in-chrome__javascript_tool` para ejecutar `getComputedStyle` (y `getBoundingClientRect` para dimensiones) sobre los elementos relevantes de `impl_url_or_component`, capturando por elemento: `color`, `backgroundColor`, `fontSize`, `fontFamily`, `fontWeight`, `width`/`height`, `borderRadius`, `borderColor`/`borderWidth`. Guardar como `impl_nodes`.

Si la extracción falla o no es viable (SPA muy dinámica, shadow DOM, elementos no localizables, etc.), continuar solo con `impl_screenshot` — pero marcar explícitamente en el reporte final (Paso 6) que la verificación de valores exactos del lado de implementación no fue posible. No fallar silenciosamente.

Flutter, iOS nativo y flows de Maestro quedan fuera del alcance de extracción de nodos de implementación (sin DOM) — para esas plataformas el análisis del Paso 3 depende solo de `impl_screenshot`.

### Paso 3 — Análisis con Claude Vision

**Si hay `design_nodes` E `impl_nodes` disponibles** (ambos lados extraídos — solo posible en web), comparar valor a valor ANTES del análisis de Vision, en vez de pedirle a Vision que estime esos valores desde las imágenes:

- `color` / `backgroundColor` (hex exacto) — `design_nodes.<elemento>.color` vs `impl_nodes.<elemento>.color`
- `fontSize`, `fontFamily`, `fontWeight`
- Dimensiones (`width`/`height` o equivalentes)
- `borderRadius`, `borderColor`/`borderWidth`

Cualquier diferencia encontrada en esta comparación de datos es un issue real — reportarlo con la severidad correspondiente (ver rúbrica del prompt abajo) sin esperar confirmación visual de Vision. El análisis de Vision sobre las imágenes sigue siendo necesario para layout, jerarquía y composición, pero para colores/tamaños/tipografía la fuente de verdad es esta comparación de datos, no la lectura de las imágenes.

Si hay `design_nodes` disponible (ver Paso 1), verificar explícitamente estas categorías contra los datos exactos del nodo antes de cerrar el análisis — son diffs que comparar solo imágenes tiende a pasar por alto:

- **Presencia/ausencia de íconos** — un ícono de más o de menos, o un ícono suelto donde el diseño lo pone dentro de un contenedor con fondo propio (o viceversa).
- **Nombre/tipo exacto de ícono** — ej. `chevron-left` vs `arrow-left`; visualmente similares a baja resolución, semánticamente distintos.
- **Orden de elementos hijos** — el mismo conjunto de elementos en otro orden (ej. texto antes que botón cuando el diseño pone botón antes que texto) se ve "parecido" en una imagen pero es un diff real.
- **Copy/texto literal exacto** — no aceptar "significado similar" como equivalente. Comparar el string literal del nodo contra el string literal implementado; dos frases con la misma intención pero texto distinto SÍ son un diff.
- **Variantes de color por contexto** — no asumir que un componente reusado se ve igual en todas sus instancias. Verificar cada instancia del componente contra su nodo real; un mismo componente puede tener variantes de color legítimamente distintas según su contenedor (ej. borde gris en un contexto, rojo en otro).
- **Color aplicado a texto vs a ícono** — cuando un título lleva texto + ícono adjunto, verificar en el nodo si el color semántico aplica a ambos o solo a uno (ej. solo el ícono lleva color, el texto es siempre neutro). No asumir que comparten color solo porque están juntos visualmente.

Enviar ambas imágenes a Claude con este prompt estructurado:

```
Primera imagen: diseño de referencia aprobado.
Segunda imagen: implementación actual en browser/app.

Analiza las diferencias visuales entre la referencia y la implementación.
Si se provee un árbol de nodos exacto de la referencia (colores hex, tamaños en
px, orden de children, texto literal, nombres de íconos), verifica cada
elemento contra esos datos exactos, no solo contra la impresión visual de la
imagen — presta atención especial a copy exacto, orden de elementos, íconos
faltantes/de más, y variantes de color según contenedor.

Clasifica cada diferencia por severidad:

- crítica: elemento faltante, jerarquía visual incorrecta, flujo roto,
  color de marca incorrecto, copy que cambia el significado o la instrucción
  al usuario, orden de elementos que altera el flujo de lectura o acción.
- menor: espaciado incorrecto (>4px), tipografía incorrecta, iconos
  incorrectos o de nombre distinto, copy con diferencia literal que no
  cambia el significado, orden de elementos cosmético sin impacto en flujo.
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

Si la referencia incluye múltiples estados o variantes (dark/light, viewports o tamaños de pantalla, estados interactivos), iterar los Pasos 1-3 por cada par referencia↔implementación correspondiente. Cuando las variantes correspondan a distintos breakpoints (mobile/tablet/desktop), cada variante debe re-ejecutar `resize_window` al `design_viewport_width` específico de ESA variante antes de capturar su `impl_screenshot` — no reusar el tamaño de ventana de la variante anterior. Consolidar un único reporte:

- El **score global = mínimo** de los scores individuales.
- La lista de issues concatena los issues de todos los pares, prefijando el `element` con la variante (ej. `[dark] Botón CTA`).

### Paso 4.5 — Validar layout heredado de componentes/CSS compartidos (cuando aplique)

Considerar este paso cuando un issue detectado en el Paso 3 es de posición o tamaño **general** de la pantalla completa (ej. contenido centrado vs. pegado a un borde, ancho máximo vs. ancho completo) en vez de un elemento puntual — esto sugiere que el layout raíz puede venir de un componente o CSS compartido entre secciones/roles, no de la pantalla evaluada en sí.

Si la implementación comparte layout/CSS de contenedor raíz entre varias pantallas del mismo rol o sección: comparar el frame `Main`/contenedor raíz de la pantalla evaluada contra el de **al menos otra pantalla relacionada** del mismo rol en la referencia. Si el patrón de layout (padding fijo vs. centrado, cap de ancho vs. sin cap) es consistente entre ambas pantallas del mismo rol, es intencional para ese rol — y un valor distinto en la implementación es un bug de "layout heredado de otro contexto", no un ajuste puntual de la pantalla evaluada. Reportarlo como tal para que la corrección se haga en el nivel compartido, no en la pantalla individual.

No aplica si no hay componentes o CSS compartidos entre roles/secciones, o si el diff es de un elemento puntual sin relación con el contenedor raíz.

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
- Con referencia `pen`, no depender solo del screenshot — extraer siempre `design_nodes` vía `batch_get` y usarlo como fuente de verdad ante conflicto con la imagen.
- Un diff de layout/posición que solo aparece por reflow responsive causado por comparar en un viewport distinto al `design_viewport_width` declarado del frame NO es un issue válido — invalida la comparación; re-capturar `impl_screenshot` con `resize_window` al width correcto antes de reportar. Esto es distinto del layout heredado de componentes/CSS compartidos (Paso 4.5): ahí la causa es CSS compartido entre roles, aquí es viewport mal alineado.

## Anti-patrones

| Anti-patrón | Corrección |
|---|---|
| Reportar "todo bien" sin haber obtenido ambos screenshots | Verificar que los Pasos 1 y 2 completaron antes del Paso 3 |
| Marcar diferencias de 1-2px como críticas | Usar la rúbrica: críticas = jerarquía/marca/flujo; subpíxel es cosmético |
| Confiar solo en el screenshot de Pencil (baja resolución) para descartar diffs de ícono, tamaño de variante o borde | Extraer `design_nodes` vía `batch_get` y verificar esos valores contra el árbol exacto |
| Confiar en Vision para verificar hex/px exactos de la implementación cuando `impl_nodes` está disponible | Comparar `design_nodes` vs `impl_nodes` valor a valor para colores, tamaños y tipografía; usar Vision solo para layout/jerarquía/composición |
| Aceptar copy "con significado similar" como equivalente al copy del diseño | Comparar el texto literal del nodo contra el texto literal implementado; diff de copy exacto es real aunque el significado coincida |
| Asumir que un componente reusado se ve igual en todas sus instancias | Verificar cada instancia contra su nodo real — las variantes de color por contexto son legítimas y deben coincidir por instancia, no por "familia" de componente |
| Evaluar un bug de layout general (centrado, ancho) solo contra la pantalla evaluada, aislada | Comparar el contenedor raíz contra al menos otra pantalla del mismo rol para distinguir bug real de layout intencional heredado (Paso 4.5) |
| Aprobar con `score < 90` sin justificar | Escribir la justificación en `summary` o marcar no aprobado |
| Detener por falta de `.pen` habiendo screenshots o Figma | Aceptar cualquiera de los tres tipos de referencia |
| Modificar código aquí mismo | Esta skill solo reporta; la corrección la decide el agente host |
| Capturar `impl_screenshot` en el tamaño de ventana que el browser tenía abierto, sin fijarlo al width del frame de diseño | `resize_window` al `design_viewport_width` exacto del frame evaluado antes de cada captura, incluida cada variante multi-frame |
