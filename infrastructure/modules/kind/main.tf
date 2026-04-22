# CloudSentinel — Kind cluster Terraform module
#
# Provisions a local Kubernetes cluster using kind (Kubernetes in Docker).
# This module is used for local development and CI pipelines.
#
# Requirements:
#   - Docker running
#   - kind binary installed
#   - kubectl installed

terraform {
  required_version = ">= 1.5.0"

  required_providers {
    kind = {
      source  = "tehcyx/kind"
      version = "~> 0.7"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.5"
    }
  }
}

# -----------------------------------------------------------------------------
# Kind cluster
# -----------------------------------------------------------------------------
resource "kind_cluster" "this" {
  name            = var.cluster_name
  node_image      = "kindest/node:v${var.kubernetes_version}"
  wait_for_ready  = true
  kubeconfig_path = pathexpand("${var.kubeconfig_dir}/${var.cluster_name}.kubeconfig")

  kind_config {
    kind        = "Cluster"
    api_version = "kind.x-k8s.io/v1alpha4"

    # Control plane node
    node {
      role = "control-plane"

      kubeadm_config_patches = [
        <<-EOT
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "cloudsentinel.io/role=control-plane"
        EOT
      ]

      # Port mappings for ingress and metrics
      extra_port_mappings {
        container_port = 80
        host_port      = var.http_port
        protocol       = "TCP"
      }
      extra_port_mappings {
        container_port = 443
        host_port      = var.https_port
        protocol       = "TCP"
      }
    }

    # Worker nodes (one per count)
    dynamic "node" {
      for_each = range(var.worker_count)
      content {
        role = "worker"
        kubeadm_config_patches = [
          <<-EOT
          kind: JoinConfiguration
          nodeRegistration:
            kubeletExtraArgs:
              node-labels: "cloudsentinel.io/role=worker,cloudsentinel.io/worker-id=${node.value}"
          EOT
        ]
      }
    }

    # Networking configuration
    networking {
      api_server_address = "127.0.0.1"
      pod_subnet         = var.pod_subnet
      service_subnet     = var.service_subnet
    }
  }
}

# -----------------------------------------------------------------------------
# Ensure kubeconfig directory exists with correct permissions
# -----------------------------------------------------------------------------
resource "local_file" "kubeconfig_copy" {
  content  = kind_cluster.this.kubeconfig
  filename = pathexpand("${var.kubeconfig_dir}/${var.cluster_name}.kubeconfig")

  file_permission      = "0600"
  directory_permission = "0700"
}
