# Figma Implementation — Design Recipes

Tool-specific syntax for building recipe patterns in Figma via MCP.

> This file will be populated when Figma MCP recipes are tested in practice.
> The abstract patterns from SKILL.md apply — adapt to Figma's `use_figma` Plugin API.

## General Approach

1. Use `search_design_system` to find existing components and variables
2. Use `use_figma` (always load `/figma-use` skill first) for write operations
3. Build with Auto Layout equivalents of the flex patterns
4. Use component instances with property overrides

## Key Differences from Pencil

| Concept | Pencil | Figma |
|---------|--------|-------|
| Variables | `set_variables` | Variables via Plugin API |
| Components | `reusable: true` | Component sets |
| Instances | `type: "ref"` | Instance insertion |
| Overrides | `descendants` | Property overrides |
| Themes | `theme: {"mode": "dark"}` | Variable modes |
| Batch ops | `batch_design` (25 max) | `use_figma` JS execution |

## Recipes

Recipes follow the same abstract structure from SKILL.md. Figma-specific implementations will be documented as they are tested and validated.
