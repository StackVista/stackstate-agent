variable "yard_id" {
  type = string
}
variable "k8s_version" {
  type = string
}
variable "k8s_runtime" {
  type = string
  validation {
    condition     = var.k8s_runtime == "dockerd" || var.k8s_runtime == "containerd"
    error_message = "The kubernetes container runtime can only be 'dockerd' or 'containerd'."
  }
}
variable "k8s_node_type" {
  default = "t2.medium"
}
variable "k8s_size" {
  default = 2
}
variable "aws_default_region" {
  default = "eu-west-1"
}

locals {
  az1          = "${var.aws_default_region}a"
  az2          = "${var.aws_default_region}b"
  cluster_name = "${var.yard_id}-cluster"
  common_tags = {
    "Environment"                                 = var.yard_id
    "kubernetes.io/cluster/${local.cluster_name}" = "shared"
  }
}
