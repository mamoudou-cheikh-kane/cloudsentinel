variable "cluster_name" {
  description = "Name of the local dev cluster"
  type        = string
  default     = "cloudsentinel-dev"
}

variable "kubernetes_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.30.0"
}

variable "worker_count" {
  description = "Number of worker nodes"
  type        = number
  default     = 2
}
