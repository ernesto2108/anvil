---
name: lint
description: Ejecutar linters y formateadores. Auto-detecta el stack (Go, React, Flutter, Swift, Python, Rust). Usar cuando el usuario diga "lint", "revisar estilo de código", "formatear código", "ejecutar eslint", "ejecutar prettier", "golangci-lint", "dart analyze", "swiftlint", "swiftformat", "ruff", "cargo clippy", o después de escribir código que debe validarse. OBLIGATORIO después de cualquier modificación de código — invocar antes de considerar cualquier tarea de código como terminada.
---

# Lint

## Gate Obligatorio Post-Código

Este skill DEBE invocarse después de cualquier modificación de código (Write, Edit en archivos fuente) antes de considerar una tarea como terminada. No es opcional.

1. Ejecutar el linter para el stack detectado
2. Ejecutar el formateador para el stack detectado
3. Si hay nuevos errores de lint → corregir inmediatamente, no dejar para una tarea separada
4. Cero nuevas violaciones es el estándar — nunca aumentar el conteo de errores de lint

Este skill maneja **únicamente estilo de código y análisis estático**. Para ejecutar tests, usar `/run-tests`.

Aplica a todos los agentes (desarrollador, tester) y ediciones directas. Nunca enviar código sin pasar este gate.

## Resolución de Comandos (Precedencia OBLIGATORIA)

Antes de ejecutar cualquier linter, resolver el comando en este orden estricto. **No saltar pasos** — el fallback genérico es el último recurso.

### 1. Makefile presente (mayor prioridad)

Si existe `Makefile` en la raíz:

```bash
grep -E '^[a-zA-Z0-9_-]+:' Makefile | sed 's/:.*//'
```

Listar todos los targets y elegir el de lint por:

- **Match por nombre**: target cuyo nombre contenga `lint`, `check`, `vet`, `analyze`, `static`, `validate`, `fmt`, `format`, `style`.
- **Match por contenido**: si ningún nombre matchea, inspeccionar el cuerpo de los targets (`make -n <target>` o leer Makefile) buscando comandos como `golangci-lint`, `eslint`, `ruff`, `cargo clippy`, `dart analyze`, `prettier`, `gofmt`.

Si hay un candidato razonable → usar `make <target>` **sin preguntar**.

Si existe Makefile pero ningún target matchea por nombre ni por contenido → **preguntar al humano una sola vez**:

> "Encontré un Makefile pero no pude identificar el target de lint. ¿Qué target debo usar? (o `skip` para usar el comando por defecto)"

Cachear la respuesta para el resto de la sesión.

### 2. `package.json` con script `lint`

Si existe `package.json` con un script `lint` definido → usar `<pm> lint` (detectar `<pm>` desde lockfile según CLAUDE.md).

### 3. Fallback — comando directo del tool

Solo si los pasos 1 y 2 no aplicaron, usar el comando directo según el stack:

| Archivo | Stack | Linter | Formateador |
|------|-------|--------|-----------|
| `go.mod` | Go | `golangci-lint run ./...` | `gofmt` (built-in) |
| `package.json` | React/Node | `<pm> exec eslint .` | `<pm> exec prettier --check .` |
| `pubspec.yaml` | Flutter | `dart analyze` | `dart format --set-exit-if-changed .` |
| `Package.swift` / `*.xcodeproj` / `*.xcworkspace` | Swift | `swiftlint lint` | `swiftformat --lint .` |
| `pyproject.toml` o `*.py` | Python | `ruff check .` | `ruff format --check .` |
| `Cargo.toml` o `*.rs` | Rust | `cargo clippy -- -D warnings` | `cargo fmt --check` |

Si se detectan múltiples stacks, lintear cada uno por separado.

## Ejecución

### Go

```bash
# Si existe .golangci.yml o .golangci.yaml, se usará automáticamente
golangci-lint run ./...
```

Si `golangci-lint` no está instalado:
```bash
go vet ./...
```

Auto-fix: `golangci-lint run --fix ./...` — luego reportar qué no se pudo corregir.

### React/Node

**Package manager:** detectar desde el lockfile según la regla de CLAUDE.md (`pnpm-lock.yaml` → pnpm, `yarn.lock` → yarn, `package-lock.json` → npm, ninguno → pnpm por defecto). Usar `pnpm exec` / `npx` / `yarn exec` para ejecutar binarios desde `node_modules/.bin`.

