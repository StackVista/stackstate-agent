//EKS Master Cluster
//This resource is the actual Kubernetes master cluster. It can take a few minutes to provision in AWS.

resource "aws_eks_cluster" "eks_cluster" {
  name     = var.k8s_cluster_name
  role_arn = aws_iam_role.eks_cluster_role.arn
  version  = var.k8s_version

  vpc_config {
    security_group_ids = [aws_security_group.eks_control_plane_sg.id]

    subnet_ids = [
      var.private_subnet_1_id,
      var.private_subnet_2_id
    ]
  }

  depends_on = [
    aws_iam_role_policy_attachment.eks_cluster_policy,
    aws_iam_role_policy_attachment.eks_service_policy,
  ]
}

data "aws_eks_cluster_auth" "eks_cluster" {
  name = aws_eks_cluster.eks_cluster.name
}

locals {
  kubeconfig = <<KUBECONFIG
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: ${aws_eks_cluster.eks_cluster.certificate_authority[0].data}
    server: ${aws_eks_cluster.eks_cluster.endpoint}
  name: kubernetes
contexts:
- context:
    cluster: kubernetes
    user: aws
  name: aws-${var.environment}
current-context: aws-${var.environment}
kind: Config
preferences: {}
users:
- name: aws
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1alpha1
      command: aws
      args:
        - "eks"
        - "get-token"
        - "--cluster-name"
        - "${var.k8s_cluster_name}"
KUBECONFIG
}
