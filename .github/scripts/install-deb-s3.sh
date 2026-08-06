#!/usr/bin/env bash
#
# Install deb-s3 and its full runtime dependency tree from a pinned, checksum
# verified manifest.
#
# deb-s3 runs with the package signing key and the pre-release AWS credentials
# in scope, so it must not be resolved at run time. Every gem is fetched at the
# exact version recorded in the manifest and verified against the SHA256 that
# RubyGems publishes for that release before anything is installed or executed.
#
# Usage: install-deb-s3.sh [manifest]
#
# Regenerating the manifest: fetch each gem and record
# "<sha256>  <name>-<version>.gem", matching the checksum published at
# https://rubygems.org/api/v1/versions/<name>.json for that version.

set -euo pipefail

MANIFEST="${1:-.github/deb-s3-gems.sha256}"

if [[ ! -f "${MANIFEST}" ]]; then
    echo "gem manifest not found: ${MANIFEST}" >&2
    exit 1
fi

MANIFEST_ABS="$(cd "$(dirname "${MANIFEST}")" && pwd)/$(basename "${MANIFEST}")"

WORKDIR="$(mktemp -d)"
trap 'rm -rf "${WORKDIR}"' EXIT

cp "${MANIFEST_ABS}" "${WORKDIR}/gems.sha256"
cd "${WORKDIR}"

while read -r _sha file; do
    [[ -n "${file:-}" ]] || continue
    name="${file%-*}"
    version="${file##*-}"
    version="${version%.gem}"
    echo "fetching ${name} ${version}"
    gem fetch "${name}" --version "${version}" --platform ruby
done < gems.sha256

echo "verifying checksums"
sha256sum --check --strict gems.sha256

echo "installing"
${GEM_INSTALL_SUDO-sudo} gem install --local --no-document --ignore-dependencies ./*.gem

# RubyGems installs versioned binstubs on some distributions (deb-s3.ruby3.2,
# deb-s33.2), so a plain "deb-s3" on PATH is not guaranteed. publish_package.sh
# invokes it by bare name, so link the canonical executable when it is missing.
if ! command -v deb-s3 >/dev/null 2>&1; then
    canonical="$(gem contents deb-s3 | grep -E '/bin/deb-s3$' | head -n 1)"
    if [[ -z "${canonical}" ]]; then
        echo "deb-s3 was installed but its executable could not be located" >&2
        exit 1
    fi
    ${GEM_INSTALL_SUDO-sudo} ln -sf "${canonical}" "${LINK_DIR:-/usr/local/bin}/deb-s3"
fi

# Smoke check: this activates the whole pinned dependency set, so a missing or
# incompatible gem fails here rather than midway through publishing.
echo "verifying deb-s3"
deb-s3 help >/dev/null
echo "deb-s3 ready: $(command -v deb-s3)"
