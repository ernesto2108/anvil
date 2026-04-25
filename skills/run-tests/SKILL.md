---
name: run-tests
description: Ejecutar tests del proyecto con detector de race conditions y cobertura. Auto-detecta el stack (Go, React, Flutter). Usar cuando el usuario diga "run tests", "test this", "check coverage", "run vitest", "go test", "flutter test", o después de implementar código que necesita verificación.
---

# Run Tests

## Auto-Detección

Detectar el stack verificando los archivos marcadores en la raíz del proyecto:

| Archivo | Stack | Comando |
|------|-------|---------|
| `go.mod` | Go | `go test ./... -race -cover -count=1` |
| `package.json` | Node/React | `<pm> exec vitest run --coverage` o `<pm> test -- --coverage` (detectar `<pm>` según CLAUDE.md — preferir `pnpm`) |
| `pubspec.yaml` | Flutter | `flutter test --coverage` |

Si se detectan múltiples stacks, ejecutar los tests de cada stack por separado.

Para Node/React: verificar `package.json` para el test runner — preferir `vitest` si está configurado, caer a `jest`, luego `<pm> test`. Detectar el package manager desde el lockfile según CLAUDE.md (`pnpm-lock.yaml` → pnpm, `yarn.lock` → yarn, `package-lock.json` → npm) y usarlo de forma consistente.

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
2. **Verificar comandos específicos del proyecto** — leer scripts de `package.json` o `Makefile` para comandos de test personalizados
3. **Ejecutar tests** — ejecutar el comando apropiado para el stack detectado
4. **Categorizar fallos** — usar la tabla de fallos de arriba
5. **Si hay errores de compilación** — detenerse, corregir código primero, luego re-ejecutar
6. **Si hay fallos de aserción** — reportar con contexto, preguntar al usuario: "Should I fix the code or update the test?"
7. **Si cobertura < 80%** — señalar paquetes con baja cobertura, preguntar al usuario si se necesita mejorar la cobertura
8. **Reportar resultados** — mostrar la tabla de resumen

Si todos pasan: reportar solo la tabla.
Si hay fallos: reportar tabla + detalles de fallos categorizados + comandos para re-ejecutar.
