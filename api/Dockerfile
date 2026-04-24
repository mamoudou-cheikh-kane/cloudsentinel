# syntax=docker/dockerfile:1.6

# -----------------------------------------------------------------------------
# Stage 1: build - install deps with Poetry 2.x (PEP 621 compatible)
# -----------------------------------------------------------------------------
FROM python:3.11-slim AS builder

ENV POETRY_VERSION=2.3.4 \
    POETRY_HOME=/opt/poetry \
    POETRY_VIRTUALENVS_IN_PROJECT=1 \
    POETRY_NO_INTERACTION=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1

RUN apt-get update \
 && apt-get install -y --no-install-recommends curl build-essential \
 && rm -rf /var/lib/apt/lists/*

RUN curl -sSL https://install.python-poetry.org | python3 - \
 && ln -s /opt/poetry/bin/poetry /usr/local/bin/poetry

WORKDIR /app

# Leverage Docker cache: copy only dependency files first.
COPY pyproject.toml poetry.lock ./
RUN poetry install --only main --no-root

# Now copy the app sources and install the project itself.
COPY app ./app
RUN poetry install --only main

# -----------------------------------------------------------------------------
# Stage 2: runtime - slim image with only what we need
# -----------------------------------------------------------------------------
FROM python:3.11-slim AS runtime

ENV PYTHONDONTWRITEBYTECODE=1 \
    PYTHONUNBUFFERED=1 \
    PATH="/app/.venv/bin:$PATH"

RUN groupadd --system app && useradd --system --gid app --home /home/app app \
 && mkdir -p /home/app /app \
 && chown -R app:app /home/app /app

WORKDIR /app

COPY --from=builder --chown=app:app /app/.venv /app/.venv
COPY --from=builder --chown=app:app /app/app /app/app

USER app

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD python -c "import urllib.request,sys; \
sys.exit(0 if urllib.request.urlopen('http://127.0.0.1:8000/healthz').status == 200 else 1)"

CMD ["uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8000"]
