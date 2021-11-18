variable "environment" {}
variable "common_tags" {}
variable "az1" {}
variable "az2" {}
variable "nat_eip_1_id" {}
variable "nat_eip_2_id" {}

variable "vpc_cidr" {
  description = "CIDR for the whole VPC"
  default     = "10.11.0.0/16"
}

variable "public_subnet_cidr" {
  description = "CIDR for the Public Subnet"
  default     = "10.11.0.0/24"
}

variable "private_subnet_cidr" {
  description = "CIDR for the Private Subnet"
  default     = "10.11.1.0/24"
}

variable "public_subnet_cidr2" {
  description = "CIDR for the Public Subnet"
  default     = "10.11.2.0/24"
}

variable "private_subnet_cidr2" {
  description = "CIDR for the Private Subnet"
  default     = "10.11.3.0/24"
}
