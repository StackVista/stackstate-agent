terraform {
  backend "s3" {}
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.10.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 3.1.0"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.1.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.3.1"
    }
  }
  required_version = ">= 1.0"
}
