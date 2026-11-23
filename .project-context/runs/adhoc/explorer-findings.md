# Exploración — findings

**Run:** adhoc
**Fecha:** 2026-09-23

## Objetivo
Evaluar si las skills `*-conventions` (go, react, flutter, swift, typescript, python, astro, rust, devops) son redundantes con el conocimiento nativo de modelos Claude actuales.

## Fuentes consultadas

| Fuente | Tipo | Origen |
|---|---|---|
| go-conventions/SKILL.md | local | /Users/ernestodiaz/projects/anvil/skills/go-conventions/SKILL.md |
| react-conventions/SKILL.md | local | /Users/ernestodiaz/projects/anvil/skills/react-conventions/SKILL.md |
| flutter-conventions/SKILL.md | local | /Users/ernestodiaz/projects/anvil/skills/flutter-conventions/SKILL.md |
| swift-conventions/SKILL.md | local | /Users/ernestodiaz/projects/anvil/skills/swift-conventions/SKILL.md |
| typescript-conventions/SKILL.md | local | /Users/ernestodiaz/projects/anvil/skills/typescript-conventions/SKILL.md |
| python-conventions/SKILL.md | local | /Users/ernestodiaz/projects/anvil/skills/python-conventions/SKILL.md |
| astro-conventions/SKILL.md | local | /Users/ernestodiaz/projects/anvil/skills/astro-conventions/SKILL.md |
| rust-conventions/SKILL.md | local | /Users/ernestodiaz/projects/anvil/skills/rust-conventions/SKILL.md |
| devops-conventions/SKILL.md | local | /Users/ernestodiaz/projects/anvil/skills/devops-conventions/SKILL.md |
| NAVIGATOR.md | local | /Users/ernestodiaz/projects/anvil/.project-context/NAVIGATOR.md (49 líneas, gate pasa) |

## Hallazgos clave
- Todas las skills go/typescript/python/rust son "dispatchers ligeros" (~40-110 líneas) que apuntan a archivos hijos (rules/, guides/, examples/) no leídos en este run — el SKILL.md raíz es solo el índice, el contenido denso vive en subarchivos no evaluados.
- react-conventions (163 líneas) y swift-conventions (102 líneas) son monolíticos con contenido inline específico (versiones de librería, tablas de decisión, checklists).
- Contenido claramente genérico de libro de texto: reglas básicas de naming, "usa const", "value types sobre clases", "errores como valores", checklists de accesibilidad estándar (WCAG).
- Contenido con valor real no inferible por el modelo solo: decisiones de stack fijadas por el equipo (Tailwind v4 sintaxis `bg-(--var)` en react-conventions:98-106; iOS 17 mínimo + Swift 6.3/Xcode 26 en swift:8,19; SwiftData vs Core Data; GoRouter obligatorio en flutter:21; `lucide-react` como única librería de iconos en react:73-92; tabla de gestión de estado por escala en react:56-64 y flutter:45-52); reglas de integración con el sistema anvil (gate post-implementación que invoca skill `/lint`, detección de anti-patrones con formato `[file:line][severity]`).

## Gaps y preguntas abiertas
- No se leyeron los archivos hijos referenciados por los dispatchers (rules/*, guides/*, detection/anti-patterns.md) — el contenido denso real de go/typescript/python/rust vive ahí, no en el SKILL.md raíz evaluado.
