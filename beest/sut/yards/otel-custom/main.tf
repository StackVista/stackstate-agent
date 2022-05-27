
// VPC, subnets
// eks cluster
// s3 bucket, API gateway, API trigger, IAM role, log group, sns topics, lambda
module "otel" {
  source = "../../hives/lambda-otel"

  environment = var.yard_id
}

// RDS

// ec2
module "ec2-stackpack" {
  source = "../../hives/ec2-stackpack"

}
