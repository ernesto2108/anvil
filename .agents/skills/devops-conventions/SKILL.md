---
name: devops-conventions
description: Convenciones y mejores prácticas de DevOps/Infraestructura. Usar cuando se escriban Dockerfiles, GitHub Actions, Terraform, manifests de K8s o configuraciones de infra en la nube.
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


# DevOps Conventions

## Cuándo Cargar

Carga esta skill cuando el usuario u orquestador solicite:
- Escribir o modificar Dockerfiles, archivos docker-compose
- Crear o actualizar workflows de GitHub Actions
- Escribir configuraciones de Terraform/OpenTofu
- Crear manifests de Kubernetes
- Configurar recursos de AWS o Google Cloud
- Configurar pipelines de CI/CD
- Revisar código de infraestructura

## Tabla de Enrutamiento

| Tarea | Cargar |
|------|------|
| Dockerfile, imágenes de contenedor | `rules/docker.md` |
| GitHub Actions, CI/CD | `rules/github-actions.md` |
| Terraform, IaC | `rules/terraform.md` |
| Manifests de Kubernetes | `rules/kubernetes.md` |
| Servicios AWS (ECS, RDS, S3, Lambda) | `guides/aws.md` |
| Google Cloud (Cloud Run, GKE, Cloud SQL) | `guides/gcp.md` |
| Argo CD, Rollouts, Workflows, Events, GitOps | `guides/argo.md` |
| Seguridad (scanning, secrets, IAM) | `rules/security.md` |

Carga solo lo que necesitas para la tarea. Se pueden cargar múltiples archivos si la tarea abarca varios aspectos (ej., Dockerfile + GitHub Actions para un pipeline de CI).

## Reglas Universales (siempre aplican)

1. **Pinea todo** — imágenes base a digest, Actions a SHA, proveedores de Terraform a restricciones de versión
2. **Sin secrets en el código** — usa secret managers, OIDC o variables de entorno desde CI; nunca hagas commit de credenciales
3. **Mínimo privilegio** — roles IAM mínimos, contenedores sin root, permisos explícitos de workflow
4. **Idempotente** — todos los scripts y configuraciones seguros de ejecutar múltiples veces
5. **Infraestructura inmutable** — reemplaza, no parches; reconstruye imágenes, no hagas SSH y corrijas
