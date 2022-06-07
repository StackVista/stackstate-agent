variable "region" {
  type    = string
  default = "eu-west-1"
}

variable "include_open_telemetry_tracing" {
  type    = string
  default = "true"
}

variable "StsAccountId" {
  type    = string
  default = "548105126730" ##Integrations Test Main
}

variable "environment" {
  type = string
}
