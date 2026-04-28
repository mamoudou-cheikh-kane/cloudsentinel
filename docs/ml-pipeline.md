# ML Pipeline

CloudSentinel ships a minimal machine-learning pipeline to detect **anomalous node behavior** from the metrics collected by the agent. This is the foundation for closing the loop: faults injected by the agent produce metric spikes, which the model learns to recognize.

!!! info "Where this fits in the bigger picture"
    In **Niveau 1** (v0.2.0), the agent injects real faults and exposes Prometheus metrics. The ML pipeline trains on those metrics so it can flag future anomalies — including the side-effects of unplanned faults like a saturated database or a crashing pod. **Niveau 2** will use this same pipeline to evaluate scenario outcomes automatically: if a chaos experiment causes anomalies the model has not seen before, the report will flag it.

## Algorithm

The pipeline uses a **scikit-learn Isolation Forest** wrapped around a `StandardScaler`. Isolation Forest is an unsupervised anomaly detection algorithm that isolates observations by randomly selecting a feature and then a split value. The number of splits required to isolate a sample becomes its anomaly score — anomalies are usually isolated much faster than normal observations.

We chose Isolation Forest because:

- :material-check: It is **unsupervised** — no need for labeled anomalies
- :material-check: It scales linearly with the number of observations
- :material-check: It works well on **multi-dimensional metrics** (CPU, memory, disk, network)
- :material-check: It is **robust to small training sets** — useful when you don't have years of historical data

## Features

The model is trained on five numerical features per observation:

| Feature | Source metric |
|---------|---------------|
| `cpu_percent` | `cloudsentinel_cpu_usage_percent` |
| `memory_percent` | `cloudsentinel_memory_usage_percent` |
| `disk_percent` | `cloudsentinel_disk_usage_percent` |
| `network_rx_kbps` | derived from `cloudsentinel_network_*_bytes_total` |
| `network_tx_kbps` | derived from `cloudsentinel_network_*_bytes_total` |

These correspond directly to the Prometheus metrics exported by the [Go agent](architecture.md#agent-internals).

## Training a model

The package bundles a synthetic dataset generator so you can train a working model with no external data:

```bash
cd ml-pipeline
poetry run python -m cloudsentinel_ml.cli generate-dataset
poetry run python -m cloudsentinel_ml.cli train
```

A sample run on 5,250 observations reaches:

| Metric | Value |
|--------|-------|
| **Precision** | 0.96 |
| **Recall** | 1.00 |
| **F1-score** | 0.98 |

Results are printed as a Rich table and the trained model is saved to `src/models/isolation_forest.joblib`.

## Predicting on new data

Once a model exists, you can classify any CSV that has the five required feature columns:

```bash
poetry run python -m cloudsentinel_ml.cli predict my_metrics.csv
```

The output is written to `src/inference/predictions.csv` with two extra columns:

- `predicted_anomaly` — 0 (normal) or 1 (anomaly)
- `anomaly_score` — float; lower values are more anomalous

## Closing the loop with the agent

The interesting workflow is when the agent and the ML pipeline run together:

```text
┌──────────────────┐    fault     ┌──────────────────┐
│  Chaos scenario  │ ───────────▶ │   Go agent       │
│  (gRPC client)   │              │   (DaemonSet)    │
└──────────────────┘              └────────┬─────────┘
                                           │ Prometheus metrics
                                           ▼
                                  ┌──────────────────┐
                                  │   Prometheus     │
                                  │  (or any TSDB)   │
                                  └────────┬─────────┘
                                           │ historical CSV / range query
                                           ▼
                                  ┌──────────────────┐
                                  │   ML Pipeline    │
                                  │ Isolation Forest │
                                  └────────┬─────────┘
                                           │ predictions
                                           ▼
                                  ┌──────────────────┐
                                  │  Scenario report │
                                  │  (Niveau 2)      │
                                  └──────────────────┘
```

A scenario produces metric perturbations. The model classifies each observation as normal or anomalous. The future Niveau 2 reports will surface these classifications as part of the experiment outcome — turning chaos engineering into **measurable, comparable experiments**.

## Replacing the synthetic dataset

In production you would replace the synthetic generator with a job that queries Prometheus for historical data:

```python
from prometheus_api_client import PrometheusConnect

conn = PrometheusConnect(url="http://prometheus.example.com")
cpu = conn.custom_query_range(
    query="cloudsentinel_cpu_usage_percent",
    start_time=...,
    end_time=...,
    step="1m",
)
# Convert to DataFrame, feed AnomalyDetector.train()
```

The `AnomalyDetector` class in `cloudsentinel_ml.model` takes any DataFrame with the right columns, so swapping the data source is a drop-in change.

## Roadmap for the ML pipeline

- :white_check_mark: **Done (v0.2.0)** — Isolation Forest training and inference on synthetic data
- :hourglass: **Niveau 2** — Automatic invocation from scenario reports (PASS/FAIL based on anomaly detection)
- :hourglass: **Niveau 3** — LSTM Autoencoder alongside Isolation Forest for richer time-series anomaly detection
- :hourglass: **Niveau 4** — Online learning: retrain automatically as the cluster's "normal" baseline drifts

## Where to go next

- :material-gamepad-variant: [Playground](playground.md) — inject faults and watch the metrics
- :material-blueprint: [Architecture](architecture.md) — how the agent exposes metrics
- :material-source-branch: [Source code](https://github.com/mamoudou-cheikh-kane/cloudsentinel) on GitHub
