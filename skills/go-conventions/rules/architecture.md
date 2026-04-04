# Architecture Rules (Architecture-Agnostic)

These rules apply regardless of whether you use hexagonal, clean, layered, or any other architecture:

1. **Directional imports** — dependencies point inward. Never import transport/infrastructure from domain
2. **Interfaces defined by consumer** — the package that USES the interface defines it, not the package that implements it
3. **One concern per file** — a file should have one reason to change
4. **DTOs separate transport from domain** — never leak domain types into HTTP/gRPC handlers or DB queries
5. **No circular imports** — if you need to import package A from B and B from A, extract a shared interface package
6. **Package boundaries = API boundaries** — keep package APIs small, unexported types large
7. **Dependency injection via constructors** — pass dependencies explicitly, never reach into global state
8. **Validation belongs in the domain, not in services** — input entities must have `Validate()` methods. Services call `entity.Validate()`, never validate fields themselves. Flow: `Handler (binding tags) → DTO.ToBusiness() → Entity.Validate() → Service (business logic only)`. See `examples/good-patterns.md` and `examples/bad-patterns.md` for validation patterns
9. **HTTP param/query names in dto constants** — path params (`g.Param("id")`) and query params (`g.Query("status")`) must use named constants from the handler's `dto/constants.go`, never inline strings. URL path param validation (TrimSpace + empty check) belongs in the handler, not the application layer. Application methods that receive already-validated string IDs should not re-validate them
10. **Domain aggregates use composition** — when a service needs to return an entity with its related data, create a new struct that embeds or composes existing entities. Never flatten or duplicate fields. Example: `OrderDetail{Order, []Item, ShippingInfo}` not a monolith struct repeating all their fields
11. **"DTO" suffix belongs at transport boundaries only** — use `DTO` in `infrastructure/input/http/dto/` (HTTP responses) and `infrastructure/output/persistencia/*/dto/` (SQL scan structs). Domain entities use descriptive names (`OrderDetail`, `AccountSummary`, `ProductFilter`) — never `OrderDTO` in the `entities/` package
