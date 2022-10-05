resource "local_file" "ansible_inventory" {
  filename = "${path.module}/ansible_inventory"
  content = yamlencode({
    all : {
      hosts : {
        local : {
          ansible_host : "localhost"
          ansible_connection : "local"
        }
        agent_redhat : {
          ansible_host : module.ec2_agent_redhat.agent_ip
          ansible_connection : "ssh"
          ansible_ssh_private_key_file : local_file.agent_redhat_rsa.filename
          ansible_user : "ec2-user"
          ansible_password : ""
        }
        agent_ubuntu : {
          ansible_host : module.ec2_agent_ubuntu.agent_ip
          ansible_connection : "ssh"
          ansible_ssh_private_key_file : local_file.agent_ubuntu_rsa.filename
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

resource "local_file" "agent_redhat_rsa" {
  // path.cwd return the full path which is needed in the ansible_inventory
  filename        = "${path.cwd}/agent_redhat_id_rsa"
  content         = module.ec2_agent_redhat.ssh_key
  file_permission = "0600"
}

resource "local_file" "agent_ubuntu_rsa" {
  // path.cwd return the full path which is needed in the ansible_inventory
  filename        = "${path.cwd}/agent_ubuntu_id_rsa"
  content         = module.ec2_agent_ubuntu.ssh_key
  file_permission = "0600"
}
