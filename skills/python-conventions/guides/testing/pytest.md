# pytest Patterns

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

## Factory Fixtures

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

## Mock at the Boundary

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

## Async Testing

```python
# pyproject.toml: asyncio_mode = "auto"
async def test_fetch_embeddings(mock_client: AsyncClient) -> None:
    result = await mock_client.post("/embed", json={"texts": ["hello"]})
    assert result.status_code == 200
    assert "embeddings" in result.json()
```

## Temporary Files and Directories

```python
def test_save_embeddings(tmp_path: Path) -> None:
    store = EmbeddingStore(tmp_path / "embeddings.npy")
    store.save(np.array([[1.0, 2.0]], dtype=np.float32))
    loaded = store.load()
    np.testing.assert_array_equal(loaded, [[1.0, 2.0]])
```

## Key Rules

- **Fixtures over setUp/tearDown** — composable, explicit dependencies
- **Parametrize for variations** — one test function, many cases
- **Mock at boundaries, not internals** — use Protocol fakes
- **`pytest.approx`** for floating point — never `==` on floats
- **`tmp_path`** for file tests — auto-cleaned by pytest
- **No `time.sleep` in tests** — use events, queues, or `asyncio.wait_for`
