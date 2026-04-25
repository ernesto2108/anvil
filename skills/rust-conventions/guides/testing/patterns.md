# Patrones de Testing

## Tests Unitarios

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parse_valid_input() {
        assert_eq!(parse("42").unwrap(), 42);
    }

    #[test]
    #[should_panic(expected = "invalid digit")]
    fn parse_invalid_panics() {
        parse("not_a_number").unwrap();
    }

    #[test]
    fn returns_result() -> Result<(), AppError> {
        let val = parse("42")?;
        assert_eq!(val, 42);
        Ok(())
    }
}
```

## Tests de Integración

```
tests/
  common/mod.rs           # shared helpers
  integration_test.rs
```

```rust
// tests/common/mod.rs
pub fn setup_test_db() -> TestDb { /* ... */ }

// tests/integration_test.rs
mod common;

#[tokio::test]
async fn test_api() {
    let app = common::spawn_app().await;
    let resp = app.client.get("/health").send().await.unwrap();
    assert_eq!(resp.status(), 200);
}
```

## Mocking Basado en Traits con mockall (OBLIGATORIO)

**Nunca escribir mocks manuales.** Los mocks manuales pueden divergir del trait real — el test pasa verde pero el código está roto. Usar `mockall` que genera mocks desde traits via proc macros, garantizando que siempre coincidan.

### Setup

```toml
# Cargo.toml
[dev-dependencies]
mockall = "0.13"
```

### Uso básico

```rust
use mockall::automock;

#[automock]
trait HttpClient: Send + Sync {
    async fn get(&self, url: &str) -> Result<Response>;
    async fn post(&self, url: &str, body: &[u8]) -> Result<Response>;
}

#[cfg(test)]
mod tests {
    use super::*;
    use mockall::predicate::*;

    #[tokio::test]
    async fn fetches_user_by_id() {
        let mut mock = MockHttpClient::new();
        mock.expect_get()
            .with(eq("https://api.example.com/users/1"))
            .returning(|_| Ok(Response::new(200, b"{\"id\":\"1\"}")))
            .times(1);

        let svc = UserService::new(mock);
        let user = svc.get_user("1").await.unwrap();
        assert_eq!(user.id, "1");
    }

    #[tokio::test]
    async fn handles_http_error() {
        let mut mock = MockHttpClient::new();
        mock.expect_get()
            .returning(|_| Err(anyhow!("connection refused")));

        let svc = UserService::new(mock);
        assert!(svc.get_user("1").await.is_err());
    }
}
```

### Reglas

- `#[automock]` en cada trait que se vaya a mockear — genera `MockTraitName` automáticamente
- Si el trait cambia, el mock se regenera en compilación — cualquier test desactualizado falla al compilar
- Usar `mockall::predicate::*` para validar argumentos (`eq`, `str::contains`, `function`)
- `.times(1)` / `.times(n)` para verificar cantidad de llamadas
- Soporta `async_trait`, genéricos y métodos estáticos
- **Si mockall no compila o falla** — NO recurrir a mocks manuales. Reportar al orquestador

## Testing Basado en Propiedades (proptest)

```rust
use proptest::prelude::*;

proptest! {
    #[test]
    fn roundtrip_serialization(input in "\\PC{1,100}") {
        let encoded = encode(&input);
        let decoded = decode(&encoded).unwrap();
        assert_eq!(input, decoded);
    }

    #[test]
    fn sort_is_idempotent(mut vec in prop::collection::vec(any::<i32>(), 0..100)) {
        vec.sort();
        let sorted = vec.clone();
        vec.sort();
        assert_eq!(vec, sorted);
    }
}
```

## Benchmarking (criterion)

```rust
// benches/my_bench.rs
use criterion::{criterion_group, criterion_main, Criterion, black_box};

fn bench_parse(c: &mut Criterion) {
    let input = "complex input";
    c.bench_function("parse", |b| b.iter(|| parse(black_box(input))));
}

criterion_group!(benches, bench_parse);
criterion_main!(benches);
```

## Snapshot Testing (insta)

```rust
use insta::{assert_snapshot, assert_json_snapshot};

#[test]
fn test_output() {
    assert_snapshot!(render_report(&test_data()));
}

#[test]
fn test_json() {
    assert_json_snapshot!(build_response(&test_input()));
}
```

## Crates Clave

- `mockall` 0.13.x — generación de mocks desde traits (OBLIGATORIO para mocking)
- `proptest` 1.x — testing basado en propiedades
- `criterion` 0.5.x — benchmarking
- `insta` 1.x — snapshot testing
- `wiremock` 0.6.x — servidor HTTP mock
- `rstest` 0.22.x — fixtures y parametrize
- `fake` 3.x — generación de datos falsos
- `assert_cmd` 2.x + `predicates` 3.x — testing de CLI
