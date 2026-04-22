"""Kind (Kubernetes in Docker) provider."""

import subprocess
from pathlib import Path

from cloudsentinel.providers.base import CloudProvider, ProvisionResult


class KindProvider(CloudProvider):
    """Provisions local kind clusters via the Terraform kind module.

    Requires: Docker running, terraform CLI, kind CLI.
    """

    _REPO_ROOT = Path(__file__).resolve().parents[3]

    @property
    def terraform_module_path(self) -> Path:
        return self._REPO_ROOT / "infrastructure" / "modules" / "kind"

    @property
    def terraform_environment_path(self) -> Path:
        return self._REPO_ROOT / "infrastructure" / "environments" / "dev"

    def provision(self) -> ProvisionResult:
        """Run terraform init + apply and return cluster info."""
        env_path = self.terraform_environment_path

        self._run_terraform(env_path, ["init", "-upgrade"])
        self._run_terraform(
            env_path,
            [
                "apply",
                "-auto-approve",
                f"-var=cluster_name={self.config.name}",
                f"-var=kubernetes_version={self.config.kubernetes_version}",
                f"-var=worker_count={self.config.worker_count}",
            ],
        )

        kubeconfig = self.get_kubeconfig()
        return ProvisionResult(
            cluster_name=self.config.name,
            kubeconfig_path=kubeconfig,
            endpoint=self._get_output(env_path, "kubeconfig_path"),
            provider="kind",
        )

    def destroy(self) -> None:
        """Run terraform destroy."""
        env_path = self.terraform_environment_path
        self._run_terraform(
            env_path,
            [
                "destroy",
                "-auto-approve",
                f"-var=cluster_name={self.config.name}",
                f"-var=kubernetes_version={self.config.kubernetes_version}",
                f"-var=worker_count={self.config.worker_count}",
            ],
        )

    def get_kubeconfig(self) -> Path:
        """Return the expected kubeconfig path."""
        return Path.home() / ".cloudsentinel" / "clusters" / f"{self.config.name}.kubeconfig"

    @staticmethod
    def _run_terraform(cwd: Path, args: list[str]) -> subprocess.CompletedProcess:
        """Run a terraform command, streaming output live."""
        cmd = ["terraform", *args]
        return subprocess.run(cmd, cwd=cwd, check=True)

    @staticmethod
    def _get_output(cwd: Path, name: str) -> str:
        """Fetch a terraform output value."""
        result = subprocess.run(
            ["terraform", "output", "-raw", name],
            cwd=cwd,
            check=True,
            capture_output=True,
            text=True,
        )
        return result.stdout.strip()
