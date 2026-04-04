# Docker Best Practices

## Base Images
- Use official or Verified Publisher images only
- Prefer minimal: `alpine`, `distroless`, or Chainguard variants
- Pin to digest for production: `FROM alpine:3.21@sha256:abcd...`
- Rebuild regularly to pick up security patches

## Multi-Stage Builds
- Separate build stage (compilers, dev deps) from runtime stage (artifacts + minimal runtime)
- Never ship build tools, package managers, or source in final image
- Create reusable base stages for shared components

## Layer Caching
- Order instructions least-changing → most-changing (OS deps before app code)
- Combine RUN: `RUN apt-get update && apt-get install -y --no-install-recommends pkg && rm -rf /var/lib/apt/lists/*`
- Copy dependency manifests first: `COPY go.mod go.sum ./` then `RUN go mod download` before source
- Sort multi-line arguments alphabetically

## Security
- Always run as non-root: create user with explicit UID/GID, use `USER` directive
- Never embed secrets or keys in image layers
- Use `.dockerignore`: exclude `.env`, `.git`, credentials, IDE files, `node_modules`
- Use `COPY` over `ADD` unless extracting archives
- Exec form for ENTRYPOINT: `ENTRYPOINT ["executable"]` (proper signal handling as PID 1)
- Scan images in CI with Trivy or Grype before pushing

## General
- One process per container
- Use `WORKDIR` with absolute paths, never `RUN cd`
- Use `EXPOSE` to document ports (not security)
- Pin package versions: `package=1.3.*`
- Always set `HEALTHCHECK` for production images

## Go Dockerfile Template

```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder
RUN apk add --no-cache ca-certificates git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server

# Runtime stage
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /app/server /server
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/server"]
```

## React/Node Dockerfile Template

```dockerfile
# Build stage
FROM node:22-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --ignore-scripts
COPY . .
RUN npm run build

# Runtime stage
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
```
