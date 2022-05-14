output "api_url" {
  value = module.otel.api_url
}

resource "local_file" "ansible_inventory" {
  filename        = "${path.module}/ansible_inventory"
  content         = <<INVENTORY

[local]
localhost ansible_connection=local

[all:vars]
yard_id=${var.yard_id}
bucket="${module.otel.bucket}"
code_zip="${module.otel.codepath}"
lambda_role_arn="${module.otel.lambda_role_arn}"
lambda_function_name="${module.otel.lambda_function_name}"
INVENTORY
  file_permission = "0777"
}
