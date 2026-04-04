# Terraform Best Practices

## Module Structure
- Standard files: `main.tf`, `variables.tf`, `outputs.tf`, `providers.tf`, `versions.tf`
- Group resources logically: `network.tf`, `compute.tf`, `database.tf`
- Root modules: direct `terraform apply` targets, contain provider config
- Child modules: reusable blocks, never contain provider config
- Templates in `templates/*.tftpl`; scripts in `scripts/`

## Naming
- Underscores for all names (matches HCL convention)
- Single-type resources named `main`; use `primary`/`secondary` to differentiate
- Don't repeat type in name: `aws_instance.web` not `aws_instance.web_instance`
- Booleans named positively: `enable_external_access`
- Numerics suffixed with units: `ram_size_gb`

## Variables & Outputs
- All variables in `variables.tf`; all outputs in `outputs.tf`
- Always include `description` and explicit `type`
- Defaults only for environment-independent values
- Outputs reference resource attributes, not input variables

## State Management
- Always remote state with locking (S3+DynamoDB, GCS, Terraform Cloud)
- Never commit `.tfstate` to version control
- One state per environment; shared infra in own state
- Use `terraform_remote_state` sparingly; prefer explicit outputs

## Plan/Apply Workflow
- `terraform fmt` + `terraform validate` in pre-commit hooks
- `terraform plan` on every PR; post output as PR comment
- `terraform apply` only from CI/CD after approval, never locally for production
- Avoid `-target` in regular workflow

## Versioning
- Pin providers: `~> 5.0` (pessimistic constraint)
- Pin modules to exact or narrow range
- Set `required_version` for CLI version
- Repo naming: `terraform-<provider>-<purpose>`

## Drift Detection
- Schedule periodic `terraform plan` runs
- Alert on any diff between state and actual
- Never manually modify Terraform-managed resources

## Project Template

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
