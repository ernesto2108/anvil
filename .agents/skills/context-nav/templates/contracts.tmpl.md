# Contratos — <ProjectName>

last_updated: <YYYY-MM-DD>

<!-- Bordes del sistema: qué expone, qué consume, qué emite -->

## REST API

### Base URL
`<http://host:port/api/vN>`

### Autenticación
`<Bearer token / API Key / ninguna — header o mecanismo>`

### Endpoints

#### GET <path>
- **Handler:** `<path>:<func>`
- **Auth:** <requerida / ninguna>
- **Query params:** `<param>` — <descripción>
- **Response:** `<Tipo>` — `{ campo: tipo, ... }`

#### POST <path>
- **Handler:** `<path>:<func>`
- **Auth:** <requerida / ninguna>
- **Body:** `<Tipo>` — `{ campo: tipo, ... }`
- **Response:** `<Tipo>`

<!-- Agregar endpoints según se detecten o implementen -->

## Message Queues / Event Streams

<!-- NATS, RabbitMQ, Redis Streams, Kafka, etc. -->

### <topic o queue name>
- **Broker:** <NATS / Redis Streams / RabbitMQ>
- **Producers:** `<path>:<func>`
- **Consumers:** `<path>:<func>`
- **Schema:** `<EventType>` — ver `<path al tipo>`
- **Garantía:** at-most-once / at-least-once / exactly-once

## Servicios externos

<!-- APIs de terceros, SDKs, servicios internos de otros repos -->

### <Nombre del servicio>
- **Cliente:** `<path>:<struct>`
- **Base URL / endpoint:** `<config key o URL>`
- **Auth:** `<mecanismo>`
- **Operaciones usadas:** `<método1>`, `<método2>`
- **Timeout configurado:** `<valor o "no configurado">`

## WebSockets

<!-- Si aplica -->

### <endpoint WS>
- **Handler:** `<path>:<func>`
- **Eventos enviados:** `<evento>` — `{ ... }`
- **Eventos recibidos:** `<evento>` — `{ ... }`

## gRPC

<!-- Si aplica -->

### <ServiceName>
- **Proto:** `<path al .proto>`
- **Server:** `<path>:<struct>`
- **Métodos:** `<método1>`, `<método2>`

## Contratos internos entre dominios

<!-- Interfaces que cruzan boundaries dentro del mismo repo -->

### <InterfaceName>
- **Definida en:** `<path>`
- **Implementada por:** `<package1>`, `<package2>`
- **Consumida por:** `<package>`
