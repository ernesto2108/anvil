# Scoring Rubric

## Scale (1-10)

| Score | Nivel | Significado |
|---|---|---|
| 9-10 | Excelente | Listo para produccion. Sin hallazgos criticos ni mejoras significativas. |
| 7-8 | Bueno | Pocos hallazgos menores. Seguro para mergear con ajustes opcionales. |
| 5-6 | Aceptable | Tiene mejoras importantes que conviene resolver antes de mergear. |
| 3-4 | Necesita trabajo | Hallazgos criticos o multiples mejoras. No recomendable mergear asi. |
| 1-2 | Riesgoso | Problemas graves de seguridad, correctitud o estabilidad. |

## Calculo del Score

El score se calcula restando penalizaciones desde 10:

| Tipo de hallazgo | Penalizacion |
|---|---|
| CRITICO | -2 por hallazgo |
| MEJORA | -0.5 por hallazgo |
| NOTA | -0 (informativo) |

Minimo: 1. No hay scores negativos.

## Categorias Evaluadas

Cada hallazgo pertenece a una categoria:

| Categoria | Que evalua |
|---|---|
| **Correctness** | Logica, edge cases, nil/null handling, race conditions |
| **Security** | Inyecciones, secrets expuestos, auth bypass, input validation |
| **Conventions** | Naming, estructura, patrones idiomaticos del stack |
| **Tests** | Cobertura de casos, calidad de assertions, edge cases sin test |
| **Performance** | N+1 queries, renders innecesarios, memory leaks, complejidad |
| **Dependencies** | Deps nuevas justificadas, versiones, licencias |
| **Infra** | Configs seguros, secrets en env, rollback, state management |
| **Lint** | Linter configurado, lint passing, sin warnings ignorados |
