# Template: architecture-backend.md

Inspired by: Stripe spec-driven design + bflorat Application View.

**Generate when:** backend work is involved.

## Template

```markdown
# Arquitectura Backend — <TASK-ID>

## Contratos API (OpenAPI)

<!-- Executable spec — YAML fragment. Agents and tools consume this directly. -->

```yaml
openapi: "3.1.0"
info:
  title: <feature name>
  version: "1.0.0"
paths:
  /api/v1/<resource>:
    post:
      summary: ...
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/<RequestDTO>"
      responses:
        "201":
          description: ...
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/<ResponseDTO>"
        "400":
          $ref: "#/components/responses/ValidationError"
        "500":
          $ref: "#/components/responses/InternalError"

components:
  schemas:
    <RequestDTO>:
      type: object
      required: [field1, field2]
      properties:
        field1:
          type: string
        field2:
          type: integer
    <ResponseDTO>:
      type: object
      properties:
        id:
          type: string
        ...
  responses:
    ValidationError:
      description: Validation failed
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/ErrorResponse"
    InternalError:
      description: Internal server error
```

## Taxonomía de errores

| Código | HTTP | Descripción | Cuándo ocurre |
|---|---|---|---|

## Casos de uso

<!-- Ports & adapters: what the system does, not how -->
- ...

## Comportamiento runtime

### <Flujo principal>

```mermaid
sequenceDiagram
  ...
```

## Estrategia de persistencia

- **Concurrencia:** ...
- **Caché:** ...
- **Reintentos / idempotencia:** ...
- **Manejo de fallos:** ...
```

## Rules

- OpenAPI spec is the source of truth for API contracts — not prose
- Use `$ref` for shared schemas — don't inline duplicate definitions
- Error taxonomy must map every error code to an HTTP status
- Sequence diagrams show the happy path + primary error path
- Persistence strategy describes behavior, not implementation (no SQL, no driver details)
- If the frontend also exists, the OpenAPI schemas here are the canonical definition — frontend derives from these
