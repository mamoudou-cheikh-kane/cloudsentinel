"""Isolation Forest model training and evaluation."""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
from typing import Any

import joblib
import numpy as np
import pandas as pd
from sklearn.ensemble import IsolationForest
from sklearn.metrics import (
    classification_report,
    confusion_matrix,
    precision_recall_fscore_support,
)
from sklearn.model_selection import train_test_split
from sklearn.preprocessing import StandardScaler

from cloudsentinel_ml.features import FEATURE_COLUMNS


@dataclass(frozen=True)
class TrainConfig:
    """Hyperparameters for the Isolation Forest."""

    n_estimators: int = 100
    contamination: float = 0.05
    random_state: int = 42
    test_size: float = 0.2


@dataclass
class TrainResult:
    """Output of a training run."""

    precision: float
    recall: float
    f1: float
    confusion_matrix: list[list[int]]
    classification_report: str


class AnomalyDetector:
    """Train/predict wrapper around Isolation Forest + StandardScaler."""

    def __init__(self, config: TrainConfig | None = None) -> None:
        self.config = config or TrainConfig()
        self.scaler: StandardScaler | None = None
        self.model: IsolationForest | None = None

    def train(self, df: pd.DataFrame) -> TrainResult:
        """Fit the model on df (must contain FEATURE_COLUMNS + is_anomaly)."""
        X = df[FEATURE_COLUMNS].values
        y = df["is_anomaly"].values

        X_train, X_test, _, y_test = train_test_split(
            X,
            y,
            test_size=self.config.test_size,
            random_state=self.config.random_state,
            stratify=y,
        )

        self.scaler = StandardScaler()
        X_train_scaled = self.scaler.fit_transform(X_train)
        X_test_scaled = self.scaler.transform(X_test)

        self.model = IsolationForest(
            n_estimators=self.config.n_estimators,
            contamination=self.config.contamination,
            random_state=self.config.random_state,
        )
        self.model.fit(X_train_scaled)

        # Isolation Forest predicts -1 for anomalies, 1 for normal. We flip it.
        y_pred_raw = self.model.predict(X_test_scaled)
        y_pred = (y_pred_raw == -1).astype(int)

        precision, recall, f1, _ = precision_recall_fscore_support(
            y_test, y_pred, average="binary", zero_division=0
        )
        cm = confusion_matrix(y_test, y_pred).tolist()
        report = classification_report(
            y_test, y_pred, target_names=["normal", "anomaly"], zero_division=0
        )

        return TrainResult(
            precision=float(precision),
            recall=float(recall),
            f1=float(f1),
            confusion_matrix=cm,
            classification_report=report,
        )

    def predict(self, df: pd.DataFrame) -> np.ndarray:
        """Return an anomaly flag (1 = anomaly) per row."""
        if self.model is None or self.scaler is None:
            raise RuntimeError("Model is not trained; call .train() or .load() first.")
        X = df[FEATURE_COLUMNS].values
        X_scaled = self.scaler.transform(X)
        return (self.model.predict(X_scaled) == -1).astype(int)

    def score(self, df: pd.DataFrame) -> np.ndarray:
        """Return an anomaly score (the lower, the more anomalous)."""
        if self.model is None or self.scaler is None:
            raise RuntimeError("Model is not trained; call .train() or .load() first.")
        X = df[FEATURE_COLUMNS].values
        X_scaled = self.scaler.transform(X)
        return self.model.decision_function(X_scaled)

    def save(self, model_path: Path) -> Path:
        """Persist scaler + model to disk."""
        if self.model is None or self.scaler is None:
            raise RuntimeError("Nothing to save — model is not trained.")
        model_path.parent.mkdir(parents=True, exist_ok=True)
        joblib.dump(
            {"scaler": self.scaler, "model": self.model, "config": self.config},
            model_path,
        )
        return model_path

    @classmethod
    def load(cls, model_path: Path) -> AnomalyDetector:
        """Load a previously saved detector."""
        bundle: dict[str, Any] = joblib.load(model_path)
        detector = cls(config=bundle["config"])
        detector.scaler = bundle["scaler"]
        detector.model = bundle["model"]
        return detector
