# Rust Anti-Patterns — Referencia de Detección

## Tabla de Anti-Patrones

| Patrón de código | Anti-Patrón | Severidad | Categoría | Corrección |
|---|---|---|---|---|
| `.unwrap()` en código de lib | unwrap-in-lib | error | reliability | Retornar `Result`, usar `?` |
| `panic!()` en código de lib | panic-in-lib | error | reliability | Retornar `Result` |
| `unsafe` sin `// SAFETY:` | undocumented-unsafe | error | safety | Documentar el invariante |
| `static mut` | static-mut | error | safety | `AtomicXxx`, `Mutex` o `LazyLock` |
| `#[allow(warnings)]` | blanket-allow | error | quality | Permitir solo lints específicos |
| `dbg!()` en código commiteado | debug-macro | error | quality | Usar `tracing::debug!` |
| `todo!()` en rama principal | todo-macro | error | quality | Implementar o trackear en un issue |
| Parámetro `String` donde funciona `&str` | unnecessary-owned | warning | performance | Aceptar `&str` o `impl AsRef<str>` |
| `.clone()` innecesario | unnecessary-clone | warning | performance | Tomar prestado o reestructurar |
| `.to_string()` en ruta caliente | hot-path-alloc | warning | performance | `Cow`, `&str` o preasignar |
| `Vec` creciendo en loop | vec-push-loop | warning | performance | Preasignar con `Vec::with_capacity` |
| `Box<dyn Error>` en API de librería | boxed-error-lib | warning | types | Enum con `thiserror` |
| `impl Display` faltante en tipos de error | error-no-display | warning | types | Derivar con `thiserror` |
| `extern "C"` desnudo (edition 2024) | bare-extern | error | edition | `unsafe extern "C"` |
| `std::env::set_var` sin unsafe (2024) | unsafe-env-set | error | edition | Envolver en `unsafe {}` |
| Bloque `unsafe` grande | wide-unsafe | warning | safety | Minimizar alcance por operación |
| `#[async_trait]` en Rust 1.75+ | legacy-async-trait | suggestion | modernization | `async fn` nativo en traits |
| `lazy_static!` en Rust 1.80+ | legacy-lazy-static | suggestion | modernization | `std::sync::LazyLock` |
| `once_cell` en Rust 1.80+ | legacy-once-cell | suggestion | modernization | `std::sync::LazyLock`/`LazyCell` |
| `expected == received` para secretos | timing-attack | error | security | `subtle::ConstantTimeEq` |
| Secreto en output de `Debug` | secret-in-debug | warning | security | Impl `Debug` personalizado con redacción |
| `println!` en código de lib/servidor | print-in-lib | warning | quality | Usar `tracing` |
| Deps con wildcard `*` | wildcard-dep | error | supply-chain | Versión exacta o con rango |
| `Cargo.lock` faltante en crate binario | no-lockfile | warning | reproducibility | Commitear `Cargo.lock` |
| Cast con `as` sin verificación | unchecked-cast | warning | correctness | `try_into()` o lints de cast de clippy |
| IO bloqueante en contexto async | blocking-in-async | error | reliability | `tokio::fs`, `spawn_blocking` |
| `select!` con future no seguro ante cancelación | cancel-unsafe-select | error | reliability | Usar métodos seguros ante cancelación |
| `tokio::spawn` fire-and-forget | untracked-spawn | warning | reliability | Usar `JoinSet` |
