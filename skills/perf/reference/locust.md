# Locust — Reference Guide

> Python-based load testing with real-time web UI and native HTML reports. Best for: Python teams, interactive monitoring, quick setup.

## Install

```bash
pip install locust
```

## Basic Locustfile

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

## Execution

```bash
# With web UI (real-time charts at http://localhost:8089)
locust -f locustfile.py --host https://web-api-qa.example.com

# Headless mode with HTML report
locust -f locustfile.py \
  --headless \
  --users 50 \
  --spawn-rate 10 \
  --run-time 30s \
  --host https://web-api-qa.example.com \
  --html report.html

# With CSV output for analysis
locust -f locustfile.py \
  --headless \
  --users 50 \
  --spawn-rate 10 \
  --run-time 30s \
  --csv results
# Produces: results_stats.csv, results_stats_history.csv, results_failures.csv
```

## Native Charts

Locust provides charts in two ways:

### Web UI (interactive)
- Run without `--headless` flag
- Open `http://localhost:8089`
- Real-time charts: RPS, response times, users
- Can adjust users during test

### HTML Report
- Use `--html report.html` flag
- Generates static HTML with embedded charts
- Includes: response time percentiles, RPS, failures table
- **Limitation:** headless HTML reports have fewer charts than the web UI

### CSV for Custom Charts
If the native HTML report is not enough, use CSV output + matplotlib:

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

## Advanced: Stages (ramping)

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

## Comparison with Other Tools

| Feature | Locust |
|---------|--------|
| Native charts | Yes — web UI + HTML report |
| Real-time dashboard | Yes (localhost:8089) |
| HTML report | Yes (`--html`) |
| Language | Python |
| Custom scenarios | Python classes |
| Distributed mode | Yes (master/worker) |
| Best for | Python teams, interactive monitoring, quick prototyping |
