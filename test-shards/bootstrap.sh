#!/bin/bash

eval "$(direnv hook bash)"
direnv allow

go install github.com/spf13/cobra/cobra

go run .
