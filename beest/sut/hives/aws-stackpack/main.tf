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
  tags = {
    Environment           = var.environment
    VantaContainsUserData = false
    VantaDescription      = "AWS Integration resources used in acceptance pipeline"
    VantaNonProd          = true
    VantaOwner            = "stackstate@stackstate.com"
    VantaUserDataStored   = "NA"
  }

  template_url = "https://stackstate-integrations-resources-eu-west-1.s3.eu-west-1.amazonaws.com/aws-topology/cloudformation/stackstate-resources-1.3.cfn.yaml"
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
  name                 = "${var.environment}-integration-user"
  permissions_boundary = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:policy/DeveloperBoundaries"
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

resource "aws_iam_role_policy" "integration_role_policy" {
  name   = "${var.environment}-agent-ec2-policy"
  role   = aws_iam_role.agent_ec2_role.id
  policy = data.aws_iam_policy_document.integration_assume_role_policy.json
}

resource "aws_iam_instance_profile" "integration_profile" {
  name       = "${var.environment}-instance-profile"
  role       = aws_iam_role.agent_ec2_role.name
  depends_on = [aws_iam_role_policy.integration_role_policy]
}

# After the create phase has ran let's stop the event bridge and remove any objects (if there was any generated in the
# s3 bucket). This solves the problem where if the prepare phase never executes then there is at least no objects
# generated in the s3 bucket (This can happen when gitlab stops a previous build without executing the prepare phase).
# If the destroy should run without the prepare triggering then it can at least still
# cleanup. At this moment we do not require a event bridge as we are not testing against it, If in the future we do
# then we will have to rethink cleaning up
# At the moment with the event bridge stopped things can never break if the prepare phase is interrupted
resource "null_resource" "aws_cf_cleanup_existing_s3_resources" {
  depends_on = [
    aws_cloudformation_stack.cfn_stackpack
  ]
  # Let's always run this when the create is executed
  triggers = {
    always_run = timestamp()
  }
  provisioner "local-exec" {
    when = create
    interpreter = ["bash", "-c"]
    command = <<-EOT
      STS_EVENT_BRIDGE_RULE_RESOURCE_ID=$(aws cloudformation describe-stack-resource --stack-name "$STACK_NAME" --logical-resource-id StsEventBridgeRule --query "StackResourceDetail.PhysicalResourceId" --output=text)
      aws events disable-rule --name "$STS_EVENT_BRIDGE_RULE_RESOURCE_ID"
      STS_LOGS_BUCKET=$(aws cloudformation describe-stack-resource --stack-name "$STACK_NAME" --logical-resource-id StsLogsBucket --query "StackResourceDetail.PhysicalResourceId" --output=text)
      sleep 180
      STS_LOGS_BUCKET_OBJECTS=$(aws s3api list-object-versions --bucket "$STS_LOGS_BUCKET" --output=json --query='{Objects: Versions[].{Key:Key,VersionId:VersionId}}')
      aws s3api delete-objects --bucket "$STS_LOGS_BUCKET" --delete "$STS_LOGS_BUCKET_OBJECTS" --output=text || true
    EOT
    environment = {
      STACK_NAME = "${var.environment}-cfn-aws-check"
    }
  }
}
