#!/bin/bash

DIR=${1:-$CI_PROJECT_DIR}

# This line is used to fix the package import paths in golang files in the agent codebase.
find "$DIR" -type d -name .git -prune -o -type f -name "*.go" -exec sed -i 's/\"github.com\/DataDog\/datadog-agent/\"github.com\/StackVista\/stackstate-agent/g' {} +

# This line is used to fix the package import paths in go.mod files in the agent codebase.
find "$DIR" -type d -name .git -prune -o -type f -name "*.mod" -exec sed -i 's/DataDog\/datadog-agent/StackVista\/stackstate-agent/g' {} +

# This line is used to fix references to the eventual installation and configuration paths in the agent codebase.
find "$DIR" -type d -name .git -prune -o -type f -name "*.go" -exec sed -i 's/\/etc\/datadog-agent/\/etc\/stackstate-agent/g' {} +
find "$DIR" -type d -name .git -prune -o -type f -name "*.go" -exec sed -i 's/\/opt\/datadog-agent/\/opt\/stackstate-agent/g' {} +

# The above will have renamed all instances of github.com/DataDog/datadog-agent to github.com/StackVista/stackstate-agent in all .go files in the agent codebase.
# But as it turns out, there is a protocol buffer that is depended on by an external package that still references the old path.
# We need to fix this for github.com/DataDog/datadog-agent/pkg/proto/pbgo/trace
# This line is used to fix the package import paths in golang files in the agent codebase, rename github.com/StackVista/stackstate-agentpkg/proto/pbgo/trace back to github.com/DataDog/datadog-agent/pkg/proto/pbgo/trace.
find "$DIR" -type d -name .git -prune -o -type f -name "*.go" -exec sed -i 's/\"github.com\/StackVista\/stackstate-agent\/pkg\/proto/\"github.com\/DataDog\/datadog-agent\/pkg\/proto/g' {} +
find "$DIR" -type d -name .git -prune -o -type f -name "*.mod" -exec sed -i 's/StackVista\/stackstate-agent\/pkg\/proto/DataDog\/datadog-agent\/pkg\/proto/g' {} +


# The following lines are used to fix ad hoc references in python files in the tasks folder.
# -------------------- This cannot be used in the pipeline -----------------------
# -------------------- The changes required must be checked in --------------------
#find $CI_PROJECT_DIR/tasks -type d -name .git -prune -o -type f -name "*.py" -exec sed -i 's/DataDog\/datadog-agent/StackVista\/stackstate-agent/g' {} +
#find $CI_PROJECT_DIR/tasks -type d -name .git -prune -o -type f -name "*.py" -exec sed -i 's/\bdatadog-agent\b/stackstate-agent/g' {} +
#find $CI_PROJECT_DIR/tasks -type d -name .git -prune -o -type f -name "*.py" -exec sed -i 's/\bDataDog\b/StackVista/g' {} +
#find $CI_PROJECT_DIR/tasks -type d -name .git -prune -o -type f -name "*.py" -exec sed -i 's/StackVista\/agent-payload/DataDog\/agent-payload/g' {} +
