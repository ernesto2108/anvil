# //go:embed, Build Tags y Builds de Escritorio (Wails)

Patrones de Go para incrustar assets en binarios, usar build tags para condicionar código por plataforma, y compilar apps de escritorio sin la propia herramienta de build del framework.

## `//go:embed` — reglas fundamentales

### Regla 1: Sin `../` en los patrones de embed

Los patrones de `//go:embed` están enraizados en el **directorio del package** donde aparece la directiva. No puedes escapar ese directorio con `../`. Esto es aplicado por el toolchain de Go en tiempo de compilación.

**Incorrecto (no compila):**
```go
// internal/dashboard/assets.go
package dashboard

import "embed"

//go:embed ../../frontend/dist  // ERROR: invalid pattern
var assets embed.FS
```

**Correcto — mueve el embed a un package cuyo directorio contenga el objetivo:**
```go
// frontendfs.go (en el root del módulo, mismo directorio que frontend/)
package myapp

import "embed"

//go:embed all:frontend/dist
var DashboardFS embed.FS
```

### Regla 2: Usa el prefijo `all:` para incluir dotfiles

Por defecto, `//go:embed` omite archivos y directorios cuyos nombres comienzan con `.` o `_`. Para bundles de assets (Vite, webpack) que pueden emitir archivos `.hidden`, prefija el patrón con `all:`:

```go
//go:embed all:frontend/dist
var assets embed.FS
```

### Regla 3: Conectar assets embebidos entre packages

Cuando el embed vive en un package raíz pero el consumidor está en `internal/`, no re-exportes leyendo archivos — pasa el `embed.FS` como parámetro. Si el consumidor se registra vía `init()` (común en comandos CLI), una var a nivel de package definida por el `init()` del package raíz es aceptable:

```go
// internal/cli/dashboard.go
package cli

import "embed"

// DashboardAssets es definida por cmd/anvil/embed_dashboard.go en tiempo de init.
var DashboardAssets embed.FS
```

```go
// cmd/anvil/embed_dashboard.go
//go:build dashboard

package main

import (
    anvilroot "github.com/org/myapp"
    "github.com/org/myapp/internal/cli"
)

func init() {
    cli.DashboardAssets = anvilroot.DashboardFS
}
```

Este es el **único uso aceptable de `init()` con efectos secundarios** en esta codebase: hacer de puente para directivas `//go:embed` entre límites de packages cuando la ruta del embed fuerza una ubicación específica de package. Documéntalo inline.

---

## Build tags para features opcionales

### Patrón: feature bajo un build tag

Cuando una feature (dashboard, debug tools, profiling) trae dependencias pesadas o CGO, condiciónala con un build tag para que el build por defecto sea ligero:

```go
// internal/dashboard/app.go
//go:build dashboard

package dashboard
// ... usa github.com/wailsapp/wails/v2, CGO, etc.
```

```go
// internal/cli/dashboard_nodashboard.go
//go:build !dashboard

package cli

import "os"

func cmdDashboard(_ *config.App) {
    output.Error("dashboard not available in this build")
    os.Exit(1)
}
```

**Invariantes:**
- Ambos archivos con tags complementarios (`dashboard` y `!dashboard`) deben exportar los mismos identificadores para que el wiring del CLI compile en ambos modos.
- Los tests para código con tag también deben usar el tag: `go test -tags dashboard ./...`.
- `go vet` debe pasar para ambos conjuntos de tags: `go vet ./...` Y `go vet -tags dashboard ./...`.

### Anti-patrón: package huérfano sin build constraints

Si un package tiene SOLO archivos protegidos por un build tag, `go build ./...` sin el tag omitirá el package silenciosamente — PERO algunas herramientas (`wails build`, `go list -json ./...`) fallarán con `build constraints exclude all Go files in /path`.

**Fix:** agrega al menos un archivo sin build constraints al package (p.ej. un `doc.go` declarando el package), O acepta que las herramientas que requieran el package deben pasar el tag.

---

