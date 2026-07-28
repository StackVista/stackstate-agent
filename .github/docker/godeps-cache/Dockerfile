# Go dependency cache image for stackstate-agent CI (STAC-25429). Derives FROM the
# datadog_build CI image and bakes the workspace's external Go module
# downloads into GOMODCACHE, so consumer jobs get a warm module cache with no
# per-job `go mod download`/network round-trips. Rebuilt only when the module graph
# hash changes -- the image tag is content-addressed
# (see .github/scripts/agent-godeps-cache-metadata.sh).
#
# No `# syntax=` frontend directive and no BuildKit secret mounts: the base image is
# pulled by the daemon after `docker login` (creds stay in the daemon, not the
# build), so the build itself needs no secrets and no Docker Hub frontend pull.
ARG BASE_IMAGE
FROM ${BASE_IMAGE}

# `read -d ''` and `pipefail` are bashisms and RUN defaults to /bin/sh (dash), where
# the download loop silently iterates zero times and publishes an empty cache.
SHELL ["/bin/bash", "-o", "pipefail", "-c"]

# Pre-download external modules for every module in the workspace, mirroring what
# `inv deps` does per module (tasks/libs/common/go.py) at the same parallelism.
# manifests/ preserves the repo's directory layout so nested go.mod files (and their
# local `replace ../` targets) resolve. No GOFLAGS override: go.work puts this in
# workspace mode, where anything other than -mod=readonly is rejected outright. Only
# the populated module cache is kept -- the manifests are removed from the final image.
COPY manifests/ /workspace/
RUN set -eux; \
    git config --global --add safe.directory '*'; \
    export PATH="${PATH}:/usr/local/go/bin"; \
    cd /workspace; \
    modules="$(find . -type f -name go.mod | wc -l)"; \
    echo "workspace modules: ${modules}"; \
    test "${modules}" -gt 0; \
    find . -type f -name go.mod -print0 \
      | xargs -0 -n1 -P8 bash -c \
          'moddir="$(dirname "$1")"; echo "go mod download :: ${moddir}"; cd "${moddir}" && go mod download' _; \
    cache="$(go env GOMODCACHE)"; \
    du -sh "${cache}"; \
    test "$(du -sm "${cache}" | cut -f1)" -gt 100; \
    rm -rf /workspace
