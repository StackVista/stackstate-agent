module "eks_vpc" {
  source = "../../hives/eks-vpc"

  environment = var.yard_id
  az1         = local.az1
  az2         = local.az2
  common_tags = local.common_tags
}

module "eks_cluster" {
  source = "../../hives/eks-cluster"

  environment         = var.yard_id
  vpc_id              = module.eks_vpc.vpc_id
  private_subnet_1_id = module.eks_vpc.private_subnet_1_id
  private_subnet_2_id = module.eks_vpc.private_subnet_2_id
  k8s_cluster_name    = local.cluster_name
  k8s_version         = var.k8s_version
  k8s_runtime         = var.k8s_runtime
  k8s_node_type       = var.k8s_node_type
  k8s_size            = var.k8s_size
}

module "receiver" {
  source = "../../hives/receiver"

  vpc_id      = module.eks_vpc.vpc_id
  subnet_id   = module.eks_vpc.private_subnet_1_id
  environment = var.yard_id
}
