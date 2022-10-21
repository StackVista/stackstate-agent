resource "local_file" "ansible_inventory" {
  filename = "${path.module}/ansible_inventory"
  content = yamlencode({
    all : {
      hosts : {
        local : {
          ansible_host : "localhost"
          ansible_connection : "local"
        }
        swarm_master : {
          ansible_host : module.ec2_swarm.swarm_master_ip
          ansible_connection : "ssh"
          ansible_ssh_private_key_file : local_file.swarm_id_rsa.filename
          ansible_user : "ubuntu"
          ansible_password : ""
        }
        swarm_worker : {
          ansible_host : module.ec2_swarm.swarm_worker_ip
          ansible_connection : "ssh"
          ansible_ssh_private_key_file : local_file.swarm_id_rsa.filename
          ansible_user : "ubuntu"
          ansible_password : ""
        }
      }
      vars : {
        yard_id : var.yard_id
        swarm_master_ip : module.ec2_swarm.swarm_master_ip
      }
    }
  })
  file_permission = "0777"
}

resource "local_file" "swarm_id_rsa" {
  // path.cwd return the full path which is needed in the ansible_inventory
  filename        = "${path.cwd}/swarm_id_rsa"
  content         = module.ec2_swarm.ssh_key
  file_permission = "0600"
}
