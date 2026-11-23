# Referencia de Estilos por Dominio (Banco Local)

Banco de estilos por industria — paletas (hex), pairings de Google Fonts, densidad/tono y productos de referencia por dominio.

**Propósito:** es el **fallback local** cuando no hay referencias inline ni investigación del explorer. El diseño por dominio nunca parte de cero. Cuando SÍ existe investigación real del explorer o referencias del humano, esa manda — este banco la complementa, no la reemplaza. Al usar este banco como origen, declararlo como "dirección basada en banco local de dominio, pendiente de validar con referencias reales".

Formato de paletas: primario / acento (+ soporte). Hex de marcas reales verificados o marcados ~aprox.

---

## 1. Fintech / Banca

| Paleta | Primario | Acento | Soporte | Sensación |
|---|---|---|---|---|
| Institucional (Stripe-like) | `#0A2540` navy | `#635BFF` blurple | `#F6F9FC`, `#00D4FF` info | trust + tech |
| Neobanco (Wise-like) | `#163300` forest | `#9FE870` lime | stone claros | dinero sin corbata |
| Trading dark (Robinhood-like) | `#0D1117` | `#00C805` ganancia | `#FF5000` pérdida | mercado en vivo |

- Nubank prueba que el violeta funciona si es identidad total: `#820AD1`.
- **Fuentes:** `Inter` + `Inter` (tabular figures — imprescindible para cifras); `IBM Plex Sans` + `IBM Plex Mono` (mono para montos = precisión); `Manrope` + `Inter` para neobancos.
- **Densidad/tono:** data-densa con cards generosas; corners 8-12px; sin playfulness en flujos de dinero; verde/rojo SOLO para ganancia/pérdida.
- **Referencias:** Stripe, Wise, Mercury, Revolut, Robinhood.

## 2. Salud / Medicina

| Paleta | Primario | Acento | Soporte | Sensación |
|---|---|---|---|---|
| Clínico calmado | `#0369A1` azul sereno | `#14B8A6` teal | `#F0F9FF` | competencia + calma |
| Telemedicina humana | `#0D9488` teal | `#FB7185` coral | stone claros | cuidado cercano |
| Farma/seguros | `#1E40AF` desat. ~C 0.12 | `#10B981` | `#F8FAFC` | institucional accesible |

- Azul = estabilidad, verde/teal = bienestar; acentos cálidos (coral, melocotón) humanizan sin perder trust cues.
- **Fuentes:** `Figtree` o `Nunito Sans` (humanistas redondeadas) + `Source Serif 4` en titulares; clásico `Lato` + `Crimson Text`.
- **Densidad/tono:** muy aireado, corners 12-16px, cuerpo ≥16px (audiencia amplia); rojo exclusivo a alertas médicas; AAA deseable, AA obligatorio.
- **Referencias:** Oscar Health, Zocdoc, One Medical, Ro.

## 3. SaaS B2B / Productividad

| Paleta | Primario | Acento | Soporte | Sensación |
|---|---|---|---|---|
| Neutral-first (Notion-like) | neutrales zinc | `#2383E2` azul único | blanco/negro tintados | herramienta invisible |
| Índigo con carácter (Linear-like) | `#5E6AD2` | mismo hue, L variable | slate teñido, dark-first | craft técnico |
| Bold identitario (Slack-like) | `#4A154B` aubergine | `#36C5F0`/`#ECB22E` | blancos limpios | marca fuerte |

- Aquí vive el mayor riesgo de "AI slop": si se usa índigo, diferenciarse por neutrales teñidos, densidad y tipografía, no por el hue.
- **Fuentes:** `Inter` sola con variación de peso (dominante en SaaS bien diseñado); `Manrope` + `Inter`; diferenciación 2025-2026: `Bricolage Grotesque` + `Inter`.
- **Densidad/tono:** compacta para power tools (grid 4/8px, corners 6-8px, filas 32-36px); aireada solo en marketing; dark mode de primera clase.
- **Referencias:** Linear, Notion, Slack, Airtable.

## 4. E-commerce / Retail

