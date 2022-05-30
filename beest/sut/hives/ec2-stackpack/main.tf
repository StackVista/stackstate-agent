data "aws_caller_identity" "current" {}

resource "aws_cloudformation_stack" "cfn_stackpack" {
  name = "cfn_stackpack"
  parameters = {
    StsAccountId = data.aws_caller_identity.current.account_id
    ExternalId = var.external_id
    MainRegion = var.region
    IncludeOpenTelemetryTracing = var.include_open_telemetry_tracing

  }
  template_body = file("${path.module}/cfn-aws-v2-stack.yml")
}


data "aws_iam_role" "awsv2-stackpack" {
  name = "StackStateAwsIntegrationRole"
  depends_on = [aws_cloudformation_stack.cfn_stackpack] #pr for cloudformation stack to output the role
}

data "aws_iam_policy_document" "policy_document" {
  statement {
    sid = "1"

    actions = [
      "sts:AssumeRole",
    ]
    resources = [
      "arn:aws:iam::*:role/StackStateAwsIntegrationRole",
    ]
  }
}

resource "aws_iam_instance_profile" "integrations_profile" {
  name = "ec2_profile"
  role = aws_iam_role.role
}

resource "aws_iam_role" "role" {
  name   = "ec2_role"
  path   = "/"
  assume_role_policy = data.aws_iam_policy_document.policy_document
}

data "aws_ami" "ubuntu" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "ec2-stackpack" {
  ami           = data.aws_ami.ubuntu.id
  instance_type = "t3.micro"
  iam_instance_profile = aws_iam_instance_profile.integrations_profile.name

}
