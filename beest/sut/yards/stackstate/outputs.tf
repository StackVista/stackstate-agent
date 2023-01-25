resource "local_file" "ansible_inventory" {
  filename = "${path.module}/ansible_inventory"
  content = yamlencode({
    all : {
      hosts : {
        local : {
          ansible_host : "localhost"
          ansible_connection : "local"
        }
      }
      vars : {
        yard_id : var.yard_id
      }
    }
  })
  file_permission = "0777"
}


resource "local_file" "get_kubeconfig" {
  filename        = "${path.module}/get_kubeconfig.sh"
  content         = <<KUBECONFIG
#!/bin/bash
set -euo pipefail

echo "Get kubeconfig for stackstate sandbox cluster"
sts-toolbox cluster connect sandbox-main.sandbox.stackstate.io
KUBECONFIG
  file_permission = "0770"
}
