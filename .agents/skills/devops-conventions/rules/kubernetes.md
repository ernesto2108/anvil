# Mejores Prácticas de Kubernetes

## Gestión de Recursos
- Siempre establece `requests` y `limits` para CPU y memoria en cada contenedor
- Usa VPA en modo de recomendación para dimensionar correctamente, luego aplica
- Usa HPA para producción; escala en métricas personalizadas cuando CPU/memoria no es suficiente

## Health Checks
- `readinessProbe`: controla el tráfico — el pod solo recibe requests cuando está listo
- `livenessProbe`: reinicia pods no saludables — úsalo con cuidado (umbrales incorrectos causan bucles de reinicio)
- `startupProbe`: protege contenedores de arranque lento — desactiva liveness/readiness hasta que el arranque tenga éxito
- Establece `initialDelaySeconds`, `periodSeconds`, `failureThreshold` basado en el tiempo de arranque real de la app

## Rolling Updates
- `strategy.type: RollingUpdate` con `maxSurge` y `maxUnavailable` explícitos
- Establece `minReadySeconds` para evitar marcar pods como listos demasiado rápido
- Usa `progressDeadlineSeconds` para auto-fallar rollouts atascados
- Define `revisionHistoryLimit` para controlar los ReplicaSets almacenados

## Secretos y Config
- Usa operadores de secretos externos (AWS SM, GCP SM, Vault) sobre secretos nativos de K8s
- Nunca almacenes secretos en manifiestos commiteados a Git
- Monta secretos como volúmenes, no como variables de entorno (las env vars se filtran en logs/crash dumps)
- Usa sealed-secrets o SOPS para gestión de secretos GitOps

## Namespaces y RBAC
- Un namespace por entorno o equipo
- ResourceQuotas y LimitRanges por namespace
- RBAC con menor privilegio; evita `cluster-admin` para cuentas de servicio de apps
- NetworkPolicies para restringir el tráfico pod-a-pod; default-deny ingress por namespace

## Labels
- Usa labels estándar de forma consistente:
  - `app.kubernetes.io/name`
  - `app.kubernetes.io/version`
  - `app.kubernetes.io/component`
  - `app.kubernetes.io/managed-by`

## General
- Trata los manifiestos como código: control de versiones, PR reviews, GitOps (ArgoCD o Flux)
- Establece `terminationGracePeriodSeconds` para que coincida con el tiempo de shutdown de la app
- Usa PodDisruptionBudgets para cargas de trabajo críticas
- Un proceso por contenedor; usa sidecars para concerns transversales

## Plantilla de Deployment

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
