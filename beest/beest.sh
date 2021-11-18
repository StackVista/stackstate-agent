#!/bin/bash

source .envrc

go build .
beest "$@"
