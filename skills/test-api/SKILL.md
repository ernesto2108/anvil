---
name: test-api
description: Test and validate API endpoints for contract compliance. Use when user says "test the API", "check endpoint", "validate response", "API contract", "curl this endpoint", or verifying HTTP status codes and response schemas.
---

Test and validate external or internal API endpoints to ensure they conform to expected contracts.

Capabilities:
- Use `curl` for basic requests
- Validate JSON structure and field types
- Check HTTP status codes and headers
- Test different auth scenarios (valid/invalid token)

Rules:
- NEVER send sensitive production data (keys/PII) in plain text
- Use placeholders for API keys
- Prefer non-destructive methods (GET/HEAD) unless necessary
- For POST/PUT/PATCH, use a development/mock environment
