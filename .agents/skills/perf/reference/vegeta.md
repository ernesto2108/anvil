# Vegeta — Guía de Referencia

> Librería y CLI de load testing HTTP basada en Go. Ideal para: proyectos Go, control preciso de tasa, resultados binarios para análisis personalizado.

## Instalación

```bash
# CLI
go install github.com/tsenart/vegeta/v12@latest
# O con brew
brew install vegeta
```

## Uso de CLI

```bash
# Ataque básico
echo 'GET https://api.example.com/health' | vegeta attack -rate=50/s -duration=30s > results.bin

# POST con body y headers
echo 'POST https://api.example.com/endpoint' | \
  vegeta attack -rate=50/s -duration=30s \
  -header "Content-Type: application/json" \
  -header "Authorization: Bearer TOKEN" \
  -body payload.json > results.bin

# Reporte
vegeta report results.bin

# Codificar a JSON (para análisis)
vegeta encode < results.bin > results.json

# Plot (histograma de texto básico)
vegeta report -type=hist[0,100ms,200ms,500ms,1s,5s] results.bin
```

## Uso de la Librería Go

```go
import vegeta "github.com/tsenart/vegeta/v12/lib"

attacker := vegeta.NewAttacker(
    vegeta.Timeout(30*time.Second),
    vegeta.Workers(10),
    vegeta.MaxWorkers(50),
)
pacer := vegeta.ConstantPacer{Freq: 50, Per: time.Second}

var metrics vegeta.Metrics
for res := range attacker.Attack(targeter, pacer, 30*time.Second, "test-name") {
    metrics.Add(res)
}
metrics.Close()
```

## Decodificación de Resultados (para gráficas)

Vegeta NO tiene gráficas nativas. Usar este patrón para decodificar y graficar con matplotlib:

```python
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import json, subprocess
from collections import defaultdict
from datetime import datetime, timedelta

def load_vegeta(bin_file):
    """Decode vegeta binary to list of dicts."""
    result = subprocess.run(
        ['vegeta', 'encode'],
        stdin=open(bin_file, 'rb'),
        capture_output=True, text=True
    )
    return [json.loads(l) for l in result.stdout.strip().split('\n') if l]

def completions_per_second(results):
    """Group completed requests by second."""
    start = min(datetime.fromisoformat(r['timestamp']) for r in results)
    buckets = defaultdict(int)
    for r in results:
        ts = datetime.fromisoformat(r['timestamp'])
        comp = ts + timedelta(seconds=r['latency'] / 1e9)
        buckets[int((comp - start).total_seconds())] += 1
    return buckets

# Load and chart
results = load_vegeta('results.bin')
comps = completions_per_second(results)
seconds = sorted(comps.keys())

fig, ax = plt.subplots(figsize=(14, 6))
ax.bar(seconds, [comps[s] for s in seconds], color='#2ecc71')
ax.axhline(y=50, color='#3498db', linestyle='--', label='Target')
ax.set_xlabel('Second')
ax.set_ylabel('Requests completed')
ax.legend()
plt.savefig('chart-completions.png', dpi=150, bbox_inches='tight')
```

## Formato Binario

Cada resultado en el archivo `.bin` (codificado en gob) contiene:
- `timestamp` — cuándo se envió el request (ISO 8601)
- `latency` — tiempo de respuesta en nanosegundos
- `code` — código de estado HTTP
- `body` — cuerpo de la respuesta (base64)
- `headers` — headers de la respuesta
- `error` — string de error si falló
- `seq` — número de secuencia

**Tiempo de completado** = `timestamp` + `latency` (para análisis de timeline).

## Comparación con Otras Herramientas

| Característica | Vegeta |
|---------|--------|
| Gráficas nativas | No — usar matplotlib |
| Dashboard en tiempo real | No |
| Reporte HTML | No |
| Resultados binarios para replay | Sí |
| Control preciso de tasa | Sí (ConstantPacer, LinearPacer) |
| Targeter personalizado (Go) | Sí — control total por request |
| Ideal para | Proyectos Go, pipelines CI, análisis personalizado |
