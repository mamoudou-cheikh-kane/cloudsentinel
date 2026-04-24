"""Tests for the root and healthz endpoints."""


def test_root_returns_service_info(client):
    response = client.get("/")
    assert response.status_code == 200
    body = response.json()
    assert body["service"] == "cloudsentinel-api"
    assert "version" in body
    assert body["docs"] == "/docs"


def test_healthz_returns_ok(client):
    response = client.get("/healthz")
    assert response.status_code == 200
    assert response.json() == {"status": "ok"}
