resource "tls_private_key" "rsa_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "agent_redhat_key_pair" {
  key_name   = "${var.environment}-agent-redhat-v2-key"
  public_key = tls_private_key.rsa_key.public_key_openssh

  tags = {
    Environment = var.environment
  }
}

resource "aws_security_group" "agent_redhat_group" {
  name   = "${var.environment}-agent-redhat-v2-sg"
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

data "aws_ami" "redhat_ami" {
  most_recent = true

  filter {
    name   = "name"
    values = ["RHEL-8.6.0_HVM-20220503-x86_64-2-Hourly2-GP2"]
  }

  owners = ["309956199498"] # amazon
}

resource "aws_instance" "agent_redhat" {
  ami                         = data.aws_ami.redhat_ami.id
  instance_type               = "t3.small"
  subnet_id                   = var.subnet_id
  associate_public_ip_address = true
  key_name                    = aws_key_pair.agent_redhat_key_pair.key_name
  vpc_security_group_ids      = [aws_security_group.agent_redhat_group.id]
  iam_instance_profile        = var.integration_profile

  tags = {
    Name                  = "${var.environment}-agent-redhat-v2"
    Environment           = var.environment
    VantaContainsUserData = false
    VantaDescription      = "Machine used used in acceptance pipeline"
    VantaNonProd          = true
    VantaOwner            = "stackstate@stackstate.com"
    VantaUserDataStored   = "NA"
  }
}