```bash
# pnpm (preferido)
pnpm exec eslint . --ext .ts,.tsx,.js,.jsx
pnpm exec prettier --check "src/**/*.{ts,tsx,js,jsx}"

# Auto-fix
pnpm exec eslint . --ext .ts,.tsx,.js,.jsx --fix
pnpm exec prettier --write "src/**/*.{ts,tsx,js,jsx}"

# npm equivalente: cambiar `pnpm exec` → `npx`
# yarn equivalente: cambiar `pnpm exec` → `yarn exec`
```

Verificar primero los scripts de `package.json` — si el proyecto define scripts `lint` / `format`, preferir `<pm> lint` (pnpm) / `npm run lint` / `yarn lint` sobre llamar directamente a los binarios.

Archivos de configuración: `.eslintrc.*` o `eslint.config.*`, `.prettierrc`. Recomendado: `eslint-config-react-app` o `@typescript-eslint`.

### Flutter

```bash
# Analizar
dart analyze

# Verificar formato
dart format --set-exit-if-changed .

# Auto-fix
dart fix --apply
dart format .
```

Configuración: `analysis_options.yaml`. Recomendado: `flutter_lints` o `very_good_analysis`.

Problemas comunes auto-corregibles: imports no utilizados, constructores `const` faltantes, preferir `final` para variables inmutables.

### Swift

SwiftLint vigila seguridad y lógica; SwiftFormat formatea estilo. Usar ambos (sin reglas de whitespace solapadas — deshabilitarlas en SwiftLint).

```bash
# Verificar
swiftlint lint
swiftformat --lint .

# Auto-fix
swiftlint lint --fix
swiftformat .

# Modo estricto para CI (cualquier warning falla)
swiftlint lint --strict
```

Configuración: `.swiftlint.yml` (recomendado opt-in `force_unwrapping`, `implicitly_unwrapped_optional`, `empty_count`, `first_where`; `excluded: [.build, Derived, Pods]`) y `.swiftformat` (`--swiftversion 6.0`, `--indent 4`, `--self remove`, `--commas always`).

Versiones fijadas por proyecto (Mint o SPM build tool plugins) — no depender del binario global de Homebrew. En proyectos legacy, generar baseline (`swiftlint lint --baseline`) para bloquear solo violaciones nuevas.

Fallback: si `swiftlint`/`swiftformat` no están instalados, **reportar y sugerir instalación** (`brew install swiftlint swiftformat` o vía Mint/SPM plugin). No hay equivalente built-in del toolchain con la misma cobertura; no continuar en silencio.

### Python

```bash
ruff check .
ruff format --check .

# Auto-fix
ruff check --fix .
ruff format .
```

Configuración: `pyproject.toml` (sección `[tool.ruff]`) o `ruff.toml`. Ruff reemplaza flake8, black e isort en una sola herramienta.

### Rust

```bash
cargo fmt --check
cargo clippy -- -D warnings

# Auto-fix
cargo fmt
cargo clippy --fix -- -D warnings
```

Configuración: `rustfmt.toml` para formato, `clippy.toml` o `[lints]` en `Cargo.toml` para clippy. El flag `-D warnings` trata cualquier warning como error.

## Flujo de trabajo

1. **Detectar stack** — verificar archivos marcadores (`go.mod`, `package.json`, `pubspec.yaml`, `Package.swift`/`*.xcodeproj`/`*.xcworkspace`, `pyproject.toml`, `Cargo.toml`)
2. **Resolver comando según precedencia** — aplicar la sección "Resolución de Comandos" (Makefile → `package.json` script → fallback directo). No saltar al fallback si hay Makefile o script `lint`.
3. **Auto-fix primero** — ejecutar el comando de corrección antes de reportar
4. **Luego verificar** — ejecutar el comando de verificación para encontrar problemas restantes
5. **Si errors > 0** — reportar errores con file:line, NO proceder a la siguiente tarea hasta que los errores estén resueltos
6. **Si solo warnings** — reportar warnings, continuar a menos que el usuario quiera corregirlos
7. **Reportar conteos** — `Errors: 3 | Warnings: 7 | Fixed: 12`

## Formato de Salida

```
Stack: Go
Linter: golangci-lint (config: .golangci.yml)

Auto-fixed: 5 issues
Remaining:
  Errors (2):
    - internal/user/service.go:45 — ineffectual assignment to err (ineffassign)
    - internal/order/handler.go:12 — error return value not checked (errcheck)
  Warnings (1):
    - internal/billing/calc.go:78 — function too complex, cyclomatic complexity 15 (cyclop)
```

Si está limpio: `Lint passed. No issues found.`
