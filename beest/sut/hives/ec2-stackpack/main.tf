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

}
