# Kubernetes Best Practices

## Resource Management
- Always set `requests` and `limits` for CPU and memory on every container
- Use VPA in recommendation mode to right-size, then apply
- Use HPA for production; scale on custom metrics when CPU/memory isn't enough

## Health Checks
- `readinessProbe`: gates traffic — pod only receives requests when ready
- `livenessProbe`: restarts unhealthy pods — use cautiously (wrong thresholds cause restart loops)
- `startupProbe`: protects slow-starting containers — disables liveness/readiness until startup succeeds
- Set `initialDelaySeconds`, `periodSeconds`, `failureThreshold` based on actual app startup time

## Rolling Updates
- `strategy.type: RollingUpdate` with explicit `maxSurge` and `maxUnavailable`
- Set `minReadySeconds` to prevent marking pods ready too quickly
- Use `progressDeadlineSeconds` to auto-fail stuck rollouts
- Define `revisionHistoryLimit` to control stored ReplicaSets

## Secrets & Config
- Use external secrets operators (AWS SM, GCP SM, Vault) over native K8s secrets
- Never store secrets in manifests committed to Git
- Mount secrets as volumes, not env vars (env vars leak in logs/crash dumps)
- Use sealed-secrets or SOPS for GitOps secret management

## Namespaces & RBAC
- One namespace per environment or team
- ResourceQuotas and LimitRanges per namespace
- RBAC with least privilege; avoid `cluster-admin` for app service accounts
- NetworkPolicies to restrict pod-to-pod traffic; default-deny ingress per namespace

## Labels
- Use standard labels consistently:
  - `app.kubernetes.io/name`
  - `app.kubernetes.io/version`
  - `app.kubernetes.io/component`
  - `app.kubernetes.io/managed-by`

## General
- Treat manifests as code: version control, PR reviews, GitOps (ArgoCD or Flux)
- Set `terminationGracePeriodSeconds` to match app shutdown time
- Use PodDisruptionBudgets for critical workloads
- One process per container; use sidecars for cross-cutting concerns

## Deployment Template

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
  labels:
    app.kubernetes.io/name: app
    app.kubernetes.io/version: "1.0.0"
spec:
  replicas: 3
  revisionHistoryLimit: 5
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  selector:
    matchLabels:
      app.kubernetes.io/name: app
  template:
    metadata:
      labels:
        app.kubernetes.io/name: app
    spec:
      serviceAccountName: app
      terminationGracePeriodSeconds: 30
      containers:
        - name: app
          image: registry/app:sha-abc123
          ports:
            - containerPort: 8080
              protocol: TCP
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 256Mi
          readinessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /health
              port: 8080
            initialDelaySeconds: 15
            periodSeconds: 20
          env:
            - name: DB_HOST
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: db-host
```
