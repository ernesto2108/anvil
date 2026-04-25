# Guía de Servicios de Google Cloud

## Cloud Run
- Por defecto para cargas de trabajo HTTP sin estado — scale-to-zero ahorra costos
- VPC egress directo (no conectores VPC) para redes privadas
- Instancias multi-contenedor (sidecars) soportadas: hasta 10 contenedores
- `min-instances > 0` para latencia sensible; `max-instances` para control de costos
- `--cpu-boost` para arranques en frío más rápidos
- Cuentas de servicio dedicadas con IAM mínimo, nunca el SA de compute por defecto

## GKE
- Modo Autopilot para la mayoría de cargas de trabajo (Google gestiona los nodos)
- Workload Identity para IAM a nivel de pod, nunca claves de SA a nivel de nodo
- Modo Standard solo cuando necesites GPU, node pools personalizados, o DaemonSets

## Cloud SQL
- IP privada con VPC peering; nunca exponer públicamente
- Backups automatizados + point-in-time recovery
- Cloud SQL Auth Proxy para conexiones seguras desde GKE/Cloud Run
- Autenticación IAM de base de datos donde sea soportado

## Artifact Registry
- Reemplaza Container Registry (deprecated)
- Habilita escaneo de vulnerabilidades
- Repositorios remotos como pull-through caches
- Pin de digest de imagen en producción

## Cloud Build
- Permisos explícitos de cuenta de servicio (por defecto cambió a mediados de 2024)
- `cloudbuild.yaml` en la raíz del repo; mantén los steps al mínimo
- Kaniko para builds de Docker (no necesita daemon)
- Cachea artefactos en Cloud Storage o Artifact Registry

## Patrones Comunes de Terraform

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
