output "receiver_ip" {
  value = aws_instance.receiver.public_ip
}
output "ssh_key" {
  value = tls_private_key.rsa_key.private_key_pem
}
