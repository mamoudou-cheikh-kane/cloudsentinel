"""Tests for provider factory and base class."""

import pytest
from cloudsentinel.config import ClusterConfig, ProviderType
from cloudsentinel.providers import KindProvider, get_provider


class TestProviderFactory:
    def test_get_kind_provider(self) -> None:
        assert get_provider("kind") is KindProvider

    def test_get_unknown_provider_raises(self) -> None:
        with pytest.raises(ValueError, match="Unknown provider"):
            get_provider("inexistant")


class TestKindProvider:
    def test_instantiation(self) -> None:
        cfg = ClusterConfig(name="test", provider=ProviderType.KIND)
        provider = KindProvider(cfg)
        assert provider.config.name == "test"

    def test_kubeconfig_path(self) -> None:
        cfg = ClusterConfig(name="test-k", provider=ProviderType.KIND)
        provider = KindProvider(cfg)
        path = provider.get_kubeconfig()
        assert path.name == "test-k.kubeconfig"
        assert "cloudsentinel" in str(path)

    def test_repr(self) -> None:
        cfg = ClusterConfig(name="repr-test", provider=ProviderType.KIND)
        provider = KindProvider(cfg)
        assert "repr-test" in repr(provider)
        assert "kind" in repr(provider)

    def test_terraform_paths_exist(self) -> None:
        cfg = ClusterConfig(name="paths-test", provider=ProviderType.KIND)
        provider = KindProvider(cfg)
        assert provider.terraform_module_path.exists()
        assert provider.terraform_environment_path.exists()
