resource "local_file" "agent_eks_aws_auth" {
  filename        = "${path.module}/agent_eks_aws_auth.yml"
  content         = module.eks_cluster.k8s_aws_auth
  file_permission = "0600"
}

resource "local_file" "get_kubeconfig" {
  filename        = "${path.module}/get_kubeconfig.sh"
  content         = <<KUBECONFIG
#!/bin/bash
set -euo pipefail

echo "Get kubeconfig for agent cluster ${local.agent_cluster_name}"
aws eks update-kubeconfig --name ${local.agent_cluster_name} --alias ${var.yard_id}

echo "Configure kubernetes agent cluster AWS authentication"
kubectl --context ${var.yard_id} apply -f ${local_file.agent_eks_aws_auth.filename}

echo "Get kubeconfig for stackstate sandbox cluster"
sts-toolbox cluster connect sandbox-main.sandbox.stackstate.io
KUBECONFIG
  file_permission = "0770"
}

//

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
        agent_k8s_runtime : var.agent_eks_runtime
        agent_k8s_version : var.agent_eks_version
      }
    }
  })
  file_permission = "0777"
}

resource "local_file" "agent_eks_node_id_rsa" {
  filename        = "${path.module}/agent_eks_node_id_rsa"
  content         = module.eks_cluster.node_ssh_key
  file_permission = "0600"
}
