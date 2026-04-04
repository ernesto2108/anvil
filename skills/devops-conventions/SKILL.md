---
name: devops-conventions
description: DevOps/Infrastructure conventions and best practices. Use when writing Dockerfiles, GitHub Actions, Terraform, K8s manifests, or cloud infra configs.
---

# DevOps Conventions

## When to Load

Load this skill when the user or orchestrator asks to:
- Write or modify Dockerfiles, docker-compose files
- Create or update GitHub Actions workflows
- Write Terraform/OpenTofu configurations
- Create Kubernetes manifests
- Configure AWS or Google Cloud resources
- Set up CI/CD pipelines
- Review infrastructure code

## Routing Table

| Task | Load |
|------|------|
| Dockerfile, container images | `rules/docker.md` |
| GitHub Actions, CI/CD | `rules/github-actions.md` |
| Terraform, IaC | `rules/terraform.md` |
| Kubernetes manifests | `rules/kubernetes.md` |
| AWS services (ECS, RDS, S3, Lambda) | `guides/aws.md` |
| Google Cloud (Cloud Run, GKE, Cloud SQL) | `guides/gcp.md` |
| Argo CD, Rollouts, Workflows, Events, GitOps | `guides/argo.md` |
| Security (scanning, secrets, IAM) | `rules/security.md` |

Load only what you need for the task. Multiple files can be loaded if the task spans concerns (e.g., Dockerfile + GitHub Actions for a CI pipeline).

## Universal Rules (always apply)

1. **Pin everything** — base images to digest, Actions to SHA, Terraform providers to version constraints
2. **No secrets in code** — use secret managers, OIDC, or env vars from CI; never commit credentials
3. **Least privilege** — minimal IAM roles, non-root containers, explicit workflow permissions
4. **Idempotent** — all scripts and configs safe to run multiple times
5. **Immutable infrastructure** — replace, don't patch; rebuild images, don't SSH and fix
