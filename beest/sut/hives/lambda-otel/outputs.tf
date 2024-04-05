output "api_url" {
  value = aws_api_gateway_deployment.deployment.invoke_url
}
output "bucket" {
  value = aws_s3_bucket.bucket.id
}
output "codepath" {
  value = aws_s3_object.code_zip.source
}
output "lambda_function_name" {
  value = aws_lambda_function.lambda.function_name
}
output "lambda_role_arn" {
  value = aws_iam_role.iam_for_lambda.arn
}
