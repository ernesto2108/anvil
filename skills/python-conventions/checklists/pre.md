# Checklist Pre-Implementación

Antes de escribir código Python, verificar:

- [ ] Versión objetivo de Python — ¿usando features de 3.12+? (`type`, `X | Y`, `batched`)
- [ ] Type hints planeados — tipos de retorno, parámetros, Protocol para interfaces
- [ ] Modelo Pydantic para cualquier input externo — API, config, datos de archivo
- [ ] Estrategia de test — qué fixtures, qué mockear (solo límites)
- [ ] Dependencias — ¿hay una solución stdlib? (`itertools.batched`, `tomllib`, `asyncio.TaskGroup`)
- [ ] Para ML/data — dtype explícito (`float32`), batch size definido, estrategia de memoria para datos grandes
- [ ] Para async — ¿usando TaskGroup? ¿Timeouts definidos? ¿Límites de concurrencia?
