# Referencia de Patrones Móviles Nativos

Composición de pantallas móviles nativas — no "web encogida a 375px". Una app nativa (iOS/Android) usa navegación, superficies y ergonomía propias de la plataforma. Diseñar mobile web responsive es otra cosa: ahí sí aplica el layout web adaptado; esta referencia es para **apps nativas** (`platform: mobile` o `both` con target de app).

Fuentes: Apple HIG (iOS 26 / Liquid Glass), Material Design 3 / M3 Expressive, NN/g, learnui.design. Unidades: `pt` = iOS, `dp`/`sp` = Android.

## Regla base

- **iOS 26**: navegación translúcida flotante (Liquid Glass) reservada a la capa de navegación — nunca a listas/cards.
- **Android 15+**: edge-to-edge obligatorio (dibujar tras las barras y aplicar `WindowInsets`).
- **Se comparte** (diseña una vez): arquitectura de información, flujo de pantallas, bottom-nav 3-5 destinos, thumb-zone, touch targets, color semántico, dark mode, validación on-blur, skeletons. **Cambia solo el "chrome" de plataforma.**

---

## 1. Navegación

### 1.1 Tab bar (iOS) vs Navigation bar (M3)

| Atributo | iOS (HIG) | Material 3 |
|---|---|---|
| Nº destinos | 2-5 (3-5 ideal) | 3-5 (mín 3, máx 5) |
| Altura | 49pt + safe area 34pt · iOS 26: cápsula flotante, inset 21pt | 80dp (baseline) · 64dp (M3 Expressive) |
| Icono | ~25×25pt, SF Symbols | 24×24dp |
| Label | 10-11pt, siempre visible | 12sp; preferir modo labeled |
| Indicador activo | Cambio de tint | Pill 64×32dp `secondaryContainer`, full-round |
| Touch target | ≥44×44pt | ≥48×48dp |

**Reglas:**
- HAZ: tab bar/navigation bar solo para destinos **top-level** del mismo nivel. Cada tab conserva su propio navigation stack (state preservation por tab).
- HAZ (iOS 26): cápsula flotante translúcida que se **minimiza al scrollear** y re-expande hacia arriba; Search se separa en isla propia a la derecha; se oculta con teclado y bajo modales.
- HAZ (M3 ≥600dp): usa **navigation rail** en tablets, no navigation bar.
- NUNCA: tab bar para acciones (crear/componer) — es navegación, no lanzadera; ni ocultarla en niveles profundos sin motivo; ni Liquid Glass en capas de contenido.

### 1.2 Navigation stack y large title

| Atributo | iOS | Android/M3 |
|---|---|---|
| Barra superior | Nav bar 44pt; large title ~96pt | Top app bar: small 64dp / medium 112dp / large 152dp |
| Título grande | Large title 34pt bold → colapsa a 17pt semibold centrado | Headline en app bar medium/large, colapsa con scroll |
| Alineación | Centrado (colapsado) | Izquierda (raíz: center opcional) |
| Back | "<" arriba-izq + **edge swipe borde izq** (obligatorio) | **System back** (ambos bordes) + Predictive Back (14+) |

**Reglas:**
- HAZ: large title en pantallas raíz de cada tab (iOS); soporta siempre el interactive pop gesture — nunca lo bloquees con gestures propios en el borde.
- HAZ (Android): diseña para Predictive Back (preview del destino); no interceptes back sin necesidad.
- NUNCA: botón back propio redundante en Android (el back-in-UI va arriba-izq; el sistema maneja el resto).

### 1.3 Regla anti-hamburguesa (CRÍTICA)

**Evidencia NN/g:** la navegación oculta reduce discoverability — usuarios ~2× menos propensos a descubrir features vs navegación visible; el uso de navegación cae 20-50% y empeora tiempo y éxito de tarea.

- NUNCA: hamburguesa como navegación **primaria** en app nativa con ≤5 destinos → usa tab bar / navigation bar.
- iOS **no tiene drawer nativo**: no lo repliques; usa tab bar + pantalla "More"/perfil para overflow.
- **M3 navigation drawer** (width 360dp, modal con scrim en compact) está **deprecado en M3 Expressive** → reemplazado por **expanded navigation rail** (colapsable ↔ expandido).
- **Excepciones legítimas del drawer:** >5 destinos de igual peso, apps tipo Gmail/Drive con muchas secciones secundarias, o navegación secundaria complementaria a la navigation bar — y aun así, en diseño nuevo 2025+ preferir nav rail expandible en pantallas medianas+.

