# Referencia de Combinación de Color (Color Craft)

Cómo **generar rampas** desde un color de marca, **combinar** primario/secundario/acento, elegir **neutrales**, derivar **dark mode** y **verificar contraste WCAG calculando** (sin herramientas web). Complementa `primitives.md` (que da las escalas) y `semantic-tokens.md` (que da el mapeo) con la teoría para no inventar valores a ojo.

Fuentes: Evil Martians (OKLCH), Tailwind v4 (valores OKLCH oficiales), Radix Colors, Refactoring UI, Material Design 3, WCAG 2.x.

---

## 1. Por qué OKLCH (no HSL)

- La **L de HSL no es perceptual**: `hsl(60 100% 50%)` (amarillo) y `hsl(240 100% 50%)` (azul) declaran la misma lightness pero el amarillo se ve mucho más claro. Variar L en HSL produce pasos desiguales y contrastes impredecibles.
- En **OKLCH** (`oklch(L C H)`) un cambio de 0.1 en L se ve igual en cualquier hue → rampas perceptualmente uniformes, generables por algoritmo.
- OKLCH mantiene el **hue estable** al variar lightness (HSL vira, especialmente azules → púrpura). Soporta gamut P3.
- Precaución única: no todo par L/C existe en sRGB → hay que hacer **gamut clamping** (reducir C hasta que sea representable, nunca tocar L).

---

## 2. Generar la rampa 50→950 desde un hex de marca

### 2.1 Curva de lightness de referencia (Tailwind v4)

Rampa `blue` oficial, usable como plantilla de L y curva de chroma para cualquier hue:

| Paso | L | C | H |
|---|---|---|---|
| 50 | 0.970 | 0.014 | 254.6 |
| 100 | 0.932 | 0.032 | 255.6 |
| 200 | 0.882 | 0.059 | 254.1 |
| 300 | 0.809 | 0.105 | 251.8 |
| 400 | 0.707 | 0.165 | 254.6 |
| 500 | 0.623 | 0.214 | 259.8 |
| 600 | 0.546 | 0.245 | 262.9 (**pico de chroma**) |
| 700 | 0.488 | 0.243 | 264.4 |
| 800 | 0.424 | 0.199 | 265.6 |
| 900 | 0.379 | 0.146 | 265.5 |
| 950 | 0.282 | 0.091 | 267.9 |

Curva de L generalizable: `0.97 / 0.93 / 0.88 / 0.81 / 0.71 / 0.62 / 0.55 / 0.49 / 0.42 / 0.38 / 0.28`. No es lineal: pasos claros comprimidos (~0.04), medios expandidos (~0.07-0.09), porque el ojo discrimina más en la zona clara.

### 2.2 Campana de chroma (regla crítica)

El chroma **NO es constante**: campana con pico en 500-600 y caída fuerte a ambos extremos. Aproximación sobre `C_max` (el chroma del paso 600):

`50 ≈ 6% · 100 ≈ 13% · 200 ≈ 24% · 300 ≈ 43% · 400 ≈ 67% · 500 ≈ 87% · 600 = 100% · 700 ≈ 99% · 800 ≈ 81% · 900 ≈ 60% · 950 ≈ 37%`

Si mantienes C constante: los claros salen chillones/imposibles en sRGB y los oscuros salen sucios.

### 2.3 Algoritmo (paso a paso)

```
1. Convertir hex de marca a OKLCH (L, C, H).
2. Localizar el paso "hogar": aquel cuya L de referencia esté más cerca de la
   L del hex (típico 500 o 600). El hex de marca SE PRESERVA EXACTO en ese paso.
3. Para cada otro paso i:
   L_i = valor de la curva de referencia
   C_i = C_marca * (campana_i / campana_hogar)
   H_i = H_marca (+/- drift opcional 2-8° hacia los oscuros, como Tailwind: 254→268)
4. Gamut clamp: si oklch(L_i C_i H_i) cae fuera de sRGB, reducir C_i hasta
   entrar (nunca tocar L — L gobierna el contraste).
5. Verificar: 600+ sobre blanco >= 4.5:1; 50-100 sirven de fondo con texto 700+.
```

### 2.4 Mapeo semántico de pasos (Radix ↔ 50-950)

| Rol | Radix | Escala 50-950 |
|---|---|---|
| Fondo de app | 1 | 50 |
| Fondo sutil de componente | 2 | 100 |
| Fondo componente normal/hover/active | 3-5 | 100-200 |
| Borde sutil (cards, separadores) | 6 | 200 |
| Borde interactivo / focus ring | 7-8 | 300 |
| Fondo sólido (botón primario) — máx chroma | 9-10 | 500-600 |
| Texto de bajo contraste | 11 | 700 |
| Texto de alto contraste | 12 | 900-950 |

Material 3 usa **tonal palettes** de 13 tonos (0=negro … 100=blanco) y roles: light primary=tone 40, container=tone 90, on-container=tone 10; dark primary=tone 80, container=tone 30.

