"""Pydantic schemas for the /scenarios endpoints."""

from __future__ import annotations

from datetime import datetime
from enum import Enum

from pydantic import BaseModel, Field, field_validator


class FaultType(str, Enum):
    """Supported fault types (mirrors orchestrator/chaos/models.py)."""

    CPU_STRESS = "cpu_stress"
    NETWORK_LATENCY = "network_latency"
    MEMORY_PRESSURE = "memory_pressure"
    DISK_FILL = "disk_fill"


class ScenarioStatus(str, Enum):
    """Lifecycle status of a scenario run."""

    PENDING = "pending"
    RUNNING = "running"
    COMPLETED = "completed"
    FAILED = "failed"


class FaultSpec(BaseModel):
    """Single fault inside a scenario."""

    type: FaultType
    duration_seconds: int = Field(ge=1, le=3600)
    parameters: dict[str, str] = Field(default_factory=dict)


class ScenarioCreateRequest(BaseModel):
    """Body of POST /scenarios."""

    name: str = Field(min_length=1, max_length=63)
    description: str = ""
    faults: list[FaultSpec] = Field(min_length=1)

    @field_validator("name")
    @classmethod
    def validate_name(cls, v: str) -> str:
        import re

        if not re.match(r"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$", v):
            raise ValueError("name must be lowercase alphanumeric with optional hyphens")
        return v


class ScenarioResponse(BaseModel):
    """Response for scenario resources."""

    id: int
    name: str
    description: str
    status: ScenarioStatus
    created_at: datetime
    updated_at: datetime


class ScenarioRunResult(BaseModel):
    """Result of a scenario execution."""

    scenario_id: int
    status: ScenarioStatus
    targets_reached: int
    injections_count: int
    rollbacks_count: int
    message: str = ""
