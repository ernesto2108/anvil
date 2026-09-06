# Reglas de Arquitectura (Independientes de la Arquitectura)

Estas reglas aplican sin importar si usas hexagonal, clean, layered, u otra arquitectura:

1. **Imports direccionales** — las dependencias apuntan hacia adentro. Nunca importar transport/infraestructura desde el dominio
2. **Interfaces definidas por el consumidor** — el paquete que USA la interfaz la define, no el que la implementa
3. **Una responsabilidad por archivo** — un archivo debe tener una sola razón para cambiar
4. **DTOs separan transport del dominio** — nunca filtrar tipos de dominio hacia handlers HTTP/gRPC o queries de DB
5. **Sin imports circulares** — si necesitas importar el paquete A desde B y B desde A, extraer un paquete de interfaz compartida
6. **Límites de paquete = límites de API** — mantener las APIs de paquete pequeñas, los tipos no exportados grandes
7. **Inyección de dependencias vía constructores** — pasar dependencias explícitamente, nunca alcanzar estado global
8. **La validación pertenece al dominio, no a los servicios** — las entidades de entrada deben tener métodos `Validate()`. Los servicios llaman `entity.Validate()`, nunca validan campos ellos mismos. Flujo: `Handler (binding tags) → DTO.ToBusiness() → Entity.Validate() → Service (business logic only)`. Ver `examples/good-patterns.md` y `examples/bad-patterns.md` para patrones de validación
9. **Nombres de param/query HTTP en constantes de dto** — los path params (`g.Param("id")`) y query params (`g.Query("status")`) deben usar constantes nombradas del `dto/constants.go` del handler, nunca strings inline. La validación de path params URL (TrimSpace + verificación de vacío) pertenece al handler, no a la capa de aplicación. Los métodos de aplicación que reciben IDs string ya validados no deben re-validarlos
10. **Los aggregates de dominio usan composición** — cuando un servicio necesita retornar una entidad con sus datos relacionados, crear un nuevo struct que embeba o componga las entidades existentes. Nunca aplanar o duplicar campos. Ejemplo: `OrderDetail{Order, []Item, ShippingInfo}` en lugar de un struct monolítico que repita todos sus campos
11. **El sufijo "DTO" pertenece solo en los límites de transport** — usar `DTO` en `infrastructure/input/http/dto/` (respuestas HTTP) y `infrastructure/output/persistencia/*/dto/` (structs de scan SQL). Las entidades de dominio usan nombres descriptivos (`OrderDetail`, `AccountSummary`, `ProductFilter`) — nunca `OrderDTO` en el paquete `entities/`
