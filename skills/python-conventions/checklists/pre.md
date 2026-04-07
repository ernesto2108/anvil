# Pre-Implementation Checklist

Before writing Python code, verify:

- [ ] Python version target — using 3.12+ features? (`type`, `X | Y`, `batched`)
- [ ] Type hints planned — return types, parameters, Protocol for interfaces
- [ ] Pydantic model for any external input — API, config, file data
- [ ] Test strategy — what fixtures, what to mock (boundary only)
- [ ] Dependencies — is there a stdlib solution? (`itertools.batched`, `tomllib`, `asyncio.TaskGroup`)
- [ ] For ML/data — dtype explicit (`float32`), batch size defined, memory strategy for large data
- [ ] For async — using TaskGroup? Timeouts defined? Concurrency limits?
