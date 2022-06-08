data "aws_caller_identity" "current" {}

resource "aws_cloudformation_stack" "cfn_stackpack" {
  name = "${var.environment}-cfn-aws-check"
  parameters = {
    StsAccountId                = data.aws_caller_identity.current.account_id
    ExternalId                  = var.environment
    MainRegion                  = var.region
    IncludeOpenTelemetryTracing = var.include_open_telemetry_tracing
  }
  // TODO  why the cloudformation template is hosted on a s3 bucket instead of being bundled as part of the stackpack resources and downloaded directly from the stackpack ?
  template_url = "https://stackstate-integrations-resources-eu-west-1.s3.eu-west-1.amazonaws.com/aws-topology/cloudformation/stackstate-resources-1.2.cfn.yaml"
}

data "aws_iam_role" "awsv2_stackpack" {
  name       = "StackStateAwsIntegrationRole"
  depends_on = [aws_cloudformation_stack.cfn_stackpack] # TODO PR for cloudformation stack to output the role
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
  name = "${var.environment}-instance-profile"
  role = aws_iam_role.role.name
}

resource "aws_iam_role" "role" {
  name               = "${var.environment}-ec2-role"
  assume_role_policy = data.aws_iam_policy_document.policy_document.json
}
