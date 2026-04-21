# Template: architecture-infra.md

Inspired by: AWS Well-Architected + bflorat Infrastructure View.

**Generate when:** infrastructure changes are involved.

## Template

```markdown
# Arquitectura de Infraestructura — <TASK-ID>

## Topología de despliegue

```mermaid
graph LR
  ...
```

## Variables de entorno y secretos

| Variable | Tipo | Descripción | Requerida |
|---|---|---|---|

## Escalabilidad

- **Triggers de escalado:** ...
- **Límites de recursos:** ...
- **Bottlenecks conocidos:** ...

## Impacto CI/CD

<!-- Pipeline changes needed -->
- ...

## Seguridad de infraestructura

- **Red:** ...
- **Secretos:** ...
- **Acceso:** ...
```

## Rules

- Deployment diagram shows services and their connections — not internal code structure
- Every env var must specify type (string, int, bool, secret) and whether it's required
- Scaling section addresses both horizontal and vertical — or explicitly says "N/A"
- CI/CD impact lists specific pipeline files that need changes
- Security section covers network boundaries, secret management, and access control
- If backend view exists, env vars here must match backend config references
