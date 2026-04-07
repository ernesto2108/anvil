# Post-Implementation Checklist

After writing Python code, verify:

- [ ] `ruff check .` passes — no lint errors
- [ ] `ruff format --check .` passes — formatting correct
- [ ] `mypy --strict` passes — no type errors
- [ ] No `Union`, `Optional`, `Dict`, `List` imports — using modern syntax
- [ ] No mutable defaults — `def f(x=None)` not `def f(x=[])`
- [ ] Exceptions chained — `raise X from original_exc`
- [ ] All resources have context managers — `async with`, `with`
- [ ] For ML: dtype is `float32`, arrays preallocated, operations vectorized
- [ ] For async: TaskGroup used, timeouts in place, no sync calls in async
- [ ] Tests pass with `pytest -x --tb=short`
- [ ] No hardcoded secrets — using env vars + pydantic-settings
