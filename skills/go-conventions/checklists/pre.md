# Pre-Implementation Checklist

Before writing Go code, verify:

- [ ] Package placement follows existing project structure
- [ ] Interfaces defined by consumer, not producer
- [ ] Error wrapping includes operation context
- [ ] Context passed as first parameter
- [ ] External calls have timeouts
- [ ] Concurrent access is protected
- [ ] Tests are table-driven with subtests
- [ ] No circular imports introduced
- [ ] Logging uses structured key-value pairs (no string concatenation)
- [ ] No sensitive data logged (passwords, tokens, PII)
- [ ] SQL queries use parameterized placeholders, not string interpolation
