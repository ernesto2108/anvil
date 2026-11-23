# Guía del Ecosistema Argo

## Argo CD — Entrega Continua GitOps

### Estructura del Repositorio
- **Separa la config del código fuente** — manifiestos en repositorio dedicado, no junto al código fuente de la app
- **Entornos como carpetas, no como ramas** — `environments/dev/`, `environments/staging/`, `environments/prod/`
- Fija todas las referencias remotas (versiones de Helm chart, bases de Kustomize) a tag o SHA

### Políticas de Sync
- **Non-prod:** `automated.selfHeal: true` + `automated.prune: true`
- **Producción:** `automated: false` — requiere gate de aprobación humana
- Excluye los campos manejados por HPA con `ignoreDifferences`:
  ```yaml
  spec:
    ignoreDifferences:
      - group: apps
        kind: Deployment
        jsonPointers:
          - /spec/replicas
  ```

### Sync Waves y Hooks
- Las waves controlan el orden: anota `argocd.argoproj.io/sync-wave: "-1"` (negativo = antes)
- Orden de ejecución: Phase > Wave > Kind > Name
- Fases de hook: PreSync (migraciones DB), Sync (deploy), PostSync (smoke tests), SyncFail (limpieza)
- Eliminación de hooks: `BeforeHookCreation` para Jobs idempotentes, `HookSucceeded` para auto-limpieza

### Multi-Cluster
- Menos de 20 clusters: hub-and-spoke (un solo Argo CD maneja todos)
- Más de 20 clusters: federado (cada cluster tiene su propio Argo CD, un meta-ArgoCD los gestiona)
- Etiqueta los clusters al registrarlos: `environment`, `region`, `tier`
- Nunca expongas la API de K8s públicamente; usa VPN o PrivateLink

### ApplicationSets
- **Cluster Generator:** La misma app a todos los clusters que coincidan con las etiquetas
- **Git Directory Generator:** Una app por directorio en monorepo (microservicios)
- **Matrix Generator:** Combina cluster + git para N-apps x M-clusters
- **Merge Generator:** Config base + overrides por cluster
- Habilita `goTemplate: true` para lógica condicional
- Agregar un nuevo cluster no debe requerir cambios en ApplicationSet

### Proyectos y RBAC
- Un AppProject por equipo/dominio — restringe repos de origen, destinos, kinds permitidos
- Los desarrolladores obtienen `sync`; el equipo de plataforma obtiene `override`, `delete`, acceso a prod
- Cuentas de servicio dedicadas para CI — nunca compartas credenciales humanas

### Gestión de Secretos
| Enfoque | Complejidad | Auto-Rotación | Mejor para |
|---|---|---|---|
| Sealed Secrets | Baja | Manual | Equipos pequeños |
| External Secrets Operator | Media | Automática | Orgs multi-cloud |
| SOPS + KSOPS | Media | Manual | Equipos usando KMS |
| Vault Plugin | Alta | Automática | Orgs con Vault |

Reglas:
- Nunca secretos en texto plano en Git
- Sealed Secrets: `scope: strict`, haz backup del key pair del controlador
- External Secrets: `ClusterSecretStore`, autenticar via IRSA/Workload Identity
- Agrega `ignoreDifferences` en campos `/data` de Secret auto-rotados

---

## Argo Workflows — Motor de Workflows Nativo de Contenedores

### DAG vs Steps
- **DAG:** Grafos de dependencias complejos con paralelismo — las tareas declaran dependencias explícitas
- **Steps:** Pipelines secuenciales/paralelos simples
- Compone workflows grandes desde DAGs reutilizables más pequeños

### Políticas de Retry
```yaml
retryStrategy:
  limit: 3
  retryPolicy: "OnFailure"
  backoff:
    duration: "10s"
    factor: 2
    maxDuration: "5m"
```
- `OnFailure`: reintentos solo en código de salida no-cero
- `Always`: reintentos en cualquier error incluyendo fallos de infra
- Siempre establece `maxDuration`

### Gestión de Recursos
- Establece `requests` + `limits` en cada contenedor de tarea
- `activeDeadlineSeconds` en workflows para prevenir ejecución indefinida
- `podPriorityClassName` para workflows críticos

