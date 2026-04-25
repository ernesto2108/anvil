---
name: design-system
description: Crear y mantener fundamentos del sistema de diseño — variables, colecciones, modos, paletas de colores, tipografía, espaciado, componentes reutilizables con variantes y temas. Usar cuando el usuario diga "design system", "design tokens", "paleta de colores", "escala tipográfica", "escala de espaciado", "crear variables", "definir colores", "configuración de temas", "modo oscuro", "librería de componentes", o cuando el agente diseñador necesite establecer fundamentos visuales antes de diseñar pantallas.
---

# Design System

> **IMPORTANTE:** Solo despachador. Carga los archivos de referencia bajo demanda. Ver tabla de enrutamiento abajo.

## Filosofía

- **Entender → Planificar → Aprobar → Construir** — primero entiende QUÉ estás diseñando (¿web? ¿app? ¿documento?), luego propón los visuales, obtén aprobación, luego construye. Nunca saltarse pasos
- **Colecciones → Componentes → Pantallas** — organiza las variables en colecciones, construye componentes a partir de esas variables, ensambla pantallas desde instancias de componentes. Nunca saltarse capas
- **Iterar, nunca reconstruir** — cuando el usuario solicita un cambio, modifica solo lo que cambió. Nunca borres el trabajo existente para comenzar desde cero
- **Los modos son de primera clase** — claro/oscuro, variantes de marca no son aspectos secundarios. Diseña para modos desde el primer día
- **Semántico sobre literal** — nombra los tokens por propósito (`color-primary`), no por valor (`blue-500`)
- **Los componentes son configurables, no duplicados** — propiedades y variantes, no 20 componentes separados

## Flujo de trabajo

### 0. Entender el Entregable (OBLIGATORIO)

**Gate:** Antes de proponer visuales, entender QUÉ estás diseñando.

Preguntar (en español): "¿Que tipo de producto es? ¿Pagina web estatica? ¿App web? ¿App movil? ¿Landing page? ¿Dashboard? ¿Documento?"

Esto determina los patrones de layout, navegación, densidad y necesidades de componentes. Un portafolio web NO es un CV en PDF. Un dashboard NO es una landing page.

#### Detección de Plataforma

Lee el campo **Scope → Platform** del PRD. Si no existe, pregunta: "¿Para qué plataforma? ¿Web, mobile, o ambos?"

- `web` → tokens web estándar
- `mobile` → carga `reference/platform-guide.md`. Usa escalas tipográficas iOS/Android, objetivos táctiles (44pt+), tamaños de fuente nativos de la plataforma
- `both` → carga `reference/platform-guide.md`. Genera AMBOS conjuntos de tokens web y móvil. Documenta el mapeo (abstracto → CSS web → iOS Swift → Android Compose) como se muestra en la guía de plataforma

### 1. Investigación e Inspiración (OBLIGATORIO)

**Gate:** Antes de proponer visuales, investiga qué funciona en el dominio. Nunca diseñes desde cero.

1. **Sitios de inspiración** — busca referencias en:
   - **Dribbble** / **Behance** — componentes de UI y casos de estudio completos
   - **Awwwards** — diseño web galardonado
   - **Mobbin** / **Screenlane** — patrones reales de apps móviles (esencial si la plataforma es `mobile` o `both`)
   - **Collectui** / **Landbook** — inspiración de UI por categorías
2. **Búsqueda específica del dominio** — busca "{dominio del proyecto} UI design" (ej., "healthcare SaaS dashboard design", "B2B workflow app mobile"). Encuentra 3-5 productos reales en el mismo dominio
3. **Investigación de fuentes** — busca en Google Fonts fuentes que coincidan con el tono del proyecto. Propone 2-3 combinaciones (encabezado + cuerpo). Considera:
   - Legibilidad en tamaños pequeños (crítico para móvil)
   - Disponibilidad de fuente variable (mejor rendimiento)
   - Soporte de idioma (si se necesita i18n)
   - Si la plataforma es `both`: verificar que la fuente funciona en web Y tiene un buen equivalente móvil (o usar la misma)
4. **Investigación de paleta de colores** — busca herramientas de paletas (Coolors, ColorHunt, Realtime Colors) para paletas que coincidan con el dominio

Documenta todos los hallazgos — alimentarán la propuesta visual.

### 1.5. Presentar Propuesta Visual (OBLIGATORIO)

**Gate:** Antes de CUALQUIER trabajo en el canvas, presenta una propuesta y obtén aprobación explícita.

