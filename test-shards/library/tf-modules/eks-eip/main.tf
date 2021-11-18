#resource "aws_eip" "nlb_eip" {
#  vpc = true
#  tags = {
#    Environment = var.environment
#    Name        = "EKS inbound IP 1"
#  }
#}
#
#resource "aws_eip" "nlb_eip_2" {
#  vpc = true
#  tags = {
#    Environment = var.environment
#    Name        = "EKS inbound IP 2"
#  }
#}

resource "aws_eip" "nat_eip" {
  vpc = true
  tags = {
    Environment = var.environment
    Name        = "EKS outbound IP 1"
  }
}

resource "aws_eip" "nat_eip_2" {
  vpc = true
  tags = {
    Environment = var.environment
    Name        = "EKS outbound IP 2"
  }
}
