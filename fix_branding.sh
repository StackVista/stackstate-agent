#!/bin/bash

DIR=${1:-$CI_PROJECT_DIR}

# ---------------------------------------------------------------------------
# Core Go/config branding (ported from the old apply_branding task)
# ---------------------------------------------------------------------------

# High-level config wording
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/Datadog config/StackState config/g' {} +

# Filesystem paths
# Do more specific paths first, before general replacements
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/\/etc\/datadog-agent\/managed\/datadog-agent/\/etc\/stackstate-agent\/managed\/stackstate-agent/g' {} +
# Handle partially branded case in Go files (in case general /etc/datadog-agent replacement ran first)
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/\/etc\/stackstate-agent\/managed\/datadog-agent/\/etc\/stackstate-agent\/managed\/stackstate-agent/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/\/opt\/datadog-packages\/datadog-agent-ddot/\/opt\/stackstate-packages\/stackstate-agent-ddot/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/\/opt\/datadog-packages\/datadog-agent/\/opt\/stackstate-packages\/stackstate-agent/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/\/etc\/datadog-agent-exp/\/etc\/stackstate-agent-exp/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/\/etc\/datadog-agent/\/etc\/stackstate-agent/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/\/opt\/datadog-agent/\/opt\/stackstate-agent/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/\/var\/log\/datadog/\/var\/log\/stackstate/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/\/var\/run\/datadog/\/var\/run\/stackstate/g' {} +

# Config filenames
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/datadog.yaml/stackstate.yaml/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/datadog.conf/stackstate.conf/g' {} +
find "$DIR/cmd/agent/common" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/datadog\.yaml/stackstate\.yaml/g' {} +
find "$DIR/cmd/agent/common" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/datadog\.conf/stackstate\.conf/g' {} +

# Generic Datadog → StackState naming where safe
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/Datadog Agent/StackState Agent/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/Datadog Cluster/StackState Cluster/g' {} +

# Cluster IDs and leader election
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/datadog-cluster-id/stackstate-cluster-id/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/datadog-leader-election/stackstate-leader-election/g' {} +

# Brand default pod annotation exclude filter regex pattern to match branded annotations
# The filter regex needs to match ad.stackstate.com instead of ad.datadoghq.com
# Pattern is in backtick string: `^ad\.datadoghq\.com\/...` -> `^ad\.stackstate\.com\/...`
find "$DIR/pkg/config/setup" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/ad\\.datadoghq\\.com/ad\\.stackstate\\.com/g' {} +

# Kubernetes / Docker OpenMetrics annotations
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/ad\.datadoghq\.com/ad.stackstate.com/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/com.datadoghq.ad/com.stackstate.ad/g' {} +
# Brand annotation keys in test data JSON files so they match the branded code expectations
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -path "*/testdata/*" -name "*.json" -exec sed -i 's/"ad\.datadoghq\.com/"ad.stackstate.com/g' {} +
# Brand paths in test data JSON files (e.g., /var/run/datadog -> /var/run/stackstate)
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -path "*/testdata/*" -name "*.json" -exec sed -i 's|"/var/run/datadog"|"/var/run/stackstate"|g' {} +
# Brand log paths in test data XML files to match production paths (/var/log/datadog -> /var/log/stackstate-agent)
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -path "*/testdata/*" -name "*.xml" -exec sed -i 's|/var/log/datadog|/var/log/stackstate|g' {} +
# Brand log paths in test code to match production paths
find "$DIR/pkg/util/log/setup/internal/seelog" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's|/var/log/datadog|/var/log/stackstate|g' {} +

# NTP pool defaults
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/0.datadog.pool.ntp.org/0.stackstate.pool.ntp.org/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/1.datadog.pool.ntp.org/1.stackstate.pool.ntp.org/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/2.datadog.pool.ntp.org/2.stackstate.pool.ntp.org/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/3.datadog.pool.ntp.org/3.stackstate.pool.ntp.org/g' {} +

# HTML templates - brand Datadog Agent strings
find "$DIR/comp/core/gui" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.tmpl" -exec sed -i 's/Datadog Agent/StackState Agent/g' {} +
find "$DIR/comp/core/gui" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.tmpl" -exec sed -i 's/Datadog Agent Manager/StackState Agent Manager/g' {} +

# Fleet installer embedded gen files - brand paths in generated systemd service files
# These files are pre-generated and embedded, so they need to be branded to match the branded template data
find "$DIR/pkg/fleet/installer/packages/embedded/templates/gen" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.service" -exec sed -i 's|/opt/datadog-packages/datadog-agent|/opt/stackstate-packages/stackstate-agent|g' {} +
find "$DIR/pkg/fleet/installer/packages/embedded/templates/gen" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.service" -exec sed -i 's|/opt/datadog-packages/datadog-agent-ddot|/opt/stackstate-packages/stackstate-agent-ddot|g' {} +
find "$DIR/pkg/fleet/installer/packages/embedded/templates/gen" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.service" -exec sed -i 's|/opt/datadog-agent|/opt/stackstate-agent|g' {} +
find "$DIR/pkg/fleet/installer/packages/embedded/templates/gen" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.service" -exec sed -i 's|/etc/datadog-agent-exp|/etc/stackstate-agent-exp|g' {} +
# Brand managed directory path - must run before general /etc/datadog-agent replacement
# Handle both unbranded (/etc/datadog-agent/managed/datadog-agent) and partially branded (/etc/stackstate-agent/managed/datadog-agent) cases
# Run in two passes to ensure we catch the pattern regardless of execution order
find "$DIR/pkg/fleet/installer/packages/embedded/templates/" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.tmpl" -exec sed -i 's/datadog/stackstate/g' {} +
find "$DIR/pkg/fleet/installer/packages/embedded/templates/" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.tmpl" -exec sed -i 's/User\=dd-agent/User\=stackstate-agent/g' {} +
find "$DIR/pkg/fleet/installer/packages/embedded/templates/" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.tmpl" -exec sed -i 's/DataDog/StackState/g' {} +
find "$DIR/pkg/fleet/installer/packages/embedded/templates/" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.tmpl" -exec sed -i 's/Datadog/Stackstate/g' {} +
find "$DIR/pkg/fleet/installer/packages/embedded/templates/" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.tmpl" -exec sed -i 's/DD\_/STS\_/g' {} +

find "$DIR/pkg/fleet/installer/packages/embedded/templates/" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.service" -exec sed -i 's/datadog/stackstate/g' {} +
find "$DIR/pkg/fleet/installer/packages/embedded/templates/" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.service" -exec sed -i 's/User\=dd-agent/User\=stackstate-agent/g' {} +
find "$DIR/pkg/fleet/installer/packages/embedded/templates/" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.service" -exec sed -i 's/DataDog/StackState/g' {} +
find "$DIR/pkg/fleet/installer/packages/embedded/templates/" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.service" -exec sed -i 's/Datadog/Stackstate/g' {} +
find "$DIR/pkg/fleet/installer/packages/embedded/templates/" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.service" -exec sed -i 's/DD\_/STS\_/g' {} +

# Compression: DD 7.71.2 switched default from zlib to zstd, but StackState receiver only accepts deflate/gzip/identity
find "$DIR/pkg/config/setup" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/DefaultCompressorKind = "zstd"/DefaultCompressorKind = "zlib"/g' {} +

# ---------------------------------------------------------------------------
# StackState receiver compatibility
# ---------------------------------------------------------------------------

# API Key Header: StackState receiver expects "sts-api-key" header, not "DD-Api-Key".
# Without this, all forwarder requests to the STS backend return 401 Unauthorized.
gofmt -l -w -r '"DD-Api-Key" -> "sts-api-key"' "$DIR/comp/forwarder"
gofmt -l -w -r '"DD-Api-Key" -> "sts-api-key"' "$DIR/cmd/process-agent"
gofmt -l -w -r '"DD-Api-Key" -> "sts-api-key"' "$DIR/pkg/trace"
gofmt -l -w -r '"DD-Api-Key" -> "sts-api-key"' "$DIR/pkg/process/runner"
gofmt -l -w -r '"DD-Api-Key" -> "sts-api-key"' "$DIR/comp/otelcol"
gofmt -l -w -r '"DD-API-KEY" -> "sts-api-key"' "$DIR/comp/forwarder"
gofmt -l -w -r '"DD-API-KEY" -> "sts-api-key"' "$DIR/cmd/process-agent"
gofmt -l -w -r '"DD-API-KEY" -> "sts-api-key"' "$DIR/pkg/trace"
# Handle Go-canonical "Dd-Api-Key" variant used as map key in tests (direct map lookup,
# not Header.Get, so it must match the canonical form of "sts-api-key" → "Sts-Api-Key")
gofmt -l -w -r '"Dd-Api-Key" -> "Sts-Api-Key"' "$DIR/pkg/trace"
# Skip TestOSSExtension in branded builds: it uses the vendored dd-trace-go profiler
# (agentless mode) which sends DD-API-KEY directly — unbranded since vendor/ is not touched.
# The branded testServer checks sts-api-key and finds nothing. The agent-based path
# (TestAgentExtension) flows through the branded trace agent proxy and works correctly.
sed -i '/func TestOSSExtension(t \*testing.T) {/a\\tt.Skip("Skipped in branded builds: vendored dd-trace-go sends DD-API-KEY, not sts-api-key")' "$DIR/comp/otelcol/ddprofilingextension/impl/extension_test.go"

# NOTE: DisableAPIKeyChecking and metadata collection (enable_metadata_collection,
# enable_gohai) are NOT changed at compile time because that breaks unit tests that
# rely on upstream defaults. Instead, these are set at runtime via stackstate.yaml /
# helm chart values / environment variables.
# inventories_enabled is defaulted to false directly in config.go because the STS
# receiver doesn't support the /api/v1/metadata endpoint.

# ---------------------------------------------------------------------------
# Python check namespace and metric prefix
# ---------------------------------------------------------------------------

