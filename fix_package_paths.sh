#!/bin/bash

DIR=${1:-$CI_PROJECT_DIR}

# This line is used to fix the package import paths in golang files in the agent codebase.
find "$DIR" -type d -name .git -prune -o -type f -name "*.go" -exec sed -i 's/\"github.com\/DataDog\/datadog-agent/\"github.com\/StackVista\/stackstate-agent/g' {} +

# This line is used to fix the package import paths in go.mod files in the agent codebase.
find "$DIR" -type d -name .git -prune -o -type f -name "*.mod" -exec sed -i 's/DataDog\/datadog-agent/StackVista\/stackstate-agent/g' {} +

# The above will have renamed all instances of github.com/DataDog/datadog-agent to github.com/StackVista/stackstate-agent in all .go files in the agent codebase.
# But as it turns out, there are packages that are depended on by external packages (e.g., OpenTelemetry) that still reference the old DataDog paths.
# We need to revert these back to DataDog paths: pkg/proto, pkg/trace, and pkg/api

# Revert pkg/proto back to DataDog
find "$DIR" -type d -name .git -prune -o -type f -name "*.go" -exec sed -i 's/\"github.com\/StackVista\/stackstate-agent\/pkg\/proto/\"github.com\/DataDog\/datadog-agent\/pkg\/proto/g' {} +
find "$DIR" -type d -name .git -prune -o -type f -name "*.mod" -exec sed -i 's/StackVista\/stackstate-agent\/pkg\/proto/DataDog\/datadog-agent\/pkg\/proto/g' {} +

# Revert pkg/trace back to DataDog (external dependency from OpenTelemetry)
find "$DIR" -type d -name .git -prune -o -type f -name "*.go" -exec sed -i 's/\"github.com\/StackVista\/stackstate-agent\/pkg\/trace/\"github.com\/DataDog\/datadog-agent\/pkg\/trace/g' {} +
find "$DIR" -type d -name .git -prune -o -type f -name "*.mod" -exec sed -i 's/StackVista\/stackstate-agent\/pkg\/trace/DataDog\/datadog-agent\/pkg\/trace/g' {} +

## Revert pkg/api back to DataDog (dependency of pkg/trace)
#find "$DIR" -type d -name .git -prune -o -type f -name "*.go" -exec sed -i 's/\"github.com\/StackVista\/stackstate-agent\/pkg\/api/\"github.com\/DataDog\/datadog-agent\/pkg\/api/g' {} +
#find "$DIR" -type d -name .git -prune -o -type f -name "*.mod" -exec sed -i 's/StackVista\/stackstate-agent\/pkg\/api/DataDog\/datadog-agent\/pkg\/api/g' {} +
#
## Special use-case for an older dependency that gets transitively pulled in through the github.com/opentelemetry/otel-collector-contrib/datadog github repo
#find . -type f \( -name "go.mod" -o -name "go.work" \) -exec sed -i 's|github\.com/StackVista/stackstate-agent/comp/core/secrets v0\.69\.2|github.com/DataDog/datadog-agent/comp/core/secrets v0.69.2|g' {} +
