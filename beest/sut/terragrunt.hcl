remote_state {
  backend = "s3"
  config = {
    region               = "eu-west-1"
    bucket               = get_env("BEEST_S3_BUCKET")
    key                  = format("%s_agentv%s/tf.tfstate", get_env("quay_user"), get_env("MAJOR_VERSION"))
    workspace_key_prefix = ""
    dynamodb_table       = get_env("BEEST_DYNAMODB_TABLE")
    encrypt              = true
    s3_bucket_tags = {
      owner = "beest"
      name  = "Terraform state storage"
    }
    dynamodb_table_tags = {
      owner = "beest"
      name  = "Terraform lock table"
    }
  }
}
