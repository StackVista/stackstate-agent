data "aws_iam_role" "administrator" {
  name = "Administrator"
}

locals {
  config-map-aws-auth = <<CONFIGMAPAWSAUTH
apiVersion: v1
kind: ConfigMap
metadata:
  name: aws-auth
  namespace: kube-system
data:
  mapRoles: |
    - rolearn: ${aws_iam_role.eks_node_role.arn}
      username: system:node:{{EC2PrivateDNSName}}
      groups:
        - system:bootstrappers
        - system:nodes
    - rolearn: ${data.aws_iam_role.administrator.arn}
      username: ${data.aws_iam_role.administrator.name}
      groups:
        - system:masters
CONFIGMAPAWSAUTH
}
