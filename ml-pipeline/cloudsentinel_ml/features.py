"""Feature engineering and synthetic dataset generation for CloudSentinel."""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path

import numpy as np
import pandas as pd

FEATURE_COLUMNS = [
    "cpu_percent",
    "memory_percent",
    "disk_percent",
    "network_rx_kbps",
    "network_tx_kbps",
]


@dataclass(frozen=True)
class SyntheticConfig:
    """Parameters for generating a synthetic metrics dataset."""

    n_normal: int = 5000
    n_anomalies: int = 250
    seed: int = 42


def generate_synthetic_dataset(config: SyntheticConfig | None = None) -> pd.DataFrame:
    """Return a DataFrame with synthetic normal + anomalous metrics."""
    cfg = config or SyntheticConfig()
    rng = np.random.default_rng(cfg.seed)

    normal = pd.DataFrame(
        {
            "cpu_percent": rng.normal(40, 10, cfg.n_normal).clip(5, 75),
            "memory_percent": rng.normal(50, 12, cfg.n_normal).clip(10, 80),
            "disk_percent": rng.normal(60, 8, cfg.n_normal).clip(20, 85),
            "network_rx_kbps": rng.normal(500, 150, cfg.n_normal).clip(50, 2000),
            "network_tx_kbps": rng.normal(400, 120, cfg.n_normal).clip(50, 2000),
            "is_anomaly": 0,
        }
    )

    anomalies = pd.DataFrame(
        {
            "cpu_percent": rng.uniform(90, 100, cfg.n_anomalies),
            "memory_percent": rng.uniform(88, 99, cfg.n_anomalies),
            "disk_percent": rng.uniform(90, 99, cfg.n_anomalies),
            "network_rx_kbps": rng.uniform(2500, 8000, cfg.n_anomalies),
            "network_tx_kbps": rng.uniform(2500, 8000, cfg.n_anomalies),
            "is_anomaly": 1,
        }
    )

    df = pd.concat([normal, anomalies], ignore_index=True)
    return df.sample(frac=1, random_state=cfg.seed).reset_index(drop=True)


def save_dataset(df: pd.DataFrame, output_path: Path) -> Path:
    """Write a dataset to disk (parent dir created if missing)."""
    output_path.parent.mkdir(parents=True, exist_ok=True)
    df.to_csv(output_path, index=False)
    return output_path


def load_dataset(input_path: Path) -> pd.DataFrame:
    """Load a previously saved dataset from disk."""
    return pd.read_csv(input_path)