---

## 3. Combinar colores (secundario y acento)

### 3.1 Armonías en UI

| Armonía | Rotación H | Veredicto |
|---|---|---|
| **Análoga** | ±30-50° | La más segura; ideal para **secundario** ("misma familia") |
| **Complementaria** | 180° | Solo como acento pequeño (10%); a partes iguales vibra |
| **Split-complementary** | 180° ± 30° | Mejor compromiso para **acento**: alto contraste sin el choque directo |
| **Triádica** | ±120° | Casi nunca en UI de producto — reservar a ilustración/marketing |
| **Monocromática** | 0° (varía L/C) | Subestimada: Notion, Vercel, Linear son casi monocromos + 1 acento |

### 3.2 Reglas operativas

- **Secundario** = rotar H ±30-40° manteniendo L/C similares; úsalo casi siempre en pasos claros (100-200) como fondos/tags — no como segundo color de botón.
- **Acento** = split-complementary (H+150° o H+210°) o complementario, C alto, **exclusivo de acciones**: CTA, links, estados activos, focus. "Contraste por escasez": si el acento solo aparece donde se actúa, el usuario aprende dónde actuar.
- **60-30-10** en interfaces: ~60% superficie neutral, ~30% primario/secundario de soporte, ~10% acento. En producto real la proporción es aún más extrema: ~90% neutral / ~8% primario / ~2% acento.
- **Máximo de hues:** 3 cromáticos intencionales (primario, secundario, acento) + 1 familia neutral + los 4 semánticos de status. Los semánticos son vocabulario funcional, no paleta decorativa. Cada hue se expande en su rampa completa (33 tokens sin romper la regla).

---

## 4. Neutrales

### 4.1 Temperatura del gris según el hue de marca

Nunca gris puro por defecto — elegir familia según el primario (nomenclatura Tailwind):

| Familia | Undertone | H aprox | Usar cuando el primario es… |
|---|---|---|---|
| `slate` | azul frío | ~250-265 | azul, índigo, violeta — look tech |
| `gray` | azul leve | ~260 | azul templado, verde azulado |
| `zinc` | casi neutro frío | ~285 | violeta, magenta, monocromo |
| `neutral` | sin cast | — | multicolor (data viz), fotografía |
| `stone` | cálido marrón | ~50-80 | naranja, amarillo, terracota, orgánico/editorial |

### 4.2 Grises teñidos

- Genera los neutrales como rampa con **H = hue del primario** y **C muy bajo**: **C entre 0.005 y 0.045 en OKLCH** (HSL: sat 2-8%). Referencia: `slate` v4 usa C 0.003-0.046 con H ~250-266 — "azul al 4%".
- Los extremos (50 y 900-950) toleran algo más de chroma que los medios. La rampa neutral baja hasta L ~0.13 (más oscuro que las cromáticas) porque el 900-950 neutral es el texto principal.
- **Prohibido mezclar dos familias de gris** en una misma UI.

---

## 5. Dark mode

1. **No invertir la rampa.** Se re-deriva: tema nuevo con la misma H de marca, no `950↔50`.
2. **Fondo base: nunca negro puro.** Rango `#101014`–`#1A1A1E` (L ≈ 0.14-0.22). Negro `#000` provoca smearing OLED y contraste doloroso.
3. **Texto: nunca blanco puro.** `#E4E4E7`–`#F4F4F5` (L ≈ 0.92-0.96) principal, secundario ~`#A1A1AA`. Blanco puro produce glare/halation.
4. **Desatura los colores de marca.** El primario sube 1-2 pasos (usa el 400 donde en light usabas el 600) y reduce C ~20-30%. Material: primary tone 40 → **tone 80**.
5. **Elevación = más claro, no sombra.** Superficies por lightness ascendente (L): base 0.15 → card 0.19 → modal 0.23 → tooltip 0.27 (hex aprox `#111113 → #1B1B1F → #242429 → #2E2E33`).
6. **Tiñe los fondos oscuros** con el hue de marca (C 0.01-0.03) en vez de gris muerto.
7. Los semánticos también se aclaran/desaturan (`red-600` → `red-400`).
8. Contraste AA se exige igual en dark; error típico: texto gris medio (L~0.6) sobre superficie L~0.25 que no llega a 4.5:1.

---

## 6. Contraste WCAG calculado a mano (SIN herramientas)

Los subagentes no tienen acceso web — el contraste se **calcula con esta fórmula**, no se asume ni se delega a WebAIM.

### 6.1 Fórmula exacta

