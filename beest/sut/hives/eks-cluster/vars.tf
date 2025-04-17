variable "arch" {
  description = "Architecture type (amd64 or arm64)"
  type        = string
  default     = "amd64"

  validation {
    condition     = contains(["amd64", "arm64"], var.arch)
    error_message = "arch must be either 'amd64' or 'arm64'."
  }
}

variable "environment" {}
variable "vpc_id" {}
variable "private_subnet_1_id" {}
variable "private_subnet_2_id" {}
variable "k8s_cluster_name" {}
variable "k8s_version" {}
variable "k8s_runtime" {}
variable "k8s_node_type" {}
variable "k8s_size" {}
