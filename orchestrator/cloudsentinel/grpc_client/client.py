"""Thin gRPC client wrapper for the CloudSentinel agent."""

from __future__ import annotations

from contextlib import contextmanager
from dataclasses import dataclass

import grpc

from cloudsentinel.chaos.models import FaultType
from cloudsentinel.grpc_client.pb import agent_pb2, agent_pb2_grpc

_FAULT_TYPE_MAP = {
    FaultType.CPU_STRESS: agent_pb2.FAULT_TYPE_CPU_STRESS,
    FaultType.NETWORK_LATENCY: agent_pb2.FAULT_TYPE_NETWORK_LATENCY,
    FaultType.MEMORY_PRESSURE: agent_pb2.FAULT_TYPE_MEMORY_PRESSURE,
    FaultType.DISK_FILL: agent_pb2.FAULT_TYPE_DISK_FILL,
}


@dataclass
class AgentEndpoint:
    """Represents a reachable agent instance."""

    node_name: str
    address: str


@dataclass
class InjectResult:
    """Result of an injection call."""

    node_name: str
    fault_id: str
    accepted: bool
    message: str


@dataclass
class RollbackResult:
    """Result of a rollback call."""

    node_name: str
    fault_id: str
    success: bool
    message: str


class AgentClient:
    """High-level gRPC client for a single agent endpoint."""

    def __init__(self, endpoint: AgentEndpoint, timeout_seconds: float = 10.0) -> None:
        self.endpoint = endpoint
        self.timeout = timeout_seconds

    @contextmanager
    def _channel(self):
        channel = grpc.insecure_channel(self.endpoint.address)
        try:
            yield channel
        finally:
            channel.close()

    def health(self) -> dict[str, str | int]:
        with self._channel() as ch:
            stub = agent_pb2_grpc.AgentServiceStub(ch)
            resp = stub.Health(agent_pb2.HealthRequest(), timeout=self.timeout)
            return {
                "node": resp.node,
                "version": resp.version,
                "uptime_seconds": resp.uptime_seconds,
            }

    def inject_fault(
        self,
        fault_type: FaultType,
        duration_seconds: int,
        parameters: dict[str, str] | None = None,
    ) -> InjectResult:
        with self._channel() as ch:
            stub = agent_pb2_grpc.AgentServiceStub(ch)
            req = agent_pb2.InjectFaultRequest(
                type=_FAULT_TYPE_MAP[fault_type],
                duration_seconds=duration_seconds,
                parameters=parameters or {},
            )
            resp = stub.InjectFault(req, timeout=self.timeout)
            return InjectResult(
                node_name=self.endpoint.node_name,
                fault_id=resp.fault_id,
                accepted=resp.accepted,
                message=resp.message,
            )

    def rollback(self, fault_id: str) -> RollbackResult:
        with self._channel() as ch:
            stub = agent_pb2_grpc.AgentServiceStub(ch)
            req = agent_pb2.RollbackRequest(fault_id=fault_id)
            resp = stub.Rollback(req, timeout=self.timeout)
            return RollbackResult(
                node_name=self.endpoint.node_name,
                fault_id=fault_id,
                success=resp.success,
                message=resp.message,
            )
