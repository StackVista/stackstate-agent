#!/bin/bash

# This line is used to fix the package import paths in golang files in the agent codebase.
find $CI_PROJECT_DIR -type d -name .git -prune -o -type f -name "*.go" -exec sed -i 's/DataDog\/datadog-agent/StackVista\/stackstate-agent/g' {} +

# This line is used to fix the package import paths in go.mod files in the agent codebase.
find $CI_PROJECT_DIR -type d -name .git -prune -o -type f -name "*.mod" -exec sed -i 's/DataDog\/datadog-agent/StackVista\/stackstate-agent/g' {} +

# The following lines are used to fix ad hoc references in python files in the tasks folder.
find $CI_PROJECT_DIR/tasks -type d -name .git -prune -o -type f -name "*.py" -exec sed -i 's/DataDog\/datadog-agent/StackVista\/stackstate-agent/g' {} +
find $CI_PROJECT_DIR/tasks -type d -name .git -prune -o -type f -name "*.py" -exec sed -i 's/\bdatadog-agent\b/stackstate-agent/g' {} +
find $CI_PROJECT_DIR/tasks -type d -name .git -prune -o -type f -name "*.py" -exec sed -i 's/\bDataDog\b/StackVista/g' {} +
find $CI_PROJECT_DIR/tasks -type d -name .git -prune -o -type f -name "*.py" -exec sed -i 's/StackVista\/agent-payload/DataDog\/agent-payload/g' {} +
