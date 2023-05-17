#!/bin/bash

if [ -z "$1" ]
  then
    sts script run --file ~/sts-query.stsl --output json
else
    sts --config "$1" script run --file ~/sts-query.stsl --output json
fi
