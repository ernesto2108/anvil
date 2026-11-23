# Locust — Guía de Referencia

> Herramienta de load testing basada en Python con UI web en tiempo real y reportes HTML nativos. Ideal para: equipos Python, monitoreo interactivo, configuración rápida.

## Instalación

```bash
pip install locust
```

## Locustfile Básico

```python
from locust import HttpUser, task, between

class BookingUser(HttpUser):
    wait_time = between(0.1, 0.5)  # wait between tasks
    host = "https://web-api-qa.example.com"

    def on_start(self):
        """Setup: runs once per simulated user."""
        self.headers = {
            "Content-Type": "application/json",
            "application-name": "insiders",
        }

    @task
    def create_booking(self):
        payload = {
            "event_id": 237769,
            "app_name": "insiders",
            "body_req": [{
                "ticket_type_id": 394839,
                "quantity": 1
            }]
        }
        self.client.post(
            "/bookings-core/api/v1/bookings/generate-booking",
            json=payload,
            headers=self.headers
        )
```

## Ejecución

```bash
# Con UI web (gráficas en tiempo real en http://localhost:8089)
locust -f locustfile.py --host https://web-api-qa.example.com

# Modo headless con reporte HTML
locust -f locustfile.py \
  --headless \
  --users 50 \
  --spawn-rate 10 \
  --run-time 30s \
  --host https://web-api-qa.example.com \
  --html report.html

# Con salida CSV para análisis
locust -f locustfile.py \
  --headless \
  --users 50 \
  --spawn-rate 10 \
  --run-time 30s \
  --csv results
# Produce: results_stats.csv, results_stats_history.csv, results_failures.csv
```

## Gráficas Nativas

Locust provee gráficas de dos formas:

### UI Web (interactiva)
- Ejecutar sin la flag `--headless`
- Abrir `http://localhost:8089`
- Gráficas en tiempo real: RPS, tiempos de respuesta, usuarios
- Se pueden ajustar usuarios durante el test

### Reporte HTML
- Usar la flag `--html report.html`
- Genera HTML estático con gráficas embebidas
- Incluye: percentiles de tiempo de respuesta, RPS, tabla de fallos
- **Limitación:** los reportes HTML en modo headless tienen menos gráficas que la UI web

### CSV para Gráficas Personalizadas
Si el reporte HTML nativo no es suficiente, usar salida CSV + matplotlib:

```python
import pandas as pd
import matplotlib.pyplot as plt

# Load the stats history CSV
df = pd.read_csv('results_stats_history.csv')
df['Timestamp'] = pd.to_datetime(df['Timestamp'], unit='s')

fig, axes = plt.subplots(3, 1, figsize=(14, 10))

# RPS
axes[0].plot(df['Timestamp'], df['Requests/s'], color='#2ecc71')
axes[0].set_ylabel('Requests/s')
axes[0].set_title('Throughput')

# Response times
axes[1].plot(df['Timestamp'], df['50%'], label='p50')
axes[1].plot(df['Timestamp'], df['95%'], label='p95')
axes[1].set_ylabel('Latency (ms)')
axes[1].legend()

# Failures
axes[2].plot(df['Timestamp'], df['Failures/s'], color='#e74c3c')
axes[2].set_ylabel('Failures/s')

plt.tight_layout()
plt.savefig('chart-locust-timeline.png', dpi=150)
```

## Avanzado: Stages (ramping)

```python
from locust import HttpUser, task, between, LoadTestShape

class StagesShape(LoadTestShape):
    stages = [
        {"duration": 30, "users": 10, "spawn_rate": 5},   # warmup
        {"duration": 60, "users": 50, "spawn_rate": 10},   # ramp to 50
        {"duration": 120, "users": 50, "spawn_rate": 50},  # sustain
        {"duration": 150, "users": 0, "spawn_rate": 10},   # ramp down
    ]

    def tick(self):
        run_time = self.get_run_time()
        for stage in self.stages:
            if run_time < stage["duration"]:
                return (stage["users"], stage["spawn_rate"])
        return None
```

## Comparación con Otras Herramientas

| Característica | Locust |
|---------|--------|
| Gráficas nativas | Sí — UI web + reporte HTML |
| Dashboard en tiempo real | Sí (localhost:8089) |
| Reporte HTML | Sí (`--html`) |
| Lenguaje | Python |
| Escenarios personalizados | Clases Python |
| Modo distribuido | Sí (master/worker) |
| Ideal para | Equipos Python, monitoreo interactivo, prototipado rápido |
