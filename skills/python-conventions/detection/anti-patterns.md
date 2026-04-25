# Python Anti-Patterns — Referencia de Detección

## Detección Pasiva

Al revisar código Python, busca estos patrones y reporta usando el formato:
`[file:line] [severity] [category] anti-pattern-name`

Por defecto, reporta solo `error` y `warning`. Reporta `suggestion` únicamente cuando el usuario pide mejorar/refactorizar.

## Tabla de Anti-Patrones

| Patrón de código | Anti-Patrón | Severidad | Categoría | Corrección |
|---|---|---|---|---|
| `def f(x=[])` or `def f(x={})` | mutable-default | error | bugs | Usar centinela `None`, asignar en el cuerpo |
| `eval(user_input)` / `exec(user_input)` | unsafe-eval | error | security | Nunca evaluar input del usuario |
| `f"SELECT * FROM {table}"` | sql-injection | error | security | Queries parametrizados: `$1` |
| `except:` (bare) | bare-except | error | reliability | Capturar excepciones específicas |
| `except Exception: pass` | swallowed-error | error | reliability | Loguear y re-lanzar o manejar |
| `import *` | wildcard-import | error | readability | Importar nombres específicos |
| `os.system(cmd)` con input del usuario | command-injection | error | security | `subprocess.run([...])` con args de lista |
| `pickle.loads(untrusted)` | insecure-deserialize | error | security | Usar JSON o formatos validados |
| `from typing import Union, Optional` (3.10+) | old-union-syntax | warning | modernization | `X \| Y` y `X \| None` |
| `Union[X, Y]` en anotaciones (3.10+) | old-union-type | warning | modernization | `X \| Y` |
| `Dict[str, int]` (3.9+) | old-generic-alias | warning | modernization | `dict[str, int]` |
| `class Foo(Generic[T]):` (3.12+) | old-generic-class | warning | modernization | `class Foo[T]:` |
| `TypeAlias = ...` (3.12+) | old-type-alias | warning | modernization | `type X = ...` |
| Pydantic `@validator` | pydantic-v1-validator | warning | deprecation | `@field_validator` + `@classmethod` |
| Pydantic `class Config:` | pydantic-v1-config | warning | deprecation | `model_config = ConfigDict(...)` |
| `asyncio.gather(*tasks)` sin manejo de errores | unstructured-concurrency | warning | reliability | `async with TaskGroup() as tg:` |
| `for idx, row in df.iterrows()` | pandas-iterrows | warning | performance | Operaciones vectorizadas |
| `df[mask]["col"] = val` | chained-assignment | warning | bugs | `df.loc[mask, "col"] = val` |
| `result = ""; for x: result += x` | string-concat-loop | warning | performance | `"".join(items)` |
| `sum([x**2 for x in data])` | list-in-aggregate | warning | memory | Generador: `sum(x**2 for x in ...)` |
| `np.array(growing_list)` | array-from-list-grow | warning | memory | Preasignar con `np.empty()` |
| Sin dtype en `np.array()` / `np.zeros()` | implicit-float64 | warning | memory | `dtype=np.float32` explícito |
| `logging.info(f"msg {var}")` | fstring-logging | suggestion | observability | structlog con contexto vinculado |
| `API_KEY = "sk-..."` | hardcoded-secret | error | security | Variable de entorno + pydantic-settings |
| `time.sleep()` en código async | sync-in-async | error | reliability | `await asyncio.sleep()` |
| `requests.get()` en código async | sync-http-in-async | error | reliability | Usar `aiohttp` o `httpx` async |
| `connect()` a nivel de módulo / I/O pesado | import-side-effect | warning | design | Inicialización lazy con `@cache` o factory |
| `class X(ABC)` para interfaz simple | abc-over-protocol | suggestion | design | `Protocol` para tipado estructural |
| `dict[str, Any]` para datos estructurados | untyped-dict | suggestion | types | `TypedDict` o modelo Pydantic |
| `inplace=True` en pandas | deprecated-inplace | suggestion | deprecation | Asignar resultado: `df = df.drop(...)` |
| `__slots__` faltante en clase con muchos datos | no-slots | suggestion | memory | `@dataclass(slots=True)` |
| `requirements.txt` como archivo principal de deps | legacy-deps | suggestion | tooling | `pyproject.toml` + uv |
