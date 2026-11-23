# Pipeline Frontend — React

## Patrón de salida

```
{architecture_path}/<project>/
├── context-summary.md      # Stack, deps, estructura, estado, APIs
├── context-modules.md      # Módulos/features con rutas, componentes, llamadas API
├── context-risks.md        # XSS, almacenamiento de tokens, deps, PII (con fragmentos de código)
├── overview.md             # C4, árbol de rutas, diagrama de estado, flujo de autenticación
├── security-audit.md       # OWASP lado cliente con puntaje
└── modules/
    └── <module-name>.md    # Uno por módulo con diagrama de flujo de usuario
```

## Esqueleto

Archivo: `<docs>/07-references/frontend-skeleton.md`

Cubre: configuración del store, query/hooks de autenticación, configuración de i18n, guards de autenticación, helpers de rutas.

## Instrucciones para el Scanner

"El esqueleto describe patrones comunes en frontends de este tipo. NO re-explorar estos patrones. Enfocarse en lo que DIFIERE: módulos/features, endpoints API, slices de estado más allá de auth, árbol de rutas con guards, integraciones externas (Sentry, Stripe, PostHog), especificidades de la herramienta de build, TypeScript vs JS. En context-summary.md, referenciar el esqueleto para patrones comunes y detallar solo las diferencias. En context-risks.md, incluir FRAGMENTOS DE CÓDIGO (5-10 líneas) por cada hallazgo para que seguridad no tenga que re-leer los archivos."

Salida: `context-summary.md`, `context-modules.md`, `context-risks.md`

## Arquitecto — Overview

Inyectar: context-summary.md INLINE. NO inyectar modules ni risks.

Generar `overview.md` con:
1. Descripcion del frontend (que hace, quien lo usa, stack)
2. Diagrama de Contexto (C4) — frontend <-> backends <-> auth <-> 3rd parties
3. Arbol de rutas completo (Mermaid graph TD) — con guards y layouts
4. Diagrama de estado global (Mermaid) — store slices y relaciones
5. Auth flow diagram (Mermaid sequence) — login -> token -> refresh -> logout
6. Dependencias externas (tabla)
7. Notas tecnicas

## Arquitecto — Detalle (módulos)

Antes de lanzar, leer `<docs>/07-references/template-module.md`.

Inyectar: template-module.md + context-modules.md INLINE. NO inyectar summary ni risks.

Instrucciones: "Usar el template como referencia de formato EXACTA. NO leer archivos de ejemplo."

Salida: `modules/*.md` (uno por módulo/feature)

## Instrucciones de Seguridad (lado cliente)

"Auditoría LADO CLIENTE de un frontend React. Enfocarse en: (1) Almacenamiento de tokens — localStorage vs HttpOnly cookies, riesgo de robo por XSS. (2) XSS — dangerouslySetInnerHTML, input sin sanitizar, inyección de URL. (3) Datos sensibles en console.log, Sentry, APM, parámetros de URL. (4) CVEs de dependencias en package.json. (5) Headers CORS/CSP. (6) Bypass de autenticación — guards de rutas, validación solo en cliente. (7) Secretos hardcodeados. context-risks.md tiene FRAGMENTOS DE CÓDIGO — NO re-leer esos archivos. Solo usar Read para rastrear dependencias entre archivos."

Salida: `security-audit.md`, archivos de bugs en `{reports_path}/bugs/` (solo critical/high)

## Tabla de inyección rápida

| Agente | Inyectar INLINE | NO inyectar |
|---|---|---|
| Scanner | frontend-skeleton.md | — |
| Arquitecto (overview) | context-summary.md | modules, risks |
| Arquitecto (modules) | template-module.md + context-modules.md | summary, risks |
| Seguridad | known-systemic-issues.md + context-risks.md + overview summary | full modules |
