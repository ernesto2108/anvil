# Patrones de pytest

## Fixtures

```python
import pytest

@pytest.fixture
def embedder() -> OpenAIEmbedder:
    return OpenAIEmbedder(api_key="test-key")

@pytest.fixture
def sample_texts() -> list[str]:
    return ["hello world", "goodbye world"]
```

## Parametrize

```python
@pytest.mark.parametrize("input_texts,expected_count", [
    (["hello"], 1),
    (["hello", "world"], 2),
    ([], 0),
])
def test_embed_returns_correct_count(
    embedder: OpenAIEmbedder,
    input_texts: list[str],
    expected_count: int,
) -> None:
    result = embedder.embed(input_texts)
    assert len(result) == expected_count
```

## Fixtures de Fábrica

```python
@pytest.fixture
def make_embedding():
    def _make(dim: int = 768, value: float = 0.0) -> list[float]:
        return [value] * dim
    return _make

def test_cosine_similarity(make_embedding) -> None:
    a = make_embedding(dim=3, value=1.0)
    b = make_embedding(dim=3, value=0.0)
    assert cosine_similarity(a, b) == pytest.approx(0.0)
```

## Mock en la Frontera

```python
# WRONG — mocking implementation details
def test_embed(mocker):
    mocker.patch("my_package.openai.Client")
    mocker.patch("my_package.openai.Client.embeddings.create")
    # brittle, tests implementation not behavior

# RIGHT — fake implementation via Protocol
class FakeEmbedder:
    def embed(self, texts: list[str]) -> list[list[float]]:
        return [[0.1] * 768 for _ in texts]

@pytest.fixture
def fake_provider() -> EmbeddingProvider:
    return FakeEmbedder()

def test_similarity_search(fake_provider: EmbeddingProvider) -> None:
    service = SimilarityService(provider=fake_provider)
    results = service.search("hello", top_k=5)
    assert len(results) <= 5
```

## Tests Async

```python
# pyproject.toml: asyncio_mode = "auto"
async def test_fetch_embeddings(mock_client: AsyncClient) -> None:
    result = await mock_client.post("/embed", json={"texts": ["hello"]})
    assert result.status_code == 200
    assert "embeddings" in result.json()
```

## Archivos y Directorios Temporales

```python
def test_save_embeddings(tmp_path: Path) -> None:
    store = EmbeddingStore(tmp_path / "embeddings.npy")
    store.save(np.array([[1.0, 2.0]], dtype=np.float32))
    loaded = store.load()
    np.testing.assert_array_equal(loaded, [[1.0, 2.0]])
```

## Reglas Clave

- **Fixtures sobre setUp/tearDown** — composables, dependencias explícitas
- **Parametrize para variaciones** — una función de test, muchos casos
- **Mock en las fronteras, no en los internos** — usar fakes de Protocol
- **`pytest.approx`** para punto flotante — nunca `==` en floats
- **`tmp_path`** para tests de archivos — limpiado automáticamente por pytest
- **Sin `time.sleep` en tests** — usar eventos, colas o `asyncio.wait_for`
