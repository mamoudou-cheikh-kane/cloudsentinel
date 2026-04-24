# CloudSentinel ML Pipeline

Isolation Forest-based anomaly detection for CloudSentinel node metrics.

## Setup

```bash
poetry install
```

## Commands

```bash
# Generate a synthetic dataset
poetry run python -m cloudsentinel_ml.cli generate-dataset

# Train the model
poetry run python -m cloudsentinel_ml.cli train

# Predict on new data
poetry run python -m cloudsentinel_ml.cli predict some_metrics.csv
```

## Tests

```bash
poetry run pytest tests/ -v
```

See the main [CloudSentinel documentation](../docs/ml-pipeline.md) for the
full pipeline description.
