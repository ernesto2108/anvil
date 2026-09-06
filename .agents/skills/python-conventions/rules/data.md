# Reglas de Datos (NumPy, Pandas, ML)

## NumPy

1. **Dtype explícito siempre** — `np.float32` no el `np.float64` por defecto. Duplica la memoria sin beneficio en embeddings
2. **Preasignar arrays** — `np.empty((n, dim), dtype=np.float32)` y luego llenar. Nunca crecer listas y convertir
3. **Vectorizar operaciones** — `np.dot`, `np.linalg.norm`, multiplicación matricial. Nunca loops Python sobre arrays
4. **Type hints** — `NDArray[np.float32]` de `numpy.typing`. Hace el dtype explícito en las firmas
5. **Evitar copias** — usar vistas (`a[:, 0]`) cuando sea posible. Las copias duplican la memoria en arrays grandes

```python
# WRONG — Python loop, float64, list grow
embeddings = []
for text in texts:
    embeddings.append(model.encode(text))
result = np.array(embeddings)  # float64, 2x peak memory

# RIGHT — preallocate, explicit dtype, batch
result = np.empty((len(texts), dim), dtype=np.float32)
for i in range(0, len(texts), batch_size):
    batch = texts[i:i + batch_size]
    result[i:i + len(batch)] = model.encode(batch)
```

## Pandas

6. **Sin `iterrows()`** — operaciones vectorizadas siempre. `df["score"] = df["a"] * df["b"]`
7. **`loc` para asignación** — `df.loc[mask, "col"] = value` no indexación encadenada
8. **Backend PyArrow** — `pd.read_parquet("data.parquet", dtype_backend="pyarrow")` para rendimiento 2x
9. **Sin inplace=True** — patrón deprecado, crea copias ocultas de todas formas

## Procesamiento por Lotes

10. **`itertools.batched`** (3.12+) — stdlib, no se necesita chunking manual
11. **Llamadas a API por lotes** — nunca una a una. Agrupar en lotes de 32-256
12. **Inserciones en DB por lotes** — `executemany` o `COPY`, nunca loop de inserciones individuales

## Gestión de Memoria

13. **Archivos mapeados en memoria** para datasets > RAM — `np.memmap` para almacenamiento de lectura intensiva y append-friendly
14. **Expresiones generadoras** — `sum(x**2 for x in data)` no `sum([x**2 for x in data])`
15. **Eliminar referencias** — `del large_array; gc.collect()` cuando se termina con temporales grandes
16. **Evitar captura en closures** — extraer solo los datos necesarios en closures, no DataFrames enteros

```python
# WRONG — captures entire df
def make_scorer(df):
    return lambda idx: df.iloc[idx]["score"]

# RIGHT — extract only what's needed
def make_scorer(scores: NDArray[np.float64]):
    return lambda idx: float(scores[idx])
```
