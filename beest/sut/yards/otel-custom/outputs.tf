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
INVENTORY
  file_permission = "0777"
}
