# Dependencies — <ProjectName>

<!-- Grafo de dependencias entre dominios. -->

last_updated: <YYYY-MM-DD>

## Grafo de dependencias

<!-- Tipo: sync (llamada directa), async (evento/queue), data (FK / esquema compartido) -->

| Dominio | Depende de | Tipo | Notas |
|---------|-----------|------|-------|
| `<dominio A>` | `<dominio B>` | sync / async / data | <qué consume y dónde> |

## Impacto de cambios

Antes de modificar un dominio, consultar la tabla del grafo para identificar los
dominios **downstream** afectados (los que dependen del dominio que vas a tocar).

- Si la dependencia es `sync` → un cambio de contrato rompe a quien depende de inmediato.
- Si la dependencia es `async` → verificar compatibilidad del payload del evento.
- Si la dependencia es `data` → verificar migraciones y FKs antes de alterar el esquema.

Listar los dominios downstream en el plan de cambio y validar cada uno antes de cerrar.

## Dependencias externas

<!-- Servicios externos que el sistema consume: APIs de terceros, DBs externas, colas -->

| Servicio externo | Tipo | Consumido por | Notas |
|------------------|------|---------------|-------|
| `<servicio>` | API REST / DB / queue | `<dominio>` | <auth, rate limits, SLA conocido> |
