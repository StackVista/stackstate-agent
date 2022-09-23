#!/bin/bash

. bootstrap_functions.sh

# if this is an additional shell, no need to execute the following
if [ `echo $$` == 1 ]; then
    install_cobra_cli
    build_beest
    generate_aws_config
fi

setup_interactive_shell
