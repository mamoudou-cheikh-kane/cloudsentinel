"""Cloud provider implementations."""

from cloudsentinel.providers.base import CloudProvider, ProvisionResult
from cloudsentinel.providers.kind import KindProvider

__all__ = ["CloudProvider", "KindProvider", "ProvisionResult"]


def get_provider(provider_type: str) -> type[CloudProvider]:
    """Factory: return provider class by name."""
    providers: dict[str, type[CloudProvider]] = {
        "kind": KindProvider,
    }
    if provider_type not in providers:
        raise ValueError(f"Unknown provider '{provider_type}'. Available: {list(providers)}")
    return providers[provider_type]
