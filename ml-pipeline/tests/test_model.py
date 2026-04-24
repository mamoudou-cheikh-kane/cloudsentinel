"""Tests for the Isolation Forest model wrapper."""

from pathlib import Path

import pytest
from cloudsentinel_ml.features import SyntheticConfig, generate_synthetic_dataset
from cloudsentinel_ml.model import AnomalyDetector, TrainConfig


@pytest.fixture
def small_dataset():
    return generate_synthetic_dataset(SyntheticConfig(n_normal=500, n_anomalies=25, seed=11))


class TestAnomalyDetectorTrain:
    def test_train_returns_metrics(self, small_dataset):
        detector = AnomalyDetector(TrainConfig(n_estimators=50, contamination=0.05))
        result = detector.train(small_dataset)
        assert 0.0 <= result.precision <= 1.0
        assert 0.0 <= result.recall <= 1.0
        assert 0.0 <= result.f1 <= 1.0

    def test_train_reaches_good_f1(self, small_dataset):
        """Synthetic anomalies are clearly separable — F1 should be > 0.8."""
        detector = AnomalyDetector(TrainConfig(n_estimators=50, contamination=0.05))
        result = detector.train(small_dataset)
        assert result.f1 > 0.8

    def test_predict_requires_training(self, small_dataset):
        detector = AnomalyDetector()
        with pytest.raises(RuntimeError, match="not trained"):
            detector.predict(small_dataset)

    def test_predict_returns_binary_array(self, small_dataset):
        detector = AnomalyDetector(TrainConfig(n_estimators=50))
        detector.train(small_dataset)
        preds = detector.predict(small_dataset)
        assert set(preds.tolist()).issubset({0, 1})
        assert len(preds) == len(small_dataset)


class TestAnomalyDetectorPersistence:
    def test_save_and_load(self, small_dataset, tmp_path: Path):
        detector = AnomalyDetector(TrainConfig(n_estimators=50))
        detector.train(small_dataset)
        model_path = tmp_path / "model.joblib"
        detector.save(model_path)
        assert model_path.exists()

        loaded = AnomalyDetector.load(model_path)
        # Predictions should match between the original and the loaded model.
        original_preds = detector.predict(small_dataset)
        loaded_preds = loaded.predict(small_dataset)
        assert (original_preds == loaded_preds).all()

    def test_save_untrained_raises(self, tmp_path: Path):
        detector = AnomalyDetector()
        with pytest.raises(RuntimeError, match="not trained"):
            detector.save(tmp_path / "model.joblib")
