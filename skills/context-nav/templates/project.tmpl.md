# Proyecto — <ProjectName>

last_updated: <YYYY-MM-DD>
task_tool: ""  # Herramienta de gestión de tareas del proyecto (valor libre, ej: Linear, Jira, Notion, ninguna)

## Objetivo

<Qué hace el sistema y qué problema resuelve — 2-4 líneas>

## Restricciones no negociables

- <restricción operativa o de diseño>

## Stack

| Componente | Tecnología | Versión |
|-----------|-----------|---------|
| Backend | Go / Python / Node | x.x |
| Frontend | React / Astro / Flutter | x.x |
| Base de datos | PostgreSQL / SQLite / Redis | x.x |
| Mensajería | NATS / RabbitMQ / Redis Streams | — |
| Infra | Docker / Kubernetes / bare | — |

## Estilo arquitectónico

<!-- Detectado desde estructura de directorios y patrones de código -->
<!-- Opciones: Hexagonal, Layered, Monolítico, Microservicios, Event-driven, CQRS, MVC -->

- **Estilo principal:** <detectado>
- **Capas:** <ej. handler → service → repository → db>
- **Módulo raíz (Go):** `<module path>`
- **Convención de paths:** `internal/<domain>/<layer>.go`

## SOLID detectado

| Principio | Estado | Observación |
|-----------|--------|-------------|
| SRP | OK / En riesgo | <top archivos > 300 líneas si aplica> |
| OCP | OK / En riesgo | <switch sobre tipos si aplica> |
| LSP | OK / No evaluado | — |
| ISP | OK / En riesgo | <interfaces grandes si aplica> |
| DIP | OK / En riesgo | <dependencias a concretos si aplica> |

## Convenciones establecidas

- <convención de nombres, estructura, patterns>
- <regla de imports o dependencias>
- <regla de tests>

## Qué NO introducir

- <patrón o librería que se descartó>
- <abstracción que no existe y no debe existir>
