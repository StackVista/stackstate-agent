#!/usr/bin/env bash
# Reproduce STAC-24773 Phase D1 pip entrypoints inside the STS build container.
# Usage (from repo root):
#   RELOCATED=true BRANDED=true ./local.sh cmd ./scripts/dev/test-python3-pip-entrypoint.sh
# Or directly:
#   docker run ... registry.tooling.stackstate.io/quay/stackstate/datadog_build_linux_x64:<tag> \
#     bash -c 'cd $PWD && ./scripts/dev/test-python3-pip-entrypoint.sh'

set -euo pipefail

REPO_ROOT="${CI_PROJECT_DIR:-$(cd "$(dirname "$0")/../.." && pwd)}"
# Branded CI runs fix_branding.sh which rewrites omnibus install_dir to /opt/stackstate-agent.
DEST="${DEST:-/opt/stackstate-agent}"
export HOME="${HOME:-/tmp/bazel-home}"

mkdir -p "$HOME"
rm -rf "$DEST"
mkdir -p "$DEST"

cd "$REPO_ROOT"

echo "=== STAC-24773 pip entrypoint probe (destdir=$DEST) ==="
echo "bazelisk: $(command -v bazelisk)"

echo ""
echo "--- @cpython//:install ---"
bazelisk run --downloader_config=/dev/null -- @cpython//:install --destdir="$DEST"

echo ""
echo "--- replace_prefix (matches python3.rb) ---"
bazelisk run --downloader_config=/dev/null -- //bazel/rules:replace_prefix \
  --prefix "$DEST/embedded" \
  "$DEST/embedded/lib/libpython3".*.so \
  "$DEST/embedded/lib/python3.13/lib-dynload"/*.so \
  "$DEST/embedded/bin/python3"*

export LD_LIBRARY_PATH="$DEST/embedded/lib${LD_LIBRARY_PATH:+:$LD_LIBRARY_PATH}"

echo ""
echo "--- embedded/bin after @cpython ---"
ls -la "$DEST/embedded/bin/" | grep -E 'pip|python' || ls -la "$DEST/embedded/bin/"

echo ""
echo "--- pip3.13 shebang from @cpython (wrong for branded install_dir) ---"
head -1 "$DEST/embedded/bin/pip3.13"

echo ""
echo "--- python3 -m pip (bundled) ---"
"$DEST/embedded/bin/python3" -m pip --version

echo ""
echo "--- pip install pip==26.0.1 (CVE fix, matches python3.rb) ---"
"$DEST/embedded/bin/python3" -m pip install pip==26.0.1

echo ""
echo "--- re-stamp pip3.* shebangs (matches python3.rb post-upgrade) ---"
for pip_script in "$DEST"/embedded/bin/pip3*; do
  [[ -f "$pip_script" && ! -L "$pip_script" ]] || continue
  sed -i "1s|.*|#!$DEST/embedded/bin/python3|" "$pip_script"
done

echo ""
echo "--- embedded/bin after pip self-upgrade ---"
ls -la "$DEST/embedded/bin/" | grep -E 'pip|python' || ls -la "$DEST/embedded/bin/"

echo ""
echo "--- entrypoint probes ---"
for cmd in pip pip3 pip3.13 python3; do
  p="$DEST/embedded/bin/$cmd"
  if [[ -L "$p" ]]; then
    echo "$cmd: symlink -> $(readlink "$p") exist=$(test -e "$p" && echo yes || echo BROKEN)"
  elif [[ -x "$p" ]]; then
    echo "$cmd: regular file (executable)"
  elif [[ -e "$p" ]]; then
    echo "$cmd: exists but not executable"
  else
    echo "$cmd: MISSING"
  fi
done

echo ""
echo "--- pip3.13 shebang after re-stamp ---"
head -1 "$DEST/embedded/bin/pip3.13"

echo ""
echo "--- direct pip3 execution ---"
if [[ -e "$DEST/embedded/bin/pip3" ]]; then
  "$DEST/embedded/bin/pip3" --version
else
  echo "pip3: cannot execute (ENOENT)"
fi

echo ""
echo "--- python3 -m pip (post-upgrade) ---"
"$DEST/embedded/bin/python3" -m pip --version

echo ""
echo "=== done ==="
