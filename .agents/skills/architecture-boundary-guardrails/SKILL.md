---
name: architecture-boundary-guardrails
description: Previene la deriva arquitectónica haciendo cumplir bounded contexts y la estructura de un caso de uso por archivo. Usar al crear nuevos servicios, mover código entre límites de dominio, o detectar imports cross-context, god-interfaces, o archivos de servicio monolíticos.
user-invocable: false
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


Previene la deriva arquitectónica haciendo cumplir bounded contexts y la estructura de un caso de uso por archivo.

Usar cuando:
- se crea o mueve código entre límites de dominio
- se implementan capas application/domain/ports en un proyecto de arquitectura limpia
- se agregan nuevos servicios, handlers, workers o repositorios

## Detección

Antes de aplicar las reglas, detectar los bounded contexts del proyecto:
1. Leer la estructura del proyecto para identificar módulos/carpetas de dominio
2. Si existe `{context_path}` (context.md del proyecto), usarlo para los límites de contexto
3. De lo contrario, inferir los contextos a partir de los directorios de dominio de nivel superior (ej. `internal/`, `src/domains/`, `packages/`)

## Reglas principales

- no mezclar bounded contexts en un mismo módulo/carpeta
- cada contexto de dominio tiene sus propias entidades, value objects y ports
- la capa application debe estar orientada a casos de uso:
  - un caso de uso por archivo
  - evitar archivos de servicio grandes que acumulen operaciones no relacionadas
- mantener los ports con alcance por contexto; evitar god-interfaces compartidas
- si un nuevo archivo toca dos contextos, detenerse y separar

## Lista de verificación pre-implementación

1. Identificar el o los bounded context(s) objetivo para la tarea
2. Confirmar que la carpeta de destino pertenece a ese contexto
3. Si el código abarca contextos, definir ports/contratos explícitos entre ellos
4. Asegurarse de que cada caso de uso tenga su propio archivo
5. Marcar los cambios con impacto arquitectónico antes de codificar refactorizaciones

## Validaciones

- sin filtración de entidades cross-context sin contratos de port
- sin archivos de aplicación monolíticos que manejen múltiples casos de uso no relacionados
- los imports se mantienen direccionales (domain <- application <- infrastructure)

## Salida

- Si existe violación: reportar archivo, tipo de violación y plan de separación
- Si está en cumplimiento: reportar "architecture guardrails OK"
