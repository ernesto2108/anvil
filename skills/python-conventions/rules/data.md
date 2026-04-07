# Data Rules (NumPy, Pandas, ML)

## NumPy

1. **Explicit dtype always** — `np.float32` not default `np.float64`. Doubles memory for no benefit in embeddings
2. **Preallocate arrays** — `np.empty((n, dim), dtype=np.float32)` then fill. Never grow lists and convert
3. **Vectorize operations** — `np.dot`, `np.linalg.norm`, matrix multiply. Never Python loops over arrays
4. **Type hints** — `NDArray[np.float32]` from `numpy.typing`. Makes dtype explicit in signatures
5. **Avoid copies** — use views (`a[:, 0]`) when possible. Copies double memory on large arrays

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

6. **No `iterrows()`** — vectorized operations always. `df["score"] = df["a"] * df["b"]`
7. **`loc` for assignment** — `df.loc[mask, "col"] = value` not chained indexing
8. **PyArrow backend** — `pd.read_parquet("data.parquet", dtype_backend="pyarrow")` for 2x performance
9. **No inplace=True** — deprecated pattern, creates hidden copies anyway

## Batch Processing

10. **`itertools.batched`** (3.12+) — stdlib, no manual chunking needed
11. **Batch API calls** — never one-by-one. Group into batches of 32-256
12. **Batch DB inserts** — `executemany` or `COPY`, never loop of single inserts

## Memory Management

13. **Memory-mapped files** for datasets > RAM — `np.memmap` for read-heavy, append-friendly storage
14. **Generator expressions** — `sum(x**2 for x in data)` not `sum([x**2 for x in data])`
15. **Delete references** — `del large_array; gc.collect()` when done with large temporaries
16. **Avoid closure capture** — extract only needed data into closures, not entire DataFrames

```python
# WRONG — captures entire df
def make_scorer(df):
    return lambda idx: df.iloc[idx]["score"]

# RIGHT — extract only what's needed
def make_scorer(scores: NDArray[np.float64]):
    return lambda idx: float(scores[idx])
```
