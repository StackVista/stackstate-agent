#!/usr/bin/env bash
# Renovate post-upgrade task: bring every workspace module to the state that
# `inv check-mod-tidy` expects. Renovate itself skips Go artifact updates.
set -euo pipefail

# Renovate installs the latest Go; tidy output depends on the version.
export GOTOOLCHAIN="go$(sed -n 's/^go \([0-9.]*\)$/\1/p' go.work)+auto"

dirs=$(go list -m -f '{{.Dir}}')
[[ -n "$dirs" ]]
mapfile -t modules <<<"$dirs"

tree_state() {
  # Renovate does not commit intent-to-add files, so hash new go.sum files directly.
  {
    git diff
    git ls-files -o --exclude-standard -- '*go.sum' | xargs -r sha256sum
  } | sha256sum
}

# Modules outside go.work that Renovate edited still need their go.sum.
tidy_outside_workspace() {
  local mod dir
  git diff --name-only HEAD -- '*go.mod' | while read -r mod; do
    dir=$(cd "$(dirname "$mod")" && pwd)
    if ! printf '%s\n' "${modules[@]}" | grep -qxF "$dir"; then
      GOWORK=off go -C "$dir" mod tidy
    fi
  done
}

# A tidy reads sibling go.mod files through local replaces, so one pass can
# leave a module stale against a sibling tidied after it.
for _ in 1 2 3 4; do
  before=$(tree_state)
  go work sync
  synced=$(tree_state)
  printf '%s\0' "${modules[@]}" | xargs -0 -P "$(nproc)" -I{} go -C {} mod tidy
  if [[ "$synced" == "$before" && "$(tree_state)" == "$before" ]]; then
    tidy_outside_workspace
    exit 0
  fi
done

echo "Go workspace did not converge after 4 passes" >&2
exit 1