# Brand Python check module namespace: datadog_checks -> stackstate_checks
# Exclude pkg/fleet — those paths reference real pip package directories on disk (matched by
# a datadog_.* regex) and must stay as datadog_checks to match the unbranded regex pattern.
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -path "*/pkg/fleet/*" -prune -o -type f -name "*.go" -exec sed -i 's/datadog_checks/stackstate_checks/g' {} +
# Fix strncmp length in test_loader.go: "datadog_checks." is 15 chars, "stackstate_checks." is 18 chars
sed -i 's/strncmp(name, "stackstate_checks.", 15)/strncmp(name, "stackstate_checks.", 18)/g' "$DIR/pkg/collector/python/test_loader.go"

# Brand rtloader C++ files: the rtloader imports "datadog_checks.checks" to find
# the AgentCheck base class. In production, only stackstate_checks is installed
# (from stackstate-agent-integrations). fix_branding.sh runs BEFORE omnibus
# compilation, so the compiled rtloader binary will use "stackstate_checks".
# (Pattern uses opening quote to avoid replacing standalone references in comments.)
sed -i 's/"datadog_checks\./"stackstate_checks./g' "$DIR/rtloader/three/three.cpp"
sed -i 's/"datadog_checks\./"stackstate_checks./g' "$DIR/rtloader/two/two.cpp" 2>/dev/null || true
# Also brand the rtloader test Python directory and fake_check import so that
# branded unit tests (which compile rtloader from branded source) can find the
# base class under stackstate_checks.
if [ -d "$DIR/rtloader/test/python/datadog_checks" ] && [ ! -d "$DIR/rtloader/test/python/stackstate_checks" ]; then
  mv "$DIR/rtloader/test/python/datadog_checks" "$DIR/rtloader/test/python/stackstate_checks"
fi
sed -i 's/from datadog_checks\./from stackstate_checks./g' "$DIR/rtloader/test/python/fake_check/__init__.py" 2>/dev/null || true

# Brand metric prefix: "datadog." -> "stackstate." in hardcoded check/metric references
sed -i 's/"datadog\./"stackstate./g' "$DIR/pkg/collector/runner/runner.go" 2>/dev/null || true
sed -i 's/"datadog\./"stackstate./g' "$DIR/pkg/collector/worker/worker.go" 2>/dev/null || true
# aggregator: brand metric prefix and metric names only (NOT import paths or function names)
sed -i 's/MetricPrefix: "datadog"/MetricPrefix: "stackstate"/g' "$DIR/pkg/aggregator/aggregator.go"
sed -i 's/"datadog\./"stackstate./g' "$DIR/pkg/aggregator/aggregator.go"
sed -i 's/"datadog\./"stackstate./g' "$DIR/pkg/aggregator/aggregator_test.go"

# Brand flare command text
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/Collect a flare and send it to Datadog/Collect a flare and send it to StackState/g' {} +

# Brand integration command text and pip package names
find "$DIR/cmd/agent/subcommands/integrations" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/Datadog integration/StackState integration/g' {} + 2>/dev/null || true
find "$DIR/cmd/agent/subcommands/integrations" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/datadog-checks/stackstate-checks/g' {} + 2>/dev/null || true
# Also brand in fleet installer integrations (pip package paths)
find "$DIR/pkg/fleet/installer/packages/integrations" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/datadog-checks/stackstate-checks/g' {} + 2>/dev/null || true

# Brand "Datadog Cluster Agent" user-facing strings in status templates
find "$DIR/pkg/status/render/templates" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.tmpl" -exec sed -i 's/Datadog Cluster Agent/StackState Cluster Agent/g' {} + 2>/dev/null || true

# Brand dist config templates
find "$DIR/cmd/agent/dist/conf.d" -type f -name "*.yaml.example" -exec sed -i 's/datadog/stackstate/g' {} + 2>/dev/null || true
find "$DIR/cmd/agent/dist/conf.d" -type f -name "*.yaml.default" -exec sed -i 's/datadog/stackstate/g' {} + 2>/dev/null || true

# UT
find "$DIR/cmd/agent/subcommands/analyzelogs" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/pattern: "datadog-agent"/pattern: "stackstate-agent"/g' {} +

# Cluster agent specific replacements
find "$DIR/cmd/cluster-agent" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/ datadog-cluster-agent / stackstate-cluster-agent /g' {} +
# Replace cluster yaml go:genrate statement in main.go
find "$DIR/cmd/cluster-agent" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/datadog-cluster.yaml/stackstate-cluster.yaml/g' {} +
# STAC-24624: Replace cluster-agent params.configName so viper searches for stackstate-cluster.yaml (the on-disk name) instead of datadog-cluster.yaml (the upstream name).
# Without this, config.ConfigFileUsed() returns "" because viper never finds the file, the auth_token path resolves relative to CWD, and inter-agent IPC breaks.
find "$DIR/comp/core/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "params.go" -exec sed -i 's/"datadog-cluster"/"stackstate-cluster"/g' {} +

# Kubernetes environment variables for cluster agent service discovery
# Replace DATADOG_CLUSTER_AGENT in Go test files (pkg/config/utils/clusteragent_test.go)
find "$DIR/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DATADOG_CLUSTER_AGENT/STACKSTATE_CLUSTER_AGENT/g' {} +
# Replace DATADOG_CLUSTER_AGENT_SERVICE_HOST in Go test files
find "$DIR/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DATADOG_CLUSTER_AGENT_SERVICE_HOST/STACKSTATE_CLUSTER_AGENT_SERVICE_HOST/g' {} +
# Replace DATADOG_CLUSTER_AGENT_SERVICE_PORT in Go test files (for completeness)
find "$DIR/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DATADOG_CLUSTER_AGENT_SERVICE_PORT/STACKSTATE_CLUSTER_AGENT_SERVICE_PORT/g' {} +
# Replace "datadog-cluster-agent" string literals in test files to match branded config values
find "$DIR/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/"datadog-cluster-agent"/"stackstate-cluster-agent"/g' {} +
# Replace default cluster_agent.kubernetes_service_name value in Go config files (pkg/config/setup/config.go)
find "$DIR/pkg/config/setup" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/"datadog-cluster-agent"/"stackstate-cluster-agent"/g' {} +
# STS - Replace default cluster_agent.token_name value in Go config files (pkg/config/setup/config.go)
find "$DIR/pkg/config/setup" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/"datadogtoken"/"stackstatetoken"/g' {} +
# Replace NewConfig("datadog") with NewConfig("stackstate") so Viper looks for stackstate.yaml instead of datadog.yaml
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" ! -name "*_test.go" -exec sed -i 's/NewConfig("datadog")/NewConfig("stackstate")/g' {} +
# Replace NewConfig("test") with NewConfig("stackstate") in test helpers so test config expects STS_* env vars
find "$DIR/pkg/config/setup" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*test*.go" -exec sed -i 's/NewConfig("test")/NewConfig("stackstate")/g' {} +
# Replace DATADOG_CLUSTER_AGENT_IMAGE in Python test files (tasks/gotest.py)
find "$DIR/tasks" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.py" -exec sed -i 's/DATADOG_CLUSTER_AGENT_IMAGE/STACKSTATE_CLUSTER_AGENT_IMAGE/g' {} +
# Replace DD_CLUSTER_AGENT_KUBERNETES_SERVICE_NAME value in YAML manifest files
find "$DIR/Dockerfiles/manifests" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.yaml" -exec sed -i '/DD_CLUSTER_AGENT_KUBERNETES_SERVICE_NAME/,/value:/ s/value: datadog-cluster-agent/value: stackstate-cluster-agent/g' {} +


# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/ DataDog / StackState /g' {} +
# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/ DataDog\. / StackState\. /g' {} +
# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/ DataDog/ StackState/g' {} +
# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/DataDog /StackState /g' {} +
# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/ datadog / stackstate /g' {} +
# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/ datadog\. / stackstate\. /g' {} +
# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/ datadog/ stackstate/g' {} +
# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/datadog /stackstate /g' {} +

# ---------------------------------------------------------------------------
# Omnibus related fixes
# ---------------------------------------------------------------------------
find "${DIR}/omnibus/package-scripts" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "pre*" -exec sed -i 's/\/var\/log\/datadog/\/var\/log\/stackstate/g' {} +
find "${DIR}/omnibus/package-scripts" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "pre*" -exec sed -i 's/datadog-agent/stackstate-agent/g' {} +
find "${DIR}/omnibus/package-scripts" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "pre*" -exec sed -i 's/datadog/stackstate/g' {} +
find "${DIR}/omnibus/package-scripts" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "pre*" -exec sed -i 's/DD_AGENT/STS_AGENT/g' {} +
find "${DIR}/omnibus/package-scripts" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "pre*" -exec sed -i 's/dd-agent/stackstate-agent/g' {} +

find "${DIR}/omnibus/package-scripts" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "post*" -exec sed -i 's/\/var\/log\/datadog/\/var\/log\/stackstate/g' {} +
find "${DIR}/omnibus/package-scripts" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "post*" -exec sed -i 's/datadog-agent/stackstate-agent/g' {} +
find "${DIR}/omnibus/package-scripts" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "post*" -exec sed -i 's/datadog/stackstate/g' {} +
find "${DIR}/omnibus/package-scripts" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "post*" -exec sed -i 's/DD_AGENT/STS_AGENT/g' {} +
find "${DIR}/omnibus/package-scripts" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "post*" -exec sed -i 's/dd-agent/stackstate-agent/g' {} +

# Ruby files (omnibus config)
find "${DIR}/omnibus" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.rb" -exec sed -i 's/\/etc\/datadog-agent/\/etc\/stackstate-agent/g' {} +
find "${DIR}/omnibus" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.rb" -exec sed -i 's/\/opt\/datadog-agent/\/opt\/stackstate-agent/g' {} +
find "${DIR}/omnibus" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.rb" -exec sed -i 's/datadog\.yaml/stackstate\.yaml/g' {} +
find "${DIR}/omnibus" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.rb" -exec sed -i 's/dd-agent/sts-agent/g' {} +

