#!/bin/bash

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=omnibus/package-scripts/gpg_signing_setup.sh
source "${script_dir}/gpg_signing_setup.sh"

# CI_PROJECT_DIR is GitLab's; GITHUB_WORKSPACE is the GitHub Actions equivalent.
PROJECT_DIR="${CI_PROJECT_DIR:-${GITHUB_WORKSPACE:-$(pwd)}}"
PKG_DIR="${PKG_DIR:-${PROJECT_DIR}/outcomes/pkg}"

if [ -z "${STACKSTATE_AGENT_VERSION:-}" ]; then
	# Pick the latest tag by default for our version.
	STACKSTATE_AGENT_VERSION=$(cat "${PROJECT_DIR}/version.txt")
	# But we will be building from the master branch in this case.
fi

echo "Signing stackstate-agent ${STACKSTATE_AGENT_VERSION}"
ls "${PKG_DIR}"/*.*

gpg_signing_setup

debsigs --sign=origin -k "${SIGNING_KEY_ID}" "${PKG_DIR}"/*.deb
