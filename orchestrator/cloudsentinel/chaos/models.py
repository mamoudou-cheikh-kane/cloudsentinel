"""Pydantic models describing chaos scenarios."""

from __future__ import annotations

import re
from enum import Enum
from pathlib import Path

import yaml
from pydantic import BaseModel, Field, field_validator


class FaultType(str, Enum):
    """Supported fault types. Mirrors the protobuf enum."""

    CPU_STRESS = "cpu_stress"
    NETWORK_LATENCY = "network_latency"
    MEMORY_PRESSURE = "memory_pressure"
    DISK_FILL = "disk_fill"


class TargetSpec(BaseModel):
    """Specifies which agents should be targeted by a scenario."""

    selector: str = Field(
        default="all",
        description="Either 'all' to target every agent, or 'labels' to filter.",
    )
    labels: dict[str, str] = Field(
        default_factory=dict,
        description="Kubernetes labels to match when selector='labels'.",
    )

    @field_validator("selector")
    @classmethod
    def validate_selector(cls, v: str) -> str:
        if v not in ("all", "labels"):
            raise ValueError("selector must be 'all' or 'labels'")
        return v


class FaultSpec(BaseModel):
    """A single fault to inject."""

    type: FaultType
    duration_seconds: int = Field(ge=1, le=3600)
    parameters: dict[str, str] = Field(default_factory=dict)


class ChaosScenario(BaseModel):
    """A complete chaos scenario loaded from YAML."""

    name: str = Field(min_length=1, max_length=63)
    description: str = ""
    targets: TargetSpec = Field(default_factory=TargetSpec)
    faults: list[FaultSpec] = Field(min_length=1)

    @field_validator("name")
    @classmethod
    def validate_name(cls, v: str) -> str:
        if not re.match(r"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$", v):
            raise ValueError("name must be lowercase alphanumeric with optional hyphens")
        return v

    @classmethod
    def from_yaml(cls, path: str | Path) -> ChaosScenario:
        """Load a scenario from a YAML file."""
        with open(path, encoding="utf-8") as f:
            data = yaml.safe_load(f)
        return cls.model_validate(data)
