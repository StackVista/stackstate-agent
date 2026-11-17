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
  # A new go.sum is untracked; intent-to-add makes it visible to git diff.
  git add -N -- '*go.sum'
  git diff | sha256sum
}

# A tidy reads sibling go.mod files through local replaces, so one pass can
# leave a module stale against a sibling tidied after it.
for _ in 1 2 3 4; do
  before=$(tree_state)
  go work sync
  synced=$(tree_state)
  printf '%s\0' "${modules[@]}" | xargs -0 -P "$(nproc)" -I{} go -C {} mod tidy
  if [[ "$synced" == "$before" && "$(tree_state)" == "$before" ]]; then
    exit 0
  fi
done

echo "Go workspace did not converge after 4 passes" >&2
exit 1
