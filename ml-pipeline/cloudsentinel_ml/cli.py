"""CloudSentinel ML command-line interface."""

from __future__ import annotations

from pathlib import Path

import pandas as pd
import typer
from rich.console import Console
from rich.panel import Panel
from rich.table import Table

from cloudsentinel_ml import __version__
from cloudsentinel_ml.features import (
    FEATURE_COLUMNS,
    SyntheticConfig,
    generate_synthetic_dataset,
    load_dataset,
    save_dataset,
)
from cloudsentinel_ml.model import AnomalyDetector, TrainConfig

app = typer.Typer(
    name="cloudsentinel-ml",
    help="CloudSentinel ML pipeline for anomaly detection.",
    no_args_is_help=True,
)
console = Console()

DEFAULT_DATASET = Path("src/features/synthetic_metrics.csv")
DEFAULT_MODEL = Path("src/models/isolation_forest.joblib")


@app.command()
def version() -> None:
    """Print the package version."""
    console.print(f"cloudsentinel-ml [bold]{__version__}[/bold]")


@app.command("generate-dataset")
def generate_dataset_cmd(
    output: Path = typer.Option(DEFAULT_DATASET, help="Output CSV path."),
    n_normal: int = typer.Option(5000, help="Number of normal observations."),
    n_anomalies: int = typer.Option(250, help="Number of anomalous observations."),
    seed: int = typer.Option(42, help="Random seed for reproducibility."),
) -> None:
    """Generate a synthetic metrics dataset and write it to disk."""
    config = SyntheticConfig(n_normal=n_normal, n_anomalies=n_anomalies, seed=seed)
    df = generate_synthetic_dataset(config)
    save_dataset(df, output)
    console.print(
        Panel.fit(
            f"[green]Dataset generated[/green]\n"
            f"Rows: {len(df)}\n"
            f"Anomalies: {int(df['is_anomaly'].sum())}\n"
            f"Path: {output}",
            title="generate-dataset",
        )
    )


@app.command()
def train(
    dataset: Path = typer.Option(DEFAULT_DATASET, help="Training CSV path."),
    model_output: Path = typer.Option(DEFAULT_MODEL, help="Where to save the model."),
    n_estimators: int = typer.Option(100, help="Number of trees."),
    contamination: float = typer.Option(0.05, help="Expected anomaly ratio."),
) -> None:
    """Train an Isolation Forest on the given dataset."""
    console.print(f"Loading dataset: [cyan]{dataset}[/cyan]")
    df = load_dataset(dataset)

    console.print("Training Isolation Forest...")
    detector = AnomalyDetector(TrainConfig(n_estimators=n_estimators, contamination=contamination))
    result = detector.train(df)

    table = Table(title="Training result", show_lines=True)
    table.add_column("Metric", style="bold")
    table.add_column("Value", justify="right")
    table.add_row("Precision", f"{result.precision:.3f}")
    table.add_row("Recall", f"{result.recall:.3f}")
    table.add_row("F1-score", f"{result.f1:.3f}")
    console.print(table)

    console.print("\nClassification report:")
    console.print(result.classification_report)

    detector.save(model_output)
    console.print(f"[green]Model saved to:[/green] {model_output}")


@app.command()
def predict(
    input_csv: Path = typer.Argument(..., help="CSV with metrics to classify."),
    model_path: Path = typer.Option(DEFAULT_MODEL, help="Trained model path."),
    output: Path = typer.Option(
        Path("src/inference/predictions.csv"), help="Where to save predictions."
    ),
) -> None:
    """Classify each row of a metrics CSV as normal or anomalous."""
    if not model_path.exists():
        console.print(f"[red]Model not found:[/red] {model_path}")
        console.print("Run [bold]cloudsentinel-ml train[/bold] first.")
        raise typer.Exit(code=1)

    detector = AnomalyDetector.load(model_path)
    df = pd.read_csv(input_csv)

    missing = [c for c in FEATURE_COLUMNS if c not in df.columns]
    if missing:
        console.print(f"[red]Missing columns in input CSV:[/red] {missing}")
        console.print(f"Expected: {FEATURE_COLUMNS}")
        raise typer.Exit(code=1)

    predictions = detector.predict(df)
    scores = detector.score(df)

    df_out = df.copy()
    df_out["predicted_anomaly"] = predictions
    df_out["anomaly_score"] = scores

    output.parent.mkdir(parents=True, exist_ok=True)
    df_out.to_csv(output, index=False)

    n_flagged = int(predictions.sum())
    console.print(
        Panel.fit(
            f"[cyan]Input:[/cyan] {input_csv} ({len(df)} rows)\n"
            f"[cyan]Flagged as anomalies:[/cyan] {n_flagged}\n"
            f"[cyan]Output:[/cyan] {output}",
            title="predict",
        )
    )


if __name__ == "__main__":
    app()