## Builds de escritorio Wails v2 (caso de estudio)

Wails v2 es un framework de Go para construir apps de escritorio con un frontend web. Su herramienta de build (`wails build`) asume un layout específico del proyecto:

- Package `main` en la **raíz del módulo**
- Directorio `frontend/` en la raíz del módulo
- `wails.json` en la raíz del módulo

Si tu proyecto tiene `main` en `cmd/<app>/` (layout estándar de Go para módulos multi-comando), `wails build` no funcionará. Debes compilar con `go build` plano y configurar manualmente los tags y flags del linker correctos.

### Build tags requeridos

Wails v2 usa build tags internos para seleccionar el modo de runtime. Sin ellos, compila un stub que retorna un error al arrancar:

```
Wails applications will not build without the correct build tags.
```

Compilar con:
```bash
go build -tags "your_feature_tag production" ./cmd/app/
```

- `production` — selecciona el runtime de producción (alternativa: `dev`)
- `your_feature_tag` — tu propio tag que protege el package del dashboard (p.ej. `dashboard`)

### Flags CGO del linker requeridos (macOS)

Wails v2 usa APIs `UIType` de `UniformTypeIdentifiers.framework` en su código Objective-C, pero NO declara el framework en directivas `#cgo LDFLAGS`. `wails build` lo inyecta vía la variable de entorno `CGO_LDFLAGS`; `go build` plano requiere que lo configures manualmente:

```bash
CGO_ENABLED=1 \
CGO_LDFLAGS="-framework UniformTypeIdentifiers" \
go build -tags "dashboard production" -o app-full ./cmd/app/
```

Sin esto, el linker falla con:
```
Undefined symbols for architecture arm64:
  "_OBJC_CLASS_$_UTType", referenced from: ...
```

### Layout de assets embebidos para Wails

Wails espera los assets del frontend como un `embed.FS` pasado a `wails.Run()`:

```go
// en el root del módulo — frontendfs.go
//go:build dashboard

package myapp

import "embed"

//go:embed all:frontend/dist
var DashboardFS embed.FS
```

El package en el root del módulo NO necesita ser `main`. Puede tener cualquier nombre de package (comúnmente el nombre del módulo). La directiva `embed` solo requiere que la ruta objetivo exista relativa al archivo.

### Ejemplo completo de target en Makefile

```make
DASHBOARD_BINARY := app-full

dashboard-frontend:
	cd frontend && npm install --ignore-scripts && npm run build

dashboard-build: dashboard-frontend
	CGO_ENABLED=1 \
	CGO_LDFLAGS="-framework UniformTypeIdentifiers" \
	go build -tags "dashboard production" -o $(DASHBOARD_BINARY) ./cmd/app/
```

### Cuándo usar `wails build` vs `go build` plano

| Escenario | Herramienta |
|---|---|
| `main` en el root del módulo, proyecto nuevo | `wails build` |
| `main` en `cmd/<app>/`, módulo multi-comando | `go build` plano con tags + CGO_LDFLAGS |
| Distribución en producción en macOS con bundle `.app` | `wails build` (maneja bundle, Info.plist, codesign) |
| Iteración en dev, binario único, cross-compile | `go build` plano |

---

## Checklist antes de entregar una feature de escritorio

- [ ] La feature está condicionada por un build tag (p.ej. `dashboard`)
- [ ] Existe un archivo stub con `//go:build !dashboard` para el mensaje de fallback
- [ ] `go build ./...` (sin tag) compila y no trae CGO ni dependencias del framework
- [ ] `go build -tags "<feature> production" ./...` compila
- [ ] `go vet` pasa para ambos conjuntos de tags
- [ ] Los tests para código con tag se ejecutan con el tag activo
- [ ] CGO_LDFLAGS documentado en el target del Makefile (no se espera del entorno)
- [ ] La directiva `//go:embed` está en un nivel de package donde la ruta objetivo se resuelve sin `../`
- [ ] Smoke test del binario en la plataforma objetivo (la ventana nativa abre, los bindings funcionan)