Incluir:
- **Referencias encontradas** (3-5 enlaces con lo que te gustó de cada una)
- **Dirección de color** (2-3 opciones que coincidan con el contexto del dominio, con vista previa completa de la escala 50→950)
- **Tipografía** (fuentes específicas de Google Fonts con enlaces, no solo "sans-serif". Muestra la razón de la combinación)
- **Tono** (densidad, esquinas, ambiente, sitio de referencia)
- **Layout** (estructura, navegación, enfoque móvil)
- **Modos** — si el usuario quiere oscuro+claro, muestra ambos en la propuesta
- **Consideraciones de plataforma** — si es `both`, explica cómo diferirán los layouts web y móvil

Formato en español. Espera aprobación explícita. Itera si es necesario.

### 2. Verificar Sistema Existente

Lee `<docs>/01-project/design-system.md`. Si existe → salta al paso 7.

### 3. Crear Colecciones de Variables

Ver `reference/primitives.md` para escalas completas. Ver `reference/semantic-tokens.md` para el mapeo.

**Colección 1: Primitives** — valores brutos, sin significado semántico, nunca consumidos directamente
**Colección 2: Semantic** (con modos) — nombres basados en propósito que aliasan los primitivos, los valores cambian por modo
**Colección 3: Component** (opcional) — con alcance a elementos específicos de UI

#### Escala de Color (OBLIGATORIO — rampa completa de 11 pasos)

Cada familia de tono DEBE tener la escala completa: **50, 100, 200, 300, 400, 500, 600, 700, 800, 900, 950**. Esto aplica a:
- **Brand primary** — la rampa del color principal de la marca
- **Brand secondary** — si el proyecto usa un color secundario
- **Neutral/Gray** — la familia gris elegida para el proyecto
- **Status colors** — rojo (danger), ámbar (warning), verde (success), azul (info)

Nunca definas solo "color primario" como un hex único. La rampa completa es necesaria para estados hover, fondos sutiles, bordes, mapeo de modo oscuro y contraste accesible.

#### Selección de Fuente (OBLIGATORIO — fuente específica, no pila del sistema)

Elige una fuente específica de **Google Fonts** (o la fuente de marca del proyecto si se proporciona). La pila de fuentes del sistema es un fallback, no una elección.

1. Selecciona 2-3 candidatos de fuente basándote en la investigación del Paso 1
2. Define el conjunto completo de pesos usados: 300 (light), 400 (normal), 500 (medium), 600 (semibold), 700 (bold)
3. Define la escala tipográfica completa: **display, 3xl, 2xl, xl, lg, base, sm, xs** — con valores en píxeles de `reference/primitives.md`
4. Si la plataforma es `both`: define tamaños web (px/rem) Y tamaños móviles (pt para iOS, sp para Android) usando `reference/platform-guide.md`

#### Tokens Específicos de Plataforma

Si el `Platform` del PRD es `mobile` o `both`:
- Carga `reference/platform-guide.md`
- Agrega tokens específicos de móvil: mínimos de objetivo táctil (44pt iOS, 48dp Android), conciencia de área segura, tamaños de fuente de la plataforma
- Documenta la tabla de mapeo abstracto → plataforma

Implementación específica de la herramienta: carga `reference/pencil-workflow.md` o `reference/figma-workflow.md`.

### 4. Construir Librería de Componentes (OBLIGATORIO)

**Gate:** Los componentes DEBEN existir ANTES de cualquier diseño de pantalla.

#### Pre-verificación: Verificar que el Sistema de Diseño Existe (BLOQUEANTE)

Antes de crear CUALQUIER componente, ejecuta `get_variables()` y verifica:

1. **Las variables de color existen** — al menos una familia de tono con rampa completa 50→950
2. **Las variables de tipografía existen** — font-family, escala de font-size (display→xs), conjunto de font-weight
3. **Las variables de espaciado existen** — al menos 4 tokens de espaciado

Si CUALQUIERA de estas verificaciones falla → **DETENTE. NO continúes.** Regresa al Paso 3 y crea primero las variables que faltan. Crear componentes sin variables lleva a valores codificados que deben reconstruirse después.

```
# Pseudocódigo de verificación
variables = get_variables()
has_colors = any variable with type "color" and name matching scale pattern (e.g., *-500, *-600)
has_typography = any variable with name matching font-size pattern (e.g., fs-*, font-*)
has_spacing = any variable with name matching spacing pattern (e.g., sp-*, space-*)

if not (has_colors and has_typography and has_spacing):
    DETENTE → "Variables del sistema de diseño incompletas. Crea variables antes que componentes."
```

