"""Tests for the /scenarios endpoints."""


def _scenario_payload(name: str = "demo-cpu", duration: int = 30) -> dict:
    return {
        "name": name,
        "description": "test scenario",
        "faults": [
            {"type": "cpu_stress", "duration_seconds": duration},
        ],
    }


class TestListScenarios:
    def test_initial_list_is_empty(self, client):
        response = client.get("/scenarios")
        assert response.status_code == 200
        assert response.json() == []


class TestCreateScenario:
    def test_valid_scenario_is_created(self, client):
        response = client.post("/scenarios", json=_scenario_payload())
        assert response.status_code == 201
        body = response.json()
        assert body["name"] == "demo-cpu"
        assert body["status"] == "pending"
        assert body["id"] >= 1

    def test_invalid_name_is_rejected(self, client):
        payload = _scenario_payload(name="Invalid_Name")
        response = client.post("/scenarios", json=payload)
        assert response.status_code == 422

    def test_empty_faults_is_rejected(self, client):
        payload = {"name": "demo", "faults": []}
        response = client.post("/scenarios", json=payload)
        assert response.status_code == 422

    def test_duration_out_of_range_is_rejected(self, client):
        payload = _scenario_payload(duration=0)
        response = client.post("/scenarios", json=payload)
        assert response.status_code == 422


class TestGetScenario:
    def test_get_existing_scenario(self, client):
        created = client.post("/scenarios", json=_scenario_payload()).json()
        response = client.get(f"/scenarios/{created['id']}")
        assert response.status_code == 200
        assert response.json()["name"] == "demo-cpu"

    def test_get_missing_scenario_returns_404(self, client):
        response = client.get("/scenarios/9999")
        assert response.status_code == 404


class TestRunScenario:
    def test_run_transitions_status_to_completed(self, client):
        created = client.post("/scenarios", json=_scenario_payload()).json()
        scenario_id = created["id"]
        assert created["status"] == "pending"

        run_response = client.post(f"/scenarios/{scenario_id}/run")
        assert run_response.status_code == 200
        assert run_response.json()["status"] == "completed"

        fetched = client.get(f"/scenarios/{scenario_id}").json()
        assert fetched["status"] == "completed"

    def test_run_missing_scenario_returns_404(self, client):
        response = client.post("/scenarios/9999/run")
        assert response.status_code == 404
