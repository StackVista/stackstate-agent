#!/bin/bash

DIR=${1:-$CI_PROJECT_DIR}

# This line is used to fix the package import paths in golang files in the agent codebase.
find "$DIR" -type d -name .git -prune -o -type f -name "*.go" -exec sed -i 's/StackVista\/stackstate-agent/DataDog\/datadog-agent/g' {} +

# This line is used to fix the package import paths in go.mod files in the agent codebase.
find "$DIR" -type d -name .git -prune -o -type f -name "*.mod" -exec sed -i 's/StackVista\/stackstate-agent/DataDog\/datadog-agent/g' {} +

# The following lines are used to fix ad hoc references in python files in the tasks folder.
# -------------------- This cannot be used in the pipeline -----------------------
# -------------------- The changes required must be checked in --------------------
#find $CI_PROJECT_DIR/tasks -type d -name .git -prune -o -type f -name "*.py" -exec sed -i 's/StackVista\/stackstate-agent/DataDog\/datadog-agent/g' {} +
#find $CI_PROJECT_DIR/tasks -type d -name .git -prune -o -type f -name "*.py" -exec sed -i 's/\bstackstate-agent\b/datadog-agent/g' {} +
#find $CI_PROJECT_DIR/tasks -type d -name .git -prune -o -type f -name "*.py" -exec sed -i 's/\bStackVista\b/DataDog/g' {} +
#find $CI_PROJECT_DIR/tasks -type d -name .git -prune -o -type f -name "*.py" -exec sed -i 's/DataDog\/agent-payload/StackVista\/agent-payload/g' {} +
