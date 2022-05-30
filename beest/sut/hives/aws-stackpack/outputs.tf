output "integration_role" {
  value = data.aws_iam_role.awsv2_stackpack.arn
}

output "integration_profile" {
  value = aws_iam_instance_profile.integrations_profile.name
}
