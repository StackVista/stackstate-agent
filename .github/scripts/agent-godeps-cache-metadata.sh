#!/usr/bin/env bash
# Computes the content-addressed tag for the stackstate-agent Go dependency cache
# image (STAC-25429), and the two refs it is reached by:
#   * push ref -> quay.io/stackstate/stackstate-agent-godeps-cache  (authoritative)
#   * pull ref -> ${REGISTRY_HOST}/quay/...                         (registry.tooling proxy)
#
# The cache has its own quay repo rather than sharing sts-ci-images with the ARC
# runner images: stackstate-agent is public, so whatever credential this workflow
# carries is reachable from any org member's same-repo branch. A dedicated repo
# keeps that reach down to a rebuildable cache.

agent_godeps_compute_metadata() {
  : "${QUAY_REGISTRY:?QUAY_REGISTRY is required}"     # quay.io
  : "${REGISTRY_HOST:?REGISTRY_HOST is required}"     # registry.tooling.stackstate.io (quay proxy)
  : "${BASE_IMAGE_TAG:?BASE_IMAGE_TAG is required}"   # datadog_build tag the cache derives FROM
  : "${BASE_IMAGE_NAME:?BASE_IMAGE_NAME is required}" # datadog_build_linux_x64 | datadog_build_linux_arm64
  : "${ARCH:?ARCH is required}"                       # amd64 | arm64

  # The Dockerfile and this script are hashed alongside the module manifests, so
  # changing how the cache is built also rotates the tag.
  #
  # The base image NAME is hashed, not just its tag: the amd64 and arm64
  # datadog_build images share a tag, so hashing the tag alone would let one
  # arch's cache satisfy the other's existence check.
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
