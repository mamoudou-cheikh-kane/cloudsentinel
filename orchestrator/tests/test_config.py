"""Tests for ClusterConfig validation."""

from pathlib import Path

import pytest
from cloudsentinel.config import ClusterConfig, ProviderType
from pydantic import ValidationError


class TestClusterConfig:
    def test_valid_config(self) -> None:
        cfg = ClusterConfig(name="my-cluster", provider=ProviderType.KIND)
        assert cfg.name == "my-cluster"
        assert cfg.provider == ProviderType.KIND
        assert cfg.worker_count == 2
        assert cfg.kubernetes_version == "1.30.0"

    def test_invalid_name_uppercase(self) -> None:
        with pytest.raises(ValidationError):
            ClusterConfig(name="MyCluster", provider=ProviderType.KIND)

    def test_invalid_name_special_chars(self) -> None:
        with pytest.raises(ValidationError):
            ClusterConfig(name="my_cluster!", provider=ProviderType.KIND)

    def test_name_too_long(self) -> None:
        with pytest.raises(ValidationError):
            ClusterConfig(name="a" * 64, provider=ProviderType.KIND)

    def test_worker_count_too_high(self) -> None:
        with pytest.raises(ValidationError):
            ClusterConfig(name="test", provider=ProviderType.KIND, worker_count=20)

    def test_worker_count_negative(self) -> None:
        with pytest.raises(ValidationError):
            ClusterConfig(name="test", provider=ProviderType.KIND, worker_count=-1)

    def test_from_yaml(self, tmp_path: Path) -> None:
        yaml_file = tmp_path / "cluster.yaml"
        yaml_file.write_text("name: test-cluster\nprovider: kind\nworker_count: 3\n")
        cfg = ClusterConfig.from_yaml(yaml_file)
        assert cfg.name == "test-cluster"
        assert cfg.worker_count == 3

    def test_default_values(self) -> None:
        cfg = ClusterConfig(name="default-test", provider=ProviderType.KIND)
        assert cfg.worker_count == 2
        assert cfg.kubernetes_version == "1.30.0"
        assert cfg.provider_options == {}