---

## 2. Superficies y overlays

### 2.1 Bottom sheets

| Atributo | iOS (sheets) | Material 3 |
|---|---|---|
| Alturas | Detents `.medium()` (~50%), `.large()`, `.custom()`; resize por arrastre | Peek → half-expanded (0.5) → expanded |
| Grabber/handle | Grabber opcional (mostrar con varios detents) | Drag handle 32×4dp, touch target 48×48 |
| Corner radius | ~10pt (iOS 26: más redondeado, concéntrico) | Extra-large 28dp superior |
| Fondo | Dimming; detent medium puede ser **non-modal** | Modal (scrim) · Standard (sin scrim) |
| Dismiss | Swipe down + X/Cancel | Swipe down, tap scrim, system back |

- HAZ: sheet con detent medium para tareas cortas que se benefician de ver contexto (filtros, compartir, detalle rápido); ofrece siempre cierre visible (X/Cancel) además del gesto.
- NUNCA: anidar sheets sobre sheets; ni navegación profunda multi-nivel dentro de un sheet (→ full-screen modal).

### 2.2 Full-screen modal vs sheet

- **Sheet** (page sheet iOS): tareas autocontenidas y cancelables — crear, editar, filtrar.
- **Full-screen** (`fullScreenCover`): flujos inmersivos/multi-paso — cámara, reproductor, onboarding. Requiere botón explícito de cierre + confirmación de descarte ("Discard changes?") si hay estado que perder.
- Decisión: ¿el usuario necesita referencia del contexto o cancela sin costo? → sheet. ¿Foco total o estado que perder? → full-screen.

### 2.3 Action sheets y menús

- iOS **action sheet**: acciones desde abajo; destructiva en **rojo y primera**, **Cancel separado abajo**. Tendencia iOS 14+→26: preferir **menus** (pull-down/context, panel translúcido) cuando la acción sale de un botón concreto — aparecen junto al elemento.
- Material: **menu** anclado o modal bottom sheet de acciones; dialogs M3 (280-560dp) solo para decisiones críticas de 2 opciones.
- NUNCA: dialog centrado para listas largas de acciones; ni más de una destructiva por menú.

---

## 3. Ergonomía

### 3.1 Thumb zone (Hoober)

| Zona | Ubicación | Qué poner |
|---|---|---|
| Verde (natural) | Tercio inferior | Acciones primarias, CTA, tab bar, FAB |
| Amarilla (estirando) | Centro y bordes medios | Interactivo secundario |
| Roja (difícil) | Tercio superior, esquina contraria a la mano | Destructivas, poco frecuentes, solo lectura |

- HAZ: CTA primario abajo-centro; interactivos primarios en el 40-50% inferior; aprovecha patrones nativos ya thumb-friendly (tab bar, bottom sheets, FAB, swipe).
- NUNCA: la única vía a una acción frecuente en la esquina superior en apps one-handed.

### 3.2 Touch targets

| Regla | iOS | Material |
|---|---|---|
| Target mínimo | 44×44pt | 48×48dp |
| Espaciado | — | ≥8dp entre targets |
| WCAG 2.2 AA | 24×24 CSS px absoluto, 44×44 recomendado | ídem |

- HAZ: si el glifo es pequeño (20-24), expande el área tappable invisible al mínimo.
- NUNCA: dos targets accionables a <8pt/dp (misclicks).

---

## 4. Listas y contenido

| Patrón | iOS | Material 3 |
|---|---|---|
| Estilos | Plain · Grouped · **Inset grouped** (default moderno, cards redondeadas margen 16pt) | one-line 56dp · two-line 72dp · three-line 88dp |
| Row mínima | 44pt | 48dp (56 cómodo) |
| Accesorios | Chevron, checkmark, switch, texto; destructiva en rojo | Leading icon/avatar 40dp, trailing switch/checkbox |
| Cards | No es patrón de sistema (usar inset grouped) | Elevated/filled/outlined, radius 12dp, contenido heterogéneo |

