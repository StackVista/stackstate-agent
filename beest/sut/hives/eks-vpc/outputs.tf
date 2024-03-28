output "vpc_id" {
  value = aws_vpc.cluster.id
}
output "private_subnet_1_id" {
  value = aws_subnet.eks_private.id
}
output "private_subnet_2_id" {
  value = aws_subnet.eks_private_2.id
}

output "public_subnet_1_id" {
  value = aws_subnet.eks_public.id
}

output "public_subnet_2_id" {
  value = aws_subnet.eks_public_2.id
}
