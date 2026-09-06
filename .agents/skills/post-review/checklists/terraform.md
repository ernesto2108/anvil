# Checklist de Revisión Terraform

## Correctness

- [ ] Recursos con `lifecycle { prevent_destroy = true }` en datos criticos (DB, storage)
- [ ] Data sources validan que el recurso existe antes de referenciarlo
- [ ] Count/for_each no produce nombres de recursos con caracteres invalidos
- [ ] Depends_on explicito solo cuando Terraform no puede inferir la dependencia
- [ ] No hay recursos huerfanos (creados pero no referenciados)
- [ ] Modulos versionados con tag especifico, no `main` o `latest`
- [ ] Outputs expuestos para valores que otros modulos necesitan

## Security

- [ ] No hay secrets en `.tf` files (usar variables con `sensitive = true`)
- [ ] No hay secrets en `terraform.tfvars` commiteados al repo
- [ ] Security groups no abren `0.0.0.0/0` en puertos sensibles (SSH, DB)
- [ ] IAM policies siguen least privilege (no `Action: "*"` ni `Resource: "*"`)
- [ ] Encryption at rest habilitado en S3, RDS, EBS
- [ ] Encryption in transit habilitado (TLS/SSL)
- [ ] State file en backend remoto con encryption (S3 + DynamoDB lock)
- [ ] No hay access keys hardcodeados — usar IAM roles

## Conventions

- [ ] Variables tienen `description` y `type` definidos
- [ ] Variables sensibles marcadas con `sensitive = true`
- [ ] Naming consistente: `{project}-{env}-{resource}` o el patron del equipo
- [ ] Archivos organizados: `main.tf`, `variables.tf`, `outputs.tf`, `providers.tf`
- [ ] Tags en todos los recursos (al menos: Name, Environment, Team, ManagedBy)
- [ ] Providers con version constraint (`~>`, `>=`)
- [ ] No hay `terraform { required_version }` sin constraint

## State & Planning

- [ ] `terraform plan` revisado antes de apply (no apply ciego)
- [ ] No hay recursos que se destruyen y recrean sin intencion (check plan output)
- [ ] State locking configurado para trabajo en equipo
- [ ] No hay `terraform import` pendiente documentado
- [ ] Workspaces o directorios separados por environment (dev, staging, prod)

## Infrastructure

- [ ] Auto-scaling configurado donde aplique
- [ ] Health checks en load balancers y target groups
- [ ] Backup y retention policies en bases de datos
- [ ] Logging habilitado (CloudWatch, CloudTrail, flow logs)
- [ ] DNS y certificados SSL gestionados via Terraform (no manual)
