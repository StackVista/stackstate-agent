//Kubernetes Masters
//It requires a few operator managed resources beforehand so that Kubernetes can properly manage
//other AWS services as well as allow inbound networking communication from your local workstation
//(if desired) and worker nodes.

//EKS Master Cluster IAM Role
//IAM role and policy to allow the EKS service to manage or retrieve data from other AWS services.
//For the latest required policy, see the EKS User Guide.

resource "aws_iam_role" "eks_cluster_role" {
  name               = "EKSClusterRole-${var.environment}"
  description        = "Allows EKS to manage clusters on your behalf."
  assume_role_policy = <<POLICY
{
   "Version":"2012-10-17",
   "Statement":[
      {
         "Effect":"Allow",
         "Principal":{
            "Service":"eks.amazonaws.com"
         },
         "Action":"sts:AssumeRole"
      }
   ]
}
POLICY
}

//https://docs.aws.amazon.com/eks/latest/userguide/service_IAM_role.html
resource "aws_iam_role_policy_attachment" "eks_cluster_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.eks_cluster_role.name
}

resource "aws_iam_role_policy_attachment" "eks_service_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSServicePolicy"
  role       = aws_iam_role.eks_cluster_role.name
}

/*
 Adding a policy to cluster IAM role that allow permissions
 required to create AWSServiceRoleForElasticLoadBalancing service-linked role by EKS during ELB provisioning
*/
data "aws_iam_policy_document" "cluster_elb_sl_role_creation" {
  statement {
    effect = "Allow"
    actions = [
      "ec2:DescribeAccountAttributes",
      "ec2:DescribeAddresses",
      "ec2:DescribeInternetGateways",
    ]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "cluster_elb_sl_role_creation" {
  name_prefix = "${var.k8s_cluster_name}-elb-sl-role-creation"
  role        = aws_iam_role.eks_cluster_role.name
  policy      = data.aws_iam_policy_document.cluster_elb_sl_role_creation.json
}
