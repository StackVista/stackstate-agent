#!/bin/bash

cat query.stql | python -m stackstate_cli.cli script execute
