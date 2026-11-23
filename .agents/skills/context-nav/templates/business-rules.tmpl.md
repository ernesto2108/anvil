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

## Modelo de autenticación y autorización

<!-- Cómo se autentica el sistema internamente y externamente -->

### Autenticación entre servicios
- **Mecanismo interno:** <ninguno — mismo cluster / mTLS / JWT interno / API key>
- **Razón:** <por qué se eligió este mecanismo>
- **Servicios que requieren auth interna:** <lista o "ninguno">

### Autenticación hacia el exterior
- **Mecanismo externo:** <JWT / API key / OAuth2 / ninguno>
- **Header utilizado:** `<Authorization: Bearer ... / X-API-Key: ...>`
- **Quién valida:** <nombre del servicio o gateway>

### Reglas de autorización
- <regla — ej: "solo el servicio de pagos puede escribir en la tabla transactions">
- <regla — ej: "los servicios internos no requieren token — están en la misma red privada">
