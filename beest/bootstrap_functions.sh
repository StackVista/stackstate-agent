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
    #sts-toolbox aws generate -p developer
    sts-toolbox aws configure profile-set -s poweruser --sso-start-url "https://d-9967196c83.awsapps.com/start/#"
}

configure_aws_beest_credentials() {
    echo "Configure AWS Beest credentials ..."

    aws sso login --sso-session SuseSSO

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
