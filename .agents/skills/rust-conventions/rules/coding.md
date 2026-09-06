# Reglas de Codificación

## Edición 2024 (Rust 1.85+)

1. **`unsafe_op_in_unsafe_fn` advierte** — las operaciones `unsafe` dentro de `unsafe fn` requieren un bloque `unsafe {}` explícito
2. **`static mut` prohibido** — usa `AtomicXxx`, `Mutex`, o `LazyLock` en su lugar
3. **`unsafe extern` requerido** — los bloques FFI deben ser `unsafe extern "C"` con anotaciones safe/unsafe
4. **`gen` está reservado** — usa `r#gen` si lo necesitas como identificador
5. **`set_var`/`remove_var` son unsafe** — la mutación de variables de entorno es ahora `unsafe` por seguridad en hilos
6. **Resolver v3 por defecto** — `resolver = "3"` en el `Cargo.toml` del workspace
7. **RPIT captura todos los lifetimes** — usa `+ use<>` para desactivar la captura implícita de lifetimes

```rust
// INCORRECTO — static mut
static mut COUNTER: u64 = 0;

// CORRECTO — atómico
use std::sync::atomic::{AtomicU64, Ordering};
static COUNTER: AtomicU64 = AtomicU64::new(0);

// CORRECTO — LazyLock para init complejo (1.80+)
use std::sync::LazyLock;
static CONFIG: LazyLock<Config> = LazyLock::new(|| Config::load());
```

## Manejo de Errores

8. **thiserror para librerías** — deriva `Error` para tipos de dominio, nunca expongas `anyhow`
9. **anyhow para aplicaciones** — `Result<T>` ergonómico, `.context()` para envolver errores
10. **miette para CLIs** — errores de diagnóstico enriquecidos con spans de código y texto de ayuda
11. **Nunca hacer panic en código de librería** — retorna `Result`, sin `.unwrap()`, sin `panic!()`
12. **Contexto en cada `?`** — `.context("qué falló")` o `.with_context(|| format!(...))`

```rust
// INCORRECTO — ? sin contexto
let data = std::fs::read_to_string(path)?;

// CORRECTO — envuelto con contexto
let data = std::fs::read_to_string(path)
    .with_context(|| format!("failed to read {path}"))?;
```

## Ownership y Borrowing

13. **`&str` sobre `String` en parámetros** — acepta referencias, retorna owned solo cuando sea necesario
14. **`Cow<'_, str>`** para ownership condicional — evita allocations cuando no son necesarias
15. **`impl AsRef<T>` para APIs flexibles** — acepta tanto owned como borrowed
16. **Clonar solo cuando sea necesario** — prefiere referencias, usa `Arc` para ownership compartido en async

```rust
// INCORRECTO — toma ownership innecesariamente
fn process(data: String) -> String { data.to_uppercase() }

// CORRECTO — toma prestada la entrada
fn process(data: &str) -> String { data.to_uppercase() }
```

## Nomenclatura

17. **snake_case** para funciones, métodos, variables, módulos, crates
18. **PascalCase** para tipos, traits, enums, structs
19. **SCREAMING_SNAKE** para constantes y estáticos
20. **`is_`, `has_`, `can_`** como prefijo para métodos que retornan bool
21. **`into_`, `as_`, `to_`** para conversiones (toma ownership, borrowing, costosa respectivamente)
22. **`try_`** como prefijo para operaciones fallidas que retornan `Result`

## Async Traits (1.75+)

23. **`async fn` nativo en traits** — ya no se necesita el proc macro `#[async_trait]`
24. **`-> impl Future + Send`** cuando necesitas el bound Send para trait objects
25. **`Pin<Box<dyn Future + Send>>`** cuando necesitas dyn dispatch

```rust
// INCORRECTO — crate async-trait (legado)
#[async_trait]
trait Repo { async fn find(&self, id: Uuid) -> Result<Entity>; }

// CORRECTO — nativo (1.75+)
trait Repo: Send + Sync {
    async fn find(&self, id: Uuid) -> Result<Entity>;
}
```
