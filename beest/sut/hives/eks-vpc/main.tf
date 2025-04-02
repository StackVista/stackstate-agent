// Described here https://docs.aws.amazon.com/vpc/latest/userguide/VPC_Scenario2.html
// beware we provision 2 AZ or fault tolerance reasons

resource "aws_vpc" "cluster" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true

  tags = merge(
    var.common_tags,
    {
      "Name" = "${var.environment}-eks-vpc"
    },
  )
}

//The below will create a ${var.public_subnet_cidr} VPC,
//two ${var.public_subnet_cidr} public subnets,
//two ${var.private_subnet_cidr} private subnets with nat instances,
//an internet gateway,
//and setup the subnet routing to route external traffic through the internet gateway

// public subnets
resource "aws_subnet" "eks_public" {
  vpc_id = aws_vpc.cluster.id

  cidr_block        = var.public_subnet_cidr
  availability_zone = var.az1

  tags = merge(
    var.common_tags,
    {
      "Name" = "${var.environment}-eks-public"
    },
  )
}

resource "aws_subnet" "eks_public_2" {
  vpc_id = aws_vpc.cluster.id

  cidr_block        = var.public_subnet_cidr2
  availability_zone = var.az2

  tags = merge(
    var.common_tags,
    {
      "Name" = "${var.environment}-eks-public-2"
    },
  )
}

# The "private" subnets below aren't actually private.
resource "aws_subnet" "eks_private" {
  vpc_id = aws_vpc.cluster.id

  cidr_block        = var.private_subnet_cidr
  availability_zone = var.az1

  tags = merge(
    var.common_tags,
    {
      "Name" = "${var.environment}-eks-private"
      "Description" = "This is not a private subnet"
    },
  )
}

resource "aws_subnet" "eks_private_2" {
  vpc_id = aws_vpc.cluster.id

  cidr_block        = var.private_subnet_cidr2
  availability_zone = var.az2

  tags = merge(
    var.common_tags,
    {
      "Name" = "${var.environment}-eks-private-2"
      "Description" = "This is not a private subnet"
    },
  )
}

// internet gateway, note: creation takes a while
resource "aws_internet_gateway" "igw" {
  vpc_id = aws_vpc.cluster.id

  tags = {
    Environment = var.environment
    Name        = "${var.environment}-igw"
  }
}


# The "private" subnets below aren't actually private. 
resource "aws_route_table" "private_route_table" {
  vpc_id = aws_vpc.cluster.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }

  tags = {
    Environment = var.environment
    Name        = "${var.environment}-private-route-table"
    Description = "This is not a private route table. The traffic goes via Internet Gateway"
  }
}

resource "aws_route_table" "private_route_table_2" {
  vpc_id = aws_vpc.cluster.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }

  tags = {
    Environment = var.environment
    Name        = "${var.environment}-private-route-table-2"
    Description = "This is not a private route table. The traffic goes via Internet Gateway"
  }
}

resource "aws_route_table" "eks_public" {
  vpc_id = aws_vpc.cluster.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }

  tags = {
    Environment = var.environment
    Name        = "${var.environment}-eks-public"
  }
}

resource "aws_route_table" "eks_public_2" {
  vpc_id = aws_vpc.cluster.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.igw.id
  }

  tags = {
    Environment = var.environment
    Name        = "${var.environment}-eks-public-2"
  }
}

// associate route tables
resource "aws_route_table_association" "eks_public" {
  subnet_id      = aws_subnet.eks_public.id
  route_table_id = aws_route_table.eks_public.id
}

resource "aws_route_table_association" "eks_public_2" {
  subnet_id      = aws_subnet.eks_public_2.id
  route_table_id = aws_route_table.eks_public_2.id
}

resource "aws_route_table_association" "eks_private" {
  subnet_id      = aws_subnet.eks_private.id
  route_table_id = aws_route_table.private_route_table.id
}

resource "aws_route_table_association" "eks-private-2" {
  subnet_id      = aws_subnet.eks_private_2.id
  route_table_id = aws_route_table.private_route_table_2.id
}
