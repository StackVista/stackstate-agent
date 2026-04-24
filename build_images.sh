#!/bin/bash
#
# Build Docker images from artifacts produced by ./local.sh all
#
# This is a convenience wrapper. The image build commands are also available
# from within ./local.sh shell as:
#   ./local.sh build_agent_image
#   ./local.sh build_cluster_agent_image
#   ./local.sh build_images
#
# Usage (from host):
#   ./build_images.sh [agent|cluster-agent|all]
#
# Environment variables:
#   REGISTRY    - Registry to tag for (e.g., registry.example.com)
#   ORG         - Organization/namespace (default: stackstate)
#   IMAGE_TAG   - Image tag (default: git short SHA)
#   ARCH        - Architecture (default: amd64)
#

set -e

WHAT="${1:-all}"

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
export ARCH="${ARCH:-amd64}"
export ORG="${ORG:-stackstate}"
export IMAGE_TAG="${IMAGE_TAG:-$(git -C "$SCRIPT_DIR" rev-parse --short HEAD 2>/dev/null || echo 'local')}"
export CI_PROJECT_DIR="$SCRIPT_DIR"

AGENT_REPO="stackstate-k8s-agent"
CLUSTER_REPO="stackstate-k8s-cluster-agent"

if [ ! -d "$SCRIPT_DIR/outcomes/Dockerfiles" ]; then
    echo "ERROR: outcomes/Dockerfiles not found. Run './local.sh all' first."
    exit 1
fi

case "$WHAT" in
    agent)
        deb_file=$(ls "$SCRIPT_DIR"/outcomes/pkg/stackstate-agent_*_${ARCH}.deb 2>/dev/null | head -1)
        [ -z "$deb_file" ] && { echo "ERROR: No .deb found"; exit 1; }
        cp "$deb_file" "$SCRIPT_DIR/outcomes/Dockerfiles/agent/"
        docker build --build-arg ARCH="${ARCH}" \
            -t "${AGENT_REPO}:${IMAGE_TAG}" \
            "$SCRIPT_DIR/outcomes/Dockerfiles/agent"
        [ -n "$REGISTRY" ] && docker tag "${AGENT_REPO}:${IMAGE_TAG}" "${REGISTRY}/${ORG}/${AGENT_REPO}:${IMAGE_TAG}"
        ;;
    cluster-agent)
        binary="$SCRIPT_DIR/bin/stackstate-cluster-agent"
        [ ! -f "$binary" ] && { echo "ERROR: binary not found at $binary"; exit 1; }
        cp "$binary" "$SCRIPT_DIR/outcomes/Dockerfiles/cluster-agent/"
        docker build -t "${CLUSTER_REPO}:${IMAGE_TAG}" \
            "$SCRIPT_DIR/outcomes/Dockerfiles/cluster-agent"
        [ -n "$REGISTRY" ] && docker tag "${CLUSTER_REPO}:${IMAGE_TAG}" "${REGISTRY}/${ORG}/${CLUSTER_REPO}:${IMAGE_TAG}"
        ;;
    all)
        "$0" agent
        "$0" cluster-agent
        ;;
    *)
        echo "Usage: $0 [agent|cluster-agent|all]"
        echo "  REGISTRY=<url> $0 all    # also tag for registry"
        exit 1
        ;;
esac

echo "Done. Images tagged: ${IMAGE_TAG}"
[ -n "$REGISTRY" ] && echo "Registry tags: ${REGISTRY}/${ORG}/${AGENT_REPO}:${IMAGE_TAG}, ${REGISTRY}/${ORG}/${CLUSTER_REPO}:${IMAGE_TAG}"
