# Mejores Prácticas de Docker

## Imágenes Base
- Usa solo imágenes oficiales o de Verified Publisher
- Prefiere mínimas: `alpine`, `distroless`, o variantes de Chainguard
- Fija al digest para producción: `FROM alpine:3.21@sha256:abcd...`
- Reconstruye regularmente para incorporar parches de seguridad

## Builds Multi-Stage
- Separa la etapa de build (compiladores, deps de dev) de la etapa de runtime (artefactos + runtime mínimo)
- Nunca envíes herramientas de build, gestores de paquetes, o código fuente en la imagen final
- Crea etapas base reutilizables para componentes compartidos

## Caché de Capas
- Ordena las instrucciones de menor a mayor cambio (deps de OS antes que código de app)
- Combina RUN: `RUN apt-get update && apt-get install -y --no-install-recommends pkg && rm -rf /var/lib/apt/lists/*`
- Copia primero los manifiestos de dependencias: `COPY go.mod go.sum ./` luego `RUN go mod download` antes del código fuente
- Ordena argumentos multi-línea alfabéticamente

## Seguridad
- Siempre ejecuta como non-root: crea usuario con UID/GID explícitos, usa directiva `USER`
- Nunca embebas secretos o claves en las capas de imagen
- Usa `.dockerignore`: excluye `.env`, `.git`, credenciales, archivos de IDE, `node_modules`
- Usa `COPY` sobre `ADD` a menos que extraigas archivos
- Forma exec para ENTRYPOINT: `ENTRYPOINT ["executable"]` (manejo correcto de señales como PID 1)
- Escanea imágenes en CI con Trivy o Grype antes de hacer push

## General
- Un proceso por contenedor
- Usa `WORKDIR` con rutas absolutas, nunca `RUN cd`
- Usa `EXPOSE` para documentar puertos (no es seguridad)
- Fija versiones de paquetes: `package=1.3.*`
- Siempre establece `HEALTHCHECK` para imágenes de producción

## Plantilla de Dockerfile para Go

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

## Plantilla de Dockerfile para React/Node

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
