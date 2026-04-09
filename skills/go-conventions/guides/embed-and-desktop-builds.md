# //go:embed, Build Tags, and Desktop Builds (Wails)

Go patterns for embedding assets into binaries, using build tags to gate platform-specific code, and compiling desktop apps without the framework's own build tool.

## `//go:embed` — core rules

### Rule 1: No `../` in embed patterns

`//go:embed` patterns are rooted at the **package directory** where the directive appears. You cannot escape that directory with `../`. This is enforced by the Go toolchain at compile time.

**Wrong (does not compile):**
```go
// internal/dashboard/assets.go
package dashboard

import "embed"

//go:embed ../../frontend/dist  // ERROR: invalid pattern
var assets embed.FS
```

**Right — move the embed to a package whose directory contains the target:**
```go
// frontendfs.go (at module root, same directory as frontend/)
package myapp

import "embed"

//go:embed all:frontend/dist
var DashboardFS embed.FS
```

### Rule 2: Use `all:` prefix to include dotfiles

By default, `//go:embed` skips files and directories whose names begin with `.` or `_`. For asset bundles (Vite, webpack) that may emit `.hidden` files, prefix the pattern with `all:`:

```go
//go:embed all:frontend/dist
var assets embed.FS
```

### Rule 3: Wiring embedded assets across packages

When the embed lives in a root package but the consumer is in `internal/`, do not re-export by reading files — pass the `embed.FS` as a parameter. If the consumer is registered via `init()` (common for CLI commands), a package-level var set by the root package's `init()` is acceptable:

```go
// internal/cli/dashboard.go
package cli

import "embed"

// DashboardAssets is set by cmd/anvil/embed_dashboard.go at init time.
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

This is the **only acceptable use of `init()` with side effects** in this codebase: bridging `//go:embed` directives across package boundaries when the embed path forces a specific package location. Document it inline.

---

## Build tags for optional features

### Pattern: feature under a build tag

When a feature (dashboard, debug tools, profiling) pulls in heavy dependencies or CGO, gate it with a build tag so the default build stays lean:

```go
// internal/dashboard/app.go
//go:build dashboard

package dashboard
// ... uses github.com/wailsapp/wails/v2, CGO, etc.
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

**Invariants:**
- Both files with complementary tags (`dashboard` and `!dashboard`) must export the same identifiers so the CLI wiring compiles in both modes.
- Tests for tagged code must also use the tag: `go test -tags dashboard ./...`.
- `go vet` must pass for both tag sets: `go vet ./...` AND `go vet -tags dashboard ./...`.

### Anti-pattern: orphan package without build constraints

If a package has ONLY files guarded by a build tag, `go build ./...` without the tag will skip the package gracefully — BUT some tooling (`wails build`, `go list -json ./...`) will fail with `build constraints exclude all Go files in /path`.

**Fix:** add at least one file without build constraints to the package (e.g., a `doc.go` declaring the package), OR accept that tooling requiring the package must pass the tag.

---

## Wails v2 desktop builds (case study)

Wails v2 is a Go framework for building desktop apps with a web frontend. Its build tool (`wails build`) assumes a specific project layout:

- `main` package at the **module root**
- `frontend/` directory at the module root
- `wails.json` at the module root

If your project has `main` in `cmd/<app>/` (standard Go layout for multi-command modules), `wails build` will not work. You must build with plain `go build` and set the right tags and linker flags manually.

### Required build tags

Wails v2 uses internal build tags to select runtime mode. Without them, it compiles a stub that returns an error at startup:

```
Wails applications will not build without the correct build tags.
```

Build with:
```bash
go build -tags "your_feature_tag production" ./cmd/app/
```

- `production` — selects the production runtime (alternative: `dev`)
- `your_feature_tag` — your own tag that guards the dashboard package (e.g., `dashboard`)

### Required CGO linker flags (macOS)

Wails v2 uses `UIType` APIs from `UniformTypeIdentifiers.framework` in its Objective-C code, but does NOT declare the framework in `#cgo LDFLAGS` directives. `wails build` injects it via `CGO_LDFLAGS` env var; plain `go build` requires you to set it manually:

```bash
CGO_ENABLED=1 \
CGO_LDFLAGS="-framework UniformTypeIdentifiers" \
go build -tags "dashboard production" -o app-full ./cmd/app/
```

Without this, the linker fails with:
```
Undefined symbols for architecture arm64:
  "_OBJC_CLASS_$_UTType", referenced from: ...
```

### Embedded assets layout for Wails

Wails expects the frontend assets as an `embed.FS` passed to `wails.Run()`:

```go
// at module root — frontendfs.go
//go:build dashboard

package myapp

import "embed"

//go:embed all:frontend/dist
var DashboardFS embed.FS
```

The package at module root does NOT need to be `main`. It can be any package name (commonly the module name). The `embed` directive only requires that the target path exists relative to the file.

### Full Makefile target example

```make
DASHBOARD_BINARY := app-full

dashboard-frontend:
	cd frontend && npm install --ignore-scripts && npm run build

dashboard-build: dashboard-frontend
	CGO_ENABLED=1 \
	CGO_LDFLAGS="-framework UniformTypeIdentifiers" \
	go build -tags "dashboard production" -o $(DASHBOARD_BINARY) ./cmd/app/
```

### When to use `wails build` vs plain `go build`

| Scenario | Tool |
|---|---|
| `main` at module root, new project | `wails build` |
| `main` in `cmd/<app>/`, multi-command module | plain `go build` with tags + CGO_LDFLAGS |
| Production distribution on macOS with `.app` bundle | `wails build` (handles bundle, Info.plist, codesign) |
| Dev iteration, single binary, cross-compile | plain `go build` |

---

## Checklist before shipping a desktop feature

- [ ] Feature is gated by a build tag (e.g., `dashboard`)
- [ ] Stub file with `//go:build !dashboard` exists for the fallback message
- [ ] `go build ./...` (no tag) compiles and does not pull CGO or framework deps
- [ ] `go build -tags "<feature> production" ./...` compiles
- [ ] `go vet` passes for both tag sets
- [ ] Tests for tagged code run with the tag set
- [ ] CGO_LDFLAGS documented in Makefile target (not expected from env)
- [ ] `//go:embed` directive is at a package level where the target path resolves without `../`
- [ ] Binary smoke test on the target platform (native window opens, bindings work)
