# Architecture Rules

## Project Structure

1. **src layout** — `src/my_package/` not flat `my_package/`. Prevents import confusion with tests
2. **`pyproject.toml` as single source** — no setup.py, setup.cfg, requirements.txt. All config in one file
3. **`py.typed` marker** — PEP 561, include in `src/my_package/` for typed packages
4. **One module = one concern** — `models.py`, `services/embedding.py`, not `utils.py` catch-all

## Toolchain

5. **uv** for package management — replaces pip, virtualenv, pip-tools. Lockfile built-in, 10-100x faster
6. **Ruff** for lint + format — replaces flake8 + isort + black + pyupgrade. Single tool, single config
7. **mypy strict mode** — `strict = true` in pyproject.toml. Catches type errors at analysis time

## Imports

8. **Absolute imports** — `from my_package.models import User` not `from .models import User` (except within sub-packages)
9. **No `import *`** — ever. Pollutes namespace, breaks static analysis
10. **TYPE_CHECKING guard** — heavy imports used only for annotations: numpy, pandas, torch

```python
from __future__ import annotations
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    import numpy as np
    from numpy.typing import NDArray
```

## Dependency Injection

11. **Constructor injection** — pass dependencies explicitly, never import and call globals

```python
# WRONG
from my_package.db import connection
class UserRepo:
    def get(self, id: str) -> User:
        return connection.fetch(...)  # global state

# RIGHT
class UserRepo:
    def __init__(self, pool: AsyncPool) -> None:
        self._pool = pool
```

12. **Protocol for dependencies** — define what you need, not what exists

```python
from typing import Protocol

class EmbeddingProvider(Protocol):
    def embed(self, texts: list[str]) -> list[list[float]]: ...
```

## Pydantic at Boundaries

13. **Pydantic v2 BaseModel** for all external data — API requests/responses, config, file parsing
14. **`ConfigDict(strict=True, frozen=True)`** — immutable, strict type coercion
15. **`field_validator` + `@classmethod`** — v2 style, not v1 `@validator`
16. **`pydantic-settings`** for env vars — `BaseSettings` with `.env` support, fails fast on missing vars

```python
from pydantic import BaseModel, Field, ConfigDict, field_validator

class SearchRequest(BaseModel):
    model_config = ConfigDict(strict=True, frozen=True)

    query: str = Field(min_length=1, max_length=1000)
    top_k: int = Field(ge=1, le=100, default=10)
```

## Ruff Configuration

```toml
[tool.ruff]
target-version = "py312"
line-length = 88
src = ["src"]

[tool.ruff.lint]
select = [
    "E", "W", "F", "I", "N", "UP", "B", "A",
    "SIM", "TCH", "RUF", "S", "PT", "C4",
    "DTZ", "T20", "PIE", "PL", "PERF", "FURB",
]
ignore = ["E501", "PLR0913"]

[tool.ruff.lint.per-file-ignores]
"tests/**" = ["S101"]
```
