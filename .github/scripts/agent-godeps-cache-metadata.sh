#!/usr/bin/env bash
# Computes the content-addressed tag for the stackstate-agent Go dependency cache
# image (STAC-25429) and the two image refs it is referenced by:
#   * push  ref -> quay.io/stackstate/stackstate-agent-godeps-cache  (authoritative)
#   * pull  ref -> ${REGISTRY_HOST}/quay/...                         (registry.tooling proxy)
#
# The cache lives in its own quay repo rather than the shared sts-ci-images, which also
# holds the ARC runner images: stackstate-agent is PUBLIC, so whatever credential this
# workflow carries is reachable from any org member's same-repo branch. A dedicated repo
# keeps that reach down to a rebuildable cache.
#
# Mirrors StackVista/StackGraph .github/scripts/stackgraph-ci-metadata.sh: the tag is
# a hash of the dependency-defining inputs, so an unchanged module graph reuses the
# same image across every branch/PR/default build (the tag *is* the cache key), while
# any change to the graph rotates the tag and a stale cache is never reused.

agent_godeps_compute_metadata() {
  : "${QUAY_REGISTRY:?QUAY_REGISTRY is required}"     # quay.io
  : "${REGISTRY_HOST:?REGISTRY_HOST is required}"     # registry.tooling.stackstate.io (quay proxy)
  : "${BASE_IMAGE_TAG:?BASE_IMAGE_TAG is required}"   # datadog_build tag the cache derives FROM
  : "${BASE_IMAGE_NAME:?BASE_IMAGE_NAME is required}" # datadog_build_linux_x64 | datadog_build_linux_arm64
  : "${ARCH:?ARCH is required}"                       # amd64 | arm64

  # Hash inputs: every module manifest (go.work + go.work.sum + modules.yml + all
  # nested go.mod/go.sum) plus the base image and the two files that define the
  # cache mechanism itself (Dockerfile + this script), so changing how the cache is
  # built also rotates the tag.
  #
  # The base image NAME is hashed, not just its tag: the amd64 and arm64
  # datadog_build images share tag 7af9194f, so hashing the tag alone would give
  # both arches the same cache tag and let one arch's image satisfy the other's
  # existence check. ARCH is in the tag as well, so the collision is impossible
  # by construction and the arch is readable off the ref.
  local godeps_hash
  godeps_hash="$(
    {
      git ls-files -z -- \
        '*go.mod' '*go.sum' 'go.work' 'go.work.sum' 'modules.yml' \
        '.github/docker/godeps-cache/Dockerfile' \
        '.github/scripts/agent-godeps-cache-metadata.sh' \
        | LC_ALL=C sort -z \
        | xargs -0 sha256sum
      printf 'base:%s:%s\n' "${BASE_IMAGE_NAME}" "${BASE_IMAGE_TAG}"
    } | sha256sum | cut -c1-16
  )"

  local image_repo="stackstate/stackstate-agent-godeps-cache"
  local image_tag="godeps-${ARCH}-${godeps_hash}"
  # shellcheck disable=SC2034  # consumed by the sourcing workflow step
  ci_image_push="${QUAY_REGISTRY}/${image_repo}:${image_tag}"
  # shellcheck disable=SC2034  # consumed by the sourcing workflow step
  ci_image="${REGISTRY_HOST}/quay/${image_repo}:${image_tag}"
}
