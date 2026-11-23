---
name: domain-entity-guardrails
description: Aplicar tipado estricto y opcionalidad explícita en entidades de dominio y value objects. Usar cuando se creen o modifiquen structs de dominio, al revisar código de la capa de dominio, o al detectar campos pointer, tipos `any` o `sql.Null*` en modelos de dominio.
user-invocable: false
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


Previene modelos de dominio frágiles mediante la aplicación de tipado estricto y opcionalidad explícita.

Usar cuando:
- se creen o modifiquen entidades de dominio o value objects
- se mapeen DTOs de DB/HTTP a modelos de dominio
- se revise código de la capa de dominio en un proyecto de arquitectura limpia

## Detección

Antes de aplicar las reglas, localiza la capa de dominio del proyecto:
1. Si existe `{context_path}` (context.md del proyecto), úsalo para encontrar las rutas de dominio
2. De lo contrario, busca patrones comunes: `internal/**/domain/`, `src/domain/`, `pkg/domain/`, `lib/domain/`
3. Aplica las reglas a cualquier struct/tipo en los directorios de dominio identificados

## Reglas

- evitar `any`/`interface{}` en entidades de dominio
- evitar campos pointer en entidades de dominio (`*string`, `*int`, `*time.Time`, etc.)
- NO introducir tipos wrapper `Optional*` en entidades de dominio
- modelar semánticas opcionales con valores cero concretos (ej., string vacío, `0`, `time.Time{}`) y documentar esa convención en comentarios de código cuando sea necesario
- representar payloads JSON con tipos concretos; preferir `json.RawMessage` en los límites a menos que exista un esquema estricto
- preferir enums/value objects sobre strings libres para campos con restricciones
- mantener los modelos de dominio agnósticos a la persistencia (sin `sql.Null*` en el dominio)

## Checklist Antes de Terminar

1. Buscar `any` en los archivos de dominio modificados
2. Buscar campos pointer en los structs de dominio modificados
3. Confirmar que las semánticas opcionales usan convenciones de valor cero (sin wrappers `Optional*`)
4. Ejecutar `gofmt` en los archivos editados

## Output

- Si se viola una regla: listar archivo + campo + corrección propuesta
- Si todo pasa: reportar "domain guardrails OK"
