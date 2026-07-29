# Go dependency cache image for stackstate-agent CI (STAC-25429). Derives FROM the
# datadog_build CI image and bakes the workspace's external Go module
# downloads into GOMODCACHE, so consumer jobs get a warm module cache with no
# per-job `go mod download`/network round-trips. Rebuilt only when the module graph
# hash changes -- the image tag is content-addressed
# (see .github/scripts/agent-godeps-cache-metadata.sh).
#
# Only the module ZIPS ($GOMODCACHE/cache/download) are kept; the extracted module
# trees are deleted before the layer closes and Go re-extracts on demand. Measured on
# a full cache, the extracted trees are ~79% of the bytes but ~97% of the *files*, and
# image pull cost here is bound by filesystem metadata, not transfer: warm-proxy pulls
# of the full-cache image spent 0.7m downloading and 4.9m extracting. Shipping ~400k
# tiny files through overlayfs cost more than it saved (+3.0m pull against ~1.45m of
# avoided in-build downloading, so ~1.5m/job net loss).
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
# the zip cache is kept -- the manifests and extracted trees are removed in the same
# RUN, so the layer diff only ever contains the final state.
#
# The prune needs `chmod -R u+w` first: Go writes the module cache read-only (dirs
# 0555, files 0444) and `rm -rf` cannot descend into it as a non-root user.
# cache/vcs holds full git clones for any module fetched outside the proxy; it is not
# needed to rebuild a module from its zip, so it goes too.
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
    echo "== populated module cache =="; \
    du -sh "${cache}"; \
    echo "files: $(find "${cache}" -type f | wc -l)"; \
    echo "== zip cache =="; \
    du -sh "${cache}/cache/download"; \
    echo "files: $(find "${cache}/cache/download" -type f | wc -l)"; \
    test "$(du -sm "${cache}/cache/download" | cut -f1)" -gt 100; \
    find "${cache}" -mindepth 1 -maxdepth 1 ! -name cache -exec chmod -R u+w {} +; \
    find "${cache}" -mindepth 1 -maxdepth 1 ! -name cache -exec rm -rf {} +; \
    if [ -d "${cache}/cache/vcs" ]; then chmod -R u+w "${cache}/cache/vcs"; rm -rf "${cache}/cache/vcs"; fi; \
    echo "== retained in image =="; \
    du -sh "${cache}"; \
    echo "files: $(find "${cache}" -type f | wc -l)"; \
    rm -rf /workspace
