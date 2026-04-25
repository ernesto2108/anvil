# Pencil — Implementación de Design Recipes

Sintaxis específica de herramienta para construir patrones de recetas en archivos .pen.

## Reglas Generales

- `include_schema: true` solo en la PRIMERA llamada a `get_editor_state` por sesión. Todas las siguientes: `include_schema: false`
- Carga guidelines UNA VEZ, no por pantalla
- Máximo 25 operaciones por llamada a `batch_design`
- Usa bindings para referencias de padre dentro del mismo batch
- Después de Copy (C), usa `descendants` para overrides — NO llamadas U() separadas en hijos copiados (los IDs cambian)

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
screen=I(document,{type:"frame",name:"Screen: Login",layout:"horizontal",x:X,y:Y,width:1440,height:900,fill:"$bg-page",placeholder:true})

// Step 2: Brand panel (4 ops)
brand=I(screen,{type:"frame",layout:"vertical",width:560,height:"fill_container",fill:"$color-primary-500",padding:48,justifyContent:"center",gap:24})
logo=I(brand,{type:"text",content:"AppName",fontFamily:"Inter",fontSize:42,fontWeight:"$fw-bold",fill:"#FFFFFF"})
tagline=I(brand,{type:"text",content:"Tagline here",fontFamily:"Inter",fontSize:"$font-size-lg",fill:"#FFFFFFCC",textGrowth:"fixed-width",width:400})
desc=I(brand,{type:"text",content:"Description",fontFamily:"Inter",fontSize:"$font-size-base",fill:"#FFFFFF99",textGrowth:"fixed-width",width:440})

// Step 3: Form panel + card (4 ops)
right=I(screen,{type:"frame",layout:"vertical",width:"fill_container",height:"fill_container",padding:48,justifyContent:"center",alignItems:"center"})
card=I(right,{type:"frame",layout:"vertical",width:400,height:"fit_content",gap:32})
title=I(card,{type:"text",content:"Título",fontFamily:"Inter",fontSize:"$font-size-2xl",fontWeight:"$fw-bold",fill:"$text-primary"})
subtitle=I(card,{type:"text",content:"Subtítulo",fontFamily:"Inter",fontSize:"$font-size-sm",fill:"$text-secondary",textGrowth:"fixed-width",width:400})

// Step 4: Form fields using component refs (5-8 ops depending on fields)
fields=I(card,{type:"frame",layout:"vertical",width:"fill_container",gap:20})
email=I(fields,{type:"ref",ref:"INPUT_GROUP_ID",width:"fill_container"})
U(email+"/LABEL_ID",{content:"Email"})
U(email+"/PLACEHOLDER_ID",{content:"correo@empresa.com"})
// ... more fields

// Step 5: Button + footer (3 ops)
btn=I(card,{type:"ref",ref:"BTN_PRIMARY_ID",width:"fill_container",justifyContent:"center"})
U(btn+"/LABEL_ID",{content:"Iniciar sesión"})
footer=I(card,{type:"frame",layout:"horizontal",width:"fill_container",justifyContent:"center",gap:4})

// Step 6: Remove placeholder
U("SCREEN_ID",{placeholder:false})
```

Total: ~18-22 ops en 2 llamadas a batch_design.

## Receta: App Shell — Pencil

```javascript
// Call 1: Structure (3 ops)
screen=I(document,{type:"frame",name:"Screen: PageName",layout:"vertical",x:X,y:Y,width:1440,height:"fit_content(900)",fill:"$bg-page",placeholder:true})
nav=I(screen,{type:"ref",ref:"NAVBAR_ID",width:"fill_container"})
main=I(screen,{type:"frame",layout:"vertical",width:"fill_container",padding:[32,48],gap:24})

// Call 2+: Content sections (varies)
// ... add content to main
```

## Receta: Fila de Tabla — Pencil

```javascript
// Standard 5-column row (7 ops)
row=I(TABLE_ID,{type:"frame",layout:"horizontal",width:"fill_container",padding:[14,20],stroke:{align:"inside",thickness:{bottom:1},fill:"$border-default"}})
c1=I(row,{type:"frame",width:120})
c1t=I(c1,{type:"text",content:"ID",fontFamily:"JetBrains Mono",fontSize:13,fill:"$text-primary"})
c2=I(row,{type:"frame",width:"fill_container"})
c2t=I(c2,{type:"text",content:"Name",fontFamily:"Inter",fontSize:"$font-size-sm",fill:"$text-primary"})
c3=I(row,{type:"frame",width:140})
c3b=I(c3,{type:"ref",ref:"BADGE_ID"})
```

## Receta: Nav Mobile — Pencil

```javascript
nav=I(screen,{type:"frame",layout:"horizontal",width:"fill_container",height:56,padding:[0,16],alignItems:"center",justifyContent:"space_between",fill:"$bg-surface",stroke:{align:"inside",thickness:{bottom:1},fill:"$border-default"}})
hamburger=I(nav,{type:"icon_font",iconFontFamily:"lucide",iconFontName:"menu",width:24,height:24,fill:"$text-primary"})
logo=I(nav,{type:"text",content:"AppName",fontFamily:"Inter",fontSize:"$font-size-lg",fontWeight:"$fw-bold",fill:"$color-primary-500"})
avatar=I(nav,{type:"frame",width:32,height:32,cornerRadius:"$radius-full",fill:"$color-primary-100",layout:"horizontal",alignItems:"center",justifyContent:"center"})
avatarIcon=I(avatar,{type:"icon_font",iconFontFamily:"lucide",iconFontName:"user",width:16,height:16,fill:"$color-primary-600"})
```

## Receta: Copia en Modo Oscuro — Pencil

```javascript
// Copy light frame with dark theme
dark=C("LIGHT_FRAME_ID",document,{name:"Dark: ScreenName",positionDirection:"bottom",positionPadding:100,theme:{"mode":"dark"}})

// If it contains a NavBar ref, override theme icon:
// First batch_get to find the nav ref ID inside the copy
// Then: U("NAV_REF_ID/THEME_ICON_ID",{iconFontName:"sun"})

// If it contains avatar dropdown ref:
// U("DROPDOWN_REF_ID/THEME_ICON_ID",{iconFontName:"sun"})
// U("DROPDOWN_REF_ID/THEME_TEXT_ID",{content:"Modo claro"})
// U("DROPDOWN_REF_ID/SWITCH_ID",{fill:"$color-primary-500",justifyContent:"end"})
```

## Errores Comunes a Evitar

1. **Nombres de iconos:** Usa `circle-check` NO `check-circle`, `circle-alert` NO `alert-circle`
2. **Mover con bindings:** `M()` requiere IDs de nodo reales, no nombres de binding. Inserta primero, obtén el ID de la respuesta, luego mueve en el siguiente batch
3. **Copy + Update descendants:** Después de `C()`, los IDs de hijos cambian. Usa `descendants` en la operación Copy misma, o `batch_get` la copia para encontrar los nuevos IDs
4. **Variable de font family:** `fontFamily: "$font-family"` puede mostrar advertencia en Pencil — usa `fontFamily: "Inter"` directamente (limitación conocida de Pencil)
5. **`alignItems: "baseline"`:** No soportado en .pen. Usa `"end"` en su lugar
6. **Disciplina con placeholder:** Establece `placeholder: true` inmediatamente al crear/copiar un frame. Elimínalo solo cuando TODO el contenido esté agregado
