# Python Anti-Patterns — Detection Reference

## Passive Detection

When reviewing Python code, scan for these patterns and report using the format:
`[file:line] [severity] [category] anti-pattern-name`

Only report `error` and `warning` by default. Report `suggestion` only when user asks to improve/refactor.

## Anti-Pattern Table

| Code Pattern | Anti-Pattern | Severity | Category | Fix |
|---|---|---|---|---|
| `def f(x=[])` or `def f(x={})` | mutable-default | error | bugs | Use `None` sentinel, assign in body |
| `eval(user_input)` / `exec(user_input)` | unsafe-eval | error | security | Never eval user input |
| `f"SELECT * FROM {table}"` | sql-injection | error | security | Parameterized queries: `$1` |
| `except:` (bare) | bare-except | error | reliability | Catch specific exceptions |
| `except Exception: pass` | swallowed-error | error | reliability | Log and re-raise or handle |
| `import *` | wildcard-import | error | readability | Import specific names |
| `os.system(cmd)` with user input | command-injection | error | security | `subprocess.run([...])` with list args |
| `pickle.loads(untrusted)` | insecure-deserialize | error | security | Use JSON or validated formats |
| `from typing import Union, Optional` (3.10+) | old-union-syntax | warning | modernization | `X \| Y` and `X \| None` |
| `Union[X, Y]` in annotations (3.10+) | old-union-type | warning | modernization | `X \| Y` |
| `Dict[str, int]` (3.9+) | old-generic-alias | warning | modernization | `dict[str, int]` |
| `class Foo(Generic[T]):` (3.12+) | old-generic-class | warning | modernization | `class Foo[T]:` |
| `TypeAlias = ...` (3.12+) | old-type-alias | warning | modernization | `type X = ...` |
| Pydantic `@validator` | pydantic-v1-validator | warning | deprecation | `@field_validator` + `@classmethod` |
| Pydantic `class Config:` | pydantic-v1-config | warning | deprecation | `model_config = ConfigDict(...)` |
| `asyncio.gather(*tasks)` without error handling | unstructured-concurrency | warning | reliability | `async with TaskGroup() as tg:` |
| `for idx, row in df.iterrows()` | pandas-iterrows | warning | performance | Vectorized operations |
| `df[mask]["col"] = val` | chained-assignment | warning | bugs | `df.loc[mask, "col"] = val` |
| `result = ""; for x: result += x` | string-concat-loop | warning | performance | `"".join(items)` |
| `sum([x**2 for x in data])` | list-in-aggregate | warning | memory | Generator: `sum(x**2 for x in ...)` |
| `np.array(growing_list)` | array-from-list-grow | warning | memory | Preallocate with `np.empty()` |
| No dtype in `np.array()` / `np.zeros()` | implicit-float64 | warning | memory | Explicit `dtype=np.float32` |
| `logging.info(f"msg {var}")` | fstring-logging | suggestion | observability | structlog with bound context |
| `API_KEY = "sk-..."` | hardcoded-secret | error | security | Environment variable + pydantic-settings |
| `time.sleep()` in async code | sync-in-async | error | reliability | `await asyncio.sleep()` |
| `requests.get()` in async code | sync-http-in-async | error | reliability | Use `aiohttp` or `httpx` async |
| Module-level `connect()` / heavy I/O | import-side-effect | warning | design | Lazy init with `@cache` or factory |
| `class X(ABC)` for simple interface | abc-over-protocol | suggestion | design | `Protocol` for structural typing |
| `dict[str, Any]` for structured data | untyped-dict | suggestion | types | `TypedDict` or Pydantic model |
| `inplace=True` in pandas | deprecated-inplace | suggestion | deprecation | Assign result: `df = df.drop(...)` |
| Missing `__slots__` on data-heavy class | no-slots | suggestion | memory | `@dataclass(slots=True)` |
| `requirements.txt` as main dep file | legacy-deps | suggestion | tooling | `pyproject.toml` + uv |
