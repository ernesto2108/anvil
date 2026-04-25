# Checklist Pre-Implementación

- [ ] ¿Edition 2024? — `edition = "2024"` en Cargo.toml, resolver = "3"
- [ ] Estrategia de errores — thiserror para lib, anyhow para app, miette para CLI
- [ ] Plan de ownership — ¿quién posee los datos? Tomar prestado donde sea posible, `Arc` para async compartido
- [ ] ¿Async o sync? — si async, runtime tokio configurado, sin bloqueo en contexto async
- [ ] `#![forbid(unsafe_code)]` a nivel de crate — permitir solo en módulos auditados
- [ ] Deps del workspace — centralizadas en `[workspace.dependencies]`
- [ ] Para CLI — clap derive, tracing para logging, indicatif para progreso
- [ ] Para blockchain — patrones anchor/alloy, tiempo constante para secretos, no_std si es necesario
- [ ] Estrategia de tests — unit + integration, proptest para invariantes, criterion para perf
