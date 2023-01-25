resource "local_file" "ansible_inventory" {
  filename = "${path.module}/ansible_inventory"
  content = yamlencode({
    all : {
      hosts : {
        local : {
          ansible_host : "localhost"
          ansible_connection : "local"
        }
        splunk : {
          ansible_host : module.ec2_splunk.splunk_ip
          ansible_connection : "ssh"
          ansible_ssh_private_key_file : local_file.splunk_id_rsa.filename
          ansible_user : "ubuntu"
          ansible_password : ""
        }
      }
      vars : {
        yard_id : var.yard_id
        agent_integration: {
          enabled: true
        }
        splunk_integration : {
          host : module.ec2_splunk.splunk_ip
          url : "https://${module.ec2_splunk.splunk_ip}:8089"
        }
      }
    }
  })
  file_permission = "0777"
}

resource "local_file" "splunk_id_rsa" {
  // path.cwd return the full path which is needed in the ansible_inventory
  filename        = "${path.cwd}/splunk_id_rsa"
  content         = module.ec2_splunk.ssh_key
  file_permission = "0600"
}
