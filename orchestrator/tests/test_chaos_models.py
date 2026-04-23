"""Tests for chaos scenario Pydantic models."""

import pytest
from cloudsentinel.chaos.models import (
    ChaosScenario,
    FaultSpec,
    FaultType,
    TargetSpec,
)
from pydantic import ValidationError


class TestTargetSpec:
    def test_default_selector_is_all(self):
        t = TargetSpec()
        assert t.selector == "all"
        assert t.labels == {}

    def test_invalid_selector_raises(self):
        with pytest.raises(ValidationError):
            TargetSpec(selector="invalid")

    def test_labels_selector(self):
        t = TargetSpec(selector="labels", labels={"role": "worker"})
        assert t.labels == {"role": "worker"}


class TestFaultSpec:
    def test_minimal_fault(self):
        f = FaultSpec(type=FaultType.CPU_STRESS, duration_seconds=60)
        assert f.type == FaultType.CPU_STRESS
        assert f.duration_seconds == 60
        assert f.parameters == {}

    def test_duration_must_be_positive(self):
        with pytest.raises(ValidationError):
            FaultSpec(type=FaultType.CPU_STRESS, duration_seconds=0)

    def test_duration_max(self):
        with pytest.raises(ValidationError):
            FaultSpec(type=FaultType.CPU_STRESS, duration_seconds=3601)


class TestChaosScenario:
    def test_minimal_scenario(self):
        s = ChaosScenario(
            name="test-scenario",
            faults=[FaultSpec(type=FaultType.CPU_STRESS, duration_seconds=30)],
        )
        assert s.name == "test-scenario"
        assert len(s.faults) == 1

    def test_invalid_name_raises(self):
        with pytest.raises(ValidationError):
            ChaosScenario(
                name="Invalid_Name",
                faults=[FaultSpec(type=FaultType.CPU_STRESS, duration_seconds=30)],
            )

    def test_empty_faults_raises(self):
        with pytest.raises(ValidationError):
            ChaosScenario(name="empty", faults=[])

    def test_from_yaml(self, tmp_path):
        yaml_content = """
name: test-cpu
description: A test scenario
targets:
  selector: all
faults:
  - type: cpu_stress
    duration_seconds: 60
    parameters:
      intensity: "50"
"""
        path = tmp_path / "scenario.yaml"
        path.write_text(yaml_content)
        s = ChaosScenario.from_yaml(path)
        assert s.name == "test-cpu"
        assert s.faults[0].parameters == {"intensity": "50"}
