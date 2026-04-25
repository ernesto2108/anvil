# Guía de Testing en Go

> **Dispatcher:** Cargar SOLO los archivos relevantes para tu alcance de test. Cada archivo es ~3-5KB.

## Tabla de Rutas

| Alcance | Archivo |
|---|---|
| Estructura de archivos de test + tests con tabla | `guides/testing/structure-tables.md` |
| Tests de HTTP handlers | `guides/testing/http-handlers.md` |
| Tests de Repositorio / DB | `guides/testing/repositories.md` |
| Mocks manuales + helpers de test | `guides/testing/helpers-mocking.md` |
| Fixtures + configuración de tests de integración | `guides/testing/fixtures-integration.md` |
| Objetivos de cobertura + benchmarks | `guides/testing/coverage-benchmarks.md` |

## Siempre cargar

- `guides/testing/structure-tables.md` — naming, sufijo de paquete, patrón table-driven (requerido para TODOS los tests de Go)
- `guides/testing/helpers-mocking.md` — mocks manuales, `t.Helper()`, sin mockery/gomock

## Cargar cuando sea relevante

- `guides/testing/http-handlers.md` — si se testean HTTP handlers o middleware
- `guides/testing/repositories.md` — si se testean queries de DB o repositorios
- `guides/testing/fixtures-integration.md` — si se escriben tests de integración o se usan fixtures testdata
- `guides/testing/coverage-benchmarks.md` — si la cobertura es parte de la tarea

## Reglas clave (resumen)

- Paquete: usar `package foo_test` (caja negra) a menos que se testee lógica no exportada
- Naming: `Test_FunctionName` o `Test_Type_Method` (guion bajo después de `Test`)
- Table-driven por defecto — cualquier función con >1 escenario recibe un loop `tests []struct{name string, ...}`
- Assertions: stdlib (`t.Fatalf`/`t.Errorf`) O testify (`require`/`assert`) — coincidir con lo que el proyecto ya usa, nunca mezclar
- `require` (testify) / `t.Fatalf` (stdlib) para verificaciones fatales; `assert` / `t.Errorf` para no-fatales
- Sin mockery, sin gomock — todos los mocks se escriben manualmente como structs que implementan la interfaz
- `t.Helper()` en cada función helper de test
- Las assertions de errores verifican el mensaje/tipo, no solo `err != nil`