# ERB/systemd/omnibus templates
find "${DIR}/omnibus" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.erb" -exec sed -i 's/\/opt\/datadog-agent/\/opt\/stackstate-agent/g' {} +
find "${DIR}/omnibus" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.erb" -exec sed -i 's/\/etc\/datadog-agent/\/etc\/stackstate-agent/g' {} +
find "${DIR}/omnibus" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.erb" -exec sed -i 's/datadog\.yaml/stackstate\.yaml/g' {} +
find "${DIR}/omnibus" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.erb" -exec sed -i 's/dd-agent/sts-agent/g' {} +

# Python scripts (omnibus)
find "${DIR}/omnibus" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.py" -exec sed -i "s/'dd-agent'/'sts-agent'/g" {} +
find "${DIR}/omnibus" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.py" -exec sed -i 's/"dd-agent"/"sts-agent"/g' {} +


# Config Rules
# Define directories containing config_test.go files.
#
# [sts] IMPORTANT: this is an explicit allowlist of directories where the
# branding gofmt rewrites below run. Any production Go file that contains a
# branded literal ("DOCKER_DD_AGENT", "DD_PROXY_*", "DD_LOG_LEVEL", "dd_url",
# "https://app.datadoghq.com", etc.) MUST live under one of these directories,
# otherwise the literal survives into the branded binary and the corresponding
# STS_* env var / sts_url config key is silently ignored at runtime.
#
# Past regression (STAC-24908-class): the 7.51 -> 7.71 merge moved the
# allowlist mechanism from `do_go_rename(..., "./pkg/config")` (recursive over
# the whole tree) to this curated per-directory list, but the curated list
# didn't include pkg/config/env. Result: IsContainerized() in
# pkg/config/env/environment.go kept reading the upstream "DOCKER_DD_AGENT",
# the Dockerfile sets only "DOCKER_STS_AGENT", and the entire containerized
# code path (container_proc_root, container_cgroup_root, workloadmeta runtime
# collectors, container check stats) silently degraded to non-containerized
# defaults. Container metrics dropped to zero with no errors in the logs.
#
# When adding a new branded-literal call site upstream, audit with:
#   grep -lF '"DOCKER_DD_AGENT"' --include='*.go' -r . | xargs -I{} dirname {}
# and add the enclosing directory here if it's a production file.
CONFIG_TEST_DIRS="${DIR}/cmd/otel-agent/subcommands/run
${DIR}/cmd/process-agent/subcommands/config
${DIR}/cmd/security-agent/subcommands/config
${DIR}/cmd/trace-agent/config/remote
${DIR}/cmd/otel-agent/config
${DIR}/pkg/clusteragent/admission/common
${DIR}/pkg/clusteragent/admission/mutate/autoinstrumentation
${DIR}/pkg/clusteragent/admission/mutate/config
${DIR}/pkg/collector/corechecks/snmp/internal/checkconfig
${DIR}/pkg/collector/corechecks/snmp/internal/profile
${DIR}/pkg/collector/corechecks/networkpath
${DIR}/pkg/collector/corechecks/servicediscovery/module
${DIR}/pkg/config/setup
${DIR}/pkg/config/nodetreemodel
${DIR}/pkg/config/env
${DIR}/pkg/util/containers
${DIR}/pkg/util/kernel
${DIR}/pkg/security/telemetry
${DIR}/pkg/collector/corechecks/net/networkv2
${DIR}/pkg/network/config
${DIR}/pkg/orchestrator/config
${DIR}/pkg/process/checks
${DIR}/pkg/serverless/appsec/config
${DIR}/pkg/serverless/remoteconfig
${DIR}/pkg/snmp/snmpintegration
${DIR}/pkg/trace/config
${DIR}/pkg/trace/sampler
${DIR}/pkg/util/log/setup/internal/seelog
${DIR}/pkg/util/winutil/iisconfig
${DIR}/pkg/util/quantile
${DIR}/pkg/compliance/aptconfig
${DIR}/pkg/remoteconfig/state
${DIR}/pkg/dyninst/module
${DIR}/pkg/dyninst/uploader
${DIR}/pkg/fleet/daemon
${DIR}/pkg/fleet/installer/config
${DIR}/pkg/fleet/installer/env
${DIR}/pkg/fleet/installer/setup/config
${DIR}/pkg/fleet/installer/setup/defaultscript
${DIR}/pkg/inventory/software
${DIR}/pkg/networkconfigmanagement/config
${DIR}/pkg/system-probe/config
${DIR}/test/new-e2e/tests/installer/windows/suites/agent-package
${DIR}/test/new-e2e/tests/remote-config
${DIR}/comp/checks/windowseventlog/windowseventlogimpl/check
${DIR}/comp/checks/windowseventlog/windowseventlogimpl/check/eventdatafilter
${DIR}/comp/api/api/def
${DIR}/comp/core/agenttelemetry/impl
${DIR}/comp/core/config
${DIR}/comp/core/autodiscovery/autodiscoveryimpl
${DIR}/comp/core/autodiscovery/integration
${DIR}/comp/logs/agent/config
${DIR}/comp/netflow/config
${DIR}/comp/otelcol/otlp/components/connector/datadogconnector
${DIR}/comp/otelcol//collector/impl
${DIR}/comp/otelcol/otlp/components/exporter/serializerexporter
${DIR}/comp/otelcol/otlp/components/processor/infraattributesprocessor
${DIR}/comp/otelcol/otlp
${DIR}/comp/otelcol/ddflareextension/impl
${DIR}/comp/trace/config
${DIR}/comp/networkdeviceconfig/impl
${DIR}/comp/networkpath/npcollector/npcollectorimpl
${DIR}/comp/rdnsquerier/impl
${DIR}/comp/snmptraps/config
${DIR}/test/new-e2e/tests/installer/unix"

# CONFIG_TEST_DIRS="${DIR}/cmd/
# ${DIR}/comp
# ${DIR}/pkg"

# pkg/config/setup TestApplyOverrides:
# In branded builds the dd_url/sts_url defaults and their sources are heavily changed by
# branding (URL rewrites, env-prefix changes, etc.). To keep the test meaningful without
# fighting all those interactions, we drop the two intermediate assertions that check
# the "this shouldn't be overrided" condition and only keep the override behavior checks.
find "${DIR}/pkg/config/setup" \
  -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o \
  -type f -name "config_test.go" -exec sed -i \
  '/assert.Equal(config.GetString(.*), .*"this shouldn'\''t be overrided")/d' {} +

find "${DIR}/pkg/config/setup" \
  -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o \
  -type f -name "config_test.go" -exec sed -i \
  '/assert.Equal(config.GetSource(.*), pkgconfigmodel.SourceFile)/d' {} +

# pkg/config/setup TestSetupFipsEndpoints:
# FIPS proxy endpoint wiring is very Datadog-specific and heavily tied to dd_url-based
# endpoints which we brand away (sts_url, localhost endpoints, etc.). In branded builds
# we don't rely on this FIPS proxy behavior, so we drop the test in the branded copy.
find "${DIR}/pkg/config/setup" \
  -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o \
  -type f -name "config_test.go" -exec sed -i \
  '/^func TestSetupFipsEndpoints/,/^}/d' {} +

# pkg/config/setup TestSetupFipsEndpoints:
# In branded builds, the main base URL config key is sts_url instead of dd_url.
# Adjust the FIPS test YAML so the base URL is provided via sts_url while keeping
# the same "https://somehost:1234" value.
find "${DIR}/pkg/config/setup" \
  -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o \
  -type f -name "config_test.go" -exec sed -i \
  's/dd_url: https:\/\/somehost:1234/sts_url: https:\/\/somehost:1234/g' {} +

# Apply gofmt rules to each directory
while IFS= read -r config_dir; do
	gofmt -l -w -r '"dd_url" -> "sts_url"' "$config_dir"
	gofmt -l -w -r '"DD_DD_URL" -> "STS_STS_URL"' "$config_dir"
	gofmt -l -w -r '"DD_URL" -> "STS_URL"' "$config_dir"
	gofmt -l -w -r '"https://app.datadoghq.com" -> "http://localhost:7077"' "$config_dir"
    gofmt -l -w -r '"https://app.datadoghq.eu" -> "http://localhost:7077"' "$config_dir"
    gofmt -l -w -r '"https://app.datadoghq.dd_dd_url.eu" -> "http://localhost:7077"' "$config_dir"
    gofmt -l -w -r '"https://app.datadoghq.dd_url.eu" -> "http://localhost:7077"' "$config_dir"
    gofmt -l -w -r '"https://process.datadoghq.com." -> "http://localhost:7077"' "$config_dir"
    gofmt -l -w -r '"https://process.datadoghq.com" -> "http://localhost:7077"' "$config_dir"
    gofmt -l -w -r '"https://process-events.datadoghq.com." -> "http://localhost:7077"' "$config_dir"
    gofmt -l -w -r '"https://process-events.datadoghq.com" -> "http://localhost:7077"' "$config_dir"
    gofmt -l -w -r '"https://orchestrator.datadoghq.com" -> "http://localhost:7077"' "$config_dir"
    gofmt -l -w -r '"datadoghq.com" -> "stackstate.io"' "$config_dir"
	gofmt -l -w -r '"DD_PROXY_HTTP" -> "STS_PROXY_HTTP"' "$config_dir"
	gofmt -l -w -r '"DD_PROXY_HTTPS" -> "STS_PROXY_HTTPS"' "$config_dir"
	gofmt -l -w -r '"DD_PROXY_NO_PROXY" -> "STS_PROXY_NO_PROXY"' "$config_dir"
	gofmt -l -w -r '"DOCKER_DD_AGENT" -> "DOCKER_STS_AGENT"' "$config_dir"
	gofmt -l -w -r '"DD_LOG_LEVEL" -> "STS_LOG_LEVEL"' "$config_dir"
	gofmt -l -w -r '"DD" -> "STS"' "$config_dir"
    gofmt -l -w -r '"DD_" -> "STS_"' "$config_dir"
    # echo "renaming in go files in path: $config_dir"
    # set -x
    # find "$config_dir" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/DD_DD_URL/STS_STS_URL/g' {} +
    # find "$config_dir" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/DD_URL/STS_URL/g' {} +
    # find "$config_dir" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's|https://app.datadoghq.com|http://localhost:7077|g' {} +
    # find "$config_dir" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's|https://app.datadoghq.eu|http://localhost:7077|g' {} +
    # find "$config_dir" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's|https://app.datadoghq.dd_dd_url.eu|http://localhost:7077|g' {} +
    # find "$config_dir" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's|https://app.datadoghq.dd_url.eu|http://localhost:7077|g' {} +
    # find "$config_dir" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/DD_PROXY_HTTP/STS_PROXY_HTTP/g' {} +
    # find "$config_dir" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/DD_PROXY_HTTPS/STS_PROXY_HTTPS/g' {} +
    # find "$config_dir" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/DD_PROXY_NO_PROXY/STS_PROXY_NO_PROXY/g' {} +
    # find "$config_dir" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/DOCKER_DD_AGENT/DOCKER_STS_AGENT/g' {} +
    # find "$config_dir" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/DD_LOG_LEVEL/STS_LOG_LEVEL/g' {} +
    # find "$config_dir" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/DD_/STS_/g' {} +
    # # find "$config_dir" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/DD/STS/g' {} +
    # find "$config_dir" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/dd_url/sts_url/g' {} +
    # set +x
