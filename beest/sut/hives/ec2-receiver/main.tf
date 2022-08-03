resource "tls_private_key" "rsa_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "receiver_key_pair" {
  key_name   = "${var.environment}-receiver-key"
  public_key = tls_private_key.rsa_key.public_key_openssh

  tags = {
    Environment = var.environment
  }
}

resource "aws_security_group" "receiver_group" {
  name   = "${var.environment}-receiver-sg"
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
    description      = "HTTPS"
    from_port        = 443
    to_port          = 443
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
  ingress {
    description      = "StackState Topic API port"
    from_port        = 7070
    to_port          = 7070
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
  ingress {
    description      = "StackState Receiver API port"
    from_port        = 7077
    to_port          = 7077
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

resource "aws_instance" "receiver" {
  ami                         = "ami-09153889bcbb64887" //Our Packer image based on Ubuntu 18.04 (EBS-Backed x86_64)
  instance_type               = "t3.large"
  subnet_id                   = var.subnet_id
  associate_public_ip_address = true
  key_name                    = aws_key_pair.receiver_key_pair.key_name
  vpc_security_group_ids      = [aws_security_group.receiver_group.id]

  tags = {
    Name                  = "${var.environment}-receiver"
    Environment           = var.environment
    VantaContainsUserData = false
    VantaDescription      = "Machine used used in acceptance pipeline"
    VantaNonProd          = true
    VantaOwner            = "beest@stackstate.com"
    VantaUserDataStored   = "NA"
  }
}