- HAZ: listas para colecciones homogéneas; cards solo cuando cada item es rico (imagen+meta+acciones). NUNCA cards dentro de cards ni cards para listas densas.
- **Swipe actions:** trailing = frecuentes + destructiva (patrón Mail); leading = positivas. Máx 2-3 por edge, **una destructiva por edge**; desactiva full-swipe si la primera acción es irrecuperable sin undo. Ofrece siempre vía alternativa visible (no son descubribles ni accesibles solos). Material prefiere **undo vía snackbar** sobre confirmación previa.
- **Pull-to-refresh:** solo en contenido ordenado por recencia (feeds, inbox). Nunca en pantallas estáticas ni como única vía si el contenido se refresca solo.
- **Infinite scroll:** solo feeds de exploración sin objetivo. En listas orientadas a tarea usa "Load more"/paginación con posición persistente. Nunca si hay footer con contenido necesario.
- **Skeletons:** para cargas de estructura predecible (espeja el layout final, shimmer sutil); spinner para acciones puntuales. **Regla 300ms:** si la carga durará <300ms, no muestres loading (el flash se percibe como glitch). Nunca skeleton que no coincide con el layout final ni skeleton+spinner simultáneos.

---

## 5. Formularios móviles

- **Keyboard avoidance:** el campo activo siempre visible sobre el teclado (iOS keyboard avoidance nativo; Android `adjustResize`/`imePadding`). El CTA de submit debe quedar visible con teclado abierto (anclado sobre el teclado o al final del scroll). Define la acción del teclado por campo: `next` en intermedios, `done`/`go`/`search` en el último. NUNCA un CTA fijo tapado por el teclado.
- **Teclado por tipo:** email → `.emailAddress`; teléfono → `.phonePad`; número → `.numberPad`; decimal/moneda → `.decimalPad`; URL → `.URL`; OTP → `.numberPad` + `.oneTimeCode`; password → autofill/keychain. Activa siempre **autofill** (`textContentType`/`autofillHints`) — la mayor reducción de fricción. Desactiva autocorrect/capitalización en emails, usernames y códigos.
- **Validación:** **on blur** (al salir del campo), no on keystroke (excepto fuerza de password o contadores). Mensaje de error **directamente debajo del campo**, concreto ("El teléfono debe tener 10 dígitos"), con icono + color (nunca solo color). **Labels arriba del campo, siempre visibles** — nunca placeholder como único label. Nunca resumen de errores solo al top ni deshabilitar submit sin explicar por qué.
- **Pickers nativos por defecto:** date picker compact (iOS) / Material date picker; 2-5 opciones visibles → segmented control / segmented buttons, no picker; steppers solo para rangos cortos (±1-2). Nunca reimplementes date pickers custom sin razón fuerte (pierdes accesibilidad, localización, familiaridad).

---

## 6. Layout móvil

### 6.1 Safe areas (valores iPhone)

| Elemento | Valor |
|---|---|
| Safe area top — Dynamic Island (14 Pro–17) | 59pt |
| Safe area top — notch (X–14) | 47pt |
| Safe area top — sin notch (SE) | 20pt |
| Safe area bottom — home indicator | 34pt |
| Tab bar iOS 26 | inset 21pt (flotante) |
| Android | edge-to-edge obligatorio (SDK 35); gesture nav ~24dp, 3-button 48dp |

- HAZ: el contenido scrolleable fluye **bajo** las barras translúcidas (edge-to-edge); los controles interactivos respetan safe areas. Nunca hardcodear el status bar — usar insets.
- NUNCA: botones/texto crítico bajo el home indicator o en la Dynamic Island; ni gestures propios en bordes del sistema.

### 6.2 Márgenes, densidad, tipografía

- Margen lateral: **16pt (iPhone) / 16dp (Android compact)**; 20pt iPad; 24dp Android medium+.
- Espaciado: múltiplos de **8** (4 para micro): 4-8-12-16-24-32-48.
- Densidad móvil > web: una columna; ancho de línea objetivo 30-40 caracteres. **1 pantalla = 1 tarea primaria**; profundidad antes que densidad.

| Rol | iOS (SF Pro) | Material |
|---|---|---|
| Large title | 34pt bold | Display Small 36sp / Headline Large 32sp |
| Título colapsado | 17pt semibold | Title Large 22sp |
| Body | **17pt** | **Body Large 16sp** |
| Secundario | 15pt · footnote 13pt | Body Medium 14sp |
| Caption | 12-13pt | Label Medium 12sp |
| Mínimo legible | 11pt | 11-12sp |

