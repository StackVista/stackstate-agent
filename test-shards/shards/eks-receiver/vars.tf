variable "environment" {}
variable "k8s_version" {}
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
  cluster_name = "${var.environment}-cluster"
  common_tags = {
    "Environment"                                 = var.environment
    "kubernetes.io/cluster/${local.cluster_name}" = "shared"
  }
}