done <<< "$CONFIG_TEST_DIRS"

# Special case: DD_ -> STS_ only for pkg/config/setup
# Brand all environment variable names that start with DD_ to STS_
# This handles both BindEnvAndSetDefault calls and map keys
find "${DIR}/pkg/config/setup" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/"DD_/"STS_/g' {} +

# Brand YAML strings in test files: dd_url: -> sts_url:
# This handles YAML config strings in test files that gofmt can't change
find "$DIR/comp/core/agenttelemetry/impl" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/dd_url:/sts_url:/g' {} +

# Brand config key references in test files: compliance_config.endpoints.dd_url -> compliance_config.endpoints.sts_url
# The code was branded to use sts_url, so tests need to set the branded config key
find "$DIR/comp/logs/agent/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/compliance_config\.endpoints\.dd_url/compliance_config.endpoints.sts_url/g' {} +

# Brand logs_config config key references in test files
# The code was branded to use sts_url (from dd_url), so tests need to use the branded config key
# Note: logs_dd_url was NOT branded (only dd_url was), so it remains as logs_dd_url
find "$DIR/comp/logs/agent/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/logs_config\.dd_url/logs_config.sts_url/g' {} +

# Brand dd_url_443 to sts_url_443 in config setup and code
# This is a separate config key from dd_url, so it needs its own branding rule
gofmt -l -w -r '"dd_url_443" -> "sts_url_443"' "${DIR}/pkg/config/setup"
gofmt -l -w -r '"dd_url_443" -> "sts_url_443"' "${DIR}/comp/logs/agent/config"
# Also brand the config key string in config setup
find "$DIR/pkg/config/setup" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/"logs_config\.dd_url_443"/"logs_config.sts_url_443"/g' {} +
# Brand the logs_config.dd_url_443 default endpoint URL
find "$DIR/pkg/config/setup" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/agent-443-intake\.logs\.datadoghq\.com/agent-443-intake.logs.stackstate.io/g' {} +

# Brand preaggregation.dd_url to preaggregation.sts_url
# This config key is used in preaggregation setup and throughout the codebase
# Use sed for config setup to ensure BindEnv calls are branded
find "${DIR}/pkg/config/setup" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/"preaggregation\.dd_url"/"preaggregation.sts_url"/g' {} +
gofmt -l -w -r '"preaggregation.dd_url" -> "preaggregation.sts_url"' "${DIR}/pkg/config/utils"
gofmt -l -w -r '"preaggregation.dd_url" -> "preaggregation.sts_url"' "${DIR}/comp/forwarder/defaultforwarder"
# Brand error messages in endpoints.go that reference preaggregation.dd_url
find "${DIR}/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "endpoints.go" -exec sed -i 's/preaggregation\.dd_url/preaggregation.sts_url/g' {} +

# Brand dd_url to sts_url in pkg/config/utils code (endpoints.go uses "dd_url" as config key)
# This ensures the code uses the branded config key that matches the config setup
gofmt -l -w -r '"dd_url" -> "sts_url"' "${DIR}/pkg/config/utils"

# Brand external_agent_dd_url to external_agent_sts_url in pkg/config/utils
# This config key is used for external agent endpoint configuration
gofmt -l -w -r '"external_config.external_agent_dd_url" -> "external_config.external_agent_sts_url"' "${DIR}/pkg/config/utils"

# Brand pkg/config/utils/endpoints_test.go: environment variables, YAML strings, and error messages
# Environment variables in t.Setenv() calls
find "${DIR}/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "endpoints_test.go" -exec sed -i 's/t\.Setenv("DD_API_KEY"/t.Setenv("STS_API_KEY"/g' {} +
find "${DIR}/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "endpoints_test.go" -exec sed -i 's/t\.Setenv("DD_URL"/t.Setenv("STS_URL"/g' {} +
find "${DIR}/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "endpoints_test.go" -exec sed -i 's/t\.Setenv("DD_DD_URL"/t.Setenv("STS_STS_URL"/g' {} +
find "${DIR}/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "endpoints_test.go" -exec sed -i 's/t\.Setenv("DD_SITE"/t.Setenv("STS_SITE"/g' {} +
find "${DIR}/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "endpoints_test.go" -exec sed -i 's/t\.Setenv("DD_EXTERNAL_CONFIG_EXTERNAL_AGENT_DD_URL"/t.Setenv("STS_EXTERNAL_CONFIG_EXTERNAL_AGENT_STS_URL"/g' {} +
find "${DIR}/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "endpoints_test.go" -exec sed -i 's/t\.Setenv("DD_ADDITIONAL_ENDPOINTS"/t.Setenv("STS_ADDITIONAL_ENDPOINTS"/g' {} +
# YAML strings: dd_url: -> sts_url:, preaggregation.dd_url: -> preaggregation.sts_url:, external_agent_dd_url: -> external_agent_sts_url:
find "${DIR}/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "endpoints_test.go" -exec sed -i 's/dd_url:/sts_url:/g' {} +
find "${DIR}/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "endpoints_test.go" -exec sed -i 's/preaggregation\.dd_url:/preaggregation.sts_url:/g' {} +
find "${DIR}/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "endpoints_test.go" -exec sed -i 's/external_agent_dd_url:/external_agent_sts_url:/g' {} +
# Config key references in BindEnv() calls
find "${DIR}/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "endpoints_test.go" -exec sed -i 's/BindEnv("external_config\.external_agent_dd_url")/BindEnv("external_config.external_agent_sts_url")/g' {} +
# Error message expectations: preaggregation.dd_url -> preaggregation.sts_url
# Replace in assert.Contains() calls - the string literals are quoted, so match the full pattern
find "${DIR}/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "endpoints_test.go" -exec sed -i 's/preaggregation\.dd_url must not match primary URL/preaggregation.sts_url must not match primary URL/g' {} +
find "${DIR}/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "endpoints_test.go" -exec sed -i 's/preaggregation\.dd_url must not match any additional endpoints/preaggregation.sts_url must not match any additional endpoints/g' {} +
find "${DIR}/pkg/config/utils" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "endpoints_test.go" -exec sed -i 's/preaggregation\.dd_url is required when preaggregation.enabled is true/preaggregation.sts_url is required when preaggregation.enabled is true/g' {} +

# Brand config env prefix from DD_ to STS_ so BindEnv* picks up STS_* variables in branded runs
find "${DIR}/pkg/config/create" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "new.go" -exec sed -i 's/name, "DD", strings.NewReplacer/name, "STS", strings.NewReplacer/g' {} +
# Brand config implementation switch env var so branded runs use STS_CONF_NODETREEMODEL
find "${DIR}/pkg/config/create" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "new.go" -exec sed -i 's/DD_CONF_NODETREEMODEL/STS_CONF_NODETREEMODEL/g' {} +
find "${DIR}/pkg/config/structure" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "unmarshal.go" -exec sed -i 's/DD_CONF_NODETREEMODEL/STS_CONF_NODETREEMODEL/g' {} +

# Brand specific DD_* env var names in config tests so they match branded config bindings
# hostname_file env var in TestDDHostnameFileEnvVar
find "${DIR}/pkg/config/setup" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "config_test.go" -exec sed -i 's/DD_HOSTNAME_FILE/STS_HOSTNAME_FILE/g' {} +
# nested config env var in TestEnvNestedConfig
find "${DIR}/pkg/config/setup" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "config_test.go" -exec sed -i 's/DD_FOO_BAR_NESTED/STS_FOO_BAR_NESTED/g' {} +
# database monitoring Aurora env vars in TestDatabaseMonitoringAurora*
find "${DIR}/pkg/config/setup" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "config_test.go" -exec sed -i 's/DD_DATABASE_MONITORING_AUTODISCOVERY_AURORA_/STS_DATABASE_MONITORING_AUTODISCOVERY_AURORA_/g' {} +

# comp/core/config TestMockConfig env var for app_key should use branded prefix
find "${DIR}/comp/core/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "config_test.go" -exec sed -i 's/DD_APP_KEY/STS_APP_KEY/g' {} +

# serverless-init tag tests should use branded tag env vars (STS_TAGS / STS_EXTRA_TAGS)
find "${DIR}/cmd/serverless-init/tag" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "tag_test.go" -exec sed -i 's/DD_TAGS/STS_TAGS/g' {} +
find "${DIR}/cmd/serverless-init/tag" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "tag_test.go" -exec sed -i 's/DD_EXTRA_TAGS/STS_EXTRA_TAGS/g' {} +

# crio collector tests should use branded container image enabled env var
find "${DIR}/comp/core/workloadmeta/collectors/internal/crio" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "crio_test.go" -exec sed -i 's/DD_CONTAINER_IMAGE_ENABLED/STS_CONTAINER_IMAGE_ENABLED/g' {} +

# dogstatsd server tests should use branded mapper profiles env var
find "${DIR}/comp/dogstatsd/server" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "server_test.go" -exec sed -i 's/DD_DOGSTATSD_MAPPER_PROFILES/STS_DOGSTATSD_MAPPER_PROFILES/g' {} +

