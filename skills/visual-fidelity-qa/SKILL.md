---
name: visual-fidelity-qa
description: >
  Compara el frame de diseño Pencil (.pen) contra la implementación
  browser/mobile usando Claude Vision. Produce JSON de issues clasificados
  por severidad y bloquea la entrega si hay críticos. Úsalo cuando termines
  de implementar UI desde un diseño aprobado, cuando una task tenga
  `Design reference` con `frame_id` y `pen_file`, o cuando el usuario diga
  "valida la fidelidad visual", "compara diseño contra implementación",
  "QA visual", "diff visual" o "verifica que coincida con el diseño".
user-invocable: false
---

# Visual Fidelity QA

> Verifica que la implementación UI coincide con el diseño aprobado en Pencil. Usa Claude Vision (multimodal) para diff semántico — no pixel-a-pixel.

## Filosofía

1. **El diseño aprobado es la fuente de verdad** — si la implementación difiere, la implementación está mal hasta que el humano diga lo contrario.
2. **Diff semántico sobre pixel-perfect** — clasifica por impacto en UX (jerarquía, marca, flujo) no por diferencias de subpíxel.
3. **Puertas explícitas** — issues críticos bloquean la entrega; menores y cosméticos se reportan pero no bloquean.

## Inputs requeridos

| Campo | Requerido | Descripción |
|---|---|---|
| `frame_id` | sí | ID del frame en el archivo `.pen` (ej. `"frame:123:456"`) |
| `pen_file` | sí | Path al archivo `.pen` (ej. `"designs/app.pen"`) |
| `impl_url_or_component` | sí | URL del browser o nombre del componente a evaluar |

Si falta alguno → DETENER y preguntar al humano antes de continuar.

## Flujo de trabajo

### Paso 1 — Obtener screenshot del diseño

Llamar a `mcp__pencil__get_screenshot` con el `frame_id` y `pen_file`. Guardar el resultado como imagen base64 (`design_screenshot`).

Si falla → reportar al humano: "No se pudo obtener screenshot del frame `{frame_id}`. Verificar que el `frame_id` y `pen_file` sean correctos." y DETENER.

### Paso 2 — Obtener screenshot de la implementación

Usar la skill `verify` para navegar a `impl_url_or_component` y tomar screenshot. Guardar como imagen base64 (`impl_screenshot`).

Si la skill `verify` no está disponible o el entorno no tiene browser/emulador:
- Reportar al humano: "No se pudo tomar screenshot de la implementación. Provee un screenshot manualmente como `impl_screenshot.png` en el directorio raíz del proyecto."
- Si el humano provee el archivo → leerlo con Read y continuar.
- Si no lo provee → DETENER.

### Paso 3 — Análisis con Claude Vision

Enviar ambas imágenes a Claude con el siguiente prompt estructurado:

```
Primera imagen: diseño de referencia aprobado en Pencil.
Segunda imagen: implementación actual en browser/app.

Analiza las diferencias visuales entre el diseño de referencia y la
implementación. Clasifica cada diferencia por severidad:

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

### Paso 4 — Evaluar resultado y decidir

Parsear el JSON de respuesta.

- **Si hay issues `críticos`:** BLOQUEAR la entrega. Reportar al humano la lista de críticos y recomendar invocar `qa-fixer` con esos issues como input.
- **Si solo hay issues `menores` o `cosméticos`:** NO bloquear. Reportar menores como recomendaciones; marcar cosméticos como aceptables.
- **Si `score >= 90` y sin críticos:** reportar "Fidelidad visual aprobada (score: {score}/100)".

### Paso 5 — Output final

Siempre producir un reporte en este formato para el handoff:

```
## Visual Fidelity QA — {fecha}

**Score:** {score}/100
**Frame de referencia:** {frame_id} en {pen_file}
**Estado:** APROBADO | BLOQUEADO

### Issues críticos ({n})
- [{element}] {description} → {suggestion}

### Issues menores ({n})
- [{element}] {description}

### Issues cosméticos ({n}) — aceptables
- [{element}] {description}
```

## Reglas

- No modificar código de producción dentro de esta skill — solo reportar.
- No evaluar lógica de negocio ni cobertura de tests.
- No rediseñar componentes ni regenerar el frame en Pencil.
- Sin `frame_id` y `pen_file` válidos no hay QA — detener y pedirlos.

## Anti-patrones

| Anti-patrón | Corrección |
|---|---|
| Reportar "todo bien" sin haber obtenido ambos screenshots | Verificar que los dos pasos 1 y 2 completaron antes del paso 3 |
| Marcar diferencias de 1-2px como críticas | Usar la rúbrica: críticas = jerarquía/marca/flujo; subpíxel es cosmético |
| Entregar con issues críticos pendientes | BLOQUEAR siempre — la entrega solo procede sin críticos |
| Modificar código aquí mismo | Recomendar `qa-fixer`; esta skill solo reporta |
