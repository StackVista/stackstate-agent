#resource "local_file" "kubeconfig" {
#  filename = "${pathexpand("~")}/.kube/config"
#  content = module.eks_cluster.kubeconfig
#  file_permission = "0600"
#}

resource "local_file" "ansible_inventory" {
  filename = "${path.module}/ansible_inventory"
  content = <<INVENTORY
[receiver]
${module.receiver.receiver_ip} ansible_connection=ssh ansible_user=ubuntu ansible_ssh_pass=
INVENTORY
  file_permission = "0777"
}

resource "local_file" "id_rsa" {
  filename = "${path.module}/id_rsa"
  content = module.receiver.ssh_key
  file_permission = "0600"
}
