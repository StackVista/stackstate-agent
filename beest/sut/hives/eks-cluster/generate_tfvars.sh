#!/bin/bash

# Ensure the ARCH environment variable is set
if [ -z "$ARCH" ]; then
  echo "Error: ARCH environment variable is not set."
  exit 1
fi

# Generate the terraform.tfvars file with the arch variable
cat > terraform.tfvars <<EOF
arch = "${ARCH}"
EOF

echo "terraform.tfvars has been generated with arch = '${ARCH}'"
