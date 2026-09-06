# Bases de Datos Vectoriales

## Cuándo Usar

RAG (Retrieval-Augmented Generation), búsqueda semántica, recomendaciones por similaridad, deduplicación de contenido, clasificación zero-shot. NO para búsqueda por texto exacto — usar search engines (Elasticsearch/Meilisearch) para eso.

## Conceptos Clave

- **Embedding**: vector de floats (768, 1536, o 3072 dimensiones) que representa significado semántico
- **Similarity metrics**: cosine (textos), L2/euclidean (imágenes), dot product (vectores normalizados)
- **ANN (Approximate Nearest Neighbor)**: algoritmos como HNSW, IVF — tradeoff recall vs velocidad
- **Metadata filtering**: filtrar por atributos antes/durante la búsqueda vectorial
- **Chunking strategy**: cómo dividir documentos antes de embedir — crítico para calidad de RAG

## Opciones y Cuándo Elegir

| Motor | Cuándo elegir | Hosting | Notas |
|---|---|---|---|
| **pgvector** | Ya tienes Postgres, <1M vectores, queries SQL mixtas | Extensión en tu Postgres | Sin servicio extra. SQL normal + operadores vectoriales |
| **Qdrant** | Opción por defecto para producción, self-hosted viable | Self-hosted o cloud | Rust, excelente performance, API REST + gRPC |
| **Weaviate** | Knowledge graphs + vectores, búsqueda híbrida | Self-hosted o cloud | Graph capabilities, módulos de vectorización integrados |
| **Pinecone** | Zero-ops, equipo pequeño sin infra | Managed only | Caro a escala, pero cero mantenimiento |
| **ChromaDB** | Solo desarrollo/prototipos | Local | NO para producción |
| **Milvus** | Escala masiva (billones de vectores) | Self-hosted o Zilliz cloud | Operacionalmente complejo |

## Modelo de Datos

### Colecciones (equivalente a tablas)
```
collection: support_kb_ada002_v1
├── vector: float[1536]          # embedding del chunk
├── payload/metadata:
│   ├── source_doc_id: string    # documento original
│   ├── chunk_index: int         # posición en el documento
│   ├── text: string             # texto original del chunk
│   ├── category: string         # para filtrado
│   └── created_at: datetime
```

### Convenciones de Naming
```
{app}_{purpose}_{model_short}_{version}
# Ejemplos:
support_kb_ada002_v1          # knowledge base con ada-002
products_search_e5large_v2    # búsqueda de productos con E5-large
docs_rag_3small_v1            # RAG con text-embedding-3-small
```

**Regla crítica**: el nombre del modelo de embedding es parte del nombre de la colección. Vectores de modelos diferentes NO son comparables.

## Migraciones / Versionado

Las vector DBs NO tienen migraciones SQL. El cambio de modelo de embedding es una **migración completa**.

### Cambio de modelo de embedding (migración mayor)
```
1. Crear collection_v2 con nuevo modelo
2. Re-embedir TODOS los documentos (batch job)
3. Dual-read: buscar en v1 y v2, comparar resultados
4. Cutover: apuntar a v2
5. Eliminar v1 después de período de validación
```

### Cambio de metadata schema (migración menor)
- Agregar campo → agregar a nuevos documentos, backfill en existentes
- Cambiar tipo → crear colección nueva (no se puede alterar schema en la mayoría)
- Eliminar campo → dejar de escribir, ignorar en queries

### Tracking obligatorio
El agente DEBE documentar para cada colección:
- Modelo de embedding usado (nombre exacto + versión)
- Dimensiones del vector
- Métrica de similaridad configurada
- Estrategia de chunking (tamaño, overlap)

## Pitfalls de Producción

| # | Pitfall | Consecuencia | Prevención |
|---|---|---|---|
| 1 | Mezclar modelos de embedding en misma colección | Resultados de búsqueda sin sentido | Modelo en nombre de colección |
| 2 | No versionar el modelo de embedding | Re-embed imposible de reproducir | Documentar modelo + versión |
| 3 | Ignorar parámetros HNSW (`ef_construction`, `m`) | Mal recall o índice lento | Tuning según volumen de datos |
| 4 | Chunks muy grandes (>1000 tokens) | Pierde granularidad semántica | 200-500 tokens con overlap de 50-100 |
| 5 | Chunks muy pequeños (<50 tokens) | Pierde contexto | Mínimo 100 tokens |
| 6 | No guardar texto original junto al vector | No puedes mostrar resultados al usuario | Siempre guardar `text` en metadata |
| 7 | Embedding uno por uno | Latencia y costo innecesarios | Batch de 100-500 por llamada al modelo |

## Optimización de Rendimiento

### Índice HNSW (default en la mayoría)
- `m=16`, `ef_construction=200` — punto de partida razonable
- Más `m` = mejor recall, más memoria y build time
- `ef_search` (query time) — subir para más precisión, bajar para más velocidad

### pgvector específico
```sql
-- Crear índice HNSW
CREATE INDEX ON items USING hnsw (embedding vector_cosine_ops)
  WITH (m = 16, ef_construction = 200);

-- Ajustar probes para queries de alta precisión
SET ivfflat.probes = 10;

-- Usar índice parcial si filtras por categoría
CREATE INDEX ON items USING hnsw (embedding vector_cosine_ops)
  WHERE category = 'support';
```

### Estrategias de chunking
| Estrategia | Cuándo usar |
|---|---|
| Fixed size (tokens) | Documentos homogéneos, texto plano |
| Paragraph/section | Documentos con estructura clara (markdown, HTML) |
| Semantic chunking | Documentos largos donde el significado varía dentro de secciones |
| Sentence-level | FAQ, Q&A, definiciones cortas |

## Patrones Comunes

### RAG básico
```
1. Usuario hace pregunta
2. Embedir pregunta con mismo modelo que la colección
3. Buscar top-K vectores más similares (K=5-10)
4. Filtrar por metadata si aplica (categoría, fecha)
5. Pasar chunks recuperados como contexto al LLM
6. LLM genera respuesta citando los chunks
```

### Búsqueda híbrida (vectorial + keyword)
```
1. Búsqueda vectorial → top 20 por similaridad semántica
2. Búsqueda keyword (BM25) → top 20 por coincidencia exacta
3. Reciprocal Rank Fusion (RRF) → combinar rankings
4. Re-rank con cross-encoder → top 5 final
```

## Drivers por Stack

| Stack | pgvector | Qdrant | Pinecone |
|---|---|---|---|
| Go | `pgvector-go` + `pgx` | `qdrant/go-client` | `pinecone-io/go-pinecone` |
| TypeScript | `pgvector/drizzle` o raw SQL | `@qdrant/js-client-rest` | `@pinecone-database/pinecone` |
| Python | `pgvector` + `psycopg` | `qdrant-client` | `pinecone` |
| Rust | `pgvector` + `sqlx` | `qdrant-client` | REST API |
