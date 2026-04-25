# Checklist Post-Implementación

- [ ] `cargo fmt --check` — formato correcto
- [ ] `cargo clippy -- -D warnings` — sin errores de lint
- [ ] `cargo test --workspace` — todos los tests pasan
- [ ] Sin `.unwrap()` ni `panic!()` en código de librería
- [ ] Cada bloque `unsafe` tiene comentario `// SAFETY:`
- [ ] Cada `?` tiene `.context()` o `.with_context()`
- [ ] Sin IO bloqueante en funciones async
- [ ] `JoinSet` para tareas spawneadas, no fire-and-forget
- [ ] `select!` solo con futures seguros ante cancelación
- [ ] Sin `static mut` — usar atomics/Mutex/LazyLock
- [ ] `cargo audit` limpio — sin vulnerabilidades conocidas
- [ ] `cargo machete` limpio — sin dependencias no utilizadas
- [ ] Para crypto: comparaciones en tiempo constante, secretos redactados del Debug
