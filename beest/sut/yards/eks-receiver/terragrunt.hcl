include "root" {
  path = find_in_parent_folders()
}

terraform {
  after_hook "setup_kubeconfig" {
    commands = ["apply"]
    execute  = ["/bin/bash", "-c", "./get_kubeconfig.sh || true"]
  }
}
