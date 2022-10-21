output "swarm_master_ip" {
  value = aws_instance.swarm_master.public_ip
}
output "swarm_worker_ip" {
  value = aws_instance.swarm_worker.public_ip
}
output "ssh_key" {
  value = tls_private_key.rsa_key.private_key_pem
}