#### Pre-verificación: Deduplicar Antes de Crear (BLOQUEANTE)

Antes de crear CUALQUIER componente nuevo, busca los existentes:

1. Ejecuta `batch_get({ patterns: [{ reusable: true }] })` para listar TODOS los componentes reutilizables existentes
2. Para cada componente que planeas crear, verifica si ya existe uno similar por nombre o estructura
3. Si existe una coincidencia → **usa el componente existente** (crea instancias con `ref`, sobreescribe via `descendants`)
4. Solo crea un nuevo componente si NINGÚN componente existente sirve al mismo propósito

```
# Pseudocódigo de deduplicación
existing = batch_get({ patterns: [{ reusable: true }] })
for each planned_component:
    match = find in existing where name or structure is similar
    if match:
        USA match (ref + descendants) → NO crear nuevo
    else:
        CREA nuevo componente
```

**Anti-patrón:** Crear "CardA" y "CardB" cuando solo difieren en contenido = 1 componente + 2 instancias con diferentes `descendants`. Si te encuentras creando dos componentes con el mismo layout pero diferente texto/imágenes, estás duplicando.

Cada componente necesita: **propiedades** (boolean, text, instance swap), **variantes** (estado, tamaño, tipo), y **todos los valores de variables**.

#### Estructura de la Librería

Dividida en **2 frames separados** posicionados a la DERECHA de las pantallas:

**Frame 1: "Design Tokens"** — referencia visual, no componentes reutilizables. Usa columnas.

```
Design Tokens
├── Typography                          ├── Colors                    ├── Icons
│   Type Scale (descendente)            │   Agrupado por familia:      │   Todos los íconos del proyecto
│   ├── fs-hero / 48  Hero Display      │   ├── Neutrals (50→950)      │   a 24px con nombre
│   ├── fs-3xl / 30   Heading 3XL       │   ├── Brand                  │   debajo de cada uno.
│   ├── fs-2xl / 24   Heading 2XL       │   ├── Status                 │   Dividir en filas
│   ├── ...hacia abajo...               │   Cada swatch muestra:       │   si no caben.
│   └── fs-xs / 11    Caption           │   ○ círculo + nombre + hex   │   Documentar nombre del
│   ─────────────                       │   Fila Claro + Fila Oscuro   │   conjunto de íconos para devs.
│   Weights                             │   lado a lado por familia    │
│   ├── 700 bold  "Quick brown fox"     │                              │
│   ├── 600 semi  "Quick brown fox"     │                              │
│   ├── 500 med   "Quick brown fox"     │                              │
│   └── 400 norm  "Quick brown fox"     │                              │
```

**Frame 2: "Components"** — componentes reutilizables agrupados por categoría.

```
Components
├── Primitives (column)        ├── Cards (column)
│   Text/Heading               │   Job/Card
│   Text/Body                  │   ├── estado default
│   Text/Caption               │   ├── (hover si aplica)
│   Text/Label                 │   Project/Card
│   Section/Header             │   Stat/Card
│   Skill/Badge                │
│   Contact/Link               │
│   Divider                    │
├── Navigation (ancho completo, debajo de columnas)
│   Navbar (full width)
│   Footer
```

#### Presentación de Tipografía

- Muestra la **escala tipográfica completa en orden descendente** — cada variable font-size de mayor a menor
- Cada línea: `nombre-variable / valor-px` (mono, apagado) + texto de muestra en ese tamaño
- Debajo de la escala, muestra los **pesos** — la misma frase en bold/semibold/medium/normal
- Muestra los nombres de la familia de fuentes (heading, body, mono) en la parte superior

#### Presentación de Color

- Agrupa por **familia**: Neutrals, Brand, Status (success/warning/danger), Accent
- Para cada familia: muestra swatches con **nombre + valor hex** debajo de cada círculo
- Para variables con **modos**: muestra la fila Claro y la fila Oscuro lado a lado (usa cambio de tema a nivel de frame)
- Si el proyecto usa escala completa (50→950), muestra toda la rampa por tono
- Si paleta mínima, muestra solo los colores semánticos usados pero igualmente agrupados por propósito

#### Íconos

