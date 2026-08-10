#!/bin/bash

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=omnibus/package-scripts/gpg_signing_setup.sh
source "${script_dir}/gpg_signing_setup.sh"

TARGET_BUCKET="${1:-}"
if [ -z "${TARGET_BUCKET}" ]; then
	echo "Missing S3 bucket parameter" >&2
	exit 1
fi

# CI_PROJECT_DIR is GitLab's; GITHUB_WORKSPACE is the GitHub Actions equivalent.
PROJECT_DIR="${CI_PROJECT_DIR:-${GITHUB_WORKSPACE:-$(pwd)}}"
PKG_DIR="${PKG_DIR:-${PROJECT_DIR}/outcomes/pkg}"

CODENAME="${2:-${CI_COMMIT_REF_NAME:-${GITHUB_REF_NAME:-}}}"
TARGET_CODENAME="${CODENAME:-dirty}"

if [ -z "${STACKSTATE_AGENT_VERSION:-}" ]; then
	STACKSTATE_AGENT_VERSION=$(cat "${PROJECT_DIR}/version.txt")
fi

echo "Publishing stackstate-agent ${STACKSTATE_AGENT_VERSION} to ${TARGET_BUCKET} (${TARGET_CODENAME})"
ls "${PKG_DIR}"/*.*

gpg_signing_setup

deb-s3 upload --sign="${SIGNING_KEY_ID}" --codename "${TARGET_CODENAME}" --bucket "${TARGET_BUCKET}" "${PKG_DIR}"/*.deb
