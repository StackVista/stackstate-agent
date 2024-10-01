enable_auto_completions() {
    source  ~/.bashrc
    echo "Enable auto-completions ..."
    complete -C '/usr/local/bin/aws_completer' aws
    eval "$(helm completion bash)"
    eval "$(sts-toolbox completion bash)"
    eval "$(sts completion bash)"
    eval "$(/go/src/app/beest completion bash)"
}

activate_direnv() {
    echo "Activate direnv ..."
    eval "$(direnv hook bash)"
}

setup_interactive_shell() {
    enable_auto_completions
    activate_direnv
}


install_cobra_cli() {
    echo "Install Cobra CLI ..."
    go install github.com/spf13/cobra/cobra
}

generate_aws_config() {
    echo "Generate AWS config StackState profiles ..."
    sts-toolbox aws generate -p developer
}

configure_aws_beest_credentials() {
    echo "Configure AWS Beest credentials ..."

    if aws-vault exec default -- aws sts get-caller-identity >/dev/null 2>&1 ; then
      echo "AWS Beest credentials is already setup and working, skipping ..."

    elif [ ! -z "$BEEST_AWS_MFA_KEY" ]; then
        echo "Configure AWS Beest credentials with aws-vault and MFA ..."

        gpgKeyName=${artifactory_user:-beest@stackstate.com}
        echo "Generating GPG key for aws-vault backend store for ${gpgKeyName}"

        echo "Key-Type: RSA
        Key-Length: 4096
        Subkey-Type: RSA
        Subkey-Length: 4096
        Name-Real: ${gpgKeyName}
        Name-Email: ${gpgKeyName}
        Expire-Date: 0
        %no-protection
        " | gpg --batch --generate-key

        gpgKey=$(gpg --list-signatures --with-colons | grep 'sig' | grep "${gpgKeyName}" | head -n 1 | cut -d':' -f5)

        echo "Init pass with gpg key to be used as aws-vault backend store"
        pass init $gpgKey

        export AWS_ACCESS_KEY_ID=$BEEST_AWS_ACCESS_KEY_ID
        export AWS_SECRET_ACCESS_KEY=$BEEST_AWS_SECRET_ACCESS_KEY
        export AWS_VAULT_BACKEND="pass"

        echo -e "#!/bin/bash\n\npass default >/dev/null 2>&1\naws-vault exec -j --region eu-west-1 default" > ~/.aws/credential_process.sh
        chmod +x ~/.aws/credential_process.sh
        echo -e "[default]\noutput=json\nregion=eu-west-1\nmfa_serial=${BEEST_AWS_MFA_KEY}\ncredential_process=/home/keeper/.aws/credential_process.sh" > ~/.aws/config

        echo "Generate AWS config StackState profiles ..."
        sts-toolbox aws generate -p developer

        aws-vault add default --env
        pass default

        # unset aws keys and set aws profile
        unset AWS_ACCESS_KEY_ID
        unset AWS_SECRET_ACCESS_KEY
        export AWS_PROFILE=stackstate-sandbox

        echo -e "#!/bin/bash\n\npass default >/dev/null 2>&1\naws-vault exec --duration=4h default echo" > ~/.aws/refresh_credentials.sh
        chmod +x ~/.aws/refresh_credentials.sh

        aws-vault exec --duration=4h default echo
    else
        echo "Configure AWS Beest credentials using key and secret ..."

        echo -e "[default]\naws_access_key_id=$BEEST_AWS_ACCESS_KEY_ID\naws_secret_access_key=$BEEST_AWS_SECRET_ACCESS_KEY" > ~/.aws/credentials
    fi

}

connect_to_stackstate_sandbox() {
    echo "Connect to StackState sandbox cluster ..."
    sts-toolbox cluster connect sandbox-main.sandbox.stackstate.io
}

build_beest() {
    echo "Build Beest ..."
    go mod tidy
    go mod vendor
    go build .
}

export -f configure_aws_beest_credentials
