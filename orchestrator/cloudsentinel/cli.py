"""CloudSentinel command-line interface."""

from pathlib import Path

import typer
from rich.console import Console
from rich.panel import Panel
from rich.table import Table

from cloudsentinel import __version__
from cloudsentinel.config import ClusterConfig
from cloudsentinel.providers import get_provider

app = typer.Typer(
    name="cloudsentinel",
    help="Multi-cloud chaos engineering orchestrator",
    add_completion=False,
    rich_markup_mode="rich",
)

console = Console()


@app.command()
def version() -> None:
    """Show CloudSentinel version."""
    console.print(f"CloudSentinel [bold cyan]{__version__}[/]")


@app.command()
def deploy(
    config: Path = typer.Option(
        ...,
        "--config",
        "-c",
        help="Path to cluster YAML config file",
        exists=True,
        file_okay=True,
        dir_okay=False,
        readable=True,
    ),
) -> None:
    """Provision a Kubernetes cluster from a YAML config."""
    console.print(Panel.fit(f"📦 Loading config from [cyan]{config}[/]", style="blue"))
    cfg = ClusterConfig.from_yaml(config)

    console.print(
        f"🚀 Deploying [bold green]{cfg.name}[/] on [bold cyan]{cfg.provider.value}[/] "
        f"with {cfg.worker_count} workers (K8s {cfg.kubernetes_version})"
    )

    provider_cls = get_provider(cfg.provider.value)
    provider = provider_cls(cfg)

    result = provider.provision()

    console.print(
        Panel.fit(
            f"✅ Cluster [bold green]{result.cluster_name}[/] ready!\n"
            f"📋 Kubeconfig: [cyan]{result.kubeconfig_path}[/]\n"
            f"💡 Run: [yellow]export KUBECONFIG={result.kubeconfig_path}[/]",
            title="Deployment Successful",
            style="green",
        )
    )


@app.command()
def destroy(
    config: Path = typer.Option(
        ...,
        "--config",
        "-c",
        help="Path to cluster YAML config file",
        exists=True,
    ),
    yes: bool = typer.Option(False, "--yes", "-y", help="Skip confirmation prompt"),
) -> None:
    """Tear down a Kubernetes cluster."""
    cfg = ClusterConfig.from_yaml(config)

    if not yes:
        confirm = typer.confirm(f"⚠️  Really destroy cluster '{cfg.name}' ({cfg.provider.value})?")
        if not confirm:
            console.print("[yellow]Aborted.[/]")
            raise typer.Exit(0)

    console.print(f"🧹 Destroying [bold red]{cfg.name}[/]...")
    provider_cls = get_provider(cfg.provider.value)
    provider = provider_cls(cfg)
    provider.destroy()

    console.print("✅ [green]Cluster destroyed successfully.[/]")


@app.command(name="list")
def list_clusters() -> None:
    """List locally-tracked clusters."""
    clusters_dir = Path.home() / ".cloudsentinel" / "clusters"

    if not clusters_dir.exists():
        console.print("[yellow]No clusters tracked yet.[/]")
        return

    table = Table(title="CloudSentinel Clusters")
    table.add_column("Cluster", style="cyan")
    table.add_column("Kubeconfig", style="dim")

    found = False
    for kc in sorted(clusters_dir.glob("*.kubeconfig")):
        name = kc.stem
        table.add_row(name, str(kc))
        found = True

    if found:
        console.print(table)
    else:
        console.print("[yellow]No clusters tracked yet.[/]")


if __name__ == "__main__":
    app()
