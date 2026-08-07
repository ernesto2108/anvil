# Pencil — Implementación de Design Recipes

Sintaxis específica de herramienta para construir patrones de recetas en archivos .pen.

> **Para patrones de composición completos**, llama `get_guidelines('guide','Web App')` o `get_guidelines('guide','Design System')` según el tipo de pantalla — no dupliques aquí lo que Pencil ya provee. Estas recetas cubren solo lo que NO está en los guidelines (Auth, App Shell, Nav Mobile, Modo Oscuro).

## API actual de Pencil (breaking changes)

La API de Pencil cambió. Usa SIEMPRE la columna "API actual". La API vieja produce errores o nodos rotos.

| Concepto | API vieja (NO usar) | API actual |
|---|---|---|
| Ícono | `type:"icon_font"`, `iconFontFamily:"lucide"`, `iconFontName:"home"` | `type:"icon"`, `library:"lucide"`, `icon:"home"` |
| Stroke | `stroke:{align:"inside",thickness:{bottom:1},fill:"$border"}` | `strokeAlignment:"inner"`, `stroke:"$--border"`, `strokeWidth:{bottom:1}` |
| Funciones | `I()`, `U()`, `R()`, `C()`, `D()` | `Insert()`, `Update()`, `Replace()`, `Copy()`, `Delete()` (los shorthands aún funcionan, pero los guidelines usan los nombres completos) |

**Stroke — detalle:**
- `strokeAlignment`: `"inner"` | `"outer"` | `"center"`
- `stroke`: el color/fill (ej. `"$--border"`)
- `strokeWidth`: número uniforme (`1`) o por lado (`{bottom:1}`, `{top:1,bottom:1}`)

## JS en batch_design

`batch_design` acepta JavaScript completo: `for` loops, arrays, object spreads, template strings. Esto reduce drásticamente las operaciones y tokens en diseños repetitivos (grids, filas de tabla, listas de métricas).

```javascript
for (const metric of [{label:"Total Users",value:"12,543"},{label:"Revenue",value:"$48.2K"}]) {
  metricCard=Insert(metrics,{type:"ref",ref:"cardId",width:"fill_container"})
  header=Replace(metricCard+"/headerSlotId",{type:"frame",layout:"vertical",gap:4,padding:24})
  Insert(header,{type:"text",content:metric.label,fill:"$--muted-foreground",fontSize:14})
  Insert(header,{type:"text",content:metric.value,fill:"$--foreground",fontSize:32,fontWeight:"600"})
}
```

Usa loops para cualquier patrón que se repita ≥3 veces. Mantén el límite de 25 operaciones efectivas por llamada.

## Convención de tokens (`$--*`)

El design system nativo de Pencil usa el prefijo `$--*`: `$--background`, `$--foreground`, `$--primary`, `$--muted-foreground`, `$--card`, `$--border`, `$--font-primary`, `$--font-secondary`, `$--radius-m`, `$--radius-pill`. Los tokens propios de un proyecto pueden tener cualquier nombre, pero al trabajar con el design system nativo de Pencil usa `$--*`.

## Reglas Generales

- `include_schema: true` solo en la PRIMERA llamada a `get_editor_state` por sesión. Todas las siguientes: `include_schema: false`
- Carga guidelines UNA VEZ, no por pantalla
- Máximo 25 operaciones por llamada a `batch_design`
- Usa bindings para referencias de padre dentro del mismo batch
- Después de `Copy`, usa `descendants` para overrides — NO llamadas `Update` separadas en hijos copiados (los IDs cambian)

## Referencia de Nombres de Iconos (Lucide)

Iconos comunes usados en B2B SaaS — nombres verificados:
```
Navigation: layout-dashboard, git-branch, play, menu, x, chevron-down, arrow-right
Actions: plus, check, trash-2, edit, search, log-out
Status: activity, circle-check, circle-alert, circle-play (NOT check-circle, alert-circle, play-circle)
UI: user, settings, moon, sun, eye, eye-off, mail, bell
```

**CRÍTICO:** Lucide v4+ usa el prefijo `circle-*` (circle-check), NO el sufijo `*-circle` (check-circle). Siempre verifica.

## Receta: Pantalla de Auth — Pencil

```javascript
// Step 1: Create screen frame (1 op)
screen=Insert(document,{type:"frame",name:"Screen: Login",layout:"horizontal",x:X,y:Y,width:1440,height:900,fill:"$--background",placeholder:true})

// Step 2: Brand panel (4 ops)
brand=Insert(screen,{type:"frame",layout:"vertical",width:560,height:"fill_container",fill:"$--primary",padding:48,justifyContent:"center",gap:24})
logo=Insert(brand,{type:"text",content:"AppName",fontFamily:"Inter",fontSize:42,fontWeight:"700",fill:"#FFFFFF"})
tagline=Insert(brand,{type:"text",content:"Tagline here",fontFamily:"Inter",fontSize:18,fill:"#FFFFFFCC",textGrowth:"fixed-width",width:400})
desc=Insert(brand,{type:"text",content:"Description",fontFamily:"Inter",fontSize:16,fill:"#FFFFFF99",textGrowth:"fixed-width",width:440})

// Step 3: Form panel + card (4 ops)
right=Insert(screen,{type:"frame",layout:"vertical",width:"fill_container",height:"fill_container",padding:48,justifyContent:"center",alignItems:"center"})
card=Insert(right,{type:"frame",layout:"vertical",width:400,height:"fit_content",gap:32})
title=Insert(card,{type:"text",content:"Título",fontFamily:"Inter",fontSize:24,fontWeight:"700",fill:"$--foreground"})
subtitle=Insert(card,{type:"text",content:"Subtítulo",fontFamily:"Inter",fontSize:14,fill:"$--muted-foreground",textGrowth:"fixed-width",width:400})

// Step 4: Form fields using component refs (5-8 ops depending on fields)
fields=Insert(card,{type:"frame",layout:"vertical",width:"fill_container",gap:20})
email=Insert(fields,{type:"ref",ref:"INPUT_GROUP_ID",width:"fill_container"})
Update(email+"/LABEL_ID",{content:"Email"})
Update(email+"/PLACEHOLDER_ID",{content:"correo@empresa.com"})
// ... more fields

// Step 5: Button + footer (3 ops)
btn=Insert(card,{type:"ref",ref:"BTN_PRIMARY_ID",width:"fill_container",justifyContent:"center"})
Update(btn+"/LABEL_ID",{content:"Iniciar sesión"})
footer=Insert(card,{type:"frame",layout:"horizontal",width:"fill_container",justifyContent:"center",gap:4})

// Step 6: Remove placeholder
Update("SCREEN_ID",{placeholder:false})
```

