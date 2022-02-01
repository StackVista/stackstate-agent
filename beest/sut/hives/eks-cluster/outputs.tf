output "node_role_arn" {
  value = aws_iam_role.eks_node_role.arn
}
output "node_ssh_priv_key" {
  value = tls_private_key.eks_rsa.private_key_pem
}
output "endpoint" {
  value = aws_eks_cluster.cluster.endpoint
}
output "ca_cert" {
  value = aws_eks_cluster.cluster.certificate_authority.0.data
}
output "token" {
  value = data.aws_eks_cluster_auth.cluster.token
}
output "kubeconfig" {
  value = local.kubeconfig
}
