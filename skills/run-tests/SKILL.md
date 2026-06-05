---
name: run-tests
description: Ejecutar tests del proyecto con detector de race conditions y cobertura. Auto-detecta el stack (Go, React, Flutter, Python, Rust). Usar cuando el usuario diga "run tests", "test this", "check coverage", "run vitest", "go test", "flutter test", "pytest", "cargo test", o después de implementar código que necesita verificación.
---

# Run Tests

## Resolución de Comandos (Precedencia OBLIGATORIA)

Antes de ejecutar tests, resolver el comando en este orden estricto. **No saltar pasos** — el fallback genérico es el último recurso.

### 1. Makefile presente (mayor prioridad)

Si existe `Makefile` en la raíz:

```bash
grep -E '^[a-zA-Z0-9_-]+:' Makefile | sed 's/:.*//'
```

Listar todos los targets y elegir el de test por:

- **Match por nombre**: target cuyo nombre contenga `test`, `spec`, `coverage`, `ci`, `verify`, `check`.
- **Match por contenido**: si ningún nombre matchea, inspeccionar el cuerpo de los targets (`make -n <target>` o leer Makefile) buscando comandos como `go test`, `vitest`, `jest`, `pytest`, `cargo test`, `flutter test`.

Si hay un candidato razonable → usar `make <target>` **sin preguntar**.

Si existe Makefile pero ningún target matchea por nombre ni por contenido → **preguntar al humano una sola vez**:

> "Encontré un Makefile pero no pude identificar el target de tests. ¿Qué target debo usar? (o `skip` para usar el comando por defecto)"

Cachear la respuesta para el resto de la sesión.

### 2. `package.json` con script `test`

Si existe `package.json` con un script `test` definido → usar `<pm> test` (detectar `<pm>` desde lockfile según CLAUDE.md).

### 3. Fallback — comando directo del tool

Solo si los pasos 1 y 2 no aplicaron, usar el comando directo según el stack:

| Archivo | Stack | Comando |
|------|-------|---------|
| `go.mod` | Go | `go test ./... -race -cover -count=1` |
| `package.json` | Node/React | `<pm> exec vitest run --coverage` o `<pm> exec jest --coverage` |
| `pubspec.yaml` | Flutter | `flutter test --coverage` |
| `pyproject.toml` | Python | `pytest --tb=short -q` |
| `Cargo.toml` | Rust | `cargo test` |

Si se detectan múltiples stacks, ejecutar los tests de cada stack por separado.

Para Node/React (cuando se cae al fallback del paso 3): verificar `package.json` para el test runner — preferir `vitest` si está configurado, caer a `jest`. Detectar el package manager desde el lockfile según CLAUDE.md (`pnpm-lock.yaml` → pnpm, `yarn.lock` → yarn, `package-lock.json` → npm) y usarlo de forma consistente.

## Ejecución

### Go
```bash
go test ./... -race -cover -count=1
```

Flags:
- `-race`: detectar race conditions (siempre activado)
- `-cover`: mostrar cobertura por paquete
- `-count=1`: deshabilitar cache de tests (garantiza ejecución fresca)

Para tests de integración (si se solicita):
```bash
go test ./... -race -cover -count=1 -tags integration
```

Para un paquete específico:
```bash
go test -race -cover -count=1 ./internal/user/...
```

### React/Node

Detectar el package manager desde el lockfile según CLAUDE.md. Los ejemplos a continuación usan `pnpm` (por defecto); cambiar a `npm` / `yarn` según lo detectado.

```bash
# Vitest (preferido)
pnpm exec vitest run --coverage        # npm: npx vitest run --coverage
                                       # yarn: yarn exec vitest run --coverage

# Fallback Jest
pnpm exec jest --coverage --passWithNoTests

# Fallback genérico (ejecuta el script "test" de package.json)
pnpm test -- --coverage                # npm: npm test -- --coverage
                                       # yarn: yarn test --coverage

# Archivo específico
pnpm exec vitest run src/path/Component.test.tsx

# Modo watch (si se solicita)
pnpm exec vitest
```

