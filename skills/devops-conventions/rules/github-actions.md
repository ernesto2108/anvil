# Mejores Prácticas de GitHub Actions

## Estructura del Workflow
- Un concern por job; mantén los jobs enfocados y rápidos
- Establece `permissions:` explícitamente en cada workflow (menor privilegio)
- Fija actions al SHA completo del commit: `uses: actions/checkout@<sha>` (no tags)
- Usa filtros de ruta: `on.push.paths` para omitir pipelines en cambios que no son código
- Usa grupos de `concurrency` para cancelar runs supersedidos en la misma rama

## Workflows Reutilizables
- Define con `on: workflow_call` en `.github/workflows/`
- Referencia: `uses: ./.github/workflows/file.yml` o `org/repo/.github/workflows/file.yml@sha`
- Usa `secrets: inherit` para la misma org; pasa explícitamente de lo contrario
- Máximo de anidamiento: 10 niveles; sin bucles
- Los permisos solo pueden mantenerse o reducirse en cadenas anidadas

## Composite Actions vs Reusable Workflows
- **Composite actions**: secuencias de pasos pequeñas y repetidas (setup, formateo, caché)
- **Reusable workflows**: pipelines completos compartidos (build + test + deploy)

## Caché
- Habilita caché de dependencias (`actions/cache` o actions de setup incorporadas)
- Caché de capas Docker: `cache-from: type=gha` / `cache-to: type=gha,mode=max`
- Paraleliza jobs independientes — elimina dependencias `needs:` innecesarias

## Secretos
- Usa OIDC para auth en la nube (sin claves de larga duración):
  - AWS: `aws-actions/configure-aws-credentials`
  - GCP: `google-github-actions/auth`
- Nunca hagas echo o prints de secretos; enmascara con `::add-mask::`
- Almacena a nivel de org para los compartidos, a nivel de repo para los específicos del proyecto
- Rota en schedule; prefiere tokens de corta duración

## Matrix Builds
- `strategy.matrix` para testing cross-platform/versión
- `fail-fast: false` cuando todas las combinaciones deben completarse
- Combina con caché para evitar instalaciones redundantes por celda

## Plantilla de Pipeline CI (Go)

```yaml
name: CI

on:
  push:
    branches: [main]
    paths: ['**.go', 'go.mod', 'go.sum', '.github/workflows/ci.yml']
  pull_request:
    branches: [main]

permissions:
  contents: read

concurrency:
  group: ci-${{ github.ref }}
  cancel-in-progress: true

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@<sha>
      - uses: actions/setup-go@<sha>
        with:
          go-version-file: go.mod
          cache: true
      - uses: golangci/golangci-lint-action@<sha>
        with:
          version: latest

  test:
    runs-on: ubuntu-latest
    needs: lint
    steps:
      - uses: actions/checkout@<sha>
      - uses: actions/setup-go@<sha>
        with:
          go-version-file: go.mod
          cache: true
      - run: go test -race -coverprofile=coverage.out ./...
      - uses: actions/upload-artifact@<sha>
        with:
          name: coverage
          path: coverage.out

  build:
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@<sha>
      - uses: docker/setup-buildx-action@<sha>
      - uses: docker/build-push-action@<sha>
        with:
          context: .
          push: false
          cache-from: type=gha
          cache-to: type=gha,mode=max
          tags: app:${{ github.sha }}
```

## Plantilla de Pipeline CD (Deploy a Cloud Run)

```yaml
name: Deploy

on:
  push:
    branches: [main]

permissions:
  contents: read
  id-token: write  # OIDC

jobs:
  deploy:
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@<sha>
      - uses: google-github-actions/auth@<sha>
        with:
          workload_identity_provider: ${{ vars.WIF_PROVIDER }}
          service_account: ${{ vars.SA_EMAIL }}
      - uses: google-github-actions/setup-gcloud@<sha>
      - run: |
          gcloud builds submit --tag $REGION-docker.pkg.dev/$PROJECT/$REPO/$IMAGE:${{ github.sha }}
          gcloud run deploy $SERVICE --image $REGION-docker.pkg.dev/$PROJECT/$REPO/$IMAGE:${{ github.sha }} --region $REGION
```
