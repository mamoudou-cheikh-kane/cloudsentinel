# CloudSentinel — Kind module variables

variable "cluster_name" {
  description = "Name of the kind cluster"
  type        = string
  default     = "cloudsentinel-local"

  validation {
    condition     = can(regex("^[a-z0-9-]+$", var.cluster_name))
    error_message = "Cluster name must contain only lowercase letters, digits, and hyphens."
  }
}

variable "kubernetes_version" {
  description = "Kubernetes version (without the 'v' prefix)"
  type        = string
  default     = "1.30.0"
}

variable "worker_count" {
  description = "Number of worker nodes"
  type        = number
  default     = 2

  validation {
    condition     = var.worker_count >= 0 && var.worker_count <= 10
    error_message = "Worker count must be between 0 and 10."
  }
}

variable "http_port" {
  description = "Host port mapped to cluster ingress port 80"
  type        = number
  default     = 8080
}

variable "https_port" {
  description = "Host port mapped to cluster ingress port 443"
  type        = number
  default     = 8443
}

variable "pod_subnet" {
  description = "CIDR range for pod IPs"
  type        = string
  default     = "10.244.0.0/16"
}

variable "service_subnet" {
  description = "CIDR range for service IPs"
  type        = string
  default     = "10.96.0.0/16"
}

variable "kubeconfig_dir" {
  description = "Directory where kubeconfig files are stored"
  type        = string
  default     = "~/.cloudsentinel/clusters"
}
