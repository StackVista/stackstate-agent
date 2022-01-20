data "aws_iam_role" "administrator" {
  name = "Administrator"
}

resource "kubernetes_config_map" "eks_aws_auth" {
  metadata {
    name      = "aws-auth"
    namespace = "kube-system"
  }
  data = {
    mapRoles = yamlencode([
      {
        rolearn  = aws_iam_role.eks_node_role.arn
        username = "system:node:{{EC2PrivateDNSName}}"
        groups   = ["system:bootstrappers", "system:nodes"]
      },
      {
        rolearn  = data.aws_iam_role.administrator.arn
        username = data.aws_iam_role.administrator.name
        groups   = ["system:masters"]
      }
    ])
  }
}
