# Reglas de Arquitectura

## Estructura del Proyecto

1. **Layout src** — `src/my_package/` no `my_package/` plano. Previene confusión de imports con los tests
2. **`pyproject.toml` como única fuente** — sin setup.py, setup.cfg, requirements.txt. Toda la configuración en un archivo
3. **Marcador `py.typed`** — PEP 561, incluir en `src/my_package/` para paquetes tipados
4. **Un módulo = una responsabilidad** — `models.py`, `services/embedding.py`, no `utils.py` como catch-all

## Cadena de Herramientas

5. **uv** para gestión de paquetes — reemplaza pip, virtualenv, pip-tools. Lockfile incorporado, 10-100x más rápido
6. **Ruff** para lint + formato — reemplaza flake8 + isort + black + pyupgrade. Herramienta única, configuración única
7. **mypy en modo strict** — `strict = true` en pyproject.toml. Detecta errores de tipos en tiempo de análisis

## Imports

8. **Imports absolutos** — `from my_package.models import User` no `from .models import User` (excepto dentro de sub-paquetes)
9. **Sin `import *`** — nunca. Contamina el namespace, rompe el análisis estático
10. **Guard `TYPE_CHECKING`** — imports pesados usados solo para anotaciones: numpy, pandas, torch

```python
from __future__ import annotations
from typing import TYPE_CHECKING

if TYPE_CHECKING:
    import numpy as np
    from numpy.typing import NDArray
```

## Inyección de Dependencias

11. **Inyección por constructor** — pasar dependencias explícitamente, nunca importar y llamar globales

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

12. **Protocol para dependencias** — definir qué necesitas, no qué existe

```python
from typing import Protocol

class EmbeddingProvider(Protocol):
    def embed(self, texts: list[str]) -> list[list[float]]: ...
```

## Pydantic en las Fronteras

13. **Pydantic v2 BaseModel** para todos los datos externos — requests/responses de API, config, parseo de archivos
14. **`ConfigDict(strict=True, frozen=True)`** — inmutable, coerción de tipos estricta
15. **`field_validator` + `@classmethod`** — estilo v2, no `@validator` de v1
16. **`pydantic-settings`** para variables de entorno — `BaseSettings` con soporte `.env`, falla rápido en vars faltantes

```python
from pydantic import BaseModel, Field, ConfigDict, field_validator

class SearchRequest(BaseModel):
    model_config = ConfigDict(strict=True, frozen=True)

    query: str = Field(min_length=1, max_length=1000)
    top_k: int = Field(ge=1, le=100, default=10)
```

## Configuración de Ruff

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
