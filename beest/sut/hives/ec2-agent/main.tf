resource "tls_private_key" "rsa_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "agent_key_pair" {
  key_name   = "${var.environment}-agent-key"
  public_key = tls_private_key.rsa_key.public_key_openssh

  tags = {
    Environment = var.environment
  }
}

resource "aws_security_group" "agent_group" {
  name   = "${var.environment}-agent-sg"
  vpc_id = var.vpc_id

  ingress {
    description      = "SSH"
    from_port        = 22
    to_port          = 22
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  ingress {
    description      = "Ping"
    from_port        = 8
    to_port          = 0
    protocol         = "icmp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  egress {
    from_port        = 0
    to_port          = 0
    protocol         = "-1"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  tags = {
    Environment = var.environment
  }
}

data "aws_ami" "ubuntu_ami" {
  most_recent = true

  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-ssd/ubuntu-focal-20.04-amd64-server-*"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }

  owners = ["099720109477"] # Canonical
}

resource "aws_instance" "agent" {
  ami                         = data.aws_ami.ubuntu_ami.id
  instance_type               = "t3.small"
  subnet_id                   = var.subnet_id
  associate_public_ip_address = true
  key_name                    = aws_key_pair.agent_key_pair.key_name
  vpc_security_group_ids      = [aws_security_group.agent_group.id]
  iam_instance_profile        = var.integration_profile

  tags = {
    Name                  = "${var.environment}-agent"
    Environment           = var.environment
    VantaOwner            = "stackstate@stackstate.com"
    VantaNonProd          = true
    VantaDescription      = "Machines used by CI pipeline"
    VantaContainsUserData = false
    VantaUserDataStored   = "NA"
    VantaNoAlert          = "This is for test isn't part of our production systems."
  }
}
