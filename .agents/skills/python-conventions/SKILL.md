---
name: python-conventions
description: Convenciones y estándares de código Python para 3.12+. Usar al escribir código Python, revisar patrones Python, o cuando el usuario mencione "python conventions", "type hints", "pydantic", "pytest", "embeddings", "numpy", "async python", o al trabajar con archivos .py.
---

<!-- GENERADO por la skill export-system. NO EDITAR A MANO.
     Fuente de verdad: agents/, skills/, commands/, CLAUDE.md.
     Los cambios hechos aquí se pierden en la próxima exportación. -->


# Python Conventions

> **IMPORTANTE:** Este archivo es un dispatcher ligero. NO cargar todos los archivos referenciados a la vez. Leer la tabla de enrutamiento abajo, identificar qué archivos son relevantes para la tarea actual y cargar SOLO esos usando la herramienta Read. Cada archivo pesa ~3-5KB. Cargar archivos innecesarios desperdicia tokens de contexto.

## Stack y Filosofía

- **Python 3.12+ primero** — type aliases, uniones `X | Y`, `match`, `itertools.batched`
- **Type hints en todas partes** — Pydantic v2, Protocol, TypedDict, `NDArray[np.float32]`
- **Ruff reemplaza todo** — una sola herramienta para lint + formato (sin flake8, black, isort)
- **uv sobre pip** — más rápido, lockfile integrado, herramienta única
- **Estructurado sobre strings** — structlog, modelos Pydantic, no f-strings y dicts
- **Vectorizar, no iterar** — operaciones numpy/pandas, preasignar, hacer todo en batch

## Red Flags (siempre detener el trabajo)

- `from typing import Union, Optional` en Python 3.10+ → usar `X | Y`, `X | None`
- Argumentos por defecto mutables (`def f(x=[])`) → error
- `eval()` o `exec()` con input del usuario → error
- f-strings en queries SQL → SQL injection
- `import *` → error
- `except:` desnudo o `except Exception:` sin re-raise → error
- `os.system()` o `subprocess.call(shell=True)` con input del usuario → error

## Detección de Anti-Patrones

**Detección pasiva:** Al revisar código Python, cargar `detection/anti-patterns.md` y escanear en busca de patrones `error` y `warning`. Reportar como `[file:line] [severity] [category] anti-pattern-name`.

**Detección activa:** Cuando el usuario pide "mejorar", "refactorizar", "optimizar" o "limpiar" — también reportar patrones de nivel `suggestion` y proponer correcciones.

## Qué Cargar

Cargar **solo** los archivos relevantes para la tarea actual:

### Rules (referencia rápida, ~2-3KB cada uno)

| Trabajando en... | Cargar |
|---|---|
| Type hints, naming, manejo de errores, sintaxis moderna | `rules/coding.md` |
| Estructura del proyecto, imports, DI, modelos Pydantic | `rules/architecture.md` |
| NumPy, Pandas, vectorización, gestión de memoria | `rules/data.md` |

### Guides (patrones detallados con código, ~3-5KB cada uno)

| Trabajando en... | Cargar |
|---|---|
| asyncio, TaskGroup, concurrencia estructurada | `guides/async/patterns.md` |
| Context managers asíncronos, timeouts, streaming | `guides/async/resources.md` |
| Fixtures de pytest, parametrize, mocking | `guides/testing/pytest.md` |
| Testing asíncrono, factory fixtures | `guides/testing/async-testing.md` |
| Embeddings, procesamiento en batch, operaciones vectoriales | `guides/ml/embeddings.md` |
| Arrays memory-mapped, datasets grandes, memoria GPU | `guides/ml/memory.md` |
| Context managers, connection pools, cleanup | `guides/cleanup/resources.md` |
| Validación de input, SQL injection, secrets | `guides/security.md` |

### Detección y Checklists

| Cuándo... | Cargar |
|---|---|
| Revisión de código | `detection/anti-patterns.md` |
| Antes de escribir código Python | `checklists/pre.md` |
| Después de escribir código Python | `checklists/post.md` |

### Ejemplos (patrones buenos y malos por dominio, ~2-3KB cada uno)

| Trabajando en... | Cargar |
|---|---|
| Type hints, sintaxis moderna, Pydantic v2 | `examples/types.md` |
| Manejo de errores, exception groups, logging | `examples/errors.md` |
| Patrones de testing, fixtures, mocking | `examples/testing.md` |
| NumPy/Pandas, vectorización, procesamiento en batch | `examples/data.md` |
| Patrones async, TaskGroup, timeouts | `examples/async.md` |

## Gate Post-Implementación

Gate de lint: invocar skill `lint` después de cualquier cambio de código.
