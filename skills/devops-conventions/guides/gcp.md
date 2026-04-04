# Google Cloud Services Guide

## Cloud Run
- Default for stateless HTTP workloads — scale-to-zero saves cost
- Direct VPC egress (not VPC connectors) for private networking
- Multi-container instances (sidecars) supported: up to 10 containers
- `min-instances > 0` for latency-sensitive; `max-instances` for cost control
- `--cpu-boost` for faster cold starts
- Dedicated service accounts with minimal IAM, never default compute SA

## GKE
- Autopilot mode for most workloads (Google manages nodes)
- Workload Identity for pod-level IAM, never node-level SA keys
- Standard mode only when you need GPU, custom node pools, or DaemonSets

## Cloud SQL
- Private IP with VPC peering; never expose publicly
- Automated backups + point-in-time recovery
- Cloud SQL Auth Proxy for secure connections from GKE/Cloud Run
- IAM database authentication where supported

## Artifact Registry
- Replaces Container Registry (deprecated)
- Enable vulnerability scanning
- Remote repositories as pull-through caches
- Image digest pinning in production

## Cloud Build
- Explicit service account permissions (default changed mid-2024)
- `cloudbuild.yaml` in repo root; keep steps minimal
- Kaniko for Docker builds (no daemon needed)
- Cache artifacts in Cloud Storage or Artifact Registry

## Common Terraform Patterns

```hcl
# Cloud Run service
resource "google_cloud_run_v2_service" "main" {
  name     = var.service_name
  location = var.region

  template {
    service_account = google_service_account.run.email

    scaling {
      min_instance_count = var.min_instances
      max_instance_count = var.max_instances
    }

    containers {
      image = "${var.region}-docker.pkg.dev/${var.project}/${var.repo}/${var.image}:${var.tag}"

      ports {
        container_port = 8080
      }

      resources {
        limits = {
          cpu    = var.cpu
          memory = var.memory
        }
        cpu_idle = true  # scale-to-zero
      }

      startup_probe {
        http_get {
          path = "/health"
        }
        initial_delay_seconds = 5
      }
    }

    vpc_access {
      egress = "PRIVATE_RANGES_ONLY"
      network_interfaces {
        network    = var.vpc_id
        subnetwork = var.subnet_id
      }
    }
  }
}
```