```
Paso 1 — Normalizar cada canal sRGB a [0,1]:  c = valor_hex / 255   (R, G, B)
Paso 2 — Linearizar cada canal:
  si c <= 0.03928:  c_lin = c / 12.92
  si c >  0.03928:  c_lin = ((c + 0.055) / 1.055) ^ 2.4
Paso 3 — Luminancia relativa:
  L = 0.2126*R_lin + 0.7152*G_lin + 0.0722*B_lin
Paso 4 — Contrast ratio (L1 = mayor, L2 = menor):
  CR = (L1 + 0.05) / (L2 + 0.05)      → rango 1:1 a 21:1
```

### 6.2 Umbrales

| Nivel | Texto normal | Texto grande (≥24px, o ≥18.66px bold) | UI components / gráficos |
|---|---|---|---|
| AA | ≥ 4.5:1 | ≥ 3:1 | ≥ 3:1 (WCAG 1.4.11) |
| AAA | ≥ 7:1 | ≥ 4.5:1 | — |

### 6.3 Ejemplos calculados

**Ej. 1 — `#2563EB` (blue-600) sobre blanco:**
```
R=37/255=0.1451 → 0.0185 ; G=99/255=0.3882 → 0.1248 ; B=235/255=0.9216 → 0.8308
L = 0.2126*0.0185 + 0.7152*0.1248 + 0.0722*0.8308 = 0.1532
CR = (1.0+0.05)/(0.1532+0.05) = 5.17  → PASA AA texto normal, falla AAA
```

**Ej. 2 — blanco sobre `#22C55E` (green-500), botón típico:**
```
R=34/255 → 0.0182 ; G=197/255 → 0.5584 ; B=94/255 → 0.1119
L = 0.4113 ;  CR = 1.05/0.4613 = 2.28  → FALLA AA
(por eso el texto sobre verdes/ámbar medios debe ser oscuro, o el fill baja a green-700)
```

**Ej. 3 — texto `#6B7280` (gray-500) sobre blanco:**
```
R=107/255 → 0.1470 ; G=114/255 → 0.1666 ; B=128/255 → 0.2140
L = 0.1659 ;  CR = 1.05/0.2159 = 4.86  → PASA AA justo
(gray-400 #9CA3AF ~2.5:1 NO sirve para texto — solo placeholders)
```

### 6.4 Heurística rápida (pre-filtro)

Vía OKLCH: **ΔL ≥ ~0.40** entre texto y fondo suele equivaler a ≥4.5:1; **ΔL ≥ ~0.30** a ≥3:1. Útil para pre-filtrar, luego **verificar con la fórmula** — la heurística no sustituye el cálculo.

**APCA (futuro WCAG 3):** contraste de lightness percibida (Lc ~0-106), dependiente de polaridad, tamaño y peso. Guías: Lc 90 ≈ body pequeño, Lc 75 ≈ body, Lc 60 ≈ texto grande, Lc 45 ≈ headlines gruesos, Lc 30 ≈ mínimo texto. Aún no normativo en 2026: cumplir WCAG 2.x, usar APCA como criterio adicional.

---

## 7. Errores que hacen que una UI parezca "generada por IA"

1. **Indigo/púrpura por defecto + gradiente violeta** (`indigo-500 → purple-600`) — es el promedio del training data. Si el dominio no pide violeta, no usar violeta; gradiente = mismo hue variando solo L/C.
2. **Saturación uniforme en toda la UI** — las UIs pro son ~90% neutrales de C bajo con chroma concentrado en 1-2 puntos de acción.
3. **Colores de status como decoración** — si el rojo aparece en un icono decorativo, deja de significar error.
4. **Grises de familias mezcladas** (slate + zinc + stone juntos) o gris puro junto a un primario cromático.
5. **Negro `#000` y blanco `#FFF` puros** como texto/fondo — usar 950 del neutral teñido y off-white.
6. **Demasiados hues** (>3 cromáticos) o acento aplicado a todo (si todo es acento, nada lo es).
7. **Texto sobre fills medios sin verificar contraste** (blanco sobre green-500/amber-400 — ver Ej. 2).
8. **Combo "AI slop":** Inter + gradiente púrpura + 3 cards con icono en grid + glassmorphism + emoji en headings. Cualquier par ya es señal.
9. **Sombras genéricas grandes y negras** en vez de sombras teñidas con el hue del fondo, sutiles y por capas.
10. **Dark mode invertido mecánicamente** (negros puros con los mismos colores saturados de light → vibración óptica).

---

**Sources:**
- Evil Martians — OKLCH in CSS: why we moved from RGB and HSL · OK, OKLCH color picker
- Tailwind CSS — Colors (valores OKLCH oficiales v4) · Radix Colors — Understanding the scale
- Refactoring UI — Building your color palette · Material Design 3 — Key colors and tones · Material — Dark theme
- Atmos — Dark mode UI best practices · Anna Filou — color scales in OKLCH · Hype4 — 60-30-10 rule
- Braingrid / 925 Studios / SmoothUI — AI slop design tells · WCAG 2.1 (1.4.3, 1.4.11) · APCA / WCAG 3 draft
</parameter>
</invoke>
