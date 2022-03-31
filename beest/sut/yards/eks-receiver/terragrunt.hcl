include "root" {
  path = find_in_parent_folders()
}

terraform {
  after_hook "serialize_kubeconfig" {
    commands = ["apply"]
    execute  = ["./get-kubeconfig"]
  }
}
