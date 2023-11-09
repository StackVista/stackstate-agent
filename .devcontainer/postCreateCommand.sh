#!/bin/sh

cat >> /home/vscode/.zshrc << EOF
[ -f /go/src/github.com/StackVista/stackstate-agent/.env ] && source /go/src/github.com/StackVista/stackstate-agent/.env
EOF

pip3 install -r requirements.txt
pip3 install virtualenv
