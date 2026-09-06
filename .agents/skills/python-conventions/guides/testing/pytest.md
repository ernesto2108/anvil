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

## Mock en la Frontera con create_autospec (OBLIGATORIO)

**Nunca escribir fakes/mocks manuales.** Las clases fake pueden divergir del Protocol real — el test pasa pero el código está roto. Usar `create_autospec` que genera un mock desde la clase/Protocol, garantizando que las firmas coincidan en runtime. Complementar con `mypy --strict` para verificación estática.

```python
# WRONG — fake manual que puede divergir del Protocol
class FakeEmbedder:
    def embed(self, texts: list[str]) -> list[list[float]]:
        return [[0.1] * 768 for _ in texts]

# WRONG — mocker.patch en internos
def test_embed(mocker):
    mocker.patch("my_package.openai.Client")  # brittle

# RIGHT — create_autospec desde el Protocol
from unittest.mock import create_autospec

class EmbeddingProvider(Protocol):
    def embed(self, texts: list[str]) -> list[list[float]]: ...
    def dimension(self) -> int: ...

@pytest.fixture
def mock_provider() -> EmbeddingProvider:
    mock = create_autospec(EmbeddingProvider, instance=True)
    mock.embed.return_value = [[0.1] * 768]
    mock.dimension.return_value = 768
    return mock

def test_similarity_search(mock_provider: EmbeddingProvider) -> None:
    service = SimilarityService(provider=mock_provider)
    results = service.search("hello", top_k=5)
    assert len(results) <= 5
    mock_provider.embed.assert_called_once()

def test_autospec_catches_wrong_args(mock_provider: EmbeddingProvider) -> None:
    # Esto FALLA en runtime — autospec detecta argumento incorrecto
    mock_provider.embed("not_a_list")  # TypeError: positional arg mismatch
```

### Reglas de mocking en Python

- `create_autospec(MyProtocol, instance=True)` para toda interfaz/Protocol — nunca clases fake manuales
- Si el Protocol cambia (nuevo método, firma diferente), `create_autospec` lo refleja automáticamente
- Ejecutar `mypy --strict` en CI para complementar con verificación estática de tipos
- `mocker.patch` solo para parchear dependencias externas (APIs, filesystem) en la frontera — nunca para internos
- **Si autospec no funciona para un caso** — NO recurrir a fakes manuales. Reportar al orquestador

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
