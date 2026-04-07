# Vegeta — Reference Guide

> Go-based HTTP load testing library and CLI. Best for: Go projects, precise rate control, binary results for custom analysis.

## Install

```bash
# CLI
go install github.com/tsenart/vegeta/v12@latest
# Or brew
brew install vegeta
```

## CLI Usage

```bash
# Basic attack
echo 'GET https://api.example.com/health' | vegeta attack -rate=50/s -duration=30s > results.bin

# POST with body and headers
echo 'POST https://api.example.com/endpoint' | \
  vegeta attack -rate=50/s -duration=30s \
  -header "Content-Type: application/json" \
  -header "Authorization: Bearer TOKEN" \
  -body payload.json > results.bin

# Report
vegeta report results.bin

# Encode to JSON (for analysis)
vegeta encode < results.bin > results.json

# Plot (basic text histogram)
vegeta report -type=hist[0,100ms,200ms,500ms,1s,5s] results.bin
```

## Go Library Usage

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

## Result Decoding (for charts)

Vegeta has NO native charts. Use this pattern to decode and chart with matplotlib:

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

## Binary Format

Each result in the `.bin` file (gob-encoded) contains:
- `timestamp` — when the request was sent (ISO 8601)
- `latency` — response time in nanoseconds
- `code` — HTTP status code
- `body` — response body (base64)
- `headers` — response headers
- `error` — error string if failed
- `seq` — sequence number

**Completion time** = `timestamp` + `latency` (for timeline analysis).

## Comparison with Other Tools

| Feature | Vegeta |
|---------|--------|
| Native charts | No — use matplotlib |
| Real-time dashboard | No |
| HTML report | No |
| Binary results for replay | Yes |
| Precise rate control | Yes (ConstantPacer, LinearPacer) |
| Custom targeter (Go) | Yes — full control per request |
| Best for | Go projects, CI pipelines, custom analysis |
