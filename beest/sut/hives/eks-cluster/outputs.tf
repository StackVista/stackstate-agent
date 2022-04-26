output "node_ssh_key" {
  value = tls_private_key.eks_rsa.private_key_pem
}
output "k8s_aws_auth" {
  value = local.config-map-aws-auth
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
