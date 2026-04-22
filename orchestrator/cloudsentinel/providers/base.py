"""Abstract base class for cloud providers."""

from abc import ABC, abstractmethod
from dataclasses import dataclass
from pathlib import Path

from cloudsentinel.config import ClusterConfig


@dataclass
class ProvisionResult:
    """Result returned after provisioning a cluster."""

    cluster_name: str
    kubeconfig_path: Path
    endpoint: str
    provider: str


class CloudProvider(ABC):
    """Abstract base class for all cloud providers.

    Each provider (kind, gke, aks, eks) implements these methods to provision,
    query, and destroy Kubernetes clusters in a consistent way.
    """

    def __init__(self, config: ClusterConfig) -> None:
        self.config = config

    @property
    @abstractmethod
    def terraform_module_path(self) -> Path:
        """Path to the Terraform module for this provider."""

    @property
    @abstractmethod
    def terraform_environment_path(self) -> Path:
        """Path to the Terraform environment that uses the module."""

    @abstractmethod
    def provision(self) -> ProvisionResult:
        """Create the cluster. Returns connection details."""

    @abstractmethod
    def destroy(self) -> None:
        """Tear down the cluster. Idempotent."""

    @abstractmethod
    def get_kubeconfig(self) -> Path:
        """Return the path to the kubeconfig for an existing cluster."""

    def __repr__(self) -> str:
        return (
            f"{self.__class__.__name__}("
            f"name={self.config.name}, "
            f"provider={self.config.provider.value})"
        )