- Muestra cada ícono usado en el proyecto a tamaño estándar (24px) con **nombre** debajo
- Divide en filas si no caben en una línea — nunca dejes que los íconos desborden el frame
- Documenta el **nombre del conjunto de íconos** (Lucide, Heroicons, Phosphor) para que los devs instalen el paquete npm correcto
- Todos los conjuntos de íconos modernos usan `currentColor` — los íconos heredan el color de texto del padre, adaptándose a los temas automáticamente
- Prefiere conjuntos con **variantes de tamaño específicas** (Heroicons 24/20/16px están redibujados, no escalados)

#### Presentación de Componentes

- Cada componente muestra sus **variantes lado a lado** cuando aplica (default, hover, disabled)
- Cada componente muestra sus **tamaños** si tiene variantes de tamaño (sm, md, lg)
- Agrupa por categoría con **etiquetas y separadores** claros
- Los componentes de estilo de texto reutilizables (Heading, Body, Caption, Label) pertenecen a Primitives, no a Typography

#### Componentes Mínimos

Estilos de texto, Button (primary/secondary), Input, Card, Section Header, Badge, Divider. Adicionales según las necesidades del proyecto.

#### La Documentación de la Librería NO es Opcional (OBLIGATORIO)

El frame "Design Tokens" DEBE estar completamente poblado antes de que comience el diseño de cualquier pantalla. Un frame de documentación de tokens vacío o parcial es un error bloqueante.

Documentación mínima requerida:
- **Paleta de colores**: Todas las familias de tonos con swatches que muestren nombre de variable + valor hex
- **Escala tipográfica**: Todos los tamaños desde Display hasta XS con muestras de texto en vivo
- **Familias de fuentes**: Muestras del conjunto de caracteres para cada familia (principal + monoespaciada)
- **Inventario de íconos**: Cada ícono usado en el proyecto a 24px con etiqueta de nombre
- **Escala de espaciado**: Barras visuales que muestren cada valor de espaciado
- **Border radius**: Cajas de muestra que muestren cada nivel de radio

Sin esta documentación, los desarrolladores no pueden implementar el sistema de diseño correctamente y codificarán valores en lugar de usar tokens.

#### Posicionamiento y Dimensionamiento

- **Siempre a la DERECHA** de las pantallas, nunca abajo ni detrás
- **Envuelve Design Tokens + Components en un frame padre de auto-layout vertical** ("Library") con gap. Esto previene superposición automáticamente — nunca uses posiciones y fijas entre frames de librería
- **El ancho debe acomodar el contenido** — calcula basándote en el número de columnas y el componente más grande
- **Verifica overflow** después de construir — si algo está recortado, amplía el frame
- Usa `snapshot_layout` para verificar después de construir

### 5. Definir Modos y Verificar Contraste

Como mínimo planifica claro/oscuro. El modo oscuro NO es inversión — las superficies elevadas se vuelven más claras.
Verifica WCAG AA (4.5:1 texto, 3:1 texto grande) para TODOS los modos.

### 6. Ensamblar Pantallas desde Componentes

Usa instancias de componentes (`ref` en Pencil, instancias en Figma). Sobreescribe el contenido via `descendants`, nunca via `U()` en el componente madre.

**Si el usuario solicita oscuro y claro:** diseña uno, copia el frame, cambia solo el tema/modo. Muestra ambos.

### 7. Validar / Extender Sistema Existente

Verifica categorías, identifica brechas, propone adiciones (no modifiques sin aprobación), señala inconsistencias.

### 8. Producir Output

Crea `<docs>/01-project/design-system.md`. Ver `reference/output-template.md`.

## Reglas de Iteración (CRÍTICO)

- **NUNCA borres trabajo para aplicar un cambio** — identifica qué cambió, modifica solo eso
- **Cambio de componente** → edita el componente madre → todas las instancias se actualizan automáticamente
- **Cambio de variable** → actualiza el valor de la variable → todos los nodos se actualizan automáticamente
- **Cambio de layout** → modifica la estructura de la sección, mantén todo lo demás
- **Si un cambio requiere reestructuración** → explica el alcance al usuario primero, obtén aprobación antes de tocar cualquier cosa

## Reglas

