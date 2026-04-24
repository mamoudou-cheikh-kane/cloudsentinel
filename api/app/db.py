"""Database setup and session management."""

from __future__ import annotations

from collections.abc import Generator

from sqlmodel import Session, SQLModel, create_engine

DATABASE_URL = "sqlite:///./cloudsentinel.db"

# check_same_thread=False is safe because we use a single Session per request.
engine = create_engine(
    DATABASE_URL,
    connect_args={"check_same_thread": False},
    echo=False,
)


def init_db() -> None:
    """Create all tables. Called once at startup."""
    # Import models to register them with SQLModel metadata.
    from app.models import scenarios_db  # noqa: F401

    SQLModel.metadata.create_all(engine)


def get_session() -> Generator[Session, None, None]:
    """Dependency-injected DB session."""
    with Session(engine) as session:
        yield session