- HAZ: soporta **Dynamic Type** (iOS) y font scaling (Android `sp`) — layouts que no rompen a 200%; máx 3 niveles visibles por pantalla, jerarquía por peso/color antes que por tamaño.
- NUNCA: tamaño fijo en píxeles; ni gris de bajo contraste (<4.5:1 en body).

---

## 7. Patrones de pantalla completos

- **Onboarding:** 3-5 pantallas máx, <60s; **Skip siempre visible** desde la primera; value-first (beneficio, no features); preferir progressive onboarding (hints contextuales) sobre tours largos; page dots + CTA persistente.
- **Permisos (pre-permission priming):** NUNCA dispares el diálogo de sistema en el primer launch. Pide **en contexto de intención** (cámara al tapear "escanear"). Usa **pre-permission screen** propia antes del diálogo de sistema (si el usuario dice "ahora no", no quemas el prompt nativo, que en iOS solo se muestra una vez). Notificaciones = el más sensible: pídelo tras un momento de valor, nunca en el splash.
- **Empty states:** icono/ilustración contenida + título que explica **por qué** está vacío + 1 CTA que lo puebla ("Añade tu primer proyecto"). Distingue vacío-primera-vez / vacío-por-filtro (CTA limpiar filtros) / vacío-por-error (CTA retry). NUNCA "No results" ni pantalla en blanco.
- **Settings:** iOS **inset grouped list** con headers/footers, toggle a la derecha, chevron para subpantallas, destructivas en rojo al final, fila de perfil arriba; Android preference screen equivalente. Máx ~7 items por sección; cambios aplican al instante (sin "Guardar" salvo forms de perfil); los settings no son navegación.

---

## 8. iOS vs Android: qué respetar / qué se comparte

### Diferencias que un diseñador DEBE respetar

| Dimensión | iOS | Android |
|---|---|---|
| Unidades | pt (@1x/@2x/@3x) | dp / sp |
| Fuente sistema | SF Pro (+ SF Symbols) | Roboto / Google Sans (+ Material Symbols) |
| Back | "<" arriba-izq + edge swipe **solo borde izq** | System back universal (ambos bordes) + Predictive Back |
| Título | Large title 34pt → centrado | Izquierda en top app bar |
| Nav inferior | Tab bar 49pt (cápsula Liquid Glass) | Navigation bar 80dp (Expressive 64dp) con pill |
| Acción flotante | **No existe FAB** — no lo uses en iOS | FAB 56dp (small 40, large 96) canónico |
| Confirmaciones | Action sheet / Alert; destructivo rojo | Dialog M3 o snackbar + Undo |
| Deshacer destructivo | Confirmación previa | Snackbar con Undo (3-10s) |
| Estética 2025+ | Liquid Glass (translúcido flotante) | M3 Expressive (shapes redondeadas, springs, drawer deprecado) |

- NUNCA: portar la UI de una plataforma a la otra tal cual (FAB en iOS, switches iOS en Android, back button flotante en Android, action sheet estilo iOS en Android).

### Qué se comparte (diseña una vez)

Arquitectura de información, flujo y jerarquía · bottom navigation 3-5 destinos · thumb-zone y CTAs abajo · touch targets 44-48 y espaciado múltiplos de 8 · color semántico, dark mode, contraste AA · onboarding corto + permisos en contexto + empty states accionables · validación on-blur + teclados por tipo + autofill · skeletons/spinners/regla 300ms · iconografía y assets de marca.

---

**Sources:**
- Apple HIG — Tab Bars / Sheets / Layout / Navigation Bars / Lists and Tables / Pickers / Settings (developer.apple.com/design/human-interface-guidelines)
- iOS 26 Design Guidelines — learnui.design · Donny Wals (tab bars iOS 26)
- Material 3 — navigation-bar, navigation-drawer, bottom-sheets, dialogs, app-bars, lists, date pickers (m3.material.io) · material-components-android · 9to5google (M3 Expressive)
- NN/g — Hamburger Menus Hurt UX Metrics · Navigation discoverability · Skeletons vs Spinners
- Thumb zone: parachutedesign.ca · dev.to (thumb zones 2025) · Formularios: Smashing Magazine · UXPin
- Onboarding/permisos/empty states: Appcues · Mobbin · eleken.co · Safe areas: Apple forums · Android 15 edge-to-edge docs
- iOS vs Android: gendesigns.ai · Medium (Key UX Differences 2025) · useyourloaf.com (swipe actions)
</parameter>
</invoke>