- **Entender primero** — saber qué estás diseñando antes de proponer colores
- **Planificar antes de píxeles** — propuesta visual aprobada antes de cualquier trabajo en el canvas
- **Colecciones siempre** — primitivos, semánticos, de componente. Nunca volcados planos
- **Variables para todo** — fuentes, pesos, tamaños, colores, espaciado, radio. Todo
- **Componentes antes de pantallas** — con propiedades/variantes, no nodos crudos
- **El color coincide con el contexto** — profesional = slate/navy, lúdico = vibrante, tech = neutro frío
- **Iterar, no reconstruir** — solicitudes de cambio = ediciones quirúrgicas, no borrar-y-rehacer
- **Librería de componentes visible** — siempre a la derecha de las pantallas, nunca oculta
- **Contraste obligatorio** — verifica todos los modos
- **Reutilizar, nunca recrear** — 4 tarjetas = 1 componente + 4 instancias

## Detección de Anti-Patrones

| Anti-Patrón | Severidad | Corrección |
|---|---|---|
| Diseñar sin entender el tipo de entregable | error | Preguntar primero qué es |
| Saltar al canvas sin propuesta | error | Presentar propuesta visual |
| Borrar trabajo para aplicar un cambio | error | Editar solo lo que cambió |
| `fontFamily:"Inter"` codificado | error | Usar variable `$font-body` |
| `fontWeight:"600"` codificado | error | Usar variable `$fw-semibold` |
| `fill:"#22C55E"` o cualquier hex codificado | error | Usar `$color-link` o variable apropiada |
| 4 tarjetas construidas manualmente | error | 1 componente + 4 instancias `ref` |
| `U()` en componente madre desde instancia | error | Usar `descendants` en el `ref` |
| Librería de componentes detrás/debajo de pantallas | error | Posicionar a la derecha, siempre visible |
| Acento rojo para CV profesional | error | Coincidir colores con el dominio |
| Solo mostrando oscuro, usuario pidió ambos | warning | Mostrar oscuro Y claro lado a lado |
| Sin verificación de contraste | error | Verificar WCAG AA 4.5:1 |
| Escribir código antes de que el diseño sea aprobado | error | Diseño → aprobar → código. Nunca saltarse |
| Datos simulados/inventados (repos falsos, stats falsos) | error | Solo mostrar datos reales. Mejor 2 reales que 4 falsos |
| Escala de color con solo 1-3 valores por tono (ej., solo `primary: #2563eb`) | error | Rampa completa de 11 pasos requerida (50→950) para cada familia de tono |
| Usar pila de fuentes del sistema sin seleccionar una fuente específica | error | Elegir una fuente específica de Google Fonts. La pila del sistema es solo fallback |
| Escala tipográfica sin tamaños (ej., solo base y heading) | error | Escala completa requerida: display, 3xl, 2xl, xl, lg, base, sm, xs |
| Diseñar pantallas antes de que existan variables + librería de componentes | error | Variables → Componentes → Pantallas. Nunca saltarse capas |
| Sin investigación/inspiración de diseño antes de proponer visuales | error | Investigar referencias en Dribbble/Behance/Mobbin antes de proponer |
| Sin tokens móviles cuando el Platform del PRD es `both` o `mobile` | error | Cargar `reference/platform-guide.md` y definir tokens específicos de plataforma |
| Componentes móviles sin objetivos táctiles de 44pt+ | error | Todos los elementos interactivos deben cumplir el tamaño mínimo de objetivo táctil |

## Limitaciones de Herramienta (Pencil)

- **Los tipos de variable son inmutables** — planifica los tipos (color/string/number) antes de crear. Si necesitas cambiar el tipo, usa un nombre de variable nuevo
- **`fontWeight` requiere tipo string** — crea variables de peso como `{"type": "string", "value": "600"}`, no número
- **Advertencias de family de fuentes** — las variables de string para `fontFamily` muestran "invalid" en Pencil. Esto es cosmético, no un error
- **Sin aliasing nativo** — las variables semánticas y primitivas son independientes. Actualiza ambas al cambiar valores

Para wrapping de texto, armonía de grilla, limitaciones de i18n y otras restricciones específicas de Pencil, ver `reference/pencil-workflow.md`.

## Archivos de Referencia

| Trabajando en... | Cargar |
|---|---|
| Escalas de valores brutos (color, tipo, espaciado, radio, sombra) | `reference/primitives.md` |
| Mapeo de tokens semánticos | `reference/semantic-tokens.md` |
| Plantilla de output para design-system.md | `reference/output-template.md` |
| Guía de plataforma (web vs mobile vs both) | `reference/platform-guide.md` |
| **Pencil** — variables, componentes, instancias | `reference/pencil-workflow.md` |
| **Figma** — colecciones, modos, variantes, Dev Mode | `reference/figma-workflow.md` |
