output "integration_role" {
  value = aws_cloudformation_stack.cfn_stackpack.outputs.StackStateIntegrationRoleArn
}

output "integration_profile" {
  value = aws_iam_instance_profile.integrations_profile.name
}
