---
name: design-recipes
description: Patrones de diseño reutilizables para construir pantallas de forma eficiente en Pencil o Figma. Reduce las operaciones por pantalla al proporcionar recetas probadas. Cargar durante la fase GATE de Ejecución de Diseño. Usar cuando se construyen pantallas a partir de componentes del sistema de diseño en cualquier herramienta de diseño.
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


# Design Recipes

> Patrones probados para ensamblar pantallas de forma eficiente. Patrones independientes de la herramienta con implementaciones específicas por herramienta.

## Cuándo Cargar

Carga esta skill durante el **GATE de Ejecución de Diseño** en el pipeline de orquestación — después de que el diseñador produce `design-spec.md` y antes de ejecutar el diseño visual.

> **Plataforma mobile nativa:** cuando el target sea una app nativa (iOS/Android), carga también `reference/mobile-patterns.md` del skill `/design-system`. Las recetas móviles de abajo son nativas — una app nativa no es una spec web encogida a 375px.

## Flujo de trabajo

### Paso 0 — Cargar Lineamientos de Pencil (si hay archivo .pen)

Antes de construir CUALQUIER pantalla, carga los lineamientos de Pencil que correspondan al tipo de proyecto:

1. **Detecta el tipo de proyecto** desde el PRD/Design Spec:
   - SaaS dashboard, panel de admin, CRM → `get_guidelines("guide", "Web App")`
   - Landing page, sitio de marketing → `get_guidelines("guide", "Landing Page")`
   - App móvil → `get_guidelines("guide", "Mobile App")`
   - Pantallas con muchos datos, tablas → `get_guidelines("guide", "Table")`
   - Trabajo de sistema de diseño → `get_guidelines("guide", "Design System")`

2. **Explora estilos visuales** — Pencil ofrece arquetipos de estilo curados. Antes de comenzar:
   - Ejecuta `get_guidelines()` para ver los estilos disponibles
   - Elige un estilo que coincida con el dominio (ej., "Soft Bento" para SaaS amigable, "Aerial Gravitas" para empresarial, "Editorial Scientific" para apps con muchos datos)
   - Cárgalo: `get_guidelines("style", "<estilo elegido>")`

3. **Aplica lineamientos junto con las recetas** — los lineamientos de Pencil definen principios (jerarquía, densidad, retroalimentación). Las recetas a continuación definen la estructura. Usa ambos.

### Paso 1 — Construir Pantallas

1. Detecta la herramienta de diseño:
   - Archivo `.pen` → carga `reference/pencil.md`
   - URL de Figma o `.fig` → carga `reference/figma.md`
2. Identifica qué tipos de pantalla vas a construir
3. Sigue la receta para cada tipo
4. Verifica con captura de pantalla después de cada pantalla

## Recetas por Tipo de Pantalla

### Receta 1: Pantalla de Autenticación (Login, Registro, Verificación)

**Patrón:** Layout dividido — panel de marca (izquierda/arriba) + panel de formulario (derecha/abajo)

**Estructura:**
```
Desktop (1440×900):
┌──────────────┬─────────────────────────┐
│ Brand Panel  │     Form Panel          │
│ 560px        │     fill_container      │
│ Primary bg   │     Centered card       │
│ Logo+tagline │     Title+fields+CTA    │
└──────────────┴─────────────────────────┘

Mobile (375×812):
┌─────────────────────┐
│ Brand Header (compact) │
│ Primary bg, 1 line     │
├─────────────────────┤
│ Form (full width)      │
│ Title+fields+CTA       │
└─────────────────────┘
```

**Componentes necesarios:** InputGroup, InputPassword, Button/Primary, Button/Ghost
**Operaciones:** ~12 desktop, ~10 mobile
**Consejo de reutilización:** Construye Login primero, luego Copia para Registro/Verificación y modifica el contenido

### Receta 2: App Shell (Nav + Contenido)

**Patrón:** Nav superior + área de contenido con scroll

**Estructura:**
```
Desktop:
┌─────────────────────────────────────────┐
│ NavBar (ref, fill_container width, 64h) │
├─────────────────────────────────────────┤
│ Content area (padding 32-48, vertical)  │
│ ├── Page header (title + actions)       │
│ ├── Content sections                    │
│ └── Pagination (if table)               │
└─────────────────────────────────────────┘

Mobile:
┌─────────────────────┐
│ Mobile Nav (56h)     │
│ ☰  Logo  👤         │
├─────────────────────┤
│ Content (padding 16) │
└─────────────────────┘
```

**Componentes necesarios:** NavBar (desktop) o MobileNav, Avatar
**Consejo de reutilización:** Construye un app shell, Cópialo para cada página, reemplaza solo el contenido

### Receta 3: Página de Tabla de Datos

**Patrón:** Encabezado + filtros + tabla + paginación

