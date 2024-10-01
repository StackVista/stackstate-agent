resource "aws_s3_bucket" "bucket" {
  bucket        = "${var.environment}-lambda-code"
  force_destroy = true

  tags = {
    Environment           = var.environment
    VantaContainsUserData = false
    VantaDescription      = "OpenTelemetry Integration resources used in acceptance pipeline"
    VantaNonProd          = true
    VantaOwner            = "stackstate@stackstate.com"
    VantaUserDataStored   = "NA"
  }
}

resource "aws_s3_bucket_acl" "bucket_acl" {
  bucket = aws_s3_bucket.bucket.id
  acl    = "private"
}

// TODO zip the file
resource "aws_s3_object" "code_zip" {
  bucket = aws_s3_bucket.bucket.id
  key    = "hello3.zip"
  source = "${path.module}/hello.zip"
  etag   = filemd5("${path.module}/hello.zip")
}

resource "aws_sns_topic" "lambda_errors" {
  name = "${var.environment}-errors"
}

data "aws_iam_policy_document" "lambda_assume" {
  policy_id = "${var.environment}-lambda-role"

  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "lambda_to_sns" {
  statement {
    effect    = "Allow"
    actions   = ["SNS:Publish"]
    resources = ["arn:aws:sns:*:*:*"]
  }
}

resource "aws_iam_policy" "lambda_to_sns" {
  name   = "${var.environment}-lambda-sns"
  policy = data.aws_iam_policy_document.lambda_to_sns.json
}

resource "aws_iam_role" "iam_for_lambda" {
  name = "${var.environment}-iam-lambda"

  assume_role_policy = data.aws_iam_policy_document.lambda_assume.json
}

resource "aws_iam_policy_attachment" "logs_sns_policy_attach" {
  name       = "${var.environment}-lambda-sns"
  roles      = [aws_iam_role.iam_for_lambda.name]
  policy_arn = aws_iam_policy.lambda_to_sns.arn
}

resource "aws_lambda_function" "lambda" {
  function_name = "${var.environment}-hello"
  s3_bucket     = aws_s3_bucket.bucket.id
  s3_key        = aws_s3_object.code_zip.key

  role = aws_iam_role.iam_for_lambda.arn

  handler = "hello.handler"
  runtime = "nodejs12.x"

  source_code_hash = filebase64sha256(aws_s3_object.code_zip.source)

  dead_letter_config {
    target_arn = aws_sns_topic.lambda_errors.arn
  }
}

resource "aws_lambda_permission" "permission" {
  statement_id  = "AllowAPIGatewayInvoke"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.lambda.function_name
  principal     = "apigateway.amazonaws.com"

  # The /*/*/* part allows invocation from any stage, method and resource path
  # within API Gateway REST API.
  source_arn = "${aws_api_gateway_rest_api.api.execution_arn}/*/*/*"
}

