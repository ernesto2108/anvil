# Argo Ecosystem Guide

## Argo CD — GitOps Continuous Delivery

### Repository Structure
- **Separate config from source code** — manifests in dedicated repo, not alongside app source
- **Environments as folders, not branches** — `environments/dev/`, `environments/staging/`, `environments/prod/`
- Pin all remote references (Helm chart versions, Kustomize bases) to tag or SHA

### Sync Policies
- **Non-prod:** `automated.selfHeal: true` + `automated.prune: true`
- **Production:** `automated: false` — require human approval gate
- Exclude HPA-managed fields with `ignoreDifferences`:
  ```yaml
  spec:
    ignoreDifferences:
      - group: apps
        kind: Deployment
        jsonPointers:
          - /spec/replicas
  ```

### Sync Waves & Hooks
- Waves control ordering: annotate `argocd.argoproj.io/sync-wave: "-1"` (negative = earlier)
- Execution order: Phase > Wave > Kind > Name
- Hook phases: PreSync (DB migrations), Sync (deploy), PostSync (smoke tests), SyncFail (cleanup)
- Hook deletion: `BeforeHookCreation` for idempotent Jobs, `HookSucceeded` for auto-cleanup

### Multi-Cluster
- Under 20 clusters: hub-and-spoke (single Argo CD manages all)
- Over 20 clusters: federated (each cluster has own Argo CD, meta-ArgoCD manages them)
- Label clusters at registration: `environment`, `region`, `tier`
- Never expose K8s API publicly; use VPN or PrivateLink

### ApplicationSets
- **Cluster Generator:** Same app to all clusters matching labels
- **Git Directory Generator:** One app per directory in monorepo (microservices)
- **Matrix Generator:** Combine cluster + git for N-apps x M-clusters
- **Merge Generator:** Base config + per-cluster overrides
- Enable `goTemplate: true` for conditional logic
- Adding a new cluster should require zero ApplicationSet changes

### Project & RBAC
- One AppProject per team/domain — restrict source repos, destinations, allowed kinds
- Developers get `sync`; platform team gets `override`, `delete`, prod access
- Dedicated service accounts for CI — never share human credentials

### Secrets Management
| Approach | Complexity | Auto-Rotation | Best For |
|---|---|---|---|
| Sealed Secrets | Low | Manual | Small teams |
| External Secrets Operator | Medium | Automatic | Multi-cloud orgs |
| SOPS + KSOPS | Medium | Manual | Teams using KMS |
| Vault Plugin | High | Automatic | Orgs with Vault |

Rules:
- Never plaintext secrets in Git
- Sealed Secrets: `scope: strict`, back up controller key pair
- External Secrets: `ClusterSecretStore`, authenticate via IRSA/Workload Identity
- Add `ignoreDifferences` on auto-rotated Secret `/data` fields

---

## Argo Workflows — Container-Native Workflow Engine

### DAG vs Steps
- **DAG:** Complex dependency graphs with parallelism — tasks declare explicit dependencies
- **Steps:** Simple sequential/parallel pipelines
- Compose large workflows from smaller reusable DAGs

### Retry Policies
```yaml
retryStrategy:
  limit: 3
  retryPolicy: "OnFailure"
  backoff:
    duration: "10s"
    factor: 2
    maxDuration: "5m"
```
- `OnFailure`: retries on non-zero exit code only
- `Always`: retries on any error including infra failures
- Always set `maxDuration`

### Resource Management
- Set `requests` + `limits` on every task container
- `activeDeadlineSeconds` on workflows to prevent indefinite execution
- `podPriorityClassName` for critical workflows

### Templates & Reusability
- `WorkflowTemplate` for common patterns, reference with `templateRef`
- `ClusterWorkflowTemplate` for org-wide patterns (notifications, cleanup)
- Parameterize everything: image tags, limits, artifact paths

### Cron Workflows
- `concurrencyPolicy`: Allow, Replace, Forbid
- Set `startingDeadlineSeconds` for missed schedules
- Always set `successfulJobsHistoryLimit` and `failedJobsHistoryLimit`

---

## Argo Rollouts — Progressive Delivery

### Strategy Selection
- **Start with Blue-Green** — simpler, full preview, single cutover
- **Graduate to Canary** when you have reliable metrics (5-15 min analysis windows)
- Not for: shared-resource apps, queue workers, infra controllers

### Canary Pattern
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

### Blue-Green Pattern
- Define `activeService` + `previewService`
- `autoPromotionEnabled: false` initially
- `scaleDownDelaySeconds` to keep old RS for quick rollback

### Analysis Templates
- Query metrics provider (Prometheus, Datadog, CloudWatch)
- Target success rate > 99%, p99 latency < 500ms, error rate < 0.1%
- Test queries via dry-runs before production

### Operations
- Lower `RevisionHistoryLimit` to 2-3 on high-volume clusters
- Hash ConfigMap content into name for config-triggered rollouts
- Integrate notifications (Slack, PagerDuty)

---

## Argo Events — Event-Driven Automation

### Architecture
EventSource (produces) → EventBus (transport) → Sensor (consumes + triggers)

### Event Sources
- Webhooks, S3, Cron, Kafka, SQS/SNS, Pub/Sub, GitHub, GitLab, NATS, AMQP
- Apply filters at EventSource level to drop irrelevant events early
- Dedicated service accounts per EventSource

### EventBus
- Default: NATS JetStream — run 3-node cluster for production
- Alternative: Kafka for orgs already running it

### Sensors & Triggers
- Combine dependencies: `A && B` (both required), `A || B` (either)
- Filters to match specific payloads (e.g., push to `main` only)
- Trigger types: Argo Workflow, HTTP, K8s object, Log
- `retryStrategy` with backoff + `dlqTrigger` for dead letter queue

### Integration with Workflows
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

## Terraform Bootstrap Pattern

Terraform owns: cluster, VPC, IAM, Argo CD install, bootstrap app-of-apps.
Argo CD owns: all workloads, namespaces, RBAC, add-ons after bootstrap.

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

Rules:
- Never use `kubernetes_manifest` for Argo-managed resources
- Pin Helm chart versions, never `latest`
- Separate Terraform state per cluster
- Exec-based auth (short-lived tokens), not static kubeconfig