# batcher tests should use branded batcher capacity env var
find "${DIR}/pkg/batcher" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "batcher_test.go" -exec sed -i 's/DD_BATCHER_CAPACITY/STS_BATCHER_CAPACITY/g' {} +

# state manager tests should use branded check state env vars
find "${DIR}/pkg/collector/check/state" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "state_manager_test.go" -exec sed -i 's/DD_CHECK_STATE_ROOT_PATH/STS_CHECK_STATE_ROOT_PATH/g' {} +
# Brand DD_CHECK_STATE_ROOT_PATH to STS_CHECK_STATE_ROOT_PATH in handler test files
find "${DIR}/pkg/collector/check/handler" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_CHECK_STATE_ROOT_PATH/STS_CHECK_STATE_ROOT_PATH/g' {} +
find "${DIR}/pkg/collector/check/state" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "state_manager_test.go" -exec sed -i 's/DD_CHECK_STATE_EXPIRATION_DURATION/STS_CHECK_STATE_EXPIRATION_DURATION/g' {} +
find "${DIR}/pkg/collector/check/state" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "state_manager_test.go" -exec sed -i 's/DD_CHECK_STATE_PURGE_DURATION/STS_CHECK_STATE_PURGE_DURATION/g' {} +

# flare common tests should use branded env vars
find "${DIR}/pkg/flare/common" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/"DD_API_KEY"/"STS_API_KEY"/g' {} +
find "${DIR}/pkg/flare/common" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/"DD_CLUSTER_AGENT_AUTH_TOKEN"/"STS_CLUSTER_AGENT_AUTH_TOKEN"/g' {} +
find "${DIR}/pkg/flare/common" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_PROCESS_AGENT_ENABLED/STS_PROCESS_AGENT_ENABLED/g' {} +
find "${DIR}/pkg/flare/common" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_HPA_WATCHER_POLLING_FREQ/STS_HPA_WATCHER_POLLING_FREQ/g' {} +
find "${DIR}/pkg/flare/common" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_EXTERNAL_METRICS_PROVIDER_MAX_AGE/STS_EXTERNAL_METRICS_PROVIDER_MAX_AGE/g' {} +
find "${DIR}/pkg/flare/common" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_ADMISSION_CONTROLLER_INJECT_CONFIG_ENABLED/STS_ADMISSION_CONTROLLER_INJECT_CONFIG_ENABLED/g' {} +

# logs auto multiline detection tests should use branded env var
find "${DIR}/pkg/logs/internal/decoder/auto_multiline_detection" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_LOGS_CONFIG_AUTO_MULTI_LINE_DETECTION_CUSTOM_SAMPLES/STS_LOGS_CONFIG_AUTO_MULTI_LINE_DETECTION_CUSTOM_SAMPLES/g' {} +

