"""REST endpoints for chaos scenarios (SQLite-backed)."""

from __future__ import annotations

from datetime import UTC, datetime

from fastapi import APIRouter, Depends, HTTPException, status
from sqlmodel import Session, select

from app.db import get_session
from app.models import (
    ScenarioCreateRequest,
    ScenarioResponse,
    ScenarioRunResult,
    ScenarioStatus,
)
from app.models.scenarios_db import ScenarioDB

router = APIRouter(prefix="/scenarios", tags=["scenarios"])


def _to_response(record: ScenarioDB) -> ScenarioResponse:
    """Convert a DB row into an API response."""
    return ScenarioResponse(
        id=record.id,
        name=record.name,
        description=record.description,
        status=ScenarioStatus(record.status),
        created_at=record.created_at,
        updated_at=record.updated_at,
    )


@router.get("", response_model=list[ScenarioResponse])
async def list_scenarios(
    session: Session = Depends(get_session),
) -> list[ScenarioResponse]:
    """List all known scenarios."""
    records = session.exec(select(ScenarioDB).order_by(ScenarioDB.id)).all()
    return [_to_response(r) for r in records]


@router.post("", response_model=ScenarioResponse, status_code=status.HTTP_201_CREATED)
async def create_scenario(
    payload: ScenarioCreateRequest,
    session: Session = Depends(get_session),
) -> ScenarioResponse:
    """Create a new scenario (status = pending)."""
    record = ScenarioDB(
        name=payload.name,
        description=payload.description,
        status=ScenarioStatus.PENDING.value,
    )
    record.set_faults([f.model_dump() for f in payload.faults])
    session.add(record)
    session.commit()
    session.refresh(record)
    return _to_response(record)


@router.get("/{scenario_id}", response_model=ScenarioResponse)
async def get_scenario(
    scenario_id: int,
    session: Session = Depends(get_session),
) -> ScenarioResponse:
    """Retrieve a single scenario."""
    record = session.get(ScenarioDB, scenario_id)
    if record is None:
        raise HTTPException(status_code=404, detail=f"scenario {scenario_id} not found")
    return _to_response(record)


@router.post("/{scenario_id}/run", response_model=ScenarioRunResult)
async def run_scenario(
    scenario_id: int,
    session: Session = Depends(get_session),
) -> ScenarioRunResult:
    """Trigger a scenario run (simulated for now)."""
    record = session.get(ScenarioDB, scenario_id)
    if record is None:
        raise HTTPException(status_code=404, detail=f"scenario {scenario_id} not found")

    record.status = ScenarioStatus.COMPLETED.value
    record.updated_at = datetime.now(UTC)
    session.add(record)
    session.commit()

    return ScenarioRunResult(
        scenario_id=scenario_id,
        status=ScenarioStatus.COMPLETED,
        targets_reached=0,
        injections_count=0,
        rollbacks_count=0,
        message="Execution is simulated; real chaos engine integration is wired in a follow-up.",
    )
