resource "local_file" "k8s_aws_auth" {
  filename        = "${path.module}/k8s_aws_auth.yml"
  content         = module.eks_cluster.k8s_aws_auth
  file_permission = "0600"
}

resource "local_file" "get_kubeconfig" {
  filename        = "${path.module}/get_kubeconfig.sh"
  content         = <<KUBECONFIG
#!/bin/bash

echo "Getting kubeconfig for cluster ${local.cluster_name}"
aws eks update-kubeconfig --name ${local.cluster_name} --alias ${var.yard_id}

echo "Configure kubernetes cluster AWS authentication"
kubectl --context ${var.yard_id} apply -f ${local_file.k8s_aws_auth.filename}
KUBECONFIG
  file_permission = "0770"
}

resource "local_file" "ansible_inventory" {
  filename        = "${path.module}/ansible_inventory"
  content         = <<INVENTORY
[receiver]
${module.receiver.receiver_ip} ansible_connection=ssh ansible_user=ubuntu ansible_ssh_pass=

[local]
localhost ansible_connection=local

[all:vars]
yard_id=${var.yard_id}
k8s_runtime=${var.k8s_runtime}
k8s_version=${var.k8s_version}
INVENTORY
  file_permission = "0777"
}

resource "local_file" "id_rsa" {
  filename        = "${path.module}/id_rsa"
  content         = module.receiver.ssh_key
  file_permission = "0600"
}

resource "local_file" "eks_node_id_rsa" {
  filename        = "${path.module}/eks_node_id_rsa"
  content         = module.eks_cluster.node_ssh_priv_key
  file_permission = "0600"
}
