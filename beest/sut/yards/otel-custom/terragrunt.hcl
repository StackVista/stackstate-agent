include "root" {
  path = find_in_parent_folders()
}

terraform {
  after_hook "setup_kubeconfig" {
    // TODO figure out how to run this only during create
    commands = ["apply"]
    execute  = ["/bin/bash", "-c", "sts-toolbox cluster connect sandbox-main.sandbox.stackstate.io"]
  }
}
