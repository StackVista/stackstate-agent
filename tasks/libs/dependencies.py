import json
import os
import sys

from invoke import Exit

from tasks.libs.releasing.json import load_release_json
from tasks.libs.releasing.version import RELEASE_JSON_DEPENDENCIES

# [sts] STS-specific dependency overlay — pinned versions for STS forks of upstream deps
# (stackstate-agent-integrations, process-agent, omnibus-software, omnibus-ruby, jmxfetch).
STACKSTATE_DEPS_FILE = "stackstate-deps.json"


def get_effective_dependencies_env():
    """
    Load dependency versions from release.json (DD) and stackstate-deps.json (STS overlay),
    with environment variable overrides. WINDOWS_* dependencies are skipped on non-Windows
    platforms.

    Precedence (lowest to highest): release.json -> stackstate-deps.json -> env vars.

    Returns:
        dict: Environment dictionary with dependency versions as strings.
    """
    release = load_release_json()
    if not (release_dependencies := release.get(RELEASE_JSON_DEPENDENCIES)):
        raise Exit(f"Could not find {RELEASE_JSON_DEPENDENCIES!r} in release.json")
    effective_dependencies_env = {}
    for key, value in release_dependencies.items():
        if key.startswith("WINDOWS_") and sys.platform != "win32":
            print(f"Ignoring {key!r} on {sys.platform}", file=sys.stderr)
            continue
        # windows runners don't accept anything else than strings in the environment when running a subprocess.
        effective_dependencies_env[key] = str(value)

    # [sts] Overlay stackstate-deps.json on top of release.json. STS pins for STS-forked
    # deps must take precedence over DD upstream values. Logged with [dep_version] so the
    # values are visible in CI job output.
    if os.path.isfile(STACKSTATE_DEPS_FILE):
        with open(STACKSTATE_DEPS_FILE) as f:
            sts_deps = json.load(f)
        print("[sts] Using StackState dependency overlay:", file=sys.stderr)
        for key, value in sts_deps.items():
            print(f"[dep_version] {key} {value}", file=sys.stderr)
            effective_dependencies_env[key] = str(value)

    # Env vars win over both sources (DD upstream behaviour preserved).
    for key in list(effective_dependencies_env.keys()):
        if override := os.getenv(key):
            print(f"Overriding {key!r}: {effective_dependencies_env[key]!r} -> {override!r}", file=sys.stderr)
            effective_dependencies_env[key] = override

    return effective_dependencies_env
