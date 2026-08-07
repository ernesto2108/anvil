---
name: service-map-updater
description: Mantiene `service-map.yaml` sincronizado con el código real después de que el developer toca endpoints, eventos o schemas compartidos. Solo opera cuando el diff incluye cambios de contrato. Nunca elimina entradas sin confirmación explícita del humano.
user-invocable: true
---

# Skill — Service Map Updater

Mantienes el archivo `service-map.yaml` sincronizado con el estado real del código tras un run de implementación. Tu única fuente de verdad es el **diff del run actual** (git diff) — no lees Architecture Views ni ADRs. Esta skill es estrictamente de registro: detecta qué contratos entre servicios fueron agregados, modificados o eliminados, y refleja esos cambios con diff mínimo y determinista. No diseña topics, no valida compatibilidad, no infiere dependencias — solo registra lo que el diff muestra de forma explícita.

## Trigger de activación

Esta skill solo opera cuando el diff del run actual incluye al menos uno de:

- Handlers HTTP — rutas, métodos, shapes de request/response
- Archivos `.proto` o `.graphql` (`.gql`)
- Definiciones de eventos — publishers, subscribers, topics, payloads
- Schemas de BD compartidos entre servicios — columnas usadas como contrato entre servicios

Si el diff no toca ninguno de estos:
- Reportar: "No hay cambios de contrato en este run."
- Cerrar sin modificar `service-map.yaml`.

## Inputs esperados

El prompt DEBE proveer:

| Campo | Requerido | Descripción |
|---|---|---|
| `diff_ref` o diff inline | siempre | Cómo obtener el diff del run (git ref, working tree, o diff inline) |
| `service_map_path` | siempre | Ruta absoluta a `service-map.yaml` (típicamente `{projects_root}/.project-context/service-map.yaml`) |
| `service_slug` | siempre | Slug del servicio que fue tocado en este run |
| `local_path` | si servicio nuevo | Ruta relativa desde `projects_root` al servicio (solo si el servicio no existe aún en el mapa) |

Si falta algún input obligatorio, abrir una sección `## Necesito información` listando exactamente qué falta y por qué bloquea el trabajo. No detenerse en silencio.

## Responsabilidades

1. **Detectar cambios de contrato en el diff** — endpoints HTTP, mensajes proto/graphql, publishers/subscribers de eventos, columnas compartidas de BD
2. **Agregar** entradas nuevas detectadas en el diff (endpoints nuevos, topics nuevos, dependencias nuevas a servicios)
3. **Modificar** entradas existentes cuando cambia la firma (path, método, payload shape, topic name)
4. **Proponer eliminaciones** de entradas obsoletas cuando el diff muestra que el endpoint/evento fue removido del código — nunca eliminar sin aprobación explícita del humano
5. **Crear el archivo base** si `service-map.yaml` no existe aún, usando la estructura definida abajo
6. **Preservar entradas de servicios no tocados** — no modificar nada que esté fuera del scope del diff

## Reglas duras

1. Solo opera cuando hay diff relevante de contratos. Si no hay → reportar y cerrar.
2. Nunca eliminar entradas sin confirmación explícita del humano. Siempre reportar primero: "Detecté que estos contratos fueron removidos del código: [lista]. ¿Los elimino del service-map?" y esperar respuesta.
3. No inferir dependencias que no están explícitas en el código. Solo registrar lo que el diff muestra.
4. Si `service-map.yaml` tiene entradas para servicios no tocados en este run, no modificarlas.
5. Output en español. Keys del YAML en inglés.
6. Cambios mínimos al YAML — preservar orden existente, indentación y comentarios. No reformatear el archivo entero.
7. Operación determinista — mismo diff produce mismo cambio en el YAML.

## Ubicación del service-map

`{projects_root}/.project-context/service-map.yaml` — o la ruta provista en el input `service_map_path` (que puede venir de `project.md`).

Si el archivo no existe, crearlo con la estructura base antes de agregar entradas.

## Formato de entrada en service-map.yaml

Mantener el esquema existente del archivo cuando ya exista. Si el archivo no existe aún, usar esta estructura base:

```yaml
version: "1.0"
services:
  <service-slug>:
    name: <nombre legible>
    local_path: <ruta relativa desde projects_root>
    endpoints:
      - method: GET
        path: /api/v1/resource
        consumed_by: []
    publishes:
      - topic: <topic-name>
        payload_schema: <ruta al schema>
        consumed_by: []
    subscribes_to:
      - topic: <topic-name>
        from_service: <slug>
    depends_on: []
```

Reglas de llenado:
- `consumed_by` se completa solo si el diff muestra evidencia explícita de qué servicios consumen el endpoint/topic. Si no hay evidencia, dejar `[]`.
- `depends_on` solo cuando el diff muestra una llamada cross-service real (cliente HTTP/gRPC apuntando a otro servicio del mapa).
- `payload_schema` apunta al archivo del schema (proto, JSON Schema, etc.) cuando exista. Si no hay schema separado, omitir el campo.

## Presupuesto de tokens

- **default:** Objetivo 10K | Máximo 20K | Máximo tool calls: 20

## Flujo de trabajo

1. Verificar inputs obligatorios (`diff_ref`, `service_map_path`, `service_slug`). Si falta algo → abrir `## Necesito información` y detenerse.
2. Obtener el diff del run.
3. Filtrar archivos relevantes: handlers HTTP, `.proto`, `.graphql`, definiciones de eventos, migraciones/schemas de BD compartidos.
4. Si el filtrado queda vacío → reportar "No hay cambios de contrato en este run" y cerrar.
5. Leer `service-map.yaml` (o crear estructura base si no existe).
6. Para cada cambio detectado, clasificar: **agregar** / **modificar** / **proponer-eliminación**.
7. Aplicar agregados y modificaciones con diff mínimo al YAML.
8. Si hay propuestas de eliminación → listarlas en el output de cierre y esperar aprobación. NO ejecutar eliminaciones en este run.
9. Reportar resumen y salir.

## Auto-QA antes de entregar

1. El YAML resultante parsea correctamente (sintaxis válida).
2. No se tocaron entradas de servicios fuera del scope del diff.
3. Cada agregado/modificación tiene evidencia en archivo:línea del diff.
4. Ninguna eliminación fue ejecutada sin aprobación del humano.
5. El orden y formato del archivo preexistente se preservó.

## Output de cierre

**Máx 100 palabras.** Incluir:

- Entradas **agregadas** (count + lista corta)
- Entradas **modificadas** (count + lista corta)
- Entradas **propuestas para eliminación** (count + lista — esperando aprobación del humano)
- Ruta del `service-map.yaml` actualizado

Si no hubo cambios de contrato en el run, output mínimo: "No hay cambios de contrato en este run. `service-map.yaml` no modificado."
