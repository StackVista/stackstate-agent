//  The region we will deploy our cluster into.
variable "region" {
  description = "Region to deploy the cluster into"
  default = "us-east-2"
}

//  The public key to use for SSH access.
variable "public_key_path" {
  default = "./okd_rsa.pub"
}

