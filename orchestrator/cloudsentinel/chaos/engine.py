"""Chaos engine that orchestrates fault injection across agents."""

from __future__ import annotations

import logging
import time
from dataclasses import dataclass
from pathlib import Path

from kubernetes import client, config

from cloudsentinel.chaos.models import ChaosScenario
from cloudsentinel.grpc_client import AgentClient, AgentEndpoint
from cloudsentinel.grpc_client.client import InjectResult, RollbackResult

logger = logging.getLogger(__name__)

DEFAULT_NAMESPACE = "cloudsentinel"
DEFAULT_LABEL_SELECTOR = "app.kubernetes.io/name=cloudsentinel-agent"


@dataclass
class ScenarioResult:
    """Outcome of a chaos scenario run."""

    scenario_name: str
    targets_reached: int
    injections: list[InjectResult]
    rollbacks: list[RollbackResult]


class ChaosEngine:
    """Coordinates fault injection across a fleet of agents."""

    def __init__(
        self,
        kubeconfig_path: str | Path | None = None,
        namespace: str = DEFAULT_NAMESPACE,
    ) -> None:
        if kubeconfig_path:
            config.load_kube_config(config_file=str(kubeconfig_path))
        else:
            config.load_kube_config()
        self.k8s = client.CoreV1Api()
        self.namespace = namespace

    def discover_agents(self, label_selector: str = DEFAULT_LABEL_SELECTOR) -> list[AgentEndpoint]:
        """Find running agent pods and return their gRPC endpoints."""
        pods = self.k8s.list_namespaced_pod(
            namespace=self.namespace,
            label_selector=label_selector,
        )
        endpoints = []
        for pod in pods.items:
            if pod.status.phase != "Running" or not pod.status.pod_ip:
                logger.warning(
                    "skipping pod %s (phase=%s, ip=%s)",
                    pod.metadata.name,
                    pod.status.phase,
                    pod.status.pod_ip,
                )
                continue
            endpoints.append(
                AgentEndpoint(
                    node_name=pod.spec.node_name or pod.metadata.name,
                    address=f"{pod.status.pod_ip}:50051",
                )
            )
        return endpoints

    def run_scenario(
        self,
        scenario: ChaosScenario,
        wait_for_completion: bool = True,
    ) -> ScenarioResult:
        """Execute a chaos scenario end-to-end."""
        logger.info("discovering agents in namespace=%s", self.namespace)
        endpoints = self.discover_agents()
        if not endpoints:
            raise RuntimeError(f"no running agents found in namespace {self.namespace}")

        logger.info("found %d agent(s)", len(endpoints))
        targeted = endpoints

        injections: list[InjectResult] = []
        for fault in scenario.faults:
            for endpoint in targeted:
                grpc_client = AgentClient(endpoint)
                try:
                    result = grpc_client.inject_fault(
                        fault_type=fault.type,
                        duration_seconds=fault.duration_seconds,
                        parameters=fault.parameters,
                    )
                    injections.append(result)
                    logger.info(
                        "injected on %s: fault_id=%s accepted=%s",
                        endpoint.node_name,
                        result.fault_id,
                        result.accepted,
                    )
                except Exception as e:
                    logger.exception("inject failed on %s: %s", endpoint.node_name, e)

        if wait_for_completion and scenario.faults:
            max_duration = max(f.duration_seconds for f in scenario.faults)
            logger.info("waiting %ds for faults to complete", max_duration)
            time.sleep(max_duration + 1)

        rollbacks: list[RollbackResult] = []
        for injection in injections:
            if not injection.accepted:
                continue
            endpoint = next(
                (e for e in targeted if e.node_name == injection.node_name),
                None,
            )
            if endpoint is None:
                continue
            grpc_client = AgentClient(endpoint)
            try:
                rb = grpc_client.rollback(injection.fault_id)
                rollbacks.append(rb)
                logger.info(
                    "rolled back on %s: fault_id=%s success=%s",
                    endpoint.node_name,
                    rb.fault_id,
                    rb.success,
                )
            except Exception as e:
                logger.exception("rollback failed on %s: %s", endpoint.node_name, e)

        return ScenarioResult(
            scenario_name=scenario.name,
            targets_reached=len(targeted),
            injections=injections,
            rollbacks=rollbacks,
        )
