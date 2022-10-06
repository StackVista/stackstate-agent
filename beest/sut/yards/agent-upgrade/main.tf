module "vpc" {
  //TODO rename vpc module
  source = "../../hives/eks-vpc"

  environment = var.yard_id
  az1         = local.az1
  az2         = local.az2
  common_tags = local.common_tags
}

module "ec2_agent_redhat" {
  source = "../../hives/ec2-agent/v2/redhat"
  environment         = var.yard_id
  vpc_id              = module.vpc.vpc_id
  subnet_id           = module.vpc.private_subnet_1_id
  integration_profile = ""
}

module "ec2_agent_ubuntu" {
  source = "../../hives/ec2-agent/v2/ubuntu"
  environment         = var.yard_id
  vpc_id              = module.vpc.vpc_id
  subnet_id           = module.vpc.private_subnet_1_id
  integration_profile = ""
}