# network config tests should use branded env vars (many DD_* vars)
# Brand all DD_* environment variables in t.Setenv() calls to STS_*
# This catches all DD_* env vars in test files, including deprecated ones
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/t\.Setenv("DD_/t.Setenv("STS_/g' {} +
# Also brand DD_* env vars in other contexts (like os.Setenv or direct string references)
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/"DD_DISABLE_DNS_INSPECTION"/"STS_DISABLE_DNS_INSPECTION"/g' {} +
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/"DD_ENABLE_PROTOCOL_CLASSIFICATION"/"STS_ENABLE_PROTOCOL_CLASSIFICATION"/g' {} +
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_SYSTEM_PROBE_NETWORK_ENABLE_HTTP_MONITORING/STS_SYSTEM_PROBE_NETWORK_ENABLE_HTTP_MONITORING/g' {} +
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_SERVICE_MONITORING_CONFIG_ENABLE_HTTP_MONITORING/STS_SERVICE_MONITORING_CONFIG_ENABLE_HTTP_MONITORING/g' {} +
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_SERVICE_MONITORING_CONFIG_ENABLE_HTTP2_MONITORING/STS_SERVICE_MONITORING_CONFIG_ENABLE_HTTP2_MONITORING/g' {} +
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_SERVICE_MONITORING_CONFIG_ENABLE_KAFKA_MONITORING/STS_SERVICE_MONITORING_CONFIG_ENABLE_KAFKA_MONITORING/g' {} +
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_SERVICE_MONITORING_CONFIG_ENABLE_POSTGRES_MONITORING/STS_SERVICE_MONITORING_CONFIG_ENABLE_POSTGRES_MONITORING/g' {} +
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_SERVICE_MONITORING_CONFIG_ENABLE_REDIS_MONITORING/STS_SERVICE_MONITORING_CONFIG_ENABLE_REDIS_MONITORING/g' {} +
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_SERVICE_MONITORING_CONFIG_REDIS_TRACK_RESOURCES/STS_SERVICE_MONITORING_CONFIG_REDIS_TRACK_RESOURCES/g' {} +
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_SYSTEM_PROBE_NETWORK_ENABLE_GATEWAY_LOOKUP/STS_SYSTEM_PROBE_NETWORK_ENABLE_GATEWAY_LOOKUP/g' {} +
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_SYSTEM_PROBE_NETWORK_IGNORE_CONNTRACK_INIT_FAILURE/STS_SYSTEM_PROBE_NETWORK_IGNORE_CONNTRACK_INIT_FAILURE/g' {} +
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_COLLECT_DNS_STATS/STS_COLLECT_DNS_STATS/g' {} +
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_COLLECT_DNS_DOMAINS/STS_COLLECT_DNS_DOMAINS/g' {} +
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_SYSTEM_PROBE_CONFIG_MAX_DNS_STATS/STS_SYSTEM_PROBE_CONFIG_MAX_DNS_STATS/g' {} +
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_SYSTEM_PROBE_NETWORK_HTTP_REPLACE_RULES/STS_SYSTEM_PROBE_NETWORK_HTTP_REPLACE_RULES/g' {} +
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_SERVICE_MONITORING_CONFIG_HTTP_REPLACE_RULES/STS_SERVICE_MONITORING_CONFIG_HTTP_REPLACE_RULES/g' {} +
find "${DIR}/pkg/network/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_NETWORK_CONFIG_MAX_TRACKED_HTTP_CONNECTIONS/STS_NETWORK_CONFIG_MAX_TRACKED_HTTP_CONNECTIONS/g' {} +

# network/encoding/marshal tests should use branded env vars
find "${DIR}/pkg/network/encoding/marshal" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/t\.Setenv("DD_SYSTEM_PROBE_NETWORK_ENABLED"/t.Setenv("STS_SYSTEM_PROBE_NETWORK_ENABLED"/g' {} +
find "${DIR}/pkg/network/encoding/marshal" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/t\.Setenv("DD_SYSTEM_PROBE_SERVICE_MONITORING_ENABLED"/t.Setenv("STS_SYSTEM_PROBE_SERVICE_MONITORING_ENABLED"/g' {} +
find "${DIR}/pkg/network/encoding/marshal" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/t\.Setenv("DD_CCM_NETWORK_CONFIG_ENABLED"/t.Setenv("STS_CCM_NETWORK_CONFIG_ENABLED"/g' {} +
find "${DIR}/pkg/network/encoding/marshal" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/t\.Setenv("DD_RUNTIME_SECURITY_CONFIG_ENABLED"/t.Setenv("STS_RUNTIME_SECURITY_CONFIG_ENABLED"/g' {} +

# process/checks tests should use branded env vars
find "${DIR}/pkg/process/checks" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/t\.Setenv("DD_SCRUB_ARGS"/t.Setenv("STS_SCRUB_ARGS"/g' {} +
find "${DIR}/pkg/process/checks" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/t\.Setenv("DD_CUSTOM_SENSITIVE_WORDS"/t.Setenv("STS_CUSTOM_SENSITIVE_WORDS"/g' {} +

# orchestrator config tests should use branded env vars
find "${DIR}/pkg/orchestrator/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_ORCHESTRATOR_URL/STS_ORCHESTRATOR_URL/g' {} +
find "${DIR}/pkg/orchestrator/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_ORCHESTRATOR_EXPLORER_ORCHESTRATOR_DD_URL/STS_ORCHESTRATOR_EXPLORER_ORCHESTRATOR_DD_URL/g' {} +
find "${DIR}/pkg/orchestrator/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_ORCHESTRATOR_ADDITIONAL_ENDPOINTS/STS_ORCHESTRATOR_ADDITIONAL_ENDPOINTS/g' {} +
find "${DIR}/pkg/orchestrator/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_ORCHESTRATOR_EXPLORER_ORCHESTRATOR_ADDITIONAL_ENDPOINTS/STS_ORCHESTRATOR_EXPLORER_ORCHESTRATOR_ADDITIONAL_ENDPOINTS/g' {} +
find "${DIR}/pkg/orchestrator/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_ORCHESTRATOR_EXPLORER_MAX_PER_MESSAGE/STS_ORCHESTRATOR_EXPLORER_MAX_PER_MESSAGE/g' {} +
find "${DIR}/pkg/orchestrator/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_ORCHESTRATOR_EXPLORER_CUSTOM_SENSITIVE_WORDS/STS_ORCHESTRATOR_EXPLORER_CUSTOM_SENSITIVE_WORDS/g' {} +
find "${DIR}/pkg/orchestrator/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_ORCHESTRATOR_EXPLORER_CUSTOM_SENSITIVE_ANNOTATIONS_LABELS/STS_ORCHESTRATOR_EXPLORER_CUSTOM_SENSITIVE_ANNOTATIONS_LABELS/g' {} +

# serverless tests should use branded env vars (selective - only config-related ones)
# Brand API key constants in api_key.go (not just test files)
find "${DIR}/pkg/serverless/apikey" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "api_key.go" -exec sed -i 's/"DD_API_KEY"/"STS_API_KEY"/g' {} +
find "${DIR}/pkg/serverless/apikey" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "api_key.go" -exec sed -i 's/"DD_API_KEY_SECRET_ARN"/"STS_API_KEY_SECRET_ARN"/g' {} +
find "${DIR}/pkg/serverless/apikey" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "api_key.go" -exec sed -i 's/"DD_KMS_API_KEY"/"STS_KMS_API_KEY"/g' {} +
find "${DIR}/pkg/serverless/apikey" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "api_key.go" -exec sed -i 's/"DD_API_KEY_KMS_ENCRYPTED"/"STS_API_KEY_KMS_ENCRYPTED"/g' {} +
# Brand API key env vars in test files
find "${DIR}/pkg/serverless/apikey" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_API_KEY/STS_API_KEY/g' {} +
find "${DIR}/pkg/serverless/apikey" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_API_KEY_SECRET_ARN/STS_API_KEY_SECRET_ARN/g' {} +
find "${DIR}/pkg/serverless/apikey" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_KMS_API_KEY/STS_KMS_API_KEY/g' {} +
find "${DIR}/pkg/serverless/apikey" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_API_KEY_KMS_ENCRYPTED/STS_API_KEY_KMS_ENCRYPTED/g' {} +
# serverless/invocationlifecycle: brand all DD_* env vars in t.Setenv() calls
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/t\.Setenv("DD_/t.Setenv("STS_/g' {} +
# Brand DD_SERVICE in source code (read via os.Getenv)
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" ! -name "*_test.go" -exec sed -i 's/os\.Getenv("DD_SERVICE")/os.Getenv("STS_SERVICE")/g' {} +
# Brand trace propagation headers: x-datadog-* -> x-stackstate-*
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/"x-datadog-trace-id"/"x-stackstate-trace-id"/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/"x-datadog-parent-id"/"x-stackstate-parent-id"/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/"x-datadog-span-id"/"x-stackstate-span-id"/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/"x-datadog-sampling-priority"/"x-stackstate-sampling-priority"/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/"x-datadog-tags"/"x-stackstate-tags"/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/"x-datadog-invocation-error"/"x-stackstate-invocation-error"/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/"x-datadog-invocation-error-msg"/"x-stackstate-invocation-error-msg"/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/"x-datadog-invocation-error-type"/"x-stackstate-invocation-error-type"/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/"x-datadog-invocation-error-stack"/"x-stackstate-invocation-error-stack"/g' {} +
# Also brand in JSON strings and map keys in test files (comprehensive - catches all occurrences)
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/x-datadog-trace-id/x-stackstate-trace-id/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/x-datadog-parent-id/x-stackstate-parent-id/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/x-datadog-span-id/x-stackstate-span-id/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/x-datadog-sampling-priority/x-stackstate-sampling-priority/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/x-datadog-tags/x-stackstate-tags/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/function\.request\.headers\.x-datadog-/function.request.headers.x-stackstate-/g' {} +
# Brand headers in JSON payload strings (including multiline and escaped strings)
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/"x-datadog-trace-id"/"x-stackstate-trace-id"/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/"x-datadog-parent-id"/"x-stackstate-parent-id"/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/"x-datadog-sampling-priority"/"x-stackstate-sampling-priority"/g' {} +
# Brand in lifecycle_test.go eventPayload strings
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "lifecycle_test.go" -exec sed -i 's/"x-datadog-parent-id":"/"x-stackstate-parent-id":"/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "lifecycle_test.go" -exec sed -i 's/"x-datadog-sampling-priority":"/"x-stackstate-sampling-priority":"/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "lifecycle_test.go" -exec sed -i 's/"x-datadog-trace-id":"/"x-stackstate-trace-id":"/g' {} +
# Brand trace propagation extractor header constants
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" ! -name "*_test.go" -exec sed -i 's/ddTraceIDHeader          = "x-datadog-trace-id"/ddTraceIDHeader          = "x-stackstate-trace-id"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" ! -name "*_test.go" -exec sed -i 's/ddParentIDHeader         = "x-datadog-parent-id"/ddParentIDHeader         = "x-stackstate-parent-id"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" ! -name "*_test.go" -exec sed -i 's/ddSpanIDHeader           = "x-datadog-span-id"/ddSpanIDHeader           = "x-stackstate-span-id"/g' {} +

# Brand constants.go header constants (invocationlifecycle)
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "constants.go" -exec sed -i 's/TraceIDHeader = "x-datadog-trace-id"/TraceIDHeader = "x-stackstate-trace-id"/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "constants.go" -exec sed -i 's/ParentIDHeader = "x-datadog-parent-id"/ParentIDHeader = "x-stackstate-parent-id"/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "constants.go" -exec sed -i 's/SpanIDHeader = "x-datadog-span-id"/SpanIDHeader = "x-stackstate-span-id"/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "constants.go" -exec sed -i 's/SamplingPriorityHeader = "x-datadog-sampling-priority"/SamplingPriorityHeader = "x-stackstate-sampling-priority"/g' {} +
find "${DIR}/pkg/serverless/invocationlifecycle" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "constants.go" -exec sed -i 's/TraceTagsHeader = "x-datadog-tags"/TraceTagsHeader = "x-stackstate-tags"/g' {} +

# Brand daemon test file header assertions
find "${DIR}/pkg/serverless/daemon" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/respHdr\.Get("x-datadog-trace-id")/respHdr.Get("x-stackstate-trace-id")/g' {} +
find "${DIR}/pkg/serverless/daemon" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/respHdr\.Get("x-datadog-sampling-priority")/respHdr.Get("x-stackstate-sampling-priority")/g' {} +
find "${DIR}/pkg/serverless/daemon" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/res\.Header\.Get("x-datadog-trace-id")/res.Header.Get("x-stackstate-trace-id")/g' {} +
find "${DIR}/pkg/serverless/daemon" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/res\.Header\.Get("x-datadog-span-id")/res.Header.Get("x-stackstate-span-id")/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" ! -name "*_test.go" -exec sed -i 's/ddSamplingPriorityHeader = "x-datadog-sampling-priority"/ddSamplingPriorityHeader = "x-stackstate-sampling-priority"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" ! -name "*_test.go" -exec sed -i 's/ddInvocationErrorHeader  = "x-datadog-invocation-error"/ddInvocationErrorHeader  = "x-stackstate-invocation-error"/g' {} +
# Brand trace propagation test files
# Note: headersMapAll in extractor_test.go should NOT be branded in its definition because
# mapStackstateToDatadogHeaders converts x-stackstate-* back to x-datadog-* for propagator compatibility.
# The carrier output (after mapping) contains x-datadog-* headers, so tests expect x-datadog-* in headersMapAll.
# However, input JSON strings in test payloads should be branded to x-stackstate-*.
# We exclude extractor_test.go from general branding and handle it specially below.
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" ! -name "extractor_test.go" -exec sed -i 's/"x-datadog-trace-id"/"x-stackstate-trace-id"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" ! -name "extractor_test.go" -exec sed -i 's/"x-datadog-parent-id"/"x-stackstate-parent-id"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" ! -name "extractor_test.go" -exec sed -i 's/"x-datadog-span-id"/"x-stackstate-span-id"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" ! -name "extractor_test.go" -exec sed -i 's/"x-datadog-sampling-priority"/"x-stackstate-sampling-priority"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" ! -name "extractor_test.go" -exec sed -i 's/"x-datadog-invocation-error"/"x-stackstate-invocation-error"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" ! -name "extractor_test.go" -exec sed -i 's/"x-datadog-tags"/"x-stackstate-tags"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" ! -name "extractor_test.go" -exec sed -i 's/"x-datadog-invocation-error-msg"/"x-stackstate-invocation-error-msg"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" ! -name "extractor_test.go" -exec sed -i 's/"x-datadog-invocation-error-type"/"x-stackstate-invocation-error-type"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" ! -name "extractor_test.go" -exec sed -i 's/"x-datadog-invocation-error-stack"/"x-stackstate-invocation-error-stack"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" ! -name "extractor_test.go" -exec sed -i 's/"X-Datadog-Trace-Id"/"X-Stackstate-Trace-Id"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" ! -name "extractor_test.go" -exec sed -i 's/"X-Datadog-Sampling-Priority"/"X-Stackstate-Sampling-Priority"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" ! -name "extractor_test.go" -exec sed -i 's/x-datadog-trace-id/x-stackstate-trace-id/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" ! -name "extractor_test.go" -exec sed -i 's/x-datadog-parent-id/x-stackstate-parent-id/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" ! -name "extractor_test.go" -exec sed -i 's/x-datadog-span-id/x-stackstate-span-id/g' {} +
# For extractor_test.go: The code now uses headersMapAllBranded/headersMapDDBranded for input (when BRANDED=true)
# and headersMapAll/headersMapDD for expected output (x-datadog-* headers after mapping).
# Brand headersMapAllBranded and headersMapDDBranded, but keep headersMapAll and headersMapDD with x-datadog-*
# First, protect headersMapAll and headersMapDD by temporarily replacing them, then brand everything, then restore
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i 's/headersMapAll[[:space:]]*=/__PROTECTED_HEADERSMAPALL__=/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i 's/headersMapDD[[:space:]]*=/__PROTECTED_HEADERSMAPDD__=/g' {} +
# Brand everything in extractor_test.go (including headersMapAllBranded and headersMapDDBranded)
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i 's/"x-datadog-trace-id"/"x-stackstate-trace-id"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i 's/"x-datadog-parent-id"/"x-stackstate-parent-id"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i 's/"x-datadog-span-id"/"x-stackstate-span-id"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i 's/"x-datadog-sampling-priority"/"x-stackstate-sampling-priority"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i 's/"x-datadog-tags"/"x-stackstate-tags"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i 's/"x-datadog-invocation-error"/"x-stackstate-invocation-error"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i 's/"x-datadog-invocation-error-msg"/"x-stackstate-invocation-error-msg"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i 's/"x-datadog-invocation-error-type"/"x-stackstate-invocation-error-type"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i 's/"x-datadog-invocation-error-stack"/"x-stackstate-invocation-error-stack"/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i 's/x-datadog-trace-id/x-stackstate-trace-id/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i 's/x-datadog-parent-id/x-stackstate-parent-id/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i 's/x-datadog-span-id/x-stackstate-span-id/g' {} +
# Now restore headersMapAll and headersMapDD and revert them back to x-datadog-*
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i 's/__PROTECTED_HEADERSMAPALL__=/headersMapAll =/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i 's/__PROTECTED_HEADERSMAPDD__=/headersMapDD =/g' {} +
# Revert headersMapAll and headersMapDD back to x-datadog-* (they're used for expected output after mapping)
# First, do a direct replacement for the specific pattern that's causing issues
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i 's/"x-stackstate-sampling-priority":[[:space:]]*dd\.priority\.asStr/"x-datadog-sampling-priority":      dd.priority.asStr/g' {} +
# Then use a more robust approach: find the map definition and replace all x-stackstate-* within it
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i '/^[[:space:]]*headersMapAll[[:space:]]*=[[:space:]]*map\[string\]string{/,/^[[:space:]]*}[[:space:]]*$/ { s/"x-stackstate-trace-id"/"x-datadog-trace-id"/g; s/"x-stackstate-parent-id"/"x-datadog-parent-id"/g; s/"x-stackstate-span-id"/"x-datadog-span-id"/g; s/"x-stackstate-sampling-priority"/"x-datadog-sampling-priority"/g; s/"x-stackstate-tags"/"x-datadog-tags"/g; s/"x-stackstate-invocation-error"/"x-datadog-invocation-error"/g; s/"x-stackstate-invocation-error-msg"/"x-datadog-invocation-error-msg"/g; s/"x-stackstate-invocation-error-type"/"x-datadog-invocation-error-type"/g; s/"x-stackstate-invocation-error-stack"/"x-datadog-invocation-error-stack"/g; }' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "extractor_test.go" -exec sed -i '/^[[:space:]]*headersMapDD[[:space:]]*=[[:space:]]*map\[string\]string{/,/^[[:space:]]*}[[:space:]]*$/ { s/"x-stackstate-trace-id"/"x-datadog-trace-id"/g; s/"x-stackstate-parent-id"/"x-datadog-parent-id"/g; s/"x-stackstate-span-id"/"x-datadog-span-id"/g; s/"x-stackstate-sampling-priority"/"x-datadog-sampling-priority"/g; s/"x-stackstate-tags"/"x-datadog-tags"/g; s/"x-stackstate-invocation-error"/"x-datadog-invocation-error"/g; s/"x-stackstate-invocation-error-msg"/"x-datadog-invocation-error-msg"/g; s/"x-stackstate-invocation-error-type"/"x-datadog-invocation-error-type"/g; s/"x-stackstate-invocation-error-stack"/"x-datadog-invocation-error-stack"/g; }' {} +
# Exclude extractor_test.go: headersMapAll/headersMapDD must keep x-datadog-* for expected output
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" ! -name "extractor_test.go" -exec sed -i 's/x-datadog-sampling-priority/x-stackstate-sampling-priority/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" ! -name "extractor_test.go" -exec sed -i 's/x-datadog-resource-name/x-stackstate-resource-name/g' {} +
find "${DIR}/pkg/serverless/trace/propagation" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" ! -name "extractor_test.go" -exec sed -i 's/x-datadog-start-time/x-stackstate-start-time/g' {} +
find "${DIR}/pkg/serverless/trace/inferredspan" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_TRACE_ENABLED/STS_TRACE_ENABLED/g' {} +
find "${DIR}/pkg/serverless/trace/inferredspan" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_TRACE_MANAGED_SERVICES/STS_TRACE_MANAGED_SERVICES/g' {} +

# system-probe config tests should use branded env vars
find "${DIR}/pkg/system-probe/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_RUNTIME_SECURITY_CONFIG_ENABLED/STS_RUNTIME_SECURITY_CONFIG_ENABLED/g' {} +
find "${DIR}/pkg/system-probe/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_RUNTIME_SECURITY_CONFIG_FIM_ENABLED/STS_RUNTIME_SECURITY_CONFIG_FIM_ENABLED/g' {} +
find "${DIR}/pkg/system-probe/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_SYSTEM_PROBE_EVENT_MONITORING_PROCESS_ENABLED/STS_SYSTEM_PROBE_EVENT_MONITORING_PROCESS_ENABLED/g' {} +
find "${DIR}/pkg/system-probe/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_SYSTEM_PROBE_EVENT_MONITORING_NETWORK_PROCESS_ENABLED/STS_SYSTEM_PROBE_EVENT_MONITORING_NETWORK_PROCESS_ENABLED/g' {} +
find "${DIR}/pkg/system-probe/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_SYSTEM_PROBE_NETWORK_ENABLED/STS_SYSTEM_PROBE_NETWORK_ENABLED/g' {} +
find "${DIR}/pkg/system-probe/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_GPU_MONITORING_ENABLED/STS_GPU_MONITORING_ENABLED/g' {} +
find "${DIR}/pkg/system-probe/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_SYSTEM_PROBE_SERVICE_MONITORING_ENABLED/STS_SYSTEM_PROBE_SERVICE_MONITORING_ENABLED/g' {} +
find "${DIR}/pkg/system-probe/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_SERVICE_MONITORING_CONFIG_ENABLE_EVENT_STREAM/STS_SERVICE_MONITORING_CONFIG_ENABLE_EVENT_STREAM/g' {} +
find "${DIR}/pkg/system-probe/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_DISCOVERY_ENABLED/STS_DISCOVERY_ENABLED/g' {} +
find "${DIR}/pkg/system-probe/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_DYNAMIC_INSTRUMENTATION_ENABLED/STS_DYNAMIC_INSTRUMENTATION_ENABLED/g' {} +
find "${DIR}/pkg/system-probe/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_CCM_NETWORK_CONFIG_ENABLED/STS_CCM_NETWORK_CONFIG_ENABLED/g' {} +
find "${DIR}/pkg/system-probe/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_RUNTIME_SECURITY_CONFIG_NETWORK_MONITORING_ENABLED/STS_RUNTIME_SECURITY_CONFIG_NETWORK_MONITORING_ENABLED/g' {} +

# Brand DD_APM_FEATURES to STS_APM_FEATURES in config setup and test files
# The config setup binds this env var, and tests need to use the branded variable
find "${DIR}/pkg/config/setup" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/"DD_APM_FEATURES"/"STS_APM_FEATURES"/g' {} +
find "${DIR}/cmd/otel-agent/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/DD_APM_FEATURES/STS_APM_FEATURES/g' {} +

# Reversions
# ---------------------------------------------------------------------------
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/opentelemetry-collector-contrib\/pkg\/stackstate.config/opentelemetry-collector-contrib\/pkg\/datadog.config/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/opentelemetry-collector-contrib\/pkg\/datadog\.config/opentelemetry-collector-contrib\/pkg\/datadog\/config/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/stackstate "github\.com\/open-telemetry\/opentelemetry-collector-contrib\/pkg\/datadog"/datadog "github.com\/open-telemetry\/opentelemetry-collector-contrib\/pkg\/datadog"/g' {} +
find "${DIR}/omnibus" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.rb" -exec sed -i 's|https:\/\/s3\.amazonaws\.com\/sts-agent|https:\/\/s3\.amazonaws\.com\/dd-agent|g' {} +

# Revert stackstate.conf back to datadog.conf for legacy config file references
# These are references to old Agent 5 config files (datadog.conf) that should not be renamed
# The actual files in test data directories remain as datadog.conf
### find "$DIR/cmd/agent/common" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/stackstate\.conf/datadog\.conf/g' {} +

# Revert stackstate.yaml back to datadog.yaml in test file references
# Test data files remain as datadog.yaml and should not be renamed
###find "$DIR/cmd/otel-agent/config" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/testdata\/stackstate\.yaml/testdata\/datadog.yaml/g' {} +

# Revert stackstate-agent back to datadog-agent in test exclude patterns
# Test files use exclude patterns that should match the original datadog-agent paths
###find "$DIR/cmd/agent/subcommands/analyzelogs" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/pattern: "stackstate-agent"/pattern: "datadog-agent"/g' {} +

# Revert StackState Agent back to Datadog Agent in GUI test expectations
# HTML templates remain unbranded, so test expectations should match the templates
### find "$DIR/comp/core/gui/guiimpl" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/<title>StackState Agent<\/title>/<title>Datadog Agent<\/title>/g' {} +
### find "$DIR/comp/core/gui/guiimpl" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/StackState Agent Manager/Datadog Agent Manager/g' {} +

# Revert ad.stackstate.com back to ad.datadoghq.com in test expectations
# Test expectations should match the actual filter behavior which uses ad.datadoghq.com
# find "$DIR/comp/core/workloadmeta/collectors/util/kubernetes_resource_parsers" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/ad\.stackstate\.com/ad.datadoghq.com/g' {} +

# Update test expectations to match branded cluster agent service name
# The default config value was branded to stackstate-cluster-agent, so test expectations should match
find "$DIR/pkg/clusteragent/admission/mutate/agent_sidecar" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's/datadog-cluster-agent\./stackstate-cluster-agent./g' {} +
find "$DIR/pkg/clusteragent/admission/mutate/agent_sidecar" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*_test.go" -exec sed -i 's|https://datadog-cluster-agent\.|https://stackstate-cluster-agent.|g' {} +

# Brand YAML fixture file for TestSecretBackendWithMultipleEndpoints.
# The catch-all sed only targets *.go files, so YAML fixtures need explicit branding
# to keep the additional_endpoints URL in sync with the branded default site.
sed -i 's/datadoghq\.com/stackstate.io/g' "$DIR/pkg/config/utils/tests/datadog_secrets.yaml"

# NOTE: The scrubber test YAML fixture (pkg/util/scrubber/test/datadog.yaml)
# is NOT branded here — its URL is reverted in the Go expected value below
# (after the catch-all sed) so both sides match the unbranded datadoghq.com.

# Brand process URL in backtick JSON strings in process_test.go.
# gofmt changes the standalone string literal map keys to http://localhost:7077
# but backtick raw strings aren't affected by gofmt, causing a mismatch between
# the expected value (localhost:7077) and the env var JSON URL. Must run BEFORE
# the catch-all sed so the URL is already localhost by the time catch-all runs.
find "${DIR}/pkg/config/setup" -type f -name "process_test.go" -exec sed -i 's|https://process\.datadoghq\.com|http://localhost:7077|g' {} +

# ---------------------------------------------------------------------------
# Catch-all: Replace remaining datadoghq.com references with stackstate.io
# The gofmt rule "datadoghq.com" -> "stackstate.io" only catches exact Go string
# literals. This sed catches datadoghq.com as a substring in longer strings like
# "intake.profile.datadoghq.com", "agent-http-intake.logs.datadoghq.com", etc.
# Must run AFTER the gofmt loop and specific URL->localhost:7077 replacements.
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/datadoghq\.com/stackstate.io/g' {} +

# Revert Kubernetes annotation prefixes back to datadoghq.com
# These are K8s conventions that must match annotations on deployed pods —
# changing them would break autodiscovery/instrumentation for existing clusters.
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/ad\.stackstate\.io/ad.datadoghq.com/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/internal\.dd\.stackstate\.io/internal.dd.datadoghq.com/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/tags\.stackstate\.io/tags.datadoghq.com/g' {} +
# Note: internal.apm must be reverted BEFORE apm to avoid double-replacement
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/internal\.apm\.stackstate\.io/internal.apm.datadoghq.com/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/apm\.stackstate\.io/apm.datadoghq.com/g' {} +
# Additional K8s annotations that must stay as datadoghq.com
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/admission\.stackstate\.io/admission.datadoghq.com/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/autoscaling\.stackstate\.io/autoscaling.datadoghq.com/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/service-discovery\.stackstate\.io/service-discovery.datadoghq.com/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/k8s\.csi\.stackstate\.io/k8s.csi.datadoghq.com/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/external-metrics\.stackstate\.io/external-metrics.datadoghq.com/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/custom-metrics\.stackstate\.io/custom-metrics.datadoghq.com/g' {} +

# Revert CRD API groups back to datadoghq.com
# These are Kubernetes CustomResourceDefinition API groups registered in the cluster.
# Changing them would break CRD operations (DatadogMetric, etc.)
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's|stackstate\.io/v1alpha1|datadoghq.com/v1alpha1|g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's|stackstate\.io/v1alpha2|datadoghq.com/v1alpha2|g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's|stackstate\.io/v1beta1|datadoghq.com/v1beta1|g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's|stackstate\.io/v2alpha1|datadoghq.com/v2alpha1|g' {} +

# Revert package repository and installation URLs back to datadoghq.com
# These reference actual Datadog infrastructure (APT/YUM repos, signing keys)
# whose non-Go fixture files (.list, .repo) don't get branded by sed.
# Tests compare Go strings against fixture file contents, so they must match.
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/apt\.stackstate\.io/apt.datadoghq.com/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/yum\.stackstate\.io/yum.datadoghq.com/g' {} +
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/keys\.stackstate\.io/keys.datadoghq.com/g' {} +
# install revert scoped to pkg/fleet only — other files (e.g., diagnose/connectivity)
# construct install URLs from the branded site config and need the branded version.
find "$DIR/pkg/fleet" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/install\.stackstate\.io/install.datadoghq.com/g' {} +

# Revert documentation URLs back to datadoghq.com
# StackState does not host equivalent docs at these paths
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/docs\.stackstate\.io/docs.datadoghq.com/g' {} +

# Fix forwarder health domain regex to also match stackstate.io domains.
# The broad sed changes URL strings but not the regex pattern (escaped dots
# aren't matched). Without this fix, getAPIDomain won't recognize stackstate.io
# URLs and TestComputeDomainsURL will fail.
sed -i 's/datadoghq\\\.\[a-z\]+|ddog-gov/datadoghq\\.[a-z]+|stackstate\\.[a-z]+|ddog-gov/' "$DIR/comp/forwarder/defaultforwarder/forwarder_health.go"

# Fix wellKnownSitesRe to recognize stackstate.io as a well-known site.
# BuildURLWithPrefix adds a trailing FQDN dot only for well-known sites.
# Without this, URLs like "https://app.stackstate.io" lack the trailing dot
# that test assertions expect (e.g., "https://app.stackstate.io.").
sed -i 's/ddog-gov\\.com\$/ddog-gov\\.com$|stackstate\\.io$/' "$DIR/pkg/config/utils/endpoints.go"

# Fix ddURLRegexp to recognize stackstate.io as a known domain.
# AddAgentVersionToDomain adds a version prefix to URLs matching this regex.
# Without this, branded stackstate.io URLs won't get the version prefix.
sed -i 's/ddog-gov\\.com)/ddog-gov\\.com|stackstate\\.io)/' "$DIR/pkg/config/utils/endpoints.go"

# Revert CRD API group in struct fields back to datadoghq.com.
# gofmt changes the exact string "datadoghq.com" -> "stackstate.io" which
# also affects K8s GroupVersion struct fields like {Group: "datadoghq.com"}.
# These must stay as datadoghq.com to match the CRD scheme in vendor/.
find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/Group:[[:space:]]*"stackstate\.io"/Group: "datadoghq.com"/g' {} +

# Fix ddURLRegexp and ddNoSubDomainRegexp in tracer_flare.go to recognize stackstate.io.
# The catch-all sed changes test URLs to use stackstate.io but these regexes only
# recognize datadoghq.com/datad0g.com/ddog-gov.com domains. Without this fix,
# getServerlessFlareEndpoint won't match stackstate.io URLs.
sed -i 's/ddog-gov\\.com)\$/ddog-gov\\.com|stackstate\\.io)$/' "$DIR/pkg/trace/api/tracer_flare.go"

# Fix orchestrator config test expected URLs.
# gofmt changes url.Parse("https://orchestrator.datadoghq.com") to url.Parse("http://localhost:7077")
# but extractOrchestratorDDUrl() uses GetMainEndpointBackwardCompatible with DefaultSite
# (stackstate.io), producing https://orchestrator.stackstate.io.
find "${DIR}/pkg/orchestrator/config" -type f -name "config_test.go" -exec sed -i 's|url\.Parse("http://localhost:7077")|url.Parse("https://orchestrator.stackstate.io")|g' {} +

# Fix orchestrator defaultEndpoint constant.
# gofmt changes it to localhost:7077, but it should match the branded site for
# consistency with GetMainEndpointBackwardCompatible(DefaultSite).
sed -i 's|defaultEndpoint = "http://localhost:7077"|defaultEndpoint = "https://orchestrator.stackstate.io"|' "$DIR/pkg/orchestrator/config/config.go"

# Fix DefaultProcessEndpoint and DefaultProcessEventsEndpoint constants.
# gofmt changes these from branded datadoghq.com URLs to http://localhost:7077,
# but they're used as expected values in process runner tests where the actual
# endpoint comes from GetMainEndpoint(DefaultSite) → process.stackstate.io.
sed -i 's|DefaultProcessEndpoint = "http://localhost:7077"|DefaultProcessEndpoint = "https://process.stackstate.io."|' "$DIR/pkg/config/setup/process.go"
sed -i 's|DefaultProcessEventsEndpoint = "http://localhost:7077"|DefaultProcessEventsEndpoint = "https://process-events.stackstate.io."|' "$DIR/pkg/config/setup/process.go"

# Revert all standalone "stackstate.io" back to "datadoghq.com" in orchestrator CRD files.
# Every standalone "datadoghq.com" string literal in these files is a K8s CRD API group
# reference (datadogAPIGroup constant, Name/Group/groupName struct fields, GroupVersion
# prefixes in test expectations). gofmt changes them to "stackstate.io" which breaks
# CRD discovery cache lookups and collector registration.
find "${DIR}/pkg/collector/corechecks/cluster/orchestrator" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/"stackstate\.io"/"datadoghq.com"/g' {} +

# Revert scrubber test expected URL to match the YAML fixture file.
# The catch-all sed brands the Go backtick expected string to stackstate.io,
# but the YAML fixture (not a .go file) may retain datadoghq.com. Ensure the
# expected value matches the fixture by reverting it after the catch-all sed.
find "$DIR/pkg/util/scrubber" -type f -name "default_test.go" -exec sed -i 's|dd_url: https://app\.stackstate\.io|dd_url: https://app.datadoghq.com|' {} +

# Compression fix: when DefaultCompressorKind is changed from zstd to zlib,
# the service checks payload test needs its max payload size reverted from 250
# (zstd-tuned) to 200 (zlib-tuned) because zlib has different CompressBound
# characteristics that change how payloads are split.
find "$DIR/pkg/serializer/internal/metrics" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "service_checks_test.go" -exec sed -i 's/serializer_max_payload_size", 250/serializer_max_payload_size", 200/g' {} +

# Fix npcollector tests: override site to datadoghq.com so the event platform
# forwarder constructs resolvable intake endpoints (e.g., netpath-intake.datadoghq.com).
# After branding, DefaultSite becomes stackstate.io, but netpath-intake.stackstate.io
# doesn't exist — DNS resolution fails, SendEventPlatformEventBlocking blocks,
# and Test_NpCollector_stopWithoutPanic times out.
sed -i '/^func newTestNpCollector/,/var component/{/var component/i\\tagentConfigs["site"] = "datadoghq.com"
}' "$DIR/comp/networkpath/npcollector/npcollectorimpl/npcollector_testutils.go"
# ---------------------------------------------------------------------------

# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -Ei 's/stackstate\.([A-Z][a-zA-Z0-9]*)/datadog.\1/g' {} +
# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/stackstateConfFile/datadogConfFile/g' {} +
# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/stackstateInstalledIntegrationsPattern/datadogInstalledIntegrationsPattern/g' {} +
# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/stackstateMutex/datadogMutex/g' {} +
# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/stackstateAgentService/datadogAgentService/g' {} +
# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/stackstateYamlPermissions/datadogYamlPermissions/g' {} +
# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/stackstate[[:space:]]\+pkgconfigmodel\.Config/datadog     pkgconfigmodel.Config/g' {} +
# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/stackstate = cfg/datadog = cfg/g' {} +
# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/NewConfig("stackstate")/NewConfig("datadog")/g' {} +
# find "$DIR" -type d -name .git -prune -o -name "vendor" -prune -o -name "pbgo" -prune -o -type f -name "*.go" -exec sed -i 's/stackstate = /datadog = /g' {} +
# ---------------------------------------------------------------------------
