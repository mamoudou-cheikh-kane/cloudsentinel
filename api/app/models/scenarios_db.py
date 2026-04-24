"""SQLModel entities persisted in the database."""

from __future__ import annotations

import json
from datetime import UTC, datetime

from sqlmodel import Field, SQLModel


class ScenarioDB(SQLModel, table=True):
    """Scenario row in SQLite."""

    __tablename__ = "scenarios"

    id: int | None = Field(default=None, primary_key=True)
    name: str = Field(index=True, max_length=63)
    description: str = ""
    status: str = Field(default="pending", index=True)
    faults_json: str = Field(default="[]")
    created_at: datetime = Field(default_factory=lambda: datetime.now(UTC))
    updated_at: datetime = Field(default_factory=lambda: datetime.now(UTC))

    def faults_list(self) -> list[dict]:
        """Decode the faults JSON blob."""
        return json.loads(self.faults_json)

    def set_faults(self, faults: list[dict]) -> None:
        """Encode a list of faults into JSON for storage."""
        self.faults_json = json.dumps(faults)
