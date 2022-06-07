module "vpc" {
  //TODO rename vpc module
  source = "../../hives/eks-vpc"

  environment = var.yard_id
  az1         = local.az1
  az2         = local.az2
  common_tags = local.common_tags
}

// eks cluster

module "lambda_otel" {
  source = "../../hives/lambda-otel"

  environment = var.yard_id
  // TODO pass vpc
}

// RDS

module "aws_stackpack_role" {
  source = "../../hives/aws-stackpack"

  environment = var.yard_id
}

module "ec2_agent" {
  source = "../../hives/ec2-agent"

  environment         = var.yard_id
  vpc_id              = module.vpc.vpc_id
  subnet_id           = module.vpc.private_subnet_1_id
  integration_profile = module.aws_stackpack_role.integration_profile
}