Total: ~18-22 ops en 2 llamadas a batch_design.

## Receta: App Shell — Pencil

```javascript
// Call 1: Structure (3 ops)
screen=Insert(document,{type:"frame",name:"Screen: PageName",layout:"vertical",x:X,y:Y,width:1440,height:"fit_content(900)",fill:"$--background",placeholder:true})
nav=Insert(screen,{type:"ref",ref:"NAVBAR_ID",width:"fill_container"})
main=Insert(screen,{type:"frame",layout:"vertical",width:"fill_container",padding:[32,48],gap:24})

// Call 2+: Content sections (varies)
// ... add content to main
```

## Receta: Fila de Tabla — Pencil

```javascript
// Standard 5-column row (7 ops)
row=Insert(TABLE_ID,{type:"frame",layout:"horizontal",width:"fill_container",padding:[14,20],strokeAlignment:"inner",stroke:"$--border",strokeWidth:{bottom:1}})
c1=Insert(row,{type:"frame",width:120})
c1t=Insert(c1,{type:"text",content:"ID",fontFamily:"JetBrains Mono",fontSize:13,fill:"$--foreground"})
c2=Insert(row,{type:"frame",width:"fill_container"})
c2t=Insert(c2,{type:"text",content:"Name",fontFamily:"Inter",fontSize:14,fill:"$--foreground"})
c3=Insert(row,{type:"frame",width:140})
c3b=Insert(c3,{type:"ref",ref:"BADGE_ID"})
```

## Receta: Nav Mobile — Pencil

```javascript
nav=Insert(screen,{type:"frame",layout:"horizontal",width:"fill_container",height:56,padding:[0,16],alignItems:"center",justifyContent:"space_between",fill:"$--card",strokeAlignment:"inner",stroke:"$--border",strokeWidth:{bottom:1}})
hamburger=Insert(nav,{type:"icon",library:"lucide",icon:"menu",width:24,height:24,fill:"$--foreground"})
logo=Insert(nav,{type:"text",content:"AppName",fontFamily:"Inter",fontSize:18,fontWeight:"700",fill:"$--primary"})
avatar=Insert(nav,{type:"frame",width:32,height:32,cornerRadius:"$--radius-pill",fill:"$--muted",layout:"horizontal",alignItems:"center",justifyContent:"center"})
avatarIcon=Insert(avatar,{type:"icon",library:"lucide",icon:"user",width:16,height:16,fill:"$--primary"})
```

## Receta: Copia en Modo Oscuro — Pencil

```javascript
// Copy light frame with dark theme
dark=Copy("LIGHT_FRAME_ID",document,{name:"Dark: ScreenName",positionDirection:"bottom",positionPadding:100,theme:{"mode":"dark"}})

// If it contains a NavBar ref, override theme icon:
// First batch_get to find the nav ref ID inside the copy
// Then: Update("NAV_REF_ID/THEME_ICON_ID",{icon:"sun"})

// If it contains avatar dropdown ref:
// Update("DROPDOWN_REF_ID/THEME_ICON_ID",{icon:"sun"})
// Update("DROPDOWN_REF_ID/THEME_TEXT_ID",{content:"Modo claro"})
// Update("DROPDOWN_REF_ID/SWITCH_ID",{fill:"$--primary",justifyContent:"end"})
```

## Errores Comunes a Evitar

1. **API vieja de iconos:** usa `type:"icon"` + `library` + `icon`, NUNCA `type:"icon_font"` + `iconFontFamily` + `iconFontName`
2. **API vieja de stroke:** usa `strokeAlignment` + `stroke` + `strokeWidth`, NUNCA `stroke:{align,thickness,fill}`
3. **Nombres de iconos:** Usa `circle-check` NO `check-circle`, `circle-alert` NO `alert-circle`
4. **Mover con bindings:** mover un nodo requiere IDs reales, no nombres de binding. Inserta primero, obtén el ID de la respuesta, luego mueve en el siguiente batch
5. **Copy + Update descendants:** Después de `Copy`, los IDs de hijos cambian. Usa `descendants` en la operación `Copy` misma, o `batch_get` la copia para encontrar los nuevos IDs
6. **Variable de font family:** `fontFamily:"$--font-primary"` puede mostrar advertencia en Pencil — si falla, usa `fontFamily:"Inter"` directamente (limitación conocida)
7. **`alignItems:"baseline"`:** No soportado en .pen. Usa `"end"` en su lugar
8. **Disciplina con placeholder:** Establece `placeholder:true` inmediatamente al crear/copiar un frame. Elimínalo solo cuando TODO el contenido esté agregado
