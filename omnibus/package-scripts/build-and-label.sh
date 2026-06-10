#!/bin/bash
# Wrapper around `docker build` for the stackstate-agent image family. Splices
# canonical SUSE Observability OCI labels via omnibus/package-scripts/oci-labels.sh.
#
# Owns the per-image title/description/component/image-name lookup keyed off the
# Dockerfile directory basename, so the existing publish_image.sh call signature
# stays unchanged. The base image label is derived from the Dockerfile FROM line
# by oci-labels.sh itself.
#
# Usage:
#   build-and-label.sh --build-tag <REPO:TAG> --dockerfile-path <DIR> \
#                      [--build-arg NAME=VALUE ...]

set -euo pipefail

build_tag=""
dockerfile_path=""
build_args=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --build-tag)        build_tag=$2;             shift 2 ;;
    --dockerfile-path)  dockerfile_path=$2;       shift 2 ;;
    --build-arg)        build_args+=("--build-arg" "$2"); shift 2 ;;
    *) echo "build-and-label.sh: unknown argument: $1" >&2; exit 64 ;;
  esac
done

for var in build_tag dockerfile_path; do
  if [[ -z "${!var}" ]]; then
    echo "build-and-label.sh: missing required --${var//_/-}" >&2
    exit 64
  fi
done

if [[ ! -f "${dockerfile_path}/Dockerfile" ]]; then
  echo "build-and-label.sh: ${dockerfile_path}/Dockerfile does not exist" >&2
  exit 1
fi

# Per-image metadata keyed off the Dockerfile directory name. Keep in sync with
# stackstate-mission-control's official-plans/docker-image-oci-labels.md §6.D.
case "$(basename "$dockerfile_path")" in
  agent)
    image_name="stackstate-k8s-agent"
    title="SUSE Observability Agent"
    description="Node-level agent collecting metrics, traces, and topology for SUSE Observability."
    component="stackstate-k8s-agent"
    ;;
  cluster-agent)
    image_name="stackstate-k8s-cluster-agent"
    title="SUSE Observability Cluster Agent"
    description="Kubernetes cluster agent collecting cluster-level topology for SUSE Observability."
    component="stackstate-k8s-cluster-agent"
    ;;
  *)
    echo "build-and-label.sh: no metadata mapping for $(basename "$dockerfile_path")" >&2
    exit 1
    ;;
esac

# BUILD_TAG has the form "<repo>:<tag>". The tag is what goes into
# org.opencontainers.image.version and the ref.name suffix.
tag="${build_tag##*:}"
helper="$(dirname "$0")/oci-labels.sh"

mapfile -t labels < <(
  "$helper" \
    --image-name "$image_name" \
    --tag "$tag" \
    --title "$title" \
    --description "$description" \
    --component "$component" \
    --dockerfile "${dockerfile_path}/Dockerfile"
)

docker build "${labels[@]}" "${build_args[@]}" \
  -t "$build_tag" \
  "$dockerfile_path"
