# Reglas de Código

## Sintaxis Moderna (Python 3.12+)

1. **Alias de tipo PEP 695** — `type Vector = list[float]` no `Vector: TypeAlias = list[float]`
2. **Uniones PEP 604** — `int | str` no `Union[int, str]`, `str | None` no `Optional[str]`
3. **Clases genéricas** — `class Stack[T]:` no `class Stack(Generic[T]):`
4. **Sentencias match** — usar `match/case` para coincidencia estructural de patrones en lugar de cadenas if/elif
5. **F-strings siempre** — nunca formato `%` ni `.format()`, los f-strings soportan comillas anidadas en 3.12+
6. **`itertools.batched`** — usar batching de stdlib (3.12+), no escribir chunking manual

## Type Hints

7. **Type hints en todas las funciones públicas** — parámetros y tipos de retorno. Helpers privados pueden omitirlos si es obvio
8. **`from __future__ import annotations`** — para referencias adelantadas y evaluación diferida
9. **Protocol sobre ABC** — tipado estructural, sin herencia necesaria. Usar `@runtime_checkable` cuando se necesite
10. **TypedDict para datos no estructurados** — no `dict[str, Any]`. Usar `NotRequired` para claves opcionales (3.11+)
11. **`NDArray[np.float32]`** — siempre dtype explícito para arrays numpy
12. **Guard `TYPE_CHECKING`** — imports pesados (numpy, pandas, torch) solo para anotaciones van en `if TYPE_CHECKING:`

## Nomenclatura

13. **snake_case** para funciones, métodos, variables, módulos
14. **PascalCase** para clases, alias de tipo, Protocols
15. **UPPER_SNAKE** para constantes a nivel de módulo
16. **Guión bajo inicial** para privados (`_helper`), nunca dunder (`__mangled`) a menos que sea necesario
17. **Sin nombres de una sola letra** excepto `i/j/k` para índices, `x/y/z` para coordenadas, `n` para conteos, `T` para genéricos

## Manejo de Errores

18. **Jerarquía de excepciones de dominio** — `class AppError(Exception)` como base, subclases específicas
19. **Encadenar excepciones** — `raise NewError("msg") from original_exc`, nunca perder el traceback
20. **Grupos de excepciones** (3.11+) — usar `ExceptionGroup` y `except*` para agregación de errores concurrentes
21. **Sin bare except** — siempre capturar excepciones específicas. `except Exception` solo con re-lanzamiento
22. **Logging estructurado** — `structlog` con contexto vinculado, no f-strings en `logging.info()`

## Clases de Datos

23. **`@dataclass(slots=True, frozen=True)`** para objetos de valor — ahorra memoria, previene mutación
24. **Pydantic `BaseModel`** en las fronteras — inputs de API, config, validación de datos externos
25. **Patrones Pydantic v2** — `field_validator` + `@classmethod`, `ConfigDict`, no la clase `Config` de v1
