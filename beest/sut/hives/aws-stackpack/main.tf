data "aws_caller_identity" "current" {}

resource "random_integer" "cnf_postfix" {
  min = 1
  max = 100000
}

resource "aws_cloudformation_stack" "cfn_stackpack" {
  name = "${var.environment}-cfn-aws-check"
  parameters = {
    StsAccountId                = data.aws_caller_identity.current.account_id
    ExternalId                  = var.environment
    MainRegion                  = var.region
    IncludeOpenTelemetryTracing = var.include_open_telemetry_tracing
    PostFix                     = "-${random_integer.cnf_postfix.id}"
  }
  on_failure   = "DELETE"
  capabilities = ["CAPABILITY_NAMED_IAM"]

  // TODO  why the cloudformation template is hosted on a s3 bucket instead of being bundled as part of the stackpack resources and downloaded directly from the stackpack ?
  #  template_url = "https://stackstate-integrations-resources-eu-west-1.s3.eu-west-1.amazonaws.com/aws-topology/cloudformation/stackstate-resources-1.2.cfn.yaml"
  template_body = file("${path.module}/stackstate-resources-1.2.cfn.yaml")
}

data "aws_iam_role" "integration_role" {
  name = aws_cloudformation_stack.cfn_stackpack.outputs.StackStateIntegrationRole
}


data "aws_iam_policy_document" "integration_assume_role_policy" {
  statement {
    actions   = ["sts:AssumeRole"]
    resources = ["arn:aws:iam::*:role/${data.aws_iam_role.integration_role.name}"]
    effect    = "Allow"
  }
}

// for a IAM User
resource "aws_iam_user" "integration_user" {
  name = "${var.environment}-integration-user"
}

resource "aws_iam_access_key" "integration_user_key" {
  user = aws_iam_user.integration_user.name
}

resource "aws_iam_user_policy" "integration_user_policy" {
  name   = "${var.environment}-integration-user-policy"
  user   = aws_iam_user.integration_user.name
  policy = data.aws_iam_policy_document.integration_assume_role_policy.json
}

// for a EC2 instance
resource "aws_iam_role" "agent_ec2_role" {
  name = "${var.environment}-agent-ec2-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy" "test_policy" {
  name = "${var.environment}-agent-ec2-policy"
  role = aws_iam_role.agent_ec2_role.id
  policy = data.aws_iam_policy_document.integration_assume_role_policy.json
}

resource "aws_iam_instance_profile" "integrations_profile" {
  name = "${var.environment}-instance-profile"
  role = aws_iam_role.agent_ec2_role.name
}
