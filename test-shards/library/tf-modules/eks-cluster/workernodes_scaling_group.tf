//Worker Node AutoScaling Group
//Now we have everything in place to create and manage EC2 instances that will serve as our worker nodes
//in the Kubernetes cluster. This setup utilizes an EC2 AutoScaling Group (ASG) rather than manually working with
//EC2 instances. This offers flexibility to scale up and down the worker nodes on demand when used in conjunction
//with AutoScaling policies (not implemented here).
//
//First, let us create a data source to fetch the latest Amazon Machine Image (AMI) that Amazon provides with an
//EKS compatible Kubernetes baked in.

data "aws_ami" "eks_node_ami" {
  filter {
    name   = "name"
    values = ["amazon-eks-node-${aws_eks_cluster.eks_cluster.version}*"]
  }

  most_recent = true
  owners      = ["602401143452"] # Amazon Account ID
}

# EKS currently documents this required userdata for EKS worker nodes to
# properly configure Kubernetes applications on the EC2 instance.
# We utilize a Terraform local here to simplify Base64 encoding this
# information into the AutoScaling Launch Configuration.
# More information: https://amazon-eks.s3-us-west-2.amazonaws.com/1.10.3/2018-06-05/amazon-eks-nodegroup.yaml
locals {
  eks-node-userdata = <<USERDATA
#!/bin/bash -xe
/etc/eks/bootstrap.sh ${var.k8s_cluster_name}
USERDATA
}

resource "tls_private_key" "eks_rsa" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "eks_node_key_pair" {
  key_name   = "eks-deployer-${var.k8s_cluster_name}"
  public_key = tls_private_key.eks_rsa.public_key_openssh
}

resource "aws_launch_configuration" "eks_launch_configuration" {
  associate_public_ip_address = true
  iam_instance_profile        = aws_iam_instance_profile.eks_node_instance_profile.name
  image_id                    = data.aws_ami.eks_node_ami.id
  instance_type               = var.k8s_node_type
  //  spot_price                  = "0.008" # TODO check - Status Reason: Max spot instance count exceeded.
  name_prefix      = "eks-${var.k8s_cluster_name}"
  security_groups  = [aws_security_group.eks_nodes_sg.id]
  user_data_base64 = base64encode(local.eks-node-userdata)
  key_name         = aws_key_pair.eks_node_key_pair.key_name

  root_block_device {
    encrypted = true
  }

  lifecycle {
    create_before_destroy = true
  }
}

//Finally, we create an AutoScaling Group that actually launches EC2 instances based on the
//AutoScaling Launch Configuration.

//NOTE: The usage of the specific kubernetes.io/cluster/* resource tag below is required for EKS
//and Kubernetes to discover and manage compute resources.

resource "aws_autoscaling_group" "eks_autoscaling_group" {
  launch_configuration = aws_launch_configuration.eks_launch_configuration.id
  desired_capacity     = var.k8s_size
  max_size             = var.k8s_size
  min_size             = 0
  name                 = "eks-${var.k8s_cluster_name}"
  vpc_zone_identifier  = [var.private_subnet_1_id, var.private_subnet_2_id]

  tag {
    key                 = "Environment"
    value               = var.environment
    propagate_at_launch = true
  }
  tag {
    key                 = "Name"
    value               = "eks-${var.k8s_cluster_name}"
    propagate_at_launch = true
  }
  tag {
    key                 = "kubernetes.io/cluster/${var.k8s_cluster_name}"
    value               = "owned"
    propagate_at_launch = true
  }
}
