# Guía de Servicios de AWS

## ECS / Fargate
- Usa Fargate para contenedores serverless; EC2 solo para GPU o AMIs personalizadas
- Establece `enable_ecs_managed_tags = true` y `propagate_tags = "SERVICE"` para seguimiento de costos
- CPU/memoria de la tarea a nivel de task definition; los límites del contenedor deben sumar los límites de la tarea
- Usa Secrets Manager o SSM Parameter Store para variables de entorno, nunca texto plano
- Habilita el circuit breaker de deployment con rollback
- ECS Exec solo para debugging, desactívalo en producción

## ECR
- Habilita escaneo de imágenes en el push
- Aplica tags inmutables para imágenes de producción
- Políticas de ciclo de vida para expirar imágenes sin tag/antiguas
- Pull-through cache para imágenes públicas (evita los rate limits de Docker Hub)

## RDS
- Multi-AZ para producción; read replicas para cargas de trabajo con muchas lecturas
- Backups automatizados con retención; prueba los restores periódicamente
- Autenticación IAM de base de datos donde sea posible; rota via Secrets Manager
- Solo subnets privadas; acceso mediante security groups, nunca público

## S3
- Bloquea el acceso público a nivel de cuenta; políticas de bucket para excepciones
- Habilita versionado + reglas de ciclo de vida para gestión de costos
- Cifrado del lado del servidor (SSE-S3 o SSE-KMS); aplica `ssl-only` en la política
- Notificaciones de eventos o EventBridge para patrones event-driven

## CloudFront
- Origin Access Control (OAC) para orígenes S3, no el OAI legacy
- Integración WAF; políticas de caché por patrón de ruta
- Políticas de caché gestionadas donde sea posible

## Lambda
- Memoria/timeout conservadores; usa AWS Power Tuning para dimensionar correctamente
- Layers para deps compartidas; imágenes de contenedor para deps grandes
- Concurrencia reservada para proteger el downstream; provisionada para latencia sensible
- Variables de entorno desde SSM/Secrets Manager, nunca hardcodeadas

## Patrones Comunes de Terraform

```hcl
# ECS Service with Fargate
resource "aws_ecs_service" "main" {
  name            = var.service_name
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.main.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

  network_configuration {
    subnets          = var.private_subnet_ids
    security_groups  = [aws_security_group.ecs.id]
    assign_public_ip = false
  }

  propagate_tags          = "SERVICE"
  enable_ecs_managed_tags = true
}
```
