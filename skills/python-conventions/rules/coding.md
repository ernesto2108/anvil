# Coding Rules

## Modern Syntax (Python 3.12+)

1. **PEP 695 type aliases** — `type Vector = list[float]` not `Vector: TypeAlias = list[float]`
2. **PEP 604 unions** — `int | str` not `Union[int, str]`, `str | None` not `Optional[str]`
3. **Generic classes** — `class Stack[T]:` not `class Stack(Generic[T]):`
4. **Match statements** — use `match/case` for structural pattern matching instead of if/elif chains
5. **F-strings always** — never `%` formatting or `.format()`, f-strings support nested quotes in 3.12+
6. **`itertools.batched`** — use stdlib batching (3.12+), don't write manual chunking

## Type Hints

7. **Type hints on all public functions** — parameters and return types. Private helpers can omit if obvious
8. **`from __future__ import annotations`** — for forward references and deferred evaluation
9. **Protocol over ABC** — structural typing, no inheritance needed. Use `@runtime_checkable` when needed
10. **TypedDict for unstructured data** — not `dict[str, Any]`. Use `NotRequired` for optional keys (3.11+)
11. **`NDArray[np.float32]`** — always explicit dtype for numpy arrays
12. **TYPE_CHECKING guard** — heavy imports (numpy, pandas, torch) only for annotations go in `if TYPE_CHECKING:`

## Naming

13. **snake_case** for functions, methods, variables, modules
14. **PascalCase** for classes, type aliases, Protocols
15. **UPPER_SNAKE** for module-level constants
16. **Leading underscore** for private (`_helper`), never dunder (`__mangled`) unless needed
17. **No single-letter names** except `i/j/k` for indices, `x/y/z` for coordinates, `n` for counts, `T` for generics

## Error Handling

18. **Domain exception hierarchy** — `class AppError(Exception)` as base, specific subclasses
19. **Chain exceptions** — `raise NewError("msg") from original_exc`, never lose traceback
20. **Exception groups** (3.11+) — use `ExceptionGroup` and `except*` for concurrent error aggregation
21. **No bare except** — always catch specific exceptions. `except Exception` only with re-raise
22. **Structured logging** — `structlog` with bound context, not f-strings in `logging.info()`

## Data Classes

23. **`@dataclass(slots=True, frozen=True)`** for value objects — saves memory, prevents mutation
24. **Pydantic `BaseModel`** at boundaries — API inputs, config, external data validation
25. **Pydantic v2 patterns** — `field_validator` + `@classmethod`, `ConfigDict`, not v1 `Config` class
