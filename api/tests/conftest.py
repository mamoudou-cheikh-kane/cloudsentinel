"""Shared pytest fixtures."""

from __future__ import annotations

from collections.abc import Generator

import pytest
from app.db import get_session
from app.main import app
from fastapi.testclient import TestClient
from sqlalchemy.pool import StaticPool
from sqlmodel import Session, SQLModel, create_engine


@pytest.fixture(name="session")
def session_fixture() -> Generator[Session, None, None]:
    """Create an in-memory DB and yield a session."""
    engine = create_engine(
        "sqlite://",
        connect_args={"check_same_thread": False},
        poolclass=StaticPool,
    )
    # Import models so the metadata is populated before create_all.
    from app.models import scenarios_db  # noqa: F401

    SQLModel.metadata.create_all(engine)
    with Session(engine) as session:
        yield session


@pytest.fixture(name="client")
def client_fixture(session: Session) -> Generator[TestClient, None, None]:
    """Return a TestClient whose DB session is the in-memory fixture."""

    def _override_get_session() -> Generator[Session, None, None]:
        yield session

    app.dependency_overrides[get_session] = _override_get_session
    with TestClient(app) as client:
        yield client
    app.dependency_overrides.clear()
