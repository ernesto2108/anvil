---
name: design-review
description: Revisar diseños existentes para evaluar calidad, jerarquía visual y patrones anti-IA. Funciona con Pencil (archivos .pen) y Figma. Usar cuando el usuario diga "revisa este diseño", "¿esto se ve bien?", "mejora el diseño", "feedback de diseño", "QA de diseño", o después de completar la ejecución del diseño visual.
---

# Design Review

> Analiza un diseño existente para evaluar su calidad, proponer mejoras y ejecutar los cambios aprobados. Independiente de la herramienta — funciona con Pencil y Figma.

## Cuándo Usar

- Después de completar la ejecución del diseño (post GATE de Ejecución de Diseño)
- Cuando el usuario solicita retroalimentación o mejoras de diseño
- Antes de la traducción de diseño a código (gate de calidad)
- Cuando los diseños "parecen generados por IA" y necesitan humanizarse

## Flujo de trabajo

### Paso 1 — Detectar Herramienta y Capturar Estado Actual

Detecta la herramienta de diseño y captura lo que existe:

**Pencil (archivo .pen):**
1. `get_editor_state` — obtener el archivo activo y la selección
2. `batch_get` con `patterns: [{ type: "frame" }]` — listar todos los frames de nivel superior (pantallas)
3. `get_screenshot` para cada pantalla — capturar el estado visual
4. `get_variables` — leer los tokens de diseño actuales

**Figma:**
1. Pedir al usuario la URL del archivo de Figma o capturas de pantalla de las pantallas a revisar
2. Si `use_figma` está disponible, inspeccionar la estructura de nodos de forma programática
3. Si no, trabajar con las capturas de pantalla que proporcione el usuario

### Paso 2 — Analizar Contra el Checklist de Calidad

Revisa cada pantalla contra estos criterios. Puntúa cada uno del 1 al 5:

#### Jerarquía Visual (peso: 25%)
- [ ] Una región dominante por pantalla — sin secciones de igual peso compitiendo entre sí
- [ ] Punto focal claro — el ojo sabe a dónde ir primero
- [ ] Jerarquía de acciones — el CTA primario es visualmente dominante, las acciones secundarias están reducidas
- [ ] Jerarquía tipográfica — saltos de tamaño claros entre niveles de encabezado

#### Espaciado y Ritmo (peso: 20%)
- [ ] Escala de espaciado consistente (sin valores de píxeles arbitrarios)
- [ ] Espacios más estrechos dentro del contenido relacionado, espacio en blanco generoso entre secciones
- [ ] Ritmo vertical — las secciones tienen espaciado intencional y variado
- [ ] Sin áreas comprimidas junto a áreas vacías (a menos que sea intencional)

#### Color y Contraste (peso: 15%)
- [ ] Usa variables de tokens de diseño, no valores hex codificados
- [ ] Contraste suficiente para la legibilidad del texto (WCAG AA: 4.5:1 texto, 3:1 texto grande)
- [ ] Color de acento reservado para acciones — no diluido en elementos decorativos
- [ ] Colores semánticos usados correctamente (error para errores, success para éxito)

#### Patrones Anti-IA (peso: 20%)
- [ ] Asimetría intencional — no todo está perfectamente centrado/reflejado
- [ ] Densidad variada entre secciones — no el mismo espaciado en todos lados
- [ ] Contenido real — sin placeholders de "Lorem ipsum", "Item 1", "User Name"
- [ ] Divulgación progresiva — características complejas reveladas gradualmente, no todas a la vez
- [ ] Variación de layout — no todas las secciones siguen el mismo patrón de cuadrícula de tarjetas

#### Completitud (peso: 20%)
- [ ] Todos los estados interactivos diseñados (dropdowns abiertos, modales visibles, menús expandidos)
- [ ] Estados de carga, vacío y error existen (no solo el flujo feliz)
- [ ] Existe versión móvil (si es responsive/ambas plataformas)
- [ ] Existe modo oscuro (si es requerido)
- [ ] Cada CTA tiene una pantalla de destino

### Paso 3 — Generar Reporte

Produce una revisión estructurada:

```markdown
## Design Review — <nombre del archivo/proyecto>

### Puntuación General: X/10

### Análisis Pantalla por Pantalla

#### <Nombre de la Pantalla>
- **Puntuación:** X/10
- **Fortalezas:** [qué funciona bien]
- **Problemas:**
  1. [problema] — severidad: alta/media/baja — corrección: [acción específica]
  2. [problema] — severidad: alta/media/baja — corrección: [acción específica]

### Resumen de Mejoras
| # | Pantalla | Problema | Severidad | Corrección Propuesta |
|---|--------|-------|----------|-------------|
| 1 | Dashboard | Cuadrícula de tarjetas de igual peso | media | Hacer la primera tarjeta de métrica 2x de ancho |
| 2 | Login | Texto placeholder genérico | alta | Usar etiquetas específicas del dominio |
| 3 | Todas las pantallas | Gap uniforme de 24px en todos lados | media | Variar: 16px dentro de secciones, 32px entre secciones |

### Prioridad Recomendada
1. [Correcciones de alta severidad primero]
2. [Correcciones medias]
3. [Elementos de pulido]
```

### Paso 4 — Proponer y Confirmar

Presenta la revisión al usuario. Pregunta qué mejoras aplicar:
- "Todas" — aplicar todo
- "Solo alta" — solo las correcciones de alta severidad
- El usuario elige elementos específicos

**NUNCA aplicar cambios sin confirmación del usuario.**

### Paso 5 — Ejecutar Cambios Aprobados

**Pencil:**
- Usa `batch_design` para aplicar cambios (U para actualizaciones, R para reemplazos)
- Verifica cada cambio con `get_screenshot`
- Máximo 25 operaciones por llamada a batch

**Figma:**
- Carga la skill `/figma-use` antes de cualquier llamada a `use_figma`
- Aplica cambios de forma programática
- Pide al usuario que verifique visualmente (no hay herramienta de captura de pantalla disponible)

Después de todos los cambios:
- Toma capturas de pantalla finales (Pencil) o pide al usuario que verifique (Figma)
- Compara antes/después
- Reporta qué cambió

## Verificaciones Específicas de Pencil

Al revisar un archivo de Pencil, también verificar:
- Carga `get_guidelines("guide", "<tipo de proyecto>")` y verifica que el diseño sigue esos principios
- Comprueba si se aplicó un estilo de Pencil — si no, sugiere uno que se adapte al dominio
- Verifica que todas las propiedades visuales usen `$variables`, no valores codificados
- Comprueba el uso de componentes — los frames crudos donde existen componentes son una señal de alerta

## Verificaciones Específicas de Figma

Al revisar un archivo de Figma:
- Comprobar el uso de auto-layout (el posicionamiento manual = frágil)
- Verificar instancias de componentes vs. copias desvinculadas
- Comprobar el uso de tokens de diseño mediante estilos/variables
- Verificar las restricciones responsive

## Reglas

- **Proponer, no imponer** — siempre muestra al usuario qué vas a cambiar y obtén aprobación
- **Ediciones quirúrgicas** — corrige problemas específicos, no reconstruyas pantallas
- **Preserva la intención** — mejora la calidad sin cambiar la dirección del diseño
- **Captura de pantalla después de cada cambio** (Pencil) — verifica que no haya efectos secundarios
- **Puntúa honestamente** — un 10/10 significa que nada necesita mejora. La mayoría de los diseños están entre 6-8
