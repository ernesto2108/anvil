# Figma — Implementación de Design Recipes

Sintaxis específica de herramienta para construir patrones de recetas en Figma via MCP.

> Este archivo se llenará cuando las recetas de Figma MCP sean probadas en la práctica.
> Los patrones abstractos de SKILL.md aplican — adáptalos a la Plugin API `use_figma` de Figma.

## Enfoque General

1. Usa `search_design_system` para encontrar componentes y variables existentes
2. Usa `use_figma` (siempre carga el skill `/figma-use` primero) para operaciones de escritura
3. Construye con equivalentes de Auto Layout de los patrones flex
4. Usa instancias de componentes con overrides de propiedades

## Diferencias Clave con Pencil

| Concepto | Pencil | Figma |
|---------|--------|-------|
| Variables | `set_variables` | Variables via Plugin API |
| Componentes | `reusable: true` | Component sets |
| Instancias | `type: "ref"` | Instance insertion |
| Overrides | `descendants` | Property overrides |
| Temas | `theme: {"mode": "dark"}` | Variable modes |
| Operaciones batch | `batch_design` (máx 25) | `use_figma` JS execution |

## Recetas

Las recetas siguen la misma estructura abstracta de SKILL.md. Las implementaciones específicas de Figma se documentarán a medida que sean probadas y validadas.
