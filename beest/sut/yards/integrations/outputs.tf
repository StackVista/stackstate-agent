resource "local_file" "ansible_inventory" {
  filename        = "${path.module}/ansible_inventory"
  content = yamlencode({
    all : {
      hosts : {
        local : {
          ansible_host : "localhost"
          ansible_connection : "local"
        }
        agent : {
          ansible_host : module.ec2_agent.agent_ip
          ansible_connection : "ssh"
          ansible_ssh_private_key_file : local_file.agent_id_rsa.filename
          ansible_user : "ubuntu"
          ansible_password : ""
        }
      }
      vars : {
        yard_id : var.yard_id
      }
    }
  })

  file_permission = "0777"
}

resource "local_file" "agent_id_rsa" {
  // path.cwd return the full path which is needed in the ansible_inventory
  filename        = "${path.cwd}/agent_id_rsa"
  content         = module.ec2_agent.ssh_key
  file_permission = "0600"
}
