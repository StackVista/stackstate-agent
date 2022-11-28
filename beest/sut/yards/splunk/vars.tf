variable "aws_default_region" {
  default = "eu-west-1"
}
variable "yard_id" {
  type = string
}
variable "runners_ip" {
  type = string
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
