"""Pydantic models for cluster configuration."""

import re
from enum import Enum
from pathlib import Path

import yaml
from pydantic import BaseModel, Field, field_validator


class ProviderType(str, Enum):
    """Supported cloud providers."""

    KIND = "kind"
    GKE = "gke"
    AKS = "aks"
    EKS = "eks"


class ClusterConfig(BaseModel):
    """Configuration for a CloudSentinel cluster deployment."""

    name: str = Field(..., description="Cluster name (lowercase, hyphens, digits)")
    provider: ProviderType = Field(..., description="Cloud provider to use")
    kubernetes_version: str = Field("1.30.0", description="K8s version without 'v' prefix")
    worker_count: int = Field(2, ge=0, le=10, description="Number of worker nodes (0-10)")
    provider_options: dict = Field(
        default_factory=dict,
        description="Provider-specific configuration overrides",
    )

    @field_validator("name")
    @classmethod
    def validate_name(cls, v: str) -> str:
        """Ensure name is DNS-compatible."""
        if not re.match(r"^[a-z0-9-]+$", v):
            raise ValueError(
                "Cluster name must contain only lowercase letters, digits, and hyphens"
            )
        if len(v) > 63:
            raise ValueError("Cluster name must be <= 63 characters")
        return v

    @classmethod
    def from_yaml(cls, path: Path) -> "ClusterConfig":
        """Load config from a YAML file."""
        with path.open() as f:
            data = yaml.safe_load(f)
        return cls(**data)