| Paleta | Primario | Acento | Soporte | Sensación |
|---|---|---|---|---|
| Marketplace (Etsy-like) | neutrales cálidos | `#F1641E` naranja CTA | blanco, stone | urgencia amable |
| Comercio confiable (Shopify-like) | `#008060` verde | `#FBF7ED` cream | `#212326` charcoal | negocio sólido |
| DTC premium | `#1A1A1A` casi negro | `#C2410C` terracota | stone 50-200 | editorial/craft |

- Regla central: **el producto es el color** — chrome neutral para que la foto domine; acento cálido solo en Add-to-cart y ofertas (Amazon reserva `#FF9900` así).
- **Fuentes:** `DM Sans` o `Plus Jakarta Sans` (catálogo limpio); DTC editorial: `Fraunces` (serif expresiva) + `Inter`.
- **Densidad/tono:** grids aireados, corners 8-12px; CTAs grandes y sticky en mobile; badges de descuento = único rojo permitido.
- **Referencias:** Shopify, Etsy, Glossier, Allbirds.

## 5. Educación

| Paleta | Primario | Acento | Soporte | Sensación |
|---|---|---|---|---|
| Académica (Coursera-like) | `#0056D2` azul | `#F59E0B` ámbar logro | slate claros | universidad online |
| Playful K-12 (Duolingo-like) | `#58CC02` verde | `#1CB0F6` + `#FFC800` | blanco, corners XL | juego que enseña |
| Edtech moderna (Quizlet-like) | `#4255FF` azul-violeta | `#FFCD1F` amarillo | neutrales fríos | estudio con energía |

- El ámbar/amarillo como color de progreso y logro (streaks, badges) es transversal del dominio.
- **Fuentes:** `Nunito` + `Open Sans` (look Duolingo); académico: `Lora` + `Source Sans 3`; moderna: `Poppins` + `Source Sans 3`.
- **Densidad/tono:** aireado; corners 12-16px (más en infantil); progreso siempre visible; único dominio B2C donde 4+ hues funcionan si se sistematizan.
- **Referencias:** Duolingo, Coursera, Khan Academy, Quizlet.

## 6. Fitness / Bienestar

| Paleta | Primario | Acento | Soporte | Sensación |
|---|---|---|---|---|
| Performance dark | `#0B0B0F` base | `#A3E635` lime eléctrico | grises fríos | rendimiento nocturno |
| Energía outdoor (Strava-like) | `#FC4C02` naranja | slate oscuros | blanco | esfuerzo y comunidad |
| Mindfulness | `#8B7EC8` lavanda ~aprox | `#86EFAC` sage | cremas cálidos | calma (Calm ~`#F47D31`) |

- Patrón 2026: dark mode por defecto, **un solo acento de alta energía** (lime, cyan, naranja neón) reservado a progreso/CTAs, y **cifras gigantes** para stats.
- **Fuentes:** `Archivo` (Expanded, pesos altos) + `Inter` — números oversized; impacto: `Bebas Neue` + `Inter`; bienestar: `Figtree` + `Lora`.
- **Densidad/tono:** performance = dark, denso, corners 8px; wellness = claro, muy aireado, corners 16-24px. Dos sub-dominios opuestos — no mezclarlos.
- **Referencias:** Strava, Whoop, Nike Training Club, Headspace.

## 7. Food / Delivery

| Paleta | Primario | Acento | Soporte | Sensación |
|---|---|---|---|---|
| Delivery bold (DoorDash-like) | `#EB1700` rojo | stone cálidos | blanco | apetito + urgencia |
| Fresh (Uber Eats-like) | `#06C167` verde | `#142328` verde-negro | blancos | fresco y rápido |
| Appetite cálido (Swiggy-like) | `#FC8019` naranja | `#1F2937` | cream `#FDF6EC` | antojo |

- Colores de apetito: rojo, naranja, verde-fresco (Zomato `#E23744`, Deliveroo `#00CCBC` outlier teal). Azul apenas existe en food — suprime el apetito.
- **Fuentes:** `Plus Jakarta Sans` o `Epilogue` (geométricas cálidas); playful: `Baloo 2` + `DM Sans`.
- **Densidad/tono:** foto de comida dominante → chrome neutral cálido (stone); cards grandes, corners 12-16px; CTA sticky gigante; tono urgente pero amable.
- **Referencias:** DoorDash, Uber Eats, Sweetgreen.