### Templates y Reutilización
- `WorkflowTemplate` para patrones comunes, referencia con `templateRef`
- `ClusterWorkflowTemplate` para patrones de toda la org (notificaciones, limpieza)
- Parametriza todo: tags de imagen, límites, rutas de artefactos

### Cron Workflows
- `concurrencyPolicy`: Allow, Replace, Forbid
- Establece `startingDeadlineSeconds` para schedules perdidos
- Siempre establece `successfulJobsHistoryLimit` y `failedJobsHistoryLimit`

---

## Argo Rollouts — Progressive Delivery

### Selección de Estrategia
- **Empieza con Blue-Green** — más simple, preview completo, un solo corte
- **Pasa a Canary** cuando tengas métricas confiables (ventanas de análisis de 5-15 min)
- No usar para: apps de recursos compartidos, queue workers, controladores de infra

### Patrón Canary
```yaml
strategy:
  canary:
    steps:
      - setWeight: 5
      - pause: { duration: 5m }
      - analysis:
          templates:
            - templateName: success-rate
      - setWeight: 25
      - pause: { duration: 5m }
      - setWeight: 50
      - pause: { duration: 5m }
```

### Patrón Blue-Green
- Define `activeService` + `previewService`
- `autoPromotionEnabled: false` inicialmente
- `scaleDownDelaySeconds` para mantener el RS antiguo para rollback rápido

### Analysis Templates
- Consulta al proveedor de métricas (Prometheus, Datadog, CloudWatch)
- Objetivo: tasa de éxito > 99%, latencia p99 < 500ms, tasa de error < 0.1%
- Prueba las consultas via dry-runs antes de producción

### Operaciones
- Reduce `RevisionHistoryLimit` a 2-3 en clusters de alto volumen
- Hashea el contenido de ConfigMap en el nombre para rollouts disparados por config
- Integra notificaciones (Slack, PagerDuty)

---

## Argo Events — Automatización Dirigida por Eventos

### Arquitectura
EventSource (produce) → EventBus (transporte) → Sensor (consume + dispara)

### Event Sources
- Webhooks, S3, Cron, Kafka, SQS/SNS, Pub/Sub, GitHub, GitLab, NATS, AMQP
- Aplica filtros a nivel de EventSource para descartar eventos irrelevantes temprano
- Cuentas de servicio dedicadas por EventSource

### EventBus
- Por defecto: NATS JetStream — ejecuta cluster de 3 nodos para producción
- Alternativa: Kafka para orgs que ya lo ejecutan

### Sensors y Triggers
- Combina dependencias: `A && B` (ambos requeridos), `A || B` (cualquiera)
- Filtros para coincidir con payloads específicos (ej., push a `main` solo)
- Tipos de trigger: Argo Workflow, HTTP, K8s object, Log
- `retryStrategy` con backoff + `dlqTrigger` para dead letter queue

### Integración con Workflows
```yaml
triggers:
  - template:
      argoWorkflow:
        operation: submit
        source:
          resource:
            spec:
              workflowTemplateRef:
                name: ci-pipeline-template
        parameters:
          - src:
              dependencyName: github-push
              dataKey: body.ref
            dest: spec.arguments.parameters.0.value
```

---

## Patrón de Bootstrap con Terraform

Terraform gestiona: cluster, VPC, IAM, instalación de Argo CD, app-of-apps de bootstrap.
Argo CD gestiona: todas las cargas de trabajo, namespaces, RBAC, add-ons después del bootstrap.

```hcl
resource "helm_release" "argocd" {
  name             = "argocd"
  repository       = "https://argoproj.github.io/argo-helm"
  chart            = "argo-cd"
  version          = "7.7.x"
  namespace        = "argocd"
  create_namespace = true
  values           = [file("values/argocd.yaml")]
}
```

Reglas:
- Nunca uses `kubernetes_manifest` para recursos gestionados por Argo
- Fija versiones de Helm chart, nunca `latest`
- Estado de Terraform separado por cluster
- Autenticación basada en exec (tokens de corta duración), no kubeconfig estático
