# Kind Cluster Module

Terraform module that provisions a local Kubernetes cluster using [kind](https://kind.sigs.k8s.io/).

## Purpose

Used for local development and CI testing in CloudSentinel. Provides a Kubernetes environment that behaves like a real cluster without cloud costs.

## Requirements

| Name | Version |
|------|---------|
| terraform | >= 1.5.0 |
| kind provider | ~> 0.7 |
| docker | running |

## Usage

```hcl
module "local_cluster" {
  source = "../../modules/kind"

  cluster_name       = "cloudsentinel-dev"
  kubernetes_version = "1.30.0"
  worker_count       = 3
}

output "kubeconfig" {
  value = module.local_cluster.kubeconfig_path
}
```

## Inputs

| Name | Type | Default | Description |
|------|------|---------|-------------|
| cluster_name | string | `cloudsentinel-local` | Name of the cluster |
| kubernetes_version | string | `1.30.0` | K8s version (no 'v' prefix) |
| worker_count | number | `2` | Number of worker nodes (0-10) |
| http_port | number | `8080` | Host port → 80 |
| https_port | number | `8443` | Host port → 443 |
| pod_subnet | string | `10.244.0.0/16` | Pod CIDR |
| service_subnet | string | `10.96.0.0/16` | Service CIDR |
| kubeconfig_dir | string | `~/.cloudsentinel/clusters` | Where to save kubeconfig |

## Outputs

| Name | Description |
|------|-------------|
| cluster_name | Cluster name |
| kubeconfig_path | Path to kubeconfig file |
| kubeconfig | Raw kubeconfig (sensitive) |
| endpoint | API server URL |
| worker_count | Actual worker count |

## Notes

- Pod and service subnets must not overlap with your host network
- Host ports 8080/8443 must be free
- Cluster creation takes ~60 seconds
- Use `terraform destroy` to clean up
