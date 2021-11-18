#!/bin/bash

source ~/.bashrc
eval "$(direnv hook bash)"
direnv allow

go install github.com/spf13/cobra/cobra

go build .

eval "$(/go/src/app/beest completion bash)"
