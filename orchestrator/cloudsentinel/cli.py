"""CloudSentinel command-line interface."""

from pathlib import Path

import typer
from rich.console import Console
from rich.panel import Panel
from rich.table import Table

from cloudsentinel import __version__
from cloudsentinel.chaos import ChaosEngine, ChaosScenario
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


@app.command(name="chaos-run")
def chaos_run(
    scenario_file: str = typer.Argument(..., help="Path to the scenario YAML file"),
    kubeconfig: str = typer.Option(
        None,
        "--kubeconfig",
        help="Path to the kubeconfig (defaults to $KUBECONFIG or ~/.kube/config)",
    ),
    namespace: str = typer.Option(
        "cloudsentinel",
        "--namespace",
        "-n",
        help="Namespace where agents run",
    ),
) -> None:
    """Run a chaos scenario against a cluster."""
    console.print(
        Panel.fit(
            f"🧪 Loading scenario: [bold]{scenario_file}[/bold]",
            border_style="cyan",
        )
    )
    try:
        scenario = ChaosScenario.from_yaml(scenario_file)
    except Exception as e:
        console.print(f"[bold red]✗ Invalid scenario:[/bold red] {e}")
        raise typer.Exit(code=1) from e

    console.print(
        f"[bold]Scenario:[/bold] {scenario.name}\n"
        f"[bold]Description:[/bold] {scenario.description or '—'}\n"
        f"[bold]Faults:[/bold] {len(scenario.faults)}"
    )

    try:
        engine = ChaosEngine(kubeconfig_path=kubeconfig, namespace=namespace)
        console.print(f"🎯 Discovering agents in namespace [cyan]{namespace}[/cyan]…")
        result = engine.run_scenario(scenario)
    except Exception as e:
        console.print(f"[bold red]✗ Scenario failed:[/bold red] {e}")
        raise typer.Exit(code=1) from e

    table = Table(title=f"Results: {result.scenario_name}", show_lines=True)
    table.add_column("Step", style="bold")
    table.add_column("Node")
    table.add_column("Fault ID")
    table.add_column("Status")
    for inj in result.injections:
        short_id = inj.fault_id[:8] + "…" if inj.fault_id else "—"
        table.add_row(
            "inject",
            inj.node_name,
            short_id,
            "✅" if inj.accepted else "❌",
        )
    for rb in result.rollbacks:
        table.add_row(
            "rollback",
            rb.node_name,
            rb.fault_id[:8] + "…",
            "✅" if rb.success else "❌",
        )
    console.print(table)

    console.print(
        Panel.fit(
            f"🏁 Scenario complete — "
            f"{len(result.injections)} injection(s), "
            f"{len(result.rollbacks)} rollback(s)",
            border_style="green",
        )
    )
