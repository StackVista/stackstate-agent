#!/bin/bash
# Emit OCI label flags for `docker build` / `docker buildx build` / kaniko.
#
# Each line of stdout is one fully-formed `--label=key=value` argument.
# Capture and splice with:
#   mapfile -t labels < <(./omnibus/package-scripts/oci-labels.sh --image-name ... )
#   docker build "${labels[@]}" ...
#
# Mirrors the canonical SUSE Observability label set documented in
# stackstate-mission-control's official-plans/docker-image-oci-labels.md §4.
# Cross-repo references (kept in sync by convention):
#   - stackgraph: scripts/oci-labels.sh (canonical bash helper)
#   - GitHub Actions composite: image-pipeline/.github/actions/apply-oci-labels
#   - sbt helper: stackstate-sbt-build/src/main/scala/com/stackstate/sbt/oci/OciLabels.scala

set -euo pipefail

product="suse-observability"
source_url="https://gitlab.com/stackvista/agent/stackstate-agent"
documentation_url="https://documentation.suse.com/cloudnative/suse-observability/latest/en/classic.html"
readme_url=""
image_name=""
tag=""
title=""
description=""
component=""
base_image=""
base_digest=""

usage() {
  cat >&2 <<'EOF'
usage: oci-labels.sh --image-name NAME --tag TAG --title TITLE \
                     --description DESC --component COMPONENT \
                     --base-image REF
                     [--base-digest sha256:...] [--readme-url URL]
                     [--product NAME] [--source-url URL] [--documentation-url URL]

When --base-digest is provided, the buildx call to resolve the base image
digest is skipped. Required when running inside a container without docker
CLI (e.g. kaniko-executor).
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --image-name)        image_name=$2;        shift 2 ;;
    --tag)               tag=$2;               shift 2 ;;
    --title)             title=$2;             shift 2 ;;
    --description)       description=$2;       shift 2 ;;
    --component)         component=$2;         shift 2 ;;
    --base-image)        base_image=$2;        shift 2 ;;
    --base-digest)       base_digest=$2;       shift 2 ;;
    --readme-url)        readme_url=$2;        shift 2 ;;
    --product)           product=$2;           shift 2 ;;
    --source-url)        source_url=$2;        shift 2 ;;
    --documentation-url) documentation_url=$2; shift 2 ;;
    -h|--help)           usage; exit 0 ;;
    *)
      echo "oci-labels.sh: unknown argument: $1" >&2
      usage
      exit 64
      ;;
  esac
done

for var in image_name tag title description component base_image; do
  if [[ -z "${!var}" ]]; then
    echo "oci-labels.sh: missing required --${var//_/-}" >&2
    usage
    exit 64
  fi
done

# Canonical reference: always the Rancher registry path, even when the artefact
# is actually pushed to quay.io. Matches apply-oci-labels and OciLabels.compute.
ref_name="registry.rancher.com/suse-observability/${image_name}:${tag}"
if [[ -z "$readme_url" ]]; then
  readme_url="${source_url%/}/-/blob/stackstate-7.71.2/README.md"
fi
revision="${CI_COMMIT_SHA:-$(git rev-parse HEAD)}"
source="${CI_PROJECT_URL:-$source_url}"
created="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

if [[ -z "$base_digest" ]]; then
  # Resolve base.digest by parsing the default `buildx imagetools inspect` output.
  # Avoids `--format` (added in buildx v0.11) and `jq`. For a multi-arch image the
  # first `Digest:` line is the manifest-list digest, which is the canonical
  # pull-by-digest reference.
  base_digest="$(docker buildx imagetools inspect "$base_image" | awk '/^Digest:[[:space:]]+sha256:/ { print $2; exit }')"
fi
if [[ "$base_digest" != sha256:* ]]; then
  echo "oci-labels.sh: could not resolve base digest for $base_image" >&2
  exit 1
fi

emit() { printf -- '--label=%s=%s\n' "$1" "$2"; }

emit "org.opencontainers.image.title"         "$title"
emit "org.opencontainers.image.description"   "$description"
emit "org.opencontainers.image.version"       "$tag"
emit "org.opencontainers.image.ref.name"      "$ref_name"
emit "org.opencontainers.image.source"        "$source"
emit "org.opencontainers.image.revision"      "$revision"
emit "org.opencontainers.image.created"       "$created"
emit "org.opencontainers.image.base.name"     "$base_image"
emit "org.opencontainers.image.base.digest"   "$base_digest"
emit "org.opencontainers.image.vendor"        "SUSE LLC"
emit "org.opencontainers.image.authors"       "suse-observability-ops@suse.com"
emit "org.opencontainers.image.url"           "$documentation_url"
emit "org.opencontainers.image.documentation" "$documentation_url"
emit "org.opensuse.reference"                 "$ref_name"
emit "io.artifacthub.package.logo-url"        ""
emit "io.artifacthub.package.readme-url"      "$readme_url"
emit "org.openbuildservice.disturl"           ""
emit "com.suse.observability.product"         "$product"
emit "com.suse.observability.component"       "$component"

if [[ -n "${CI_JOB_URL:-}" ]]; then
  emit "published-by"                         "$CI_JOB_URL"
fi
