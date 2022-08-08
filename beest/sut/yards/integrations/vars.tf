variable "yard_id" {
  type = string
}

locals {
  cluster_name = "${var.yard_id}-cluster"
  common_tags = {
    "Environment"                                 = var.yard_id
    "kubernetes.io/cluster/${local.cluster_name}" = "shared"
  }
}
