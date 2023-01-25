output "api_url" {
  value = module.lambda_otel.api_url
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
        bucket : module.lambda_otel.bucket
        code_zip : module.lambda_otel.codepath
        lambda_role_arn : module.lambda_otel.lambda_role_arn
        lambda_function_name : module.lambda_otel.lambda_function_name
        agent_integration: {
          enabled: true
        }
        aws_integration : {
          region : var.aws_default_region
          cf_stack_name : module.aws_stackpack_role.stack_name
          integration_iam_role : module.aws_stackpack_role.integration_role
          integration_access_key : module.aws_stackpack_role.access_key_id
          integration_secret_key : module.aws_stackpack_role.secret
        }
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
