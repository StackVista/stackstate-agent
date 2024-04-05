variable "region" {
  type = string
}

variable "include_open_telemetry_tracing" {
  type    = string
  default = "true"
}

variable "environment" {
  type = string
}
