"""Tests for the features module."""

from pathlib import Path

import pandas as pd
from cloudsentinel_ml.features import (
    FEATURE_COLUMNS,
    SyntheticConfig,
    generate_synthetic_dataset,
    load_dataset,
    save_dataset,
)


class TestGenerateSyntheticDataset:
    def test_default_size(self):
        df = generate_synthetic_dataset()
        assert len(df) == 5250
        assert df["is_anomaly"].sum() == 250

    def test_custom_size(self):
        config = SyntheticConfig(n_normal=100, n_anomalies=20, seed=7)
        df = generate_synthetic_dataset(config)
        assert len(df) == 120
        assert df["is_anomaly"].sum() == 20

    def test_required_columns_present(self):
        df = generate_synthetic_dataset()
        for col in FEATURE_COLUMNS:
            assert col in df.columns
        assert "is_anomaly" in df.columns

    def test_seed_is_reproducible(self):
        df1 = generate_synthetic_dataset(SyntheticConfig(seed=123))
        df2 = generate_synthetic_dataset(SyntheticConfig(seed=123))
        pd.testing.assert_frame_equal(df1, df2)

    def test_anomalies_are_higher_than_normals(self):
        df = generate_synthetic_dataset()
        normal_cpu = df[df["is_anomaly"] == 0]["cpu_percent"].mean()
        anomaly_cpu = df[df["is_anomaly"] == 1]["cpu_percent"].mean()
        assert anomaly_cpu > normal_cpu + 30  # clear separation


class TestSaveLoadDataset:
    def test_roundtrip(self, tmp_path: Path):
        df = generate_synthetic_dataset(SyntheticConfig(n_normal=50, n_anomalies=5))
        out_path = tmp_path / "sub" / "data.csv"
        save_dataset(df, out_path)
        assert out_path.exists()

        loaded = load_dataset(out_path)
        assert len(loaded) == len(df)
        assert list(loaded.columns) == list(df.columns)

    def test_save_creates_parent_dir(self, tmp_path: Path):
        df = generate_synthetic_dataset(SyntheticConfig(n_normal=10, n_anomalies=2))
        out_path = tmp_path / "deeply" / "nested" / "data.csv"
        save_dataset(df, out_path)
        assert out_path.exists()
