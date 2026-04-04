# GitHub Actions Best Practices

## Workflow Structure
- One concern per job; keep jobs focused and fast
- Set `permissions:` explicitly on every workflow (least privilege)
- Pin actions to full commit SHA: `uses: actions/checkout@<sha>` (not tags)
- Use path filters: `on.push.paths` to skip pipelines for non-code changes
- Use `concurrency` groups to cancel superseded runs on same branch

## Reusable Workflows
- Define with `on: workflow_call` in `.github/workflows/`
- Reference: `uses: ./.github/workflows/file.yml` or `org/repo/.github/workflows/file.yml@sha`
- Use `secrets: inherit` for same-org; explicitly forward otherwise
- Max nesting: 10 levels; no loops
- Permissions can only be maintained or reduced in nested chains

## Composite Actions vs Reusable Workflows
- **Composite actions**: small repeated step sequences (setup, formatting, caching)
- **Reusable workflows**: full shared pipelines (build + test + deploy)

## Caching
- Enable dependency caching (`actions/cache` or built-in setup actions)
- Docker layer caching: `cache-from: type=gha` / `cache-to: type=gha,mode=max`
- Parallelize independent jobs — remove unnecessary `needs:` dependencies

## Secrets
- Use OIDC for cloud auth (no long-lived keys):
  - AWS: `aws-actions/configure-aws-credentials`
  - GCP: `google-github-actions/auth`
- Never echo or print secrets; mask with `::add-mask::`
- Store at org level for shared, repo level for project-specific
- Rotate on schedule; prefer short-lived tokens

## Matrix Builds
- `strategy.matrix` for cross-platform/version testing
- `fail-fast: false` when all combinations must complete
- Combine with caching to avoid redundant installs per cell

## CI Pipeline Template (Go)

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

## CD Pipeline Template (Deploy to Cloud Run)

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
