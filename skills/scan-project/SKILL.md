---
name: scan-project
disable-model-invocation: true
description: Escanear la estructura del repo y escribir context.md en el vault con el objetivo del producto y un snapshot técnico. Usar al iniciar una nueva sesión, cuando el usuario diga "scan the project", "what stack is this", "analyze the repo", o cuando context.md esté ausente o desactualizado.
---

# Scan Project

Descubrir la estructura REAL del repositorio y las herramientas utilizadas. NO asumir ninguna arquitectura ni nombres de carpetas. Reflejar el proyecto exactamente como existe.

## Paso 1: Objetivo del Producto

Si el contexto del objetivo del producto está ausente o desactualizado en `{context_path}` (resolver desde `~/.claude/project-registry.md` — ver vault-setup path table), hacer estas preguntas primero:
1. "What is the project objective in 3-6 lines?"
2. "What non-negotiable rules must I always respect?"

## Paso 2: Detectar Stacks

Verificar estos archivos marcadores para determinar qué stacks están presentes:

| Archivo | Stack | Qué recopilar |
|------|-------|-----------------|
| `go.mod` | Go | Versión de Go, módulos, ubicaciones de `*_test.go`, `.golangci.*` |
| `package.json` | Node/React | Versión de Node, framework (Next/Vite/CRA), ubicaciones de `*.test.tsx`, config de eslint |
| `pubspec.yaml` | Flutter | Versión de Dart, dependencias, ubicaciones de `*_test.dart`, `analysis_options.yaml` |
| `Cargo.toml` | Rust | Edition, dependencias |
| `requirements.txt` / `pyproject.toml` | Python | Versión de Python, framework |

Múltiples stacks pueden coexistir (ej. backend Go + frontend React).

## Paso 3: Recopilar Información

Para TODOS los stacks:
1. Árbol de directorios (profundidad 3)
2. Hints de CI / runtime — `Dockerfile`, `docker-compose.*`, `.github/workflows/*`
3. Archivos de configuración — `Makefile`, `taskfile`, scripts
4. Ignorar: `<docs>/`, `.git/`, `vendor/`, `node_modules/`, `dist/`, `build/`, `tmp/`, `.next/`

### Específico de Go
- Leer `go.mod` — versión y dependencias
- Buscar `*_test.go` — listar ubicaciones de tests
- Buscar `.golangci.yml` / `.golangci.yaml`
- Buscar estructura de `internal/`, `cmd/`, `pkg/`

### Específico de React/Node
- Leer `package.json` — scripts, dependencies, devDependencies
- Detectar framework: Next.js (`next.config.*`), Vite (`vite.config.*`), CRA (`react-scripts`)
- Buscar `*.test.tsx`, `*.test.ts`, `*.spec.tsx`
- Buscar config de eslint (`.eslintrc.*`, `eslint.config.*`)
- Buscar config de prettier (`.prettierrc.*`)
- Verificar `tsconfig.json`

### Específico de Flutter
- Leer `pubspec.yaml` — dependencies, dev_dependencies
- Buscar `*_test.dart`
- Verificar `analysis_options.yaml`
- Verificar `l10n.yaml` (localización)
- Verificar estructura de `lib/`, `test/`, `integration_test/`

## Paso 4: Escribir Salida

Escribir ÚNICAMENTE en `{context_path}` (sobreescribir si existe). Resolver la ruta desde vault-setup path table:
- Obsidian vault: `<docs>/01-project/context.md`
- Linear+Outline o `.workspace/`: `.workspace/context.md`
Nunca eliminar secciones técnicas al agregar contexto del objetivo del producto; conservar ambas.

Usar el template de `output-template.md` — incluir solo las secciones para los stacks detectados.

## Paso 5: Bootstrap de Context Navigator (modo deep o primer scan)

Si se ejecuta en `mode: deep` O si `.project-context/NAVIGATOR.md` no existe en el proyecto:

1. Cargar `skills/context-nav/bootstrap.md` — define las firmas de código a buscar por stack
2. Ejecutar los greps de inferencia de patrones del bootstrap (Paso 2 del bootstrap)
3. Ejecutar los greps de contratos del bootstrap (Paso 3 del bootstrap)
4. Detectar bounded contexts desde estructura de directorios (Paso 4 del bootstrap)
5. Ejecutar detección SOLID (Paso 5 del bootstrap)
6. Escribir todos los archivos en `.project-context/` usando los templates de `skills/context-nav/templates/`
7. Marcar `coverage: bootstrap` en `.project-context/NAVIGATOR.md`

**Si `.project-context/NAVIGATOR.md` ya existe con `coverage: bootstrap` o superior** y el diff con el último commit es < 3 días: saltar el Paso 5 — el contexto está fresco.

## Checklist de Acciones

- [ ] Leer archivos marcadores (`go.mod`, `package.json`, `pubspec.yaml`)
- [ ] Listar directorios (profundidad 3)
- [ ] Por stack detectado: recopilar versión, deps, archivos de test, config del linter
- [ ] Buscar archivos CI (`.github/workflows/*`, `Dockerfile`, `docker-compose.*`)
- [ ] Buscar herramientas de build (`Makefile`, `taskfile.*`)
- [ ] Preguntar sobre el objetivo del producto si está ausente
- [ ] Escribir `{context_path}`
- [ ] Si mode: deep o sin `.project-context/`: bootstrap de `.project-context/` según Paso 5
