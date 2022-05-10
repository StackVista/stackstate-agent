output "api_url" {
  value = aws_api_gateway_deployment.deployment.invoke_url
}
output "bucket" {
  value = aws_s3_bucket.bucket.id
}
output "codepath" {
  value = aws_s3_object.code_zip.source
}
