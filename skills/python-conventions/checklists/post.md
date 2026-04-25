# Checklist Post-Implementación

Después de escribir código Python, verificar:

- [ ] `ruff check .` pasa — sin errores de lint
- [ ] `ruff format --check .` pasa — formato correcto
- [ ] `mypy --strict` pasa — sin errores de tipos
- [ ] Sin `Union`, `Optional`, `Dict`, `List` imports — usar sintaxis moderna
- [ ] Sin defaults mutables — `def f(x=None)` no `def f(x=[])`
- [ ] Excepciones encadenadas — `raise X from original_exc`
- [ ] Todos los recursos tienen context managers — `async with`, `with`
- [ ] Para ML: dtype es `float32`, arrays pre-allocados, operaciones vectorizadas
- [ ] Para async: se usa TaskGroup, timeouts establecidos, sin llamadas sync en async
- [ ] Tests pasan con `pytest -x --tb=short`
- [ ] Sin secrets hardcodeados — usar env vars + pydantic-settings
