#!/bin/bash

. bootstrap_functions.sh

# if this is an additional shell, no need to execute the following
if [ `echo $$` == 1 ]; then
    install_cobra_cli
    build_beest
    generate_aws_config
fi

# we export the function so we can call it in the .envrc, after BEEST_AWS variables have been set
export -f configure_aws_beest_credentials

setup_interactive_shell
