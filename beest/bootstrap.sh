#!/bin/bash

go install github.com/spf13/cobra/cobra
complete -C '/usr/local/bin/aws_completer' aws
eval "$(sts-toolbox completion bash)"
eval "$(sts completion bash)"

source ~/.bashrc
eval "$(direnv hook bash)"
direnv allow

go build .
eval "$(/go/src/app/beest completion bash)"