Solución de problemas: warnings de act — envolver actualizaciones de estado en `act()`. Problemas async — usar `waitFor()` en lugar de timeouts manuales. Providers faltantes — envolver el componente en los context providers necesarios.

### Flutter
```bash
# Tests unitarios + de widgets
flutter test --coverage

# Archivo específico
flutter test test/path/to_test.dart

# Tests de integración
flutter test integration_test/

# Con reporter expandido
flutter test --reporter expanded
```

Solución de problemas: fallos de golden — `flutter test --update-goldens`. Tests inestables — verificar `pumpAndSettle()` faltante o futures sin resolver. Ver cobertura: `genhtml coverage/lcov.info -o coverage/html`.

### Python

```bash
pytest --tb=short -q

# Con cobertura
pytest --tb=short -q --cov=. --cov-report=term-missing

# Archivo específico
pytest path/to/test_file.py -v
```

Flags:
- `--tb=short`: tracebacks concisos
- `-q`: output limpio (quiet)
- `--cov=.`: cobertura del directorio actual
- `--cov-report=term-missing`: muestra líneas no cubiertas

Configuración: `pyproject.toml` (sección `[tool.pytest.ini_options]`) o `pytest.ini`.

### Rust

```bash
cargo test

# Con output de tests que pasan
cargo test -- --nocapture

# Test específico
cargo test nombre_del_test
```

Por defecto `cargo test` ejecuta tests unitarios + de integración + doc-tests. Usa `--release` para tests con optimizaciones.

## Analizar Resultados

### Categorizar fallos

Cuando los tests fallan, categorizar cada fallo:

| Categoría | Señal | Acción |
|----------|--------|--------|
| **Error de compilación** | `cannot find`, `undefined`, `syntax error` | Corregir código primero |
| **Fallo de aserción** | `expected X got Y`, `assert`, `require` | Verificar lógica o actualizar test |
| **Timeout** | `context deadline exceeded`, `test timed out` | Verificar deadlocks, aumentar timeout |
| **Race condition** | `DATA RACE`, `concurrent map` | Agregar sincronización |
| **Inestable** | Pasa al reintentar | Investigar dependencias de timing |

### Verificación de cobertura

Reportar el porcentaje de cobertura. Señalar si está por debajo del 80% en paquetes de lógica de negocio:

```
Coverage: 73.2% — BELOW THRESHOLD (80%)
Low coverage:
  - internal/billing: 45.2%
  - internal/notification: 61.0%
```

### Re-ejecutar fallos

Sugerir comando para re-ejecutar solo los tests fallidos:

```bash
# Go — ejecutar test específico que falló
go test -race -run TestFailingName ./internal/pkg/...

# Vitest (cambiar `pnpm exec` por `npx` / `yarn exec` según lockfile)
pnpm exec vitest run --reporter=verbose path/to/failing.test.ts

# Flutter
flutter test test/failing_test.dart
```

## Formato de Salida

Resumir resultados en una tabla:

```
| Metric    | Result |
|-----------|--------|
| Passed    | 142    |
| Failed    | 3      |
| Skipped   | 5      |
| Coverage  | 84.1%  |
| Duration  | 12.3s  |
| Races     | 0      |
```

## Flujo de Trabajo

1. **Detectar stack** — verificar archivos marcadores
2. **Resolver comando según precedencia** — aplicar la sección "Resolución de Comandos" (Makefile → `package.json` script → fallback directo). No saltar al fallback si hay Makefile o script `test`.
3. **Ejecutar tests** — ejecutar el comando resuelto
4. **Categorizar fallos** — usar la tabla de fallos de arriba
5. **Si hay errores de compilación** — detenerse, corregir código primero, luego re-ejecutar
6. **Si hay fallos de aserción** — reportar con contexto, preguntar al usuario: "Should I fix the code or update the test?"
7. **Si cobertura < 80%** — señalar paquetes con baja cobertura, preguntar al usuario si se necesita mejorar la cobertura
8. **Reportar resultados** — mostrar la tabla de resumen

Si todos pasan: reportar solo la tabla.
Si hay fallos: reportar tabla + detalles de fallos categorizados + comandos para re-ejecutar.
