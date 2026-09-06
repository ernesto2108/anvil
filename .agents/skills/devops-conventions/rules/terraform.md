# Mejores Prácticas de Terraform

## Estructura de Módulos
- Archivos estándar: `main.tf`, `variables.tf`, `outputs.tf`, `providers.tf`, `versions.tf`
- Agrupa recursos lógicamente: `network.tf`, `compute.tf`, `database.tf`
- Módulos raíz: objetivos directos de `terraform apply`, contienen config de provider
- Módulos hijo: bloques reutilizables, nunca contienen config de provider
- Templates en `templates/*.tftpl`; scripts en `scripts/`

## Nomenclatura
- Guiones bajos para todos los nombres (coincide con la convención HCL)
- Recursos de tipo único llamados `main`; usa `primary`/`secondary` para diferenciar
- No repitas el tipo en el nombre: `aws_instance.web` no `aws_instance.web_instance`
- Booleanos con nombre positivo: `enable_external_access`
- Numéricos con sufijo de unidades: `ram_size_gb`

## Variables y Outputs
- Todas las variables en `variables.tf`; todos los outputs en `outputs.tf`
- Siempre incluye `description` y `type` explícito
- Defaults solo para valores independientes del entorno
- Los outputs referencian atributos de recursos, no variables de input

## Gestión de Estado
- Siempre estado remoto con locking (S3+DynamoDB, GCS, Terraform Cloud)
- Nunca commitees `.tfstate` al control de versiones
- Un estado por entorno; infra compartida en su propio estado
- Usa `terraform_remote_state` con moderación; prefiere outputs explícitos

## Flujo de Plan/Apply
- `terraform fmt` + `terraform validate` en hooks de pre-commit
- `terraform plan` en cada PR; publica el output como comentario del PR
- `terraform apply` solo desde CI/CD después de aprobación, nunca localmente para producción
- Evita `-target` en el flujo de trabajo regular

## Versionado
- Fija providers: `~> 5.0` (restricción pesimista)
- Fija módulos a rango exacto o estrecho
- Establece `required_version` para la versión del CLI
- Nombres de repos: `terraform-<provider>-<purpose>`

## Detección de Drift
- Programa ejecuciones periódicas de `terraform plan`
- Alerta ante cualquier diferencia entre el estado y la realidad
- Nunca modifiques manualmente recursos gestionados por Terraform

## Plantilla de Proyecto

```
infrastructure/
├── environments/
│   ├── dev/
│   │   ├── main.tf
│   │   ├── variables.tf
│   │   ├── outputs.tf
│   │   ├── terraform.tfvars
│   │   └── backend.tf
│   ├── staging/
│   └── production/
├── modules/
│   ├── networking/
│   ├── compute/
│   └── database/
└── versions.tf
```
