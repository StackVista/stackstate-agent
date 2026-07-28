# Go dependency cache image for stackstate-agent CI (STAC-25429). Derives FROM the
# private datadog_build CI image and bakes the workspace's external Go module
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

# Pre-download external modules for every module in the workspace. manifests/
# preserves the repo's directory layout so nested go.mod files (and their local
# `replace ../` targets) resolve. GOFLAGS=-mod=mod lets `go mod download` reconcile
# go.sum in-container without a committed change; only the populated module cache is
# kept -- the manifests themselves are removed from the final image.
COPY manifests/ /workspace/
RUN set -eux; \
    git config --global --add safe.directory '*'; \
    export PATH="${PATH}:/usr/local/go/bin"; \
    export GOFLAGS=-mod=mod; \
    cd /workspace; \
    find . -type f -name go.mod -print0 \
      | while IFS= read -r -d '' modfile; do \
          moddir="$(dirname "${modfile}")"; \
          echo "go mod download :: ${moddir}"; \
          ( cd "${moddir}" && go mod download ); \
        done; \
    rm -rf /workspace
