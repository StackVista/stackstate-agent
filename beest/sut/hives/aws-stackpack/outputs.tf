output "integration_role" {
  value = data.aws_iam_role.integration_role.arn
}

output "integration_profile" {
  value = aws_iam_instance_profile.integrations_profile.name
}

output "user_arn" {
  value = aws_iam_user.integration_user.arn
}

output "secret" {
  value = aws_iam_access_key.integration_user_key.secret
}

output "access_key_id" {
  value = aws_iam_access_key.integration_user_key.id
}
