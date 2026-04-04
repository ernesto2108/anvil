# Pencil Implementation — Design Recipes

Tool-specific syntax for building recipe patterns in .pen files.

## General Rules

- `include_schema: true` only on FIRST `get_editor_state` call per session. All subsequent: `include_schema: false`
- Load guidelines ONCE, not per screen
- Max 25 operations per `batch_design` call
- Use bindings for parent references within same batch
- After Copy (C), use `descendants` for overrides — NOT separate U() calls on copied children (IDs change)

## Icon Name Reference (Lucide)

Common icons used in B2B SaaS — verified names:
```
Navigation: layout-dashboard, git-branch, play, menu, x, chevron-down, arrow-right
Actions: plus, check, trash-2, edit, search, log-out
Status: activity, circle-check, circle-alert, circle-play (NOT check-circle, alert-circle, play-circle)
UI: user, settings, moon, sun, eye, eye-off, mail, bell
```

**CRITICAL:** Lucide v4+ uses `circle-*` prefix (circle-check), NOT `*-circle` suffix (check-circle). Always verify.

## Recipe: Auth Screen — Pencil

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

Total: ~18-22 ops across 2 batch_design calls.

## Recipe: App Shell — Pencil

```javascript
// Call 1: Structure (3 ops)
screen=I(document,{type:"frame",name:"Screen: PageName",layout:"vertical",x:X,y:Y,width:1440,height:"fit_content(900)",fill:"$bg-page",placeholder:true})
nav=I(screen,{type:"ref",ref:"NAVBAR_ID",width:"fill_container"})
main=I(screen,{type:"frame",layout:"vertical",width:"fill_container",padding:[32,48],gap:24})

// Call 2+: Content sections (varies)
// ... add content to main
```

## Recipe: Table Row — Pencil

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

## Recipe: Mobile Nav — Pencil

```javascript
nav=I(screen,{type:"frame",layout:"horizontal",width:"fill_container",height:56,padding:[0,16],alignItems:"center",justifyContent:"space_between",fill:"$bg-surface",stroke:{align:"inside",thickness:{bottom:1},fill:"$border-default"}})
hamburger=I(nav,{type:"icon_font",iconFontFamily:"lucide",iconFontName:"menu",width:24,height:24,fill:"$text-primary"})
logo=I(nav,{type:"text",content:"AppName",fontFamily:"Inter",fontSize:"$font-size-lg",fontWeight:"$fw-bold",fill:"$color-primary-500"})
avatar=I(nav,{type:"frame",width:32,height:32,cornerRadius:"$radius-full",fill:"$color-primary-100",layout:"horizontal",alignItems:"center",justifyContent:"center"})
avatarIcon=I(avatar,{type:"icon_font",iconFontFamily:"lucide",iconFontName:"user",width:16,height:16,fill:"$color-primary-600"})
```

## Recipe: Dark Mode Copy — Pencil

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

## Common Mistakes to Avoid

1. **Icon names:** Use `circle-check` NOT `check-circle`, `circle-alert` NOT `alert-circle`
2. **Move with bindings:** `M()` requires actual node IDs, not binding names. Insert first, get the ID from response, then Move in next batch
3. **Copy + Update descendants:** After `C()`, child IDs change. Use `descendants` in the Copy operation itself, or `batch_get` the copy to find new IDs
4. **Font family variable:** `fontFamily: "$font-family"` may show warning in Pencil — use `fontFamily: "Inter"` directly (known Pencil limitation)
5. **`alignItems: "baseline"`:** Not supported in .pen. Use `"end"` instead
6. **Placeholder discipline:** Set `placeholder: true` immediately when creating/copying a frame. Remove only when ALL content is added
