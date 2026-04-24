"""FastAPI application entry point for CloudSentinel."""

from contextlib import asynccontextmanager

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware

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

# Allow the dashboard (http://localhost:3000) to call this API from the browser.
app.add_middleware(
    CORSMiddleware,
    allow_origins=[
        "http://localhost:3000",
        "http://127.0.0.1:3000",
    ],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
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
