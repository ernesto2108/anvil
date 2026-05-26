---
name: service-map
description: Conciencia de dependencias entre servicios para microservicios. Úsalo al modificar endpoints, esquemas de BD, contratos compartidos, o cualquier código que otros servicios consumen. Se activa con "service map", "quién usa este endpoint", "análisis de impacto", "cross-service", "verificación de dependencias", "qué servicios dependen de", "antes de refactorizar endpoint", o cuando se trabaja en un proyecto con archivo service-map.yaml.
---

# Service Map — Conciencia de Dependencias entre Servicios

## Propósito

Prevenir cambios que rompan la compatibilidad entre microservicios verificando las dependencias **antes** de modificar endpoints, esquemas de BD, contratos compartidos o comunicación entre servicios.

## Ubicación del Service Map

```
.project-context/service-map.yaml
```

El service map vive siempre como archivo local en `.project-context/service-map.yaml` (o donde el repo lo tenga). Si no existe un service map, solicitar al usuario que cree uno usando la plantilla en `service-map-template.yaml`.

Si el proyecto tiene `task_tool` configurado (campo de `.project-context/project.md`) y hay que registrar el análisis de impacto fuera del repo, **describir al humano** qué crear en su herramienta — nunca ejecutar acciones en ella.

## Flujo Pre-Cambio

**Paso 1 — Identificar qué está cambiando**
- Endpoint HTTP (ruta, esquema de solicitud/respuesta, códigos de estado)
- Tabla/columna de BD (cambio de esquema, migración)
- Evento/mensaje (estructura del payload, topic)
- Librería/paquete compartido
- Variable de entorno o configuración

**Paso 2 — Consultar el service map**
- Buscar el recurso en service-map.yaml
- Identificar todos los consumidores y el propietario
- Resolver rutas locales: `projects_root` + `local_path` del servicio

**Paso 3 — Reportar el impacto**
```
## Análisis de Impacto
**Cambiando:** [qué]
**Propietario:** [servicio]
**Consumidores:** [lista]

### Cambios que rompen compatibilidad:
- [qué podría romperse y dónde]

### Cambios seguros:
- [aditivos/sin ruptura de compatibilidad]

### Enfoque recomendado:
- [pasos para cambiar de forma segura]
```

**Paso 4 — Inspeccionar servicios afectados**
- Usar rutas resueltas para localizar los repositorios consumidores en disco
- Leer el código del consumidor para encontrar DÓNDE existe la dependencia
- Reportar el archivo exacto y la línea
- Si el repositorio no se encuentra en disco, advertir al usuario

**Paso 5 — Si hay incertidumbre, PREGUNTAR**
- Nunca asumir que un cambio es seguro — verificar o preguntar

## Reglas de Seguridad de Cambios

**Siempre seguros:** nuevo campo de respuesta opcional, nuevo endpoint, nuevo topic de evento, nueva columna en BD con valor por defecto

**Potencialmente disruptivos (requieren verificación con el consumidor):** eliminar/renombrar campo de respuesta, cambiar tipo de campo, cambiar URL/método, validación más estricta, cambiar tipo de columna en BD, cambiar payload del evento

**Siempre disruptivos (deploy coordinado):** eliminar endpoint, cambiar autenticación, renombrar tabla compartida en BD, cambiar nombre del topic del evento

## Patrones Recomendados

**Versionado de endpoints:** mantener v1 sin cambios, agregar v2

**Expand-and-contract para BD:** agregar nueva columna → migrar consumidores → backfill → eliminar la anterior

**Flujo de deprecación:** marcar en service-map.yaml con `status: deprecated`, `deprecated_since`, `replacement`, `consumed_by`

## Conciencia Cross-Stack

| Stack | Qué verificar |
|-------|--------------|
| Backend (Go) | Endpoints HTTP, protos gRPC, esquemas de BD, productores de eventos |
| Frontend (React) | Llamadas a API, tipos compartidos, configuraciones de entorno |
| Mobile (Flutter) | Llamadas a API, contratos de notificaciones push, esquemas de deep link |
| Infraestructura | Variables de entorno, secretos, DNS, rutas del balanceador de carga |

## Cuándo Activar

- El usuario modifica endpoints, APIs, esquemas de BD, protos, tipos compartidos
- Siempre preguntar: "¿Algún otro servicio consume lo que estoy por cambiar?"
- Si sí → consultar el service map. Si no hay mapa → preguntar al usuario.

## Referencia del Esquema

Ver `service-map-template.yaml` para el esquema completo. Secciones clave: `services`, `shared_databases`, `events`, `shared_contracts`.
