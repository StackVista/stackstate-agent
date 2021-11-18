#module "eks_eip" {
#  source = "../../library/tf-modules/eks-eip"
#
#  environment = var.environment
#}
#
#module "eks_vpc" {
#  source = "../../library/tf-modules/eks-vpc"
#
#  environment  = var.environment
#  az1          = local.az1
#  az2          = local.az2
#  common_tags  = local.common_tags
#  nat_eip_1_id = module.eks_eip.nat_eip_1_id
#  nat_eip_2_id = module.eks_eip.nat_eip_2_id
#}
#
#module "eks_cluster" {
#  source = "../../library/tf-modules/eks-cluster"
#
#  environment         = var.environment
#  vpc_id              = module.eks_vpc.vpc_id
#  private_subnet_1_id = module.eks_vpc.private_subnet_1_id
#  private_subnet_2_id = module.eks_vpc.private_subnet_2_id
#  k8s_cluster_name    = local.cluster_name
#  k8s_version         = var.k8s_version
#  k8s_node_type       = var.k8s_node_type
#  k8s_size            = var.k8s_size
#}

module "receiver" {
  source = "../../library/tf-modules/receiver"

  #  subnet_id = module.eks_vpc.private_subnet_1_id
  subnet_id   = "subnet-fa36adb2" // TODO temporary eu-west-1a, testing ansible connection
  environment = var.environment
}
