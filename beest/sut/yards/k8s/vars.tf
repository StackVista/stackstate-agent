variable "aws_default_region" {
  default = "eu-west-1"
}
variable "yard_id" {
  type = string
}
variable "agent_eks_version" {
  type = string
}
variable "agent_eks_runtime" {
  type = string
  validation {
    condition     = var.agent_eks_runtime == "dockerd" || var.agent_eks_runtime == "containerd"
    error_message = "The kubernetes container runtime can only be 'dockerd' or 'containerd'."
  }
}
variable "agent_eks_node_type" {
  default = "t3.xlarge"
}
variable "agent_eks_size" {
  default = 1
}
variable "runners_ip" {
  type = string
}

locals {
  az1                = "${var.aws_default_region}a"
  az2                = "${var.aws_default_region}b"
  agent_cluster_name = "${var.yard_id}-cluster"
  common_tags = {
    "Environment"                                       = var.yard_id
    "kubernetes.io/cluster/${local.agent_cluster_name}" = "shared"
  }
}
