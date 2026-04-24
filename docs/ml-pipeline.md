# ML Pipeline

CloudSentinel ships a minimal machine-learning pipeline to detect
anomalous node behavior from the metrics collected by the agent.

## Algorithm

The pipeline uses a **scikit-learn Isolation Forest** wrapped around a
`StandardScaler`. Isolation Forest is an unsupervised anomaly detection
algorithm that isolates observations by randomly selecting a feature and
then a split value. The number of splits required to isolate a sample
becomes its anomaly score — anomalies are usually isolated much faster
than normal observations.

## Features

The model is trained on five numerical features per observation:

- `cpu_percent`
- `memory_percent`
- `disk_percent`
- `network_rx_kbps`
- `network_tx_kbps`

These correspond directly to the Prometheus metrics exported by the Go
agent (`cpu_usage_percent`, `memory_usage_percent`, etc.).

## Training a model

The package bundles a synthetic dataset generator so you can train a
working model with no external data:

```bash
cd ml-pipeline
poetry run python -m cloudsentinel_ml.cli generate-dataset
poetry run python -m cloudsentinel_ml.cli train
```

A sample run on 5,250 observations reaches:

| Metric    | Value |
|-----------|-------|
| Precision | 0.96  |
| Recall    | 1.00  |
| F1-score  | 0.98  |

Results are printed as a Rich table and the trained model is saved to
`src/models/isolation_forest.joblib`.

## Predicting on new data

Once a model exists, you can classify any CSV that has the five required
feature columns:

```bash
poetry run python -m cloudsentinel_ml.cli predict my_metrics.csv
```

The output is written to `src/inference/predictions.csv` with two extra
columns: `predicted_anomaly` (0/1) and `anomaly_score` (float; lower is
more anomalous).

## Replacing the synthetic dataset

In production you would replace the synthetic generator with a job that
queries Prometheus for historical data:

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

The `AnomalyDetector` class in `cloudsentinel_ml.model` takes any
DataFrame with the right columns, so swapping the data source is a
drop-in change.
