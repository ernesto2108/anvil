# Business Rules — <ProjectName>

<!-- Invariantes de negocio que cruzan dominios. -->

last_updated: <YYYY-MM-DD>

## Invariantes globales

<!-- Reglas que aplican a todo el sistema, sin importar el dominio -->

### <Nombre de la regla>
- **Regla:** <enunciado de la invariante, ej. "el precio nunca puede ser negativo">
- **Dónde se aplica:** `<path>:<line>` (validación / guard)
- **Por qué:** <razón de negocio>

## Reglas por dominio

<!-- Una subsección por dominio con sus reglas específicas -->

### <dominio>
- **<Nombre de la regla>:** <enunciado>
  - Dónde: `<path>:<line>`
  - Por qué: <razón de negocio>

## Reglas cross-dominio

<!-- Reglas que involucran más de un dominio -->

### <Nombre de la regla>
- **Regla:** <enunciado, ej. "un pedido no puede cerrarse sin al menos un ítem">
- **Dominios involucrados:** <dominio A>, <dominio B>
- **Dónde se aplica:** `<path>:<line>`
- **Por qué:** <razón de negocio>
