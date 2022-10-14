//EKS Master Cluster
//This resource is the actual Kubernetes master cluster. It can take a few minutes to provision in AWS.

resource "aws_eks_cluster" "cluster" {
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

data "aws_eks_cluster_auth" "cluster" {
  name = aws_eks_cluster.cluster.name
}

resource "aws_eks_addon" "ebs_csi_driver" {
  count = tonumber(var.k8s_version) >= 1.23 ? 1 : 0
  addon_name   = "aws-ebs-csi-driver"
  cluster_name = aws_eks_cluster.cluster.name
  addon_version = "v1.11.4-eksbuild.1"
}
