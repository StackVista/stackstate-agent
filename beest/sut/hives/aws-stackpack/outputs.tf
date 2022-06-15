output "integration_role" {
  value = data.aws_iam_role.integration_role.arn
}

output "integration_profile" {
  value = aws_iam_instance_profile.integrations_profile.name
}
