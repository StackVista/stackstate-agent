#!/bin/sh

sa_token=`cat /var/run/secrets/kubernetes.io/serviceaccount/token`

mkdir -p /root/.stackstate/cli
cat > /root/.stackstate/cli/conf.yaml <<EOF
instances:
  default:
    base_api:
      url: "http://stackstate-router:8080"
      token_auth:
        token: "$sa_token"
    receiver_api:
      url: "http://stackstate-router:8080/receiver"
    admin_api:
      url: "http://stackstate-router:8080/admin"
      token_auth:
        token: "$sa_token"
    clients:
      default:
        api_key: "API_KEY"
        hostname: "hostname"
        internal_hostname: "internal_hostname"
EOF

# this will keep the container running forever without exiting after it starts
sleep infinity
