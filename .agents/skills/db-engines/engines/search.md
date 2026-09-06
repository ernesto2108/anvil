# Search Engines

## Cuándo Usar

Búsqueda full-text con relevancia, facets/filtros combinados, autocomplete, búsqueda multilingüe. Un search engine NO reemplaza la DB principal — es un **índice secundario sincronizado**. La fuente de verdad siempre es la DB principal.

## Opciones y Cuándo Elegir

| Motor | Cuándo elegir | Notas |
|---|---|---|
| **Elasticsearch** | Queries analíticas complejas, agregaciones, logs (ELK stack) | Flexible, maduro, operacionalmente complejo. OpenSearch como fork open-source |
| **Meilisearch** | Búsqueda de productos/contenido, UX out-of-the-box | Fácil de operar, typo-tolerant por defecto, excelente para e-commerce |
| **Typesense** | Similar a Meilisearch, más estricto en tipos | Buen performance, geo-search integrado, hosting managed disponible |

## Conceptos Clave

### Mapping (Schema del Índice)
El mapping define tipos de campos. Es la decisión más importante y la más difícil de cambiar después.

| Tipo | Uso | Ejemplo |
|---|---|---|
| `text` | Analizado, búsqueda full-text | Nombre de producto, descripción |
| `keyword` | No analizado, filtros y agregaciones exactas | Status, categoría, email |
| `integer/float` | Numéricos, rangos | Precio, cantidad |
| `date` | Timestamps, rangos de fecha | created_at |
| `boolean` | Filtros binarios | is_active |
| `nested` | Objetos con queries independientes | Array de variantes con precio + color |
| `geo_point` | Coordenadas geográficas | Ubicación de tienda |

**Regla crítica**: `text` vs `keyword` — confundirlos rompe queries. Un campo `status` como `text` tokeniza "in_progress" en ["in", "progress"]. Debe ser `keyword`.

### Analyzers (Cómo se Tokeniza el Texto)
```
Input: "The quick brown fox"

Standard analyzer: ["the", "quick", "brown", "fox"]
Spanish analyzer:  ["quick", "brown", "fox"] (remueve stopwords ES)
Keyword analyzer:  ["The quick brown fox"] (sin tokenización)
Custom analyzer:   ["quick", "brown", "fox"] + sinónimos + stemming
```

Para idiomas con morfología rica (español, alemán): configurar analyzer del idioma explícitamente.

## Modelo de Datos

### Diseño de Índice
```json
// Elasticsearch mapping
{
  "mappings": {
    "properties": {
      "title":       {"type": "text", "analyzer": "spanish"},
      "title_exact": {"type": "keyword"},
      "description": {"type": "text", "analyzer": "spanish"},
      "category":    {"type": "keyword"},
      "price":       {"type": "float"},
      "tags":        {"type": "keyword"},
      "location":    {"type": "geo_point"},
      "created_at":  {"type": "date"},
      "is_active":   {"type": "boolean"}
    }
  }
}
```

### Convenciones de Naming

**Índices:**
```
{app}_{entity}_{env}_v{n}
# Ejemplos:
store_products_prod_v3
blog_articles_staging_v1
support_tickets_prod_v2

# NUNCA usar alias como nombre base de índice real
# Alias: store_products_prod → apunta a store_products_prod_v3
```

**Campos:**
```
# snake_case, mismos nombres que la DB fuente
product_name, category_id, created_at

# Multi-field para text + keyword:
title        → text (búsqueda)
title.raw    → keyword (sort, aggregation)
```

## Migraciones — Alias + Reindex

Los search engines NO soportan ALTER mapping en campos existentes. El patrón estándar es **alias swap**:

### Flujo de migración
```
1. Crear índice nuevo: store_products_prod_v4 (con nuevo mapping)
2. Reindex desde v3:
   POST _reindex
   {"source": {"index": "store_products_prod_v3"},
    "dest":   {"index": "store_products_prod_v4"}}
3. Verificar que v4 tiene todos los docs
4. Swap alias:
   POST _aliases
   {"actions": [
     {"remove": {"index": "store_products_prod_v3", "alias": "store_products_prod"}},
     {"add":    {"index": "store_products_prod_v4", "alias": "store_products_prod"}}
   ]}
5. Eliminar v3 después de período de validación
```

### Cambios seguros (sin reindex)
- Agregar campo nuevo al mapping
- Agregar sub-field a campo existente (`title.autocomplete`)

