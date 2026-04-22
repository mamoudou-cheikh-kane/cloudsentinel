# CloudSentinel — Development environment
#
# Provisions a local kind cluster for day-to-day development.

terraform {
  required_version = ">= 1.5.0"

  # For dev, we keep state local. Production envs will use S3+DynamoDB.
  backend "local" {
    path = "terraform.tfstate"
  }
}

module "local_cluster" {
  source = "../../modules/kind"

  cluster_name       = var.cluster_name
  kubernetes_version = var.kubernetes_version
  worker_count       = var.worker_count
}

output "cluster_name" {
  value = module.local_cluster.cluster_name
}

output "kubeconfig_path" {
  value = module.local_cluster.kubeconfig_path
}
