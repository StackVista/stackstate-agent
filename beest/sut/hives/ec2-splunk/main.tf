resource "tls_private_key" "rsa_key" {
  algorithm = "RSA"
  rsa_bits  = 4096
}

resource "aws_key_pair" "splunk_key_pair" {
  key_name   = "${var.environment}-splunk-key"
  public_key = tls_private_key.rsa_key.public_key_openssh

  tags = {
    Environment = var.environment
  }
}

data "http" "local_ip" {
  url = "https://ipv4.icanhazip.com"
}

resource "aws_security_group" "splunk_group" {
  name   = "${var.environment}-splunk-sg"
  vpc_id = var.vpc_id

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["${var.runners_ip}/32"]
  }

  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["${chomp(data.http.local_ip.body)}/32"]
  }

  ingress {
    description      = "Splunk User Interface"
    from_port        = 8000
    to_port          = 8000
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
  ingress {
    description      = "Splunk API"
    from_port        = 8089
    to_port          = 8089
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
    description      = "StackState Simulator Receiver API port"
    from_port        = 7078
    to_port          = 7078
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

resource "aws_instance" "splunk" {
  ami                         = "ami-0e24b531109ae5895" //Our Packer image based on Ubuntu 18.04 (EBS-Backed x86_64)
  instance_type               = "t3.large"
  subnet_id                   = var.subnet_id
  associate_public_ip_address = true
  key_name                    = aws_key_pair.splunk_key_pair.key_name
  vpc_security_group_ids      = [aws_security_group.splunk_group.id]

  root_block_device {
    volume_size = 30 # in GB <<----- I increased this!
    volume_type = "gp3"
    encrypted   = true
  }

  tags = {
    Name                  = "${var.environment}-splunk"
    Environment           = var.environment
    VantaContainsUserData = false
    VantaDescription      = "Machine used used in acceptance pipeline"
    VantaNonProd          = true
    VantaOwner            = "beest@stackstate.com"
    VantaUserDataStored   = "NA"
  }
}
