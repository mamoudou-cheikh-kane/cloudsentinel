# CloudSentinel - Local development & demo Makefile
#
# Usage:
#   make help           Show this help
#   make demo           Run the end-to-end demo on a local Kind cluster
#   make demo-clean     Tear down the demo cluster
#   make build-agent    Build the agent Docker image
#   make test-agent     Run the agent test suite
#   make test-all       Run every Go and Python test suite

.PHONY: help demo demo-clean build-agent test-agent test-all check-prereqs

CLUSTER_NAME ?= cloudsentinel-demo
AGENT_IMAGE  ?= cloudsentinel-agent:0.2.0

help:
	@echo "CloudSentinel - available targets"
	@echo ""
	@echo "  make demo           Spin up Kind, deploy the agent, show metrics"
	@echo "  make demo-clean     Delete the demo Kind cluster"
	@echo "  make build-agent    Build the agent Docker image ($(AGENT_IMAGE))"
	@echo "  make test-agent     Run the Go agent test suite"
	@echo "  make test-all       Run every test suite (Go + Python)"
	@echo "  make check-prereqs  Verify docker, kind, kubectl are installed"
	@echo ""

check-prereqs:
	@bash scripts/check-prereqs.sh

build-agent: check-prereqs
	@echo ">>> Building agent image $(AGENT_IMAGE)"
	@docker build -t $(AGENT_IMAGE) ./agent

demo: check-prereqs build-agent
	@bash scripts/demo.sh $(CLUSTER_NAME) $(AGENT_IMAGE)

demo-clean:
	@echo ">>> Deleting Kind cluster $(CLUSTER_NAME)"
	@kind delete cluster --name $(CLUSTER_NAME) || true

test-agent:
	@echo ">>> Running Go agent tests"
	@cd agent && go test ./... -v

test-all:
	@echo ">>> Running Go agent tests"
	@cd agent && go test ./... -v
	@echo ""
	@echo ">>> Running Python orchestrator tests"
	@cd orchestrator && poetry run pytest -v
	@echo ""
	@echo ">>> Running Python API tests"
	@cd api && poetry run pytest -v
	@echo ""
	@echo ">>> Running Python ML pipeline tests"
	@cd ml-pipeline && poetry run pytest -v
