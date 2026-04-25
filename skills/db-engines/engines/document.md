# Bases de Datos de Documentos

## Cuándo Usar

Entidades con estructura variable, jerarquías naturales (order → items → products), acceso mayormente por ID primario, esquemas que evolucionan rápidamente. NO para datos altamente relacionales con JOINs complejos — usar relacional.

## Opciones y Cuándo Elegir

| Motor | Cuándo elegir | Notas |
|---|---|---|
| **MongoDB** | Uso general, estructura flexible, queries ricas | El más maduro. Atlas para managed |
| **DynamoDB** | Escala masiva en AWS, latencia predecible | Partition key + sort key definen TODO. Sin GSI no puedes filtrar por otros campos |
| **Firestore** | Apps mobile/real-time, sincronización offline | Límite 1 write/doc/second en transacciones. Real-time listeners nativos |

## Modelo de Datos

### Embedding vs Referencing — La Decisión Fundamental

| Patrón | Cuándo usar | Ejemplo |
|---|---|---|
| **Embedding** (subdocumento) | Dato siempre se lee junto, relación 1:pocos, dato no cambia independientemente | Order con items embebidos |
| **Referencing** (ID externo) | Dato cambia independientemente, relación muchos:muchos, documento crecería sin límite | User con referencia a company_id |

**Regla de oro**: modela para los **patrones de acceso**, no para la normalización.

### Ejemplo de documento bien diseñado (MongoDB)
```json
{
  "_id": "order_abc123",
  "_schema_version": 2,
  "customer_id": "cust_789",
  "status": "shipped",
  "items": [
    {
      "product_id": "prod_456",
      "name": "Widget Pro",
      "quantity": 2,
      "unit_price": 29.99
    }
  ],
  "shipping_address": {
    "street": "123 Main St",
    "city": "Austin",
    "state": "TX",
    "zip": "78701"
  },
  "created_at": "2025-01-15T10:30:00Z",
  "updated_at": "2025-01-16T14:22:00Z"
}
```

### Convenciones de Naming

**Colecciones** (equivalente a tablas):
```
# Plural, snake_case
user_profiles
order_items
api_keys
audit_logs
```

**Campos**:
```
# snake_case consistente con el stack (Go/Python)
# O camelCase si el stack es TypeScript — elegir UNO y no mezclar
created_at, updated_at          # timestamps
is_active, has_verified          # booleanos
_schema_version                  # campo de versión (prefijo _ = interno)
```

**DynamoDB — Tabla única (Single Table Design)**:
```
PK: USER#123              SK: PROFILE
PK: USER#123              SK: ORDER#456
PK: ORDER#456             SK: ITEM#789
```

## Migraciones / Versionado

Document DBs no tienen `ALTER TABLE`. El esquema evoluciona con los datos.

### Estrategia: Schema Versioning + Lazy Migration

```javascript
// Campo _schema_version en cada documento
{
  "_schema_version": 2,
  "full_name": "John Doe"  // v2: fusionó first_name + last_name
}

// Lazy migration al leer
function getUser(doc) {
  if (doc._schema_version < 2) {
    doc = migrateV1toV2(doc);
    await collection.replaceOne({_id: doc._id}, doc);
  }
  return doc;
}
```

### Tipos de cambio y estrategia

| Cambio | Estrategia | Riesgo |
|---|---|---|
| Agregar campo nuevo | Escribir en nuevos docs, default en lecturas | Bajo |
| Renombrar campo | Lazy migration + dual-read temporal | Medio |
| Cambiar tipo de campo | Batch migration + código que maneja ambos tipos | Alto |
| Eliminar campo | Dejar de leer/escribir, cleanup batch posterior | Bajo |
| Reestructurar subdocumento | Versión nueva + lazy migration | Alto |

### Batch Migration (para cambios que no pueden ser lazy)
```javascript
// Rate-limited para no saturar la DB
const cursor = collection.find({_schema_version: {$lt: 2}});
let batch = [];
for await (const doc of cursor) {
  batch.push({
    updateOne: {
      filter: {_id: doc._id},
      update: {$set: migrateV1toV2(doc)}
    }
  });
  if (batch.length >= 500) {
    await collection.bulkWrite(batch);
    batch = [];
    await sleep(100); // rate limiting
  }
}
```

## Pitfalls de Producción

| # | Pitfall | Consecuencia | Prevención |
|---|---|---|---|
| 1 | Arrays que crecen sin límite | Document growth, fragmentación, 16MB limit (MongoDB) | Límite explícito o referencing |
| 2 | Sin índices antes de producción | Full collection scan en queries | Crear índices para cada patrón de query |
| 3 | DynamoDB hot partition | Throttling, latencia errática | Distribuir partition key, evitar IDs secuenciales |
| 4 | Firestore OR entre campos | No soportado — requiere múltiples queries + merge en cliente | Campos compuestos o denormalización |
| 5 | Transacciones cross-colección | Costo doble en RCU/WCU (DynamoDB), límites estrictos (Firestore) | Minimizar transacciones, diseñar para single-doc writes |
| 6 | No usar projection | Transfiere documentos completos aunque solo necesites 2 campos | Siempre especificar campos necesarios |
| 7 | Índices compuestos con orden incorrecto | Índice no se usa para la query | Orden = selectividad descendente |

## Índices

### MongoDB
```javascript
// Índice compuesto — orden importa
db.orders.createIndex({customer_id: 1, created_at: -1});

// Índice parcial — solo documentos activos
db.orders.createIndex(
  {status: 1},
  {partialFilterExpression: {status: {$ne: "archived"}}}
);

// Índice de texto para búsqueda
db.products.createIndex({name: "text", description: "text"});

// TTL index — auto-elimina documentos expirados
db.sessions.createIndex({expires_at: 1}, {expireAfterSeconds: 0});
```

### DynamoDB
```
# Tabla principal: PK + SK obligatorios
# Global Secondary Index (GSI): para queries por otros campos
# Local Secondary Index (LSI): mismo PK, diferente SK — solo al crear tabla
```

## Optimización de Rendimiento

- **Projection**: seleccionar solo campos necesarios — reduce network y memoria
- **Bulk operations**: `insertMany`, `bulkWrite` en vez de loops
- **Read preference** (MongoDB): leer de secondaries para queries analíticas
- **DynamoDB**: `BatchGetItem` y `BatchWriteItem`, evitar `Scan`
- **Firestore**: queries compuestas requieren índices compuestos — crearlos proactivamente
- **Connection pooling**: MongoDB mantiene pool por defecto, configurar `maxPoolSize`

## Drivers por Stack

| Stack | MongoDB | DynamoDB | Firestore |
|---|---|---|---|
| Go | `go.mongodb.org/mongo-driver` | `aws-sdk-go-v2/service/dynamodb` | `cloud.google.com/go/firestore` |
| TypeScript | `mongodb` driver o `mongoose` (ODM) | `@aws-sdk/client-dynamodb` | `firebase-admin` |
| Python | `pymongo` o `motor` (async) | `boto3` | `google-cloud-firestore` |
| Rust | `mongodb` crate | `aws-sdk-dynamodb` | REST API |

**Notas por driver:**
- **Mongoose** (TS): ODM con schemas validados — bueno para proyectos medianos+, overhead para microservicios simples
- **Motor** (Python): wrapper async sobre pymongo — usar con FastAPI/asyncio
- **Beanie** (Python): ODM async con Pydantic — tipado fuerte para FastAPI
