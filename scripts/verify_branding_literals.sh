#!/usr/bin/env bash
# Verify fix_branding.sh rewrote containerization-related upstream literals.
# Run after fix_branding.sh in branded CI jobs.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${ROOT}"

failed=0

check_no_match() {
  local label="$1"
  local pattern="$2"
  shift 2
  local files=("$@")
  if matches=$(grep -nE "${pattern}" "${files[@]}" 2>/dev/null || true); then
    if [ -n "${matches}" ]; then
      echo "ERROR: ${label} still present after fix_branding.sh:"
      echo "${matches}"
      failed=1
    fi
  fi
}

# IsContainerized() must read DOCKER_STS_AGENT (direct sed + CONFIG_TEST_DIRS gofmt).
check_no_match "DOCKER_DD_AGENT getenv in environment.go" \
  'Getenv\("DOCKER_DD_AGENT"\)' \
  pkg/config/env/environment.go

# Container env-var filtering and security telemetry match on the branded name.
check_no_match "DOCKER_DD_AGENT literal in container/security paths" \
  '"DOCKER_DD_AGENT"' \
  pkg/util/containers/env_vars_filter.go \
  pkg/security/telemetry/telemetry.go \
  pkg/util/kernel/fs.go \
  pkg/collector/corechecks/net/networkv2/network.go

if [ "${failed}" -ne 0 ]; then
  echo "Branding verification failed. Add missing directories to CONFIG_TEST_DIRS in fix_branding.sh."
  exit 1
fi

echo "Branding literal verification passed."