## 8. Real Estate

| Paleta | Primario | Acento | Soporte | Sensación |
|---|---|---|---|---|
| Portal masivo (Zillow-like) | `#006AFF` azul | `#0D9488` o coral | slate claros | búsqueda confiable |
| Luxury/brokerage | `#0F172A` charcoal | `#B08D57` bronce | blancos cálidos, serif | patrimonio |
| Proptech friendly | `#0D9488` teal | `#F59E0B` ámbar | stone | compra sin fricción |

- La fotografía manda: UI neutral y aireada, color solo en precio, CTA y estados de listing.
- **Fuentes:** portal: `Public Sans` o `Inter`; luxury: `Playfair Display` o `Cormorant Garamond` + `Lato`; clásico: `Libre Baskerville` + `IBM Plex Sans`.
- **Densidad/tono:** aireado, fotos edge-to-edge, corners 8-12px (0-4px en luxury — corners rectos leen premium); mapas desaturados para que los pins destaquen.
- **Referencias:** Zillow, Compass, Redfin.

## 9. Entretenimiento / Social

| Paleta | Primario | Acento | Soporte | Sensación |
|---|---|---|---|---|
| Streaming dark (Netflix-like) | `#141414` base | `#E50914` rojo | grises neutros | cine |
| Audio dark (Spotify-like) | `#121212` base | `#1DB954` verde | `#B3B3B3` texto | inmersión |
| Comunidad vibrante | `#5865F2` Discord / `#9146FF` Twitch | `#FE2C55` TikTok | dark canvas | pertenencia |

- Único dominio donde acentos ultra-saturados sobre negro son la norma — el contenido (portadas, video) aporta el color; el chrome es negro casi puro con UN acento.
- **Fuentes:** `Space Grotesk` o `Outfit` (geométricos con personalidad) + `Inter`/`DM Sans`; `Figtree` como proxy libre del Circular de Spotify.
- **Densidad/tono:** dark por defecto, denso en media grids, corners 8-16px; motion y hover son identidad; jerarquía por imagen, no por color de UI.
- **Referencias:** Spotify, Netflix, Twitch, Discord.

## 10. Developer Tools

| Paleta | Primario | Acento | Soporte | Sensación |
|---|---|---|---|---|
| Monochrome (Vercel-like) | negro/blanco | `#0070F3` azul único | grises neutros puros | precisión mínima |
| Terminal dark (Supabase-like) | `#1C1C1C` | `#3ECF8E` verde terminal | grises verdes | hacker-friendly |
| Docs light (GitHub-like) | blanco / `#0D1117` dark | `#0969DA` azul | `#57606A` texto sec. | estándar de industria |

- El respeto del dev se gana con contención: 1 acento, neutrales impecables, dark mode perfecto y **monospace para todo lo técnico** (código, IDs, rutas, hashes).
- **Fuentes:** `Geist` + `Geist Mono` (look Vercel); `Inter` + `JetBrains Mono`; carácter: `Space Grotesk` + `IBM Plex Mono`.
- **Densidad/tono:** la más compacta (filas 28-32px, corners 6-8px, grid 4px); tono sobrio, cero marketing-speak dentro del producto; syntax highlighting es la única zona multicolor.
- **Referencias:** Vercel, GitHub, Supabase, Stripe (docs/dashboard).

---

**Sources:**
- Matt Medley / Mantlr — Google Font pairings for UI 2025 · InspoAI — fintech palettes · ThinkPod — healthcare colors 2025
- Canvas Builder — fitness dark trends 2026 · Updivision — UI color trends 2026 · shadcn/ui — Tailwind colors reference
- Marcas reales verificadas (Stripe, Wise, Robinhood, Linear, Notion, Duolingo, Strava, Spotify, Vercel, etc.)
</parameter>
</invoke>
