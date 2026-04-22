# CloudSentinel — Kind module outputs

output "cluster_name" {
  description = "Name of the created cluster"
  value       = kind_cluster.this.name
}

output "kubeconfig_path" {
  description = "Path to the kubeconfig file"
  value       = local_file.kubeconfig_copy.filename
}

output "kubeconfig" {
  description = "Raw kubeconfig content (sensitive)"
  value       = kind_cluster.this.kubeconfig
  sensitive   = true
}

output "endpoint" {
  description = "Kubernetes API server endpoint"
  value       = kind_cluster.this.endpoint
}

output "client_certificate" {
  description = "Client certificate (PEM)"
  value       = kind_cluster.this.client_certificate
  sensitive   = true
}

output "client_key" {
  description = "Client private key (PEM)"
  value       = kind_cluster.this.client_key
  sensitive   = true
}

output "cluster_ca_certificate" {
  description = "Cluster CA certificate (PEM)"
  value       = kind_cluster.this.cluster_ca_certificate
  sensitive   = true
}

output "worker_count" {
  description = "Number of worker nodes created"
  value       = var.worker_count
}
