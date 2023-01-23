terraform {
  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "~> 2.1.0"
    }
  }
  required_version = ">= 1.0"
}


variable "yard_id" {
  type = string
}

variable "runners_ip" {
  type = string
}

resource "local_file" "ansible_inventory" {
  filename = "${path.module}/ansible_inventory"
  content = yamlencode({
    all : {
      hosts : {
        local : {
          ansible_host : "localhost"
          ansible_connection : "local"
        }
      }
      vars : {
        yard_id : var.yard_id
      }
    }
  })
  file_permission = "0777"
}
