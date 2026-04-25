# Guía de Seguridad

## Validación de Input (Pydantic)

```python
from pydantic import BaseModel, Field, field_validator, ConfigDict

class SearchRequest(BaseModel):
    model_config = ConfigDict(strict=True, frozen=True)

    query: str = Field(min_length=1, max_length=1000)
    top_k: int = Field(ge=1, le=100, default=10)

    @field_validator("query")
    @classmethod
    def sanitize_query(cls, v: str) -> str:
        return v.strip()
```

## Prevención de Inyección SQL

```python
# WRONG — f-string in SQL
sql = f"SELECT * FROM items WHERE name = '{query}'"

# RIGHT — parameterized queries
sql = "SELECT * FROM items WHERE name = $1"
await pool.fetch(sql, query)
```

## Gestión de Secretos

```python
from pydantic_settings import BaseSettings
from pydantic import Field, ConfigDict

class Settings(BaseSettings):
    model_config = ConfigDict(env_file=".env")

    openai_api_key: str = Field(alias="OPENAI_API_KEY")
    db_url: str = Field(alias="DATABASE_URL")
    debug: bool = False

settings = Settings()  # fails fast if missing
```

## Inyección de Comandos

```python
# WRONG
import os
os.system(f"convert {filename}")  # shell injection

# RIGHT
import subprocess
subprocess.run(["convert", filename], check=True)  # list args, no shell
```

## Pickle / Deserialización

```python
# WRONG — pickle loads arbitrary code
data = pickle.loads(untrusted_bytes)

# RIGHT — use safe formats
data = json.loads(untrusted_bytes)
# or: validated Pydantic model
model = MyModel.model_validate_json(untrusted_bytes)
```

## Análisis de Dependencias

```bash
# In CI pipeline
uv pip audit              # if using uv
pip-audit --strict        # standalone
safety check              # alternative
```
