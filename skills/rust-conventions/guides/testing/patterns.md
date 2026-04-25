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

## Mocking Basado en Traits

```rust
// RIGHT — trait at boundary, mock in tests
trait HttpClient: Send + Sync {
    async fn get(&self, url: &str) -> Result<Response>;
}

#[cfg(test)]
struct MockClient {
    responses: std::collections::HashMap<String, Response>,
}

#[cfg(test)]
impl HttpClient for MockClient {
    async fn get(&self, url: &str) -> Result<Response> {
        self.responses.get(url).cloned()
            .ok_or_else(|| anyhow!("not found"))
    }
}
```

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

- `proptest` 1.x — testing basado en propiedades
- `criterion` 0.5.x — benchmarking
- `insta` 1.x — snapshot testing
- `wiremock` 0.6.x — servidor HTTP mock
- `rstest` 0.22.x — fixtures y parametrize
- `fake` 3.x — generación de datos falsos
- `assert_cmd` 2.x + `predicates` 3.x — testing de CLI