### Cambios que requieren reindex
- Cambiar tipo de campo (`text` → `keyword`)
- Cambiar analyzer
- Cambiar número de shards

### Meilisearch / Typesense
Más flexibles — permiten actualizar settings sin reindex en la mayoría de casos:
```javascript
// Meilisearch: actualizar searchable attributes
await index.updateSearchableAttributes(['title', 'description', 'tags']);
// Typesense: recrear colección si cambia schema de campos
```

## Sincronización con DB Principal

El índice es una **copia**. Necesitas estrategia de sync:

| Estrategia | Latencia | Complejidad | Cuándo usar |
|---|---|---|---|
| **Sync en write** | Baja (ms) | Media | App escribe a DB + índice en misma operación |
| **CDC (Change Data Capture)** | Baja (s) | Alta | Debezium/Kafka → índice. Mejor para múltiples consumers |
| **Cron/batch** | Alta (min-hrs) | Baja | Reindex periódico completo. Solo para datos que cambian poco |
| **Event-driven** | Baja (s) | Media | Evento de dominio → consumer que actualiza índice |

**Regla**: si la búsqueda es feature core del producto, usar CDC o sync en write. Si es secondary, cron está bien.

## Pitfalls de Producción

| # | Pitfall | Consecuencia | Prevención |
|---|---|---|---|
| 1 | Cambiar mapping después de tener datos | Reindex completo necesario | Planificar mapping antes de indexar |
| 2 | `_source` disabled | No puedes recuperar el documento original | Mantener `_source` habilitado |
| 3 | Indexar todos los campos | Índice gigante, queries lentas | Solo indexar campos que se buscan/filtran |
| 4 | No versionar el mapping | Reindex no reproducible | Mapping como archivo en el repo |
| 5 | Split-brain en clusters ES | Datos duplicados o perdidos | `minimum_master_nodes` correctamente configurado |
| 6 | Búsqueda sin paginación | Timeout en resultados grandes | Siempre `from`+`size` o `search_after` |
| 7 | Sync desincronizado con DB | Resultados de búsqueda obsoletos | Monitorear lag de sync, reconciliación periódica |
| 8 | Shards demasiado pequeños | Overhead de cluster | 10-50GB por shard como regla general |

## Optimización de Rendimiento

### Indexación
- **Bulk API** para indexación masiva — nunca documento por documento
- **Refresh interval**: `30s` en lugar del default `1s` durante bulk indexing
- Desactivar replicas durante reindex masivo, re-activar después

### Queries
- **`filter` context vs `query` context**: filters son cacheados y no calculan score — usar para filtros exactos
- **`bool` query**: combinar `must` (score) + `filter` (sin score, cacheado)
- **Highlighting**: solo pedir si se muestra al usuario — es costoso
- **Aggregations**: usar `composite` aggregation para paginar facets grandes

### Relevance Tuning
```json
// Boost por campo — título más importante que descripción
{
  "multi_match": {
    "query": "wireless headphones",
    "fields": ["title^3", "description^1", "tags^2"]
  }
}

// Function score — ajustar por popularidad/fecha
{
  "function_score": {
    "query": {"match": {"title": "headphones"}},
    "functions": [
      {"field_value_factor": {"field": "popularity", "modifier": "log1p"}}
    ]
  }
}
```

### Meilisearch / Typesense específico
- Configurar `searchableAttributes` en orden de prioridad (no buscar en todo)
- Configurar `filterableAttributes` explícitamente — solo campos que se filtran
- `sortableAttributes` solo para campos que necesitan sort — cada uno consume memoria

## Drivers por Stack

| Stack | Elasticsearch | Meilisearch | Typesense |
|---|---|---|---|
| Go | `elastic/go-elasticsearch` u `olivere/elastic` | `meilisearch/meilisearch-go` | `typesense/typesense-go` |
| TypeScript | `@elastic/elasticsearch` | `meilisearch` | `typesense` |
| Python | `elasticsearch-py` | `meilisearch` | `typesense` |
| Rust | `elasticsearch-rs` | `meilisearch-sdk` | REST API |

**Notas:**
- `olivere/elastic` tiene API más ergonómica que el driver oficial de ES — pero puede ir detrás en features nuevas
- Meilisearch y Typesense tienen SDKs muy simples — REST API como fallback universal