**Estructura:**
```
Área de contenido:
├── Encabezado de página: Título (izquierda) + botón Primary (derecha)
├── Fila de filtros: Dropdown selects (horizontal)
├── Tarjeta de tabla (surface bg, rounded, border):
│   ├── Fila de encabezado (neutral-50 bg): etiquetas de columnas
│   ├── Filas de datos: celdas con texto/badges
│   └── Última fila: sin borde inferior
└── Paginación: texto informativo (izquierda) + botones prev/next (derecha)
```

**Patrón de fila de tabla (CRÍTICO):**
```
Row (horizontal, fill_container, padding [14, 20], bottom border)
├── Cell 1 (frame, width: fill_container o fixed) → content
├── Cell 2 (frame, width: fill_container o fixed) → content
├── Cell N (frame, width: fixed para badges/dates) → badge/text
```

**Guía de ancho de columnas:**
- Nombre/título: fill_container
- Descripción: fill_container
- Badge de estado: 120-140px
- Fecha: 140-160px
- Conteo/número: 80-100px
- Acciones: 80-100px

**Operaciones:** ~25 para encabezado+tabla+3 filas. Dividir en 2 llamadas.

### Receta 4: Página de Detalle

**Patrón:** Breadcrumb + encabezado con acciones + contenido multi-columna

**Estructura:**
```
Área de contenido:
├── Breadcrumb: padre > actual
├── Encabezado: título + badge + botón editar
├── Texto de descripción
├── Fila de metadatos (horizontal, gap 24)
├── Layout de dos columnas (horizontal, gap 24):
│   ├── Columna izquierda (fill_container): info principal
│   └── Columna derecha (fill_container): info secundaria
└── Sección de elementos relacionados
```

### Receta 5: Wizard/Stepper

**Patrón:** Breadcrumb + stepper + tarjeta de formulario + botones de navegación

**Estados del stepper:**
- Completado: círculo verde + ícono check + texto verde + línea verde
- Activo: círculo primary + número + texto primary
- Pendiente: círculo neutral + número + texto secundario + línea neutral

**Estructura:**
```
Área de contenido:
├── Breadcrumb
├── Título
├── Stepper (horizontal, 3 pasos)
├── Tarjeta de formulario (surface bg, rounded, padding 32):
│   ├── Título de sección
│   ├── Descripción
│   ├── Campos del formulario
│   └── Acciones: Atrás (izquierda) + Siguiente/Crear (derecha)
```

**Consejo de reutilización:** Construye el paso 1, Copia para los pasos 2-3, actualiza los estados del stepper + contenido del formulario

### Receta 6: Lista de Tarjetas Móvil (reemplaza tablas)

**Patrón:** Pila vertical de tarjetas en lugar de filas de tabla

**Estructura de tarjeta:**
```
Card (surface bg, rounded-lg, border, padding 16):
├── Fila superior (horizontal, space_between):
│   ├── Nombre/título (semibold)
│   └── Badge de estado
├── Descripción (texto secundario, xs)
└── Metadatos (texto disabled, xs)
```

### Recetas Móviles Nativas (iOS / Android)

> Estas recetas asumen **app nativa**. Cargar `reference/mobile-patterns.md` de `/design-system` para valores y reglas de plataforma. Unidades: `pt` iOS, `dp`/`sp` Android.

#### Receta M1: Tab Bar Shell (app shell móvil por defecto)

**Patrón:** El shell por defecto de una app nativa — tab bar inferior de 3-5 destinos top-level. **Reemplaza al menú hamburguesa como default** (ver evidencia NN/g en mobile-patterns.md).

**Estructura:**
```
Pantalla (390×844):
├── Safe area top (Dynamic Island 59pt / notch 47pt)
├── Nav bar / large title (44pt compacta · ~96pt con large title 34pt bold)
├── Contenido (single-column, scroll vertical, margen lateral 16)
└── Tab bar (fondo capa nav):
    iOS: 49pt + safe area 34pt · 3-5 tabs · icono ~25pt SF Symbol + label 10-11pt
         · indicador = tint color · (iOS 26: cápsula flotante, inset 21pt)
    M3:  80dp (Expressive 64dp) · icono 24dp + label 12sp · pill 64×32 secondaryContainer
```

**Reglas:** cada tab conserva su propio navigation stack; tab bar solo para navegación top-level (nunca para acciones "crear"); touch target ≥44pt iOS / ≥48dp Android. Si >5 destinos → replantear IA, no meter hamburguesa.

#### Receta M2: Bottom Sheet

**Patrón:** Superficie deslizante desde abajo para tareas cortas con contexto visible.

**Estructura:**
```
Bottom sheet:
├── Grabber/handle superior (iOS opcional · M3 drag handle 32×4dp, target 48×48)
├── Corner radius superior (iOS ~10pt · M3 28dp)
├── Contenido (detent medium ~50% o large full)
├── Cierre visible (X / Cancel) además del swipe down
└── Scrim/dimming del fondo (modal) — non-modal solo si no tapa contenido crítico
```

