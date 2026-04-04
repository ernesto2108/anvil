# AWS Services Guide

## ECS / Fargate
- Use Fargate for serverless containers; EC2 only for GPU or custom AMIs
- Set `enable_ecs_managed_tags = true` and `propagate_tags = "SERVICE"` for cost tracking
- Task CPU/memory at task definition level; container limits must sum to task limits
- Use Secrets Manager or SSM Parameter Store for env vars, never plain-text
- Enable deployment circuit breaker with rollback
- ECS Exec for debugging only, disable in production

## ECR
- Enable image scanning on push
- Enforce immutable tags for production images
- Lifecycle policies to expire untagged/old images
- Pull-through cache for public images (avoid Docker Hub rate limits)

## RDS
- Multi-AZ for production; read replicas for read-heavy workloads
- Automated backups with retention; test restores periodically
- IAM database authentication where possible; rotate via Secrets Manager
- Private subnets only; access through security groups, never public

## S3
- Block public access at account level; bucket policies for exceptions
- Enable versioning + lifecycle rules for cost management
- Server-side encryption (SSE-S3 or SSE-KMS); enforce `ssl-only` in policy
- Event notifications or EventBridge for event-driven patterns

## CloudFront
- Origin Access Control (OAC) for S3 origins, not legacy OAI
- WAF integration; cache policies per path pattern
- Managed cache policies where possible

## Lambda
- Memory/timeout conservative; use AWS Power Tuning to right-size
- Layers for shared deps; container images for large deps
- Reserved concurrency to protect downstream; provisioned for latency-sensitive
- Env vars from SSM/Secrets Manager, never hardcoded

## Common Terraform Patterns

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
