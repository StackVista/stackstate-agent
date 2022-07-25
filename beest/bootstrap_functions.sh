enable_auto_completions() {
    echo "Enable auto-completions ..."
    complete -C '/usr/local/bin/aws_completer' aws
    eval "$(sts-toolbox completion bash)"
    eval "$(sts completion bash)"
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
    mkdir -p ~/.aws
    touch ~/.aws/config
    sts-toolbox aws generate -p developer
}

configure_aws_beest_credentials() {
    echo "Configure AWS Beest credentials ..."
    echo -e "[default]\naws_access_key_id=$BEEST_AWS_ACCESS_KEY_ID\naws_secret_access_key=$BEEST_AWS_SECRET_ACCESS_KEY" > ~/.aws/credentials
}

build_beest() {
    echo "Build Beest ..."
    go build .
    eval "$(/go/src/app/beest completion bash)"
}