**Reglas:** detents medium para filtros/compartir/detalle rápido; nunca sheets anidados ni navegación multi-nivel dentro (→ full-screen modal).

#### Receta M3: Formulario Móvil

**Patrón:** Form single-column con teclado gestionado y CTA visible.

**Estructura:**
```
├── Labels ARRIBA del campo, siempre visibles (nunca placeholder como único label)
├── Campos (row mínima 44pt/48dp), teclado por tipo (email/phone/number/OTP), autofill ON
├── Validación ON-BLUR: error directamente debajo del campo (icono + color + texto concreto)
├── Acción del teclado: `next` en intermedios, `done`/`go` en el último
└── CTA de submit VISIBLE con teclado abierto (anclado sobre el teclado o al final del scroll)
```

**Reglas:** el campo activo siempre visible sobre el teclado; nunca CTA fijo tapado por el teclado; pickers nativos por defecto.

#### Receta M4: Pantalla de Detalle Móvil

**Patrón:** Nav bar con back + contenido single-column + acciones en thumb zone.

**Estructura:**
```
├── Nav bar: back "<" arriba-izq (iOS: + edge swipe borde izq · Android: system back)
│   + título (iOS centrado · Android izquierda)
├── Contenido single-column (imagen/hero → meta → cuerpo, margen 16, espaciado múltiplos de 8)
└── Acciones primarias en el tercio inferior (thumb zone verde);
    destructivas fuera de la zona verde o tras confirmación
```

**Reglas:** una tarea primaria por pantalla; profundidad antes que densidad; respetar safe areas (nada crítico bajo el home indicator 34pt).

### Receta 7: Menú Hamburguesa (SOLO web responsive / mobile web)

> **Alcance:** válida SOLO para **web responsive / mobile web**. En **apps nativas** NO uses hamburguesa como navegación primaria — usa la **Receta M1 (Tab Bar Shell)**. Evidencia NN/g (ver mobile-patterns.md): la navegación oculta reduce la discoverability ~2× y baja el uso de navegación 20-50%.

**Patrón:** Overlay de pantalla completa con secciones

**Estructura:**
```
Pantalla completa (375×812):
├── Barra de nav: X (cerrar) + Logo + Avatar
├── Contenido (vertical, padding):
│   ├── SECCIÓN: "NAVEGACIÓN" (label, uppercase, xs)
│   │   ├── Item de menú (activo: primary bg + texto primary)
│   │   ├── Item de menú (ícono + texto)
│   │   └── Item de menú
│   ├── Divider
│   ├── SECCIÓN: "APARIENCIA"
│   │   └── Toggle de tema (ícono + texto + switch)
│   ├── Divider
│   ├── SECCIÓN: "CUENTA"
│   │   ├── Profile
│   │   ├── Settings
│   │   └── Logout (color error)
├── Spacer (fill_container)
└── Barra de usuario (inferior): avatar + nombre + email + badge de rol
```

### Receta 8: Dropdown de Avatar (Desktop)

**Patrón:** Tarjeta flotante anclada al avatar

**Estructura:**
```
Dropdown (260w, surface bg, rounded-lg, shadow-lg):
├── Fila de info de usuario: avatar + nombre/email + badge de rol
├── Divider
├── Items de menú: ícono (18px) + texto
│   ├── Profile
│   ├── Settings
│   └── Toggle de tema (ícono + texto + switch)
├── Divider
└── Logout (color error)
```

**Posicionamiento:** posición absoluta, x = nav_width - dropdown_width - 24, y = nav_height - 4

## Receta de Modo Oscuro

1. Construye primero la versión clara
2. Copia el frame: `C("lightFrameId", document, {name: "Dark: ...", positionDirection: "bottom", positionPadding: 100, theme: {"mode": "dark"}})`
3. Sobreescribe los elementos específicos de la herramienta:
   - Toggle de tema: ícono luna→sol, texto "Modo oscuro"→"Modo claro", switch OFF→ON
4. Verifica con captura de pantalla — revisa el contraste en badges y texto

## Reglas de Eficiencia

1. **Componentes PRIMERO** — construye todos los componentes reutilizables antes de cualquier pantalla
2. **Copia, no reconstruyas** — la primera pantalla de cada tipo se construye, las variantes se copian
3. **Máximo 25 ops por lote** — divide las pantallas grandes en secciones lógicas
4. **Verifica después de cada pantalla** — captura de pantalla para detectar problemas temprano
5. **Usa refs, no frames crudos** — todo patrón repetido debe ser una instancia de componente
6. **Agrupa actualizaciones relacionadas** — agrupa todas las sobreescrituras para una instancia en una sola llamada
