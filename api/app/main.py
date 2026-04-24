"""FastAPI application entry point for CloudSentinel."""

from contextlib import asynccontextmanager

from fastapi import FastAPI

from app import __version__
from app.db import init_db
from app.routers import scenarios_router


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Run at startup: create the DB schema."""
    init_db()
    yield


app = FastAPI(
    title="CloudSentinel API",
    description="HTTP API for managing chaos engineering scenarios.",
    version=__version__,
    contact={
        "name": "CloudSentinel",
        "url": "https://github.com/mamoudou-cheikh-kane/cloudsentinel",
    },
    lifespan=lifespan,
)

app.include_router(scenarios_router)


@app.get("/", tags=["meta"])
async def root() -> dict[str, str]:
    """Return a welcome message."""
    return {
        "service": "cloudsentinel-api",
        "version": __version__,
        "docs": "/docs",
    }


@app.get("/healthz", tags=["meta"])
async def healthz() -> dict[str, str]:
    """Liveness probe."""
    return {"status": "ok"}
