# Upstream Merge Guide

This document covers the periodic task of merging upstream Datadog Agent changes into the StackState Agent fork. This is not day-to-day work — see [CLAUDE.md](CLAUDE.md) for normal development workflows.

## Overview

The StackState Agent is a fork of the Datadog Agent. Periodically, upstream Datadog releases are merged into the fork to pick up new features, bug fixes, and dependency updates. This is a large, intensive task that touches most of the codebase.

**Fork structure:**
- Main branch is named `stackstate-<DD-version>` after the DD version it tracks (e.g., `stackstate-7.71.2`).
- Each merge produces a new main branch; the previous one is left in place as historical reference.
- A set of named scaffolding branches (`base-*`, `common-ancestor-*`, `backport-*`, `merged-*-to-*`) is used to make the merge tractable — see "Pre-merge: branch setup" below.
- A clean "compare copy" of the repo at a sibling path is useful for diffing post-merge fix-ups against the raw merge point.

## Build & Test Infrastructure

### local.sh

The `local.sh` script orchestrates containerized builds. Key steps:
- `PREP` — rsyncs source into the container, runs `fix_package_paths.sh` (if relocated), runs `fix_branding.sh` (if branded)
- `DEPS_DEB` — installs dependencies, runs `inv deps`, regenerates vendor
- `BUILD_CLUSTER_AGENT` / `BUILD_AGENT` — compiles binaries
- `BUILD_DEB` — builds the .deb package via omnibus
- `UNIT_TESTS` — builds with race detector, runs full test suite

The build container image is `registry.tooling.stackstate.io/quay/stackstate/datadog_build_linux_x64`.

### GitLab CI

- Pipeline structure: parent pipeline triggers bridge jobs, which spawn child pipelines (`agent-x86`, `agent-arm`)
- API base: `https://gitlab.com/api/v4/projects/<PROJECT_ID>`
- Auth: `Authorization: Bearer $GITLAB_TOKEN` (token stored in `.env`)
- Use `[cluster-agent]` in commit messages to run only cluster-agent pipeline steps
- The `branded_unit_tests` job runs `fix_branding.sh` then the full test suite
- The `unbranded_unit_tests` job runs tests without branding (baseline comparison)
- Jobs have `retry: max: 2, when: always` — any single test failure triggers up to 2 retries

#### STS deviations in the unit-test invocation

Both `branded_unit_tests` and `unbranded_unit_tests` in `.gitlab-ci-agent.yml` invoke `inv -e test` with two STS-specific flags that diverge from upstream defaults:

- `--build-exclude=$STS_UT_BUILD_EXCLUDE` — drops build tags for features StackState does not ship in the cluster-agent / node-agent images. The current set is `oracle,trivy,trivy_no_javadb,nvml,jetson,bundle_installer,systemd`. **If a future upstream merge introduces a new heavy build tag for a feature StackState doesn't surface (e.g., a new database integration, GPU/hardware support, vendor SDK), consider adding it to this list to keep CI time bounded.** Service-discovery integrations (`consul`, `etcd`, `zk`, `ncm`) are deliberately kept in.
- `--timeout=600` — bumps Go's per-package test timeout from 180s to 600s. Required because we run `go clean -modcache` at job start, so subprocess-heavy tests like `pkg/collector/corechecks/servicediscovery/apm.TestGoDetector` (which shells out to `go build` four times to compile fixture binaries) can blow the default 3-minute timeout on a busy runner. Don't drop this without first confirming the modcache wipe is also gone.

## Integrations repo: CI runner image (embedded Python bump)

The Python integrations (`stackstate-agent-integrations`, a **separate** repo)
run in production under the CPython the agent **embeds via omnibus**, not the
system Python. The source of truth for that version is the agent's
`omnibus/config/software/python3.rb` → `default_version` (currently `3.13.14`).
The integrations repo mirrors it in `.python-version` and is supposed to run its
CI on a matching interpreter.

Historically the integrations CI **runner image and venv lagged** behind — stuck
on 3.11 while the embedded Python moved to 3.13.14 — so integration tests ran on
a different Python (and therefore different resolved dependency versions) than
the agent actually ships. **When an upstream merge bumps `python3.rb`
`default_version`, bump the integrations runner image + venv to the same version
as part of the merge.** This is upstream-merge work, not an ad-hoc task.

Files to change in `stackstate-agent-integrations` (replace the old `3.11` /
`py311` with the new `X.Y` / `py3XY`):

- `.setup-scripts/image/Dockerfile` — `FROM registry.tooling.stackstate.io/docker/python:<X.Y.Z>-bullseye`
- `.setup-scripts/image/Makefile` — runner image tag suffix `-py3XY` (the `build`/`push`/`tag_latest`/`push_latest` targets)
- `.setup-scripts/setup_env.sh` — `virtualenv --python=python3.XY` (and the `lib/python3.XY/site-packages` echo)
- `.gitlab-ci.yml` — the `image:` tag (`stackstate-agent-integrations-runner:<DATE>-py3XY`) and `PYTHON_VERSION:`
- `.python-version` — the pinned interpreter (often already bumped ahead of the image)

Then rebuild and publish the runner image:

```
make -C .setup-scripts/image build push tag_latest push_latest
```

This pushes to `registry.tooling.stackstate.io` and **needs registry
credentials** — coordinate with whoever owns SUSE/registry infra (the same people
as the build container image in "Build & Test Infrastructure"). Finally point
`.gitlab-ci.yml`'s `image:` at the new `<DATE>-py3XY` tag the Makefile produced.

- **Why it matters:** the runner image resolves `pydantic` (and the other base
  deps in `stackstate_checks_base/requirements.in`) for *its* Python. A stale
  image hides bugs that only appear on the production interpreter / dependency
  set.
- **Symptom if skipped:** a check passes integrations CI but fails in the shipped
  agent. Concrete case — **STAC-25137** (Rabobank Dynatrace): a pydantic v2
  `dict_type` validation bug only reproduced on the production stack (py3.13 /
  pydantic 2.12.5) while integrations CI was still on py3.11, so CI stayed green.
- **Worked resolution (STAC-25137):** the fix is **three-legged** and all three
  legs ship together or not at all:
  1. **Code fix in `stackstate-agent-integrations`** on a ticket branch
     (`STAC-25137`): tolerant-union validator on
     `HostProperties.customHostMetadata` so dict, list, and str inputs all parse
     successfully. Released as integrations tag **`7.78.2-3`**.
  2. **Agent pin bump** in `stackstate-deps.json`:
     `STACKSTATE_INTEGRATIONS_VERSION: 7.78.2-2 → 7.78.2-3`. Without this the
     agent still ships the broken integrations build.
  3. **Embedded Python bump** in `omnibus/config/software/python3.rb`
     (`3.13.13 → 3.13.14`) so the production interpreter matches what
     integrations CI now tests against. Stale embedded Python is what let the
     bug slip past CI in the first place; bumping it without re-verifying CI
     would re-open the same blind spot on the next merge.

  Customer-validated outcome: Rabobank confirmed missing hosts re-appeared in
  topology once the agent shipping all three changes rolled out.

## Omnibus → Bazel migration (STAC-24773)

Datadog 7.78 moved most native dependencies from `omnibus/config/software/*.rb` to Bazel. STAC-24773 migrates the fork on branch `STAC-24773-bazel-migration` (MR !426). **After that work lands, do not blindly restore deleted `.rb` files from `stackstate-<prev>` during the next upstream merge** — many files were intentionally removed in favor of `bazelisk run` install lines.

**Canonical handoff:** [docs/dev/stac-24773-bazel-migration.md](docs/dev/stac-24773-bazel-migration.md) — status, CI gating, `--downloader_config=/dev/null`, `replace_prefix` pattern, Phase D scope, parallel tickets.

**Still omnibus after Phase B+C:** `python3.rb` (Phase D → `@cpython`, **same STAC-24773 ticket**), `cacerts`, `openssl3`, FIPS provider, and python3's transitive chain (`libffi`, `zlib`, …).

## Branding: datadoghq.com to stackstate.io

All branding transformations live in `fix_branding.sh`. This script runs at build time and must NOT be applied as permanent local code changes — the source tree stays close to upstream for easier future merges.

### Strategy: Broad replacement + comprehensive reverts

1. **gofmt rule**: `gofmt -r '"datadoghq.com" -> "stackstate.io"'` — changes exact standalone Go string literals (e.g., `DefaultSite`). Also applies other gofmt rules for localhost:7077 URL substitutions in specific directories.
2. **Catch-all sed**: `sed 's/datadoghq\.com/stackstate.io/g'` on all `*.go` files — catches `datadoghq.com` as a substring in URLs like `api.datadoghq.com`, `intake.profile.datadoghq.com`, etc.
3. **Targeted reverts** — patterns that must NOT be branded are reverted back to `datadoghq.com`.

### Key insight: escaped dots in Go regex strings

The catch-all sed does NOT match `datadoghq\.com` (with backslash-dot) in source files, because `\.` in the file is two characters (backslash + dot), not a literal dot. This means:
- Go regex patterns like `ad\.datadoghq\.com` are NOT changed by the sed
- But string constants like `"ad.datadoghq.com/"` ARE changed
- This creates regex/constant mismatches that must be fixed by reverting the string constants

### Patterns that must be reverted (in fix_branding.sh)

**K8s annotations** (must stay `datadoghq.com` — K8s protocol):
- `ad`, `internal.dd`, `tags`, `apm`, `internal.apm`
- `admission`, `autoscaling`, `service-discovery`, `k8s.csi`, `external-metrics`, `custom-metrics`

**CRD API groups** (K8s CustomResourceDefinition registrations):
- Version suffixes: `v1alpha1`, `v1alpha2`, `v1beta1`, `v2alpha1`
- Standalone `"datadoghq.com"` in orchestrator CRD files (Group, Name, groupName fields, `datadogAPIGroup` constant)

**Package repository URLs** (reference real Datadog infrastructure):
- `apt`, `yum`, `keys` — global revert
- `install` — scoped to `pkg/fleet/` only (diagnose/connectivity needs branded URLs)

**Documentation URLs**: `docs.datadoghq.com`

**Regex patterns** (must add `stackstate.io` as recognized domain):
- `wellKnownSitesRe` in `pkg/config/utils/endpoints.go` — trailing FQDN dot
- `ddURLRegexp` in `pkg/config/utils/endpoints.go` — `AddAgentVersionToDomain`
- `ddURLRegexp` + `ddNoSubDomainRegexp` in `pkg/trace/api/tracer_flare.go` — separate file from endpoints.go
- Forwarder health domain regex in `comp/forwarder/defaultforwarder/forwarder_health.go`

**Constants overridden by gofmt to `localhost:7077`** (must be fixed to branded URLs):
- `DefaultProcessEndpoint` → `https://process.stackstate.io.`
- `DefaultProcessEventsEndpoint` → `https://process-events.stackstate.io.`
- `defaultEndpoint` (orchestrator) → `https://orchestrator.stackstate.io`
- Test expected values using `url.Parse` in orchestrator `config_test.go`

**YAML fixture files** (catch-all sed only targets `*.go`):
- `pkg/config/utils/tests/datadog_secrets.yaml` — branded explicitly
- `pkg/util/scrubber/test/datadog.yaml` — NOT branded; Go expected value reverted instead

**Compression**: `serializer_max_payload_size` 250 → 200 (zstd → zlib `CompressBound` difference)

**Test DNS resolution**: npcollector tests override `site` to `datadoghq.com` so the event platform forwarder constructs resolvable intake endpoints (`netpath-intake.datadoghq.com` instead of `netpath-intake.stackstate.io`)

### Adding new branding patterns

When upstream introduces new `datadoghq.com` references, most are handled automatically by the catch-all sed. You only need to add to `fix_branding.sh` when:
1. A reference must NOT be branded (add a revert)
2. A Go regex pattern needs to recognize `stackstate.io` (add the domain to the regex)
3. A non-`.go` fixture file needs branding (add explicit sed for that file)
4. A `gofmt` rule produces `localhost:7077` but the correct value is a branded URL (add a fixup)

### CONFIG_TEST_DIRS allowlist: audit after every merge

`fix_branding.sh` rewrites a fixed set of branded Go literals (`"DOCKER_DD_AGENT"`, `"DD_PROXY_*"`, `"DD_LOG_LEVEL"`, `"dd_url"`, `"DD_URL"`, `"DD_DD_URL"`, `"https://app.datadoghq.com/eu"`, the `process`/`process-events`/`orchestrator.datadoghq.com` URLs). These rewrites only run inside the directories listed in the `CONFIG_TEST_DIRS` variable. **Any production Go file outside the allowlist keeps the upstream literal verbatim** — at runtime the agent then looks for `DOCKER_DD_AGENT` / `DD_LOG_LEVEL` / `dd_url` instead of the STS-branded equivalent that's actually set, and the affected code path silently degrades.

This bit us in the 7.51 → 7.71 merge. The 7.51 branding script used `do_go_rename(..., "./pkg/config")` — recursive over the whole tree. The 7.71 rewrite replaced it with the curated per-directory list and dropped `pkg/config/env` from coverage. `pkg/config/env/environment.go:IsContainerized()` kept reading `DOCKER_DD_AGENT`, the Dockerfile sets only `DOCKER_STS_AGENT`, so the function returned `false` in production. Consequence: `container_proc_root`/`container_cgroup_root` defaulted to the agent container's own (empty) paths, workloadmeta's containerd/crio collectors refused to start ("Agent is not running on containerd"), and the `container` check emitted **0 metric samples** with no errors logged. Only the kubelet check still worked, because it uses the kubelet HTTP API and doesn't depend on cgroup paths. See the comment block above `CONFIG_TEST_DIRS` in `fix_branding.sh` for the full failure mode.

**After every merge, audit each branded literal:**

```bash
# Find any production .go file containing the literal that isn't under CONFIG_TEST_DIRS
grep -lF '"DOCKER_DD_AGENT"' -r --include='*.go' --exclude-dir=vendor . | \
  xargs -I{} dirname {} | sort -u
```

Repeat for `"DD_PROXY_HTTP"`, `"DD_PROXY_HTTPS"`, `"DD_PROXY_NO_PROXY"`, `"DD_LOG_LEVEL"`, `"dd_url"`, `"DD_URL"`, `"DD_DD_URL"`. Any directory in the output that isn't covered (including parent coverage — `gofmt` recurses) needs to be added to `CONFIG_TEST_DIRS`. Skip `_test.go`-only dirs and files with `//go:build ... test` (they don't ship in the production binary).

**Verification after patching the allowlist:** dry-run `gofmt` on each new directory and confirm it lists the file you expected:

```bash
gofmt -l -r '"DOCKER_DD_AGENT" -> "DOCKER_STS_AGENT"' pkg/config/env
# -> should print pkg/config/env/environment.go
```

**Verification after a branded build:** the binary should contain *only* the branded forms.

```bash
strings bin/agent/agent | grep -E '^DOCKER_(DD|STS)_AGENT$' | sort -u
# expected: DOCKER_STS_AGENT only
```

The same `strings` sanity check applies to `DD_LOG_LEVEL` / `STS_LOG_LEVEL` etc. when those literals matter for the helm-chart-provided env vars.

**E2E regression:** `beest/tests/k8s/test_receiver_metrics.py::test_container_metrics` queries container CPU/memory metrics (`memRss`, `systemPct`, etc.) via PromQL. It fails when `IsContainerized()` is unbranded because the container check emits zero samples while kubelet metrics may still pass. Strengthening that test's non-zero assertions is the primary guard against this class of silent branding regression.

## Path Relocation (fix_package_paths.sh)

When `RELOCATED=true`, the source is moved from the Datadog import path to the StackState path:
- `github.com/DataDog/datadog-agent` → `github.com/StackVista/stackstate-agent`

This involves rewriting Go import paths, cleaning the module cache, removing `go.sum` and `vendor`, then re-syncing `go work` and re-vendoring.

## StackState-Specific Code That Can Be Lost During Merge

Upstream Datadog merges can silently drop StackState-specific code blocks (usually marked with `// sts begin` / `// sts end` or `// [sts]` comments). These are modifications to upstream files that don't exist in Datadog's codebase. **After every merge, verify these are still present:**

### Delete `.github/` after every merge (mirroring to github)
- **Directory:** `.github/` (entire tree)
- **What:** STS deletes `.github/` from the repo to keep the GitLab→GitHub mirror working. DD's Actions workflows / dependabot config / CODEOWNERS interfere with the mirror push (Actions/permissions checks). Every upstream merge will re-introduce the directory and may add new files inside it, so a fresh `git rm -r .github` is needed after each merge.
- **Why:** mirror reliability — `.github/` content is irrelevant to STS (we use GitLab CI exclusively; the canonical pipelines live in `.gitlab-ci*.yml`).
- **Reference commits on stackstate-7.71.2:** `7f85fc9c2c` (move to `github_disabled/`) + `8cd65c57b3` (delete `github_disabled/`). On future merges, prefer a single `git rm -r .github` commit over cherry-picking those two — DD adds files inside `.github/` between releases that the original move-commit doesn't cover, so cherry-picking leaves leftovers.
- **Symptom if missed:** mirror push to github.com fails; the github-side `master` falls behind GitLab `stackstate-<DD-version>`.

### Tagger: `kube_cluster_name` on all pod tags
- **File:** `comp/core/tagger/collectors/workloadmeta_extract.go` (old path: `pkg/tagger/collectors/workloadmeta_extract.go`)
- **What:** Adds `kube_cluster_name` tag (from `clustername.GetClusterName()`) to all Kubernetes pod tags
- **Why:** vmagent relabel rules derive `cluster_name`, `_k8s_cluster_`, and `_scope_` labels from this tag. Without it, the StackState UI cannot display CPU/memory metrics for containers because MetricBindings use `${tags.cluster-name}` to scope queries.
- **Symptom if missing:** Container CPU/memory columns empty in StackState UI; `cluster_name`, `_k8s_cluster_`, `_scope_` labels absent from all container/kubelet metrics in VictoriaMetrics.
- **Note:** Datadog doesn't need this because they use `expected_tags_duration` to inject host tags at flush time. StackState relies on the tagger instead.

### Config: serializer compression override
- **File:** `pkg/config/setup/config.go` — `DefaultCompressorKind` constant (handled by `fix_branding.sh`, NOT in `stackstate()`)
- **What:** `fix_branding.sh` changes `DefaultCompressorKind = "zstd"` to `"zlib"` and adjusts `serializer_max_payload_size` in tests from 250 → 200
- **Why:** The StackState receiver does not support zstd decompression. It returns HTTP 400 for zstd-compressed payloads, silently breaking host metadata ingestion (`/intake/` endpoint).
- **Important:** Do NOT override this in the `stackstate()` function — it must be done via `fix_branding.sh` because the payload size test tuning (250 vs 200) must match the compressor. Branded tests get both changes; unbranded tests keep zstd + 250.
- **Symptom if missing:** Receiver returns 400 for all agent payloads; host metadata not ingested; metric enrichment stops.

### Config: other StackState defaults
- **File:** `pkg/config/setup/config.go`, in the `stackstate()` function
- **What:** Various StackState-specific defaults (skip leader election, batcher config, transactional forwarder, check state, etc.)
- **Why:** These configure StackState-specific components and disable Datadog-only features.

### Resources metadata provider: disabled (STAC-24623)
- **File:** `cmd/agent/subcommands/run/command.go` — `fx.Supply(resourcesimpl.Disabled())` is supplied before `metadata.Bundle()`.
- **What:** Suppresses the gohai-derived "resources" payload that the node-agent would otherwise post to `/intake/` every 5 minutes (`comp/metadata/resources/resourcesimpl/resources.go`, `defaultCollectInterval = 300s`).
- **Why:** The StackState receiver decodes the payload through `case class Intake` (`stackstate-receiver-project/.../apimodel/Intake.scala`) which mandates a top-level `internalHostname: String`. The resources payload places hostname under `meta.host` instead, so spray-json returns 400 with `"Object is missing required member 'internalHostname'"`. Even when the field is added, the receiver's `Intake.resources: Option[Resources]` is parsed but **never read** by any processor — the payload is wasted bandwidth. 7.51.1 prod has been silently 400ing on this for years; rather than perpetuate the noise, we disable the producer.
- **Do NOT replace this with a serializer-side `internalHostname` injection.** Earlier rebase commits (`2881df138d`, `d9f478c698`) added that injection; it was removed in `8802b7a3` and replaced with this provider-disable in STAC-24623. Generic post-marshal byte injection is the wrong layer — STS payloads carry `internalHostname` structurally (see `pkg/batcher/batcher.go:150`, `comp/metadata/host/hostimpl/utils/common.go:20`, `pkg/serializer/internal/metrics/events.go`).
- **Pattern to watch in future merges:** Any new metadata payload component added to `comp/metadata/` that calls `serializer.SendMetadata` / `SendProcessesMetadata` must either embed `hostMetadataUtils.CommonPayload` (which has `InternalHostname`) or be disabled if the receiver doesn't consume it. Grep `SendMetadata\|SendProcessesMetadata\|SendHostMetadata\|SendAgentchecksMetadata` for new call sites.
- **Cluster-agent and dogstatsd are unaffected:** cluster-agent does not wire `metadata.Bundle()`; dogstatsd already supplies `Disabled()` upstream (`cmd/dogstatsd/subcommands/start/command.go:161`).

### Config: `use_v2_api.series` override
- **File:** `pkg/config/setup/config.go`, in `serializer()` function
- **What:** Forces `use_v2_api.series` to `false`
- **Why:** The StackState receiver only supports the v1 series API.

### Config: `inventories_enabled` override
- **File:** `pkg/config/setup/common_settings.go` (around line 825)
- **What:** Defaults `inventories_enabled` to `false` (DD's default is `true`).
- **Why:** The StackState receiver doesn't expose `/api/v1/metadata`. Leaving inventories enabled means the host, hostgpu, inventoryagent, inventoryhost, inventorychecks, packagesigning, systemprobe, securityagent, haagent (and `hostsysteminfo` in 7.78+) bundle modules all POST payloads that get 404'd. Wasted CPU + bandwidth + log spam.
- **Symptom if missing:** hundreds of `/api/v1/metadata` 404 ERROR lines per beest run (baseline with the override is ~14-21 from the cluster-agent's own pre-payload calls).
- **Re-enable via config/env if needed.**

### Config: `agent_telemetry.enabled` override
- **File:** `pkg/config/setup/common_settings.go` (around line 1416)
- **What:** Defaults `agent_telemetry.enabled` to `false` (DD's default is `true`).
- **Why:** STS doesn't operate an `instrumentation-telemetry-intake.<site>` endpoint. Leaving it enabled means agents try to flush self-telemetry to a DNS-unresolvable hostname every 30s+15min.
- **Symptom if missing:** `failed to flush agent telemetry session: ... dial tcp: lookup instrumentation-telemetry-intake.stackstate.io.: no such host` ERROR lines on every flush.
- **Re-enable via config/env if needed (and provide an actual intake endpoint).**

### Config: `process_config.process_discovery.enabled` override
- **File:** `pkg/config/setup/process.go` (around line 162)
- **What:** Defaults `process_config.process_discovery.enabled` to `false` (DD default is `true`).
- **Why:** STS receiver doesn't expose `/api/v1/discovery` — the process_discovery check posts there every 4h and the agent gets DNS lookup failures for `https://process.<site>/api/v1/discovery`. Previously masked by the helm chart's deprecated `STS_PROCESS_AGENT_ENABLED=false` which globally disabled the process agent. helm-charts commits `28ac7745`+`92742361` correctly moved to granular `STS_PROCESS_CONFIG_{PROCESS,CONTAINER}_COLLECTION_ENABLED=false`, but those only cover two of the three internal gates — process_discovery had to be disabled at the agent level.
- **Symptom if missing:** `Post "https://process.stackstate.io./api/v1/discovery": dial tcp: lookup process.stackstate.io.: no such host` ERROR lines every check cycle on both node-agent and cluster-check-agent.

### Config: silent value-only STS overrides in `common_settings.go` (9 keys)

DD periodically flips these from STS's preferred value back to DD's default. None had `[sts]` comments in 7.71.2, so they're invisible to `[sts]`-grep sweeps. Restore each with a new `[sts]` comment.

| Key | STS value | DD default | Symptom if reverted |
|---|---|---|---|
| `orchestrator_explorer.enabled` | `false` | `true` | 8 RBAC "Failed to watch" + orchestrator.stackstate.io DNS |
| `remote_configuration.enabled` | `false` | `true` | "mkdir /opt/stackstate-agent/run/remote-config.db: permission denied" |
| `container_image.enabled` | `false` | `true` | container-image SBOM payloads to absent intake |
| `container_lifecycle.enabled` | `false` | `true` | container lifecycle events to absent intake |
| `cluster_checks.advanced_dispatching_enabled` | `false` | `true` | "cannot get runner IP from http headers" WARN |
| `cluster_checks.rebalance_with_utilization` | `false` | `true` | pairs with advanced_dispatching |
| `kubernetes_kubelet_host` | `os.Getenv("STS_KUBERNETES_KUBELET_HOST")` | `""` | slow kubelet auto-discovery |
| `kubelet_cache_pods_duration` | `5` | `0` | no /pods cache, more kubelet load |
| `disk_check.use_core_loader` | `false` | `true` | DD's new core loader vs STS Python check |

**Sweep on every merge** with the value-diff script in memory `upstream-merge-validation-pitfalls.md` §3h. The `[sts]`-grep sweep alone is insufficient — silent value-only diffs survive it.

### Topology event serialization
- **File:** `pkg/serializer/internal/metrics/events.go`
- **What:** Serializes `EventContext` field in event payloads
- **Why:** StackState topology events require the context field for proper processing.

### Config: connectivity checker disabled
- **File:** `pkg/config/setup/config.go`, in the `stackstate()` function
- **File:** `comp/connectivitychecker/impl/connectivitychecker.go`
- **What:** `connectivity_checker.enabled` defaults to `false`; the component skips lifecycle/timer registration when disabled.
- **Why:** DD 7.71.2 added a periodic connectivity checker that probes all DD endpoints every 10 minutes. The STS receiver doesn't support many of these endpoints, causing 404s in receiver logs. The `// sts begin/end` guard in `NewComponent` must be preserved.

### RTLoader branding
- **File:** `fix_branding.sh` (applied at build time)
- **What:** Brands C++ rtloader files (header paths, module names)
- **Why:** Python checks loaded via rtloader won't work if the C++ layer references `datadog_agent` instead of `stackstate_agent`.

### Config: `forwarder_max_concurrent_requests` override
- **File:** `pkg/config/setup/config.go`
- **What:** Default must stay at `1` (StackState override). DD upstream changed it from unset to `10` in commit `f4b1c7cc17`.
- **Why:** With concurrent requests > 1, topology snapshot batches (`SnapshotStart` → data → `SnapshotStop`) can arrive out of order at the receiver, causing `DuplicateSnapshotItem` errors in the sync processor. The larger the cluster, the more batches per snapshot, the more likely reordering occurs.
- **Symptom if wrong:** Topology sync thrashing on large clusters — `DuplicateSnapshotItem` and `ComponentForRelationMissing` errors in the sync processor, create/delete churn on topology components.

### Cluster-agent telemetry wrapper: `Write` override to suppress "superfluous WriteHeader" warnings
- **File:** `pkg/clusteragent/api/handler_telemetry.go`
- **What:** `telemetryWriterWrapper` embeds `http.ResponseWriter` and overrides `WriteHeader` to record per-endpoint API telemetry and APM span tags. DD does not override `Write`. When a handler calls `Write` before `WriteHeader`, Go's stdlib implicitly calls `WriteHeader(200)` on the underlying writer — but the wrapper's own `wroteHeader` flag stays `false`, so the next explicit `WriteHeader` forwards a duplicate and Go logs `http: superfluous response.WriteHeader call from ... handler_telemetry.go:87` at WARN level. STS adds a `Write` override that calls `wrapper.WriteHeader(http.StatusOK)` first if needed, keeping the flag in sync.
- **Why:** Customer-facing log noise (military, banking) — the response is correct, but the warning shows up several times per minute as node-agents poll the cluster-agent. Made more visible in DD 7.78 by `ce01e57d3f` / `03a7f6e570` (cluster-agent APM tracing additions), which expanded the wrapper's usage across ~25 API endpoints.
- **Symptom if reverted:** Recurring `Error from the agent http API server: http: superfluous response.WriteHeader call from ... (handler_telemetry.go:87)` WARN bursts in the cluster-agent log, especially around the node-agent's periodic `getCheckConfigs` / `getAllEndpointsCheckConfigs` polls.
- **Regression test:** `TestWithTelemetryWrapper_WriteBeforeWriteHeader_NoDuplicateHeader` in `handler_telemetry_test.go` — counts `WriteHeader` forwards on a wrapping `countingResponseWriter`; must equal 1.

### Check-context "Log receiver not provided" downgraded from Warn to Info
- **Files:** `pkg/collector/aggregator/check_context.go`, `pkg/collector/python/check_context.go`
- **What:** Both files emit `"Log receiver not provided. Logs from integrations will not be collected."` at startup when no log receiver is wired in. STS doesn't configure the integration-logs pipeline by default, so this fires every agent boot — and fires *twice* because STS retains the legacy python-package check context alongside DD 7.78's new aggregator-package one (see `pkg/collector/python/loader.go:109-117` for the rationale). STS downgrades both to `log.Info`.
- **Why:** Customer-facing log noise. Same regulated-customer driver as the `telemetryWriterWrapper` patch — this is the expected STS configuration state, not a problem worth flagging at WARN.
- **Symptom if reverted:** Two `WARN ... Log receiver not provided. Logs from integrations will not be collected.` lines per agent startup, one from each package.

### Cluster-agent server.go: errorLog filter for "superfluous WriteHeader"
- **File:** `cmd/cluster-agent/api/server.go` — `superfluousWriteHeaderFilteringWriter` struct + wrapping of `logWriter` before it's passed to `stdLog.New(...)`.
- **What:** Wraps the `io.Writer` passed to `http.Server.ErrorLog` so any line containing `"superfluous response.WriteHeader call"` is dropped before reaching seelog. This is a belt-and-suspenders companion to the `telemetryWriterWrapper.Write` override above. The structural override fixes the wrapper's bookkeeping; this filter catches double-WriteHeader paths the wrapper can't reach (panic recovery middleware, direct `ResponseWriter.Write` bypasses, anything downstream that calls `WriteHeader` on the underlying writer twice).
- **Why:** The `Write` override alone does not silence every path that produces the WARN. In the 7.78.2 cycle, sandbox soak revealed that the WARN was actually firing through a *cluster-checks dispatch* path tied to a failing check (aws_topology), which goes through different middleware than the wrapper covers. The filter is a small, low-risk insurance policy that guarantees the warning never appears regardless of which handler path produced it.
- **Symptom if reverted:** WARN bursts return on any handler that triggers a double-WriteHeader, *especially* when a cluster-check fails to load and the cluster-agent's reporting handshake with the checks-agent fires repeatedly.
- **Verify on the merged branch:** `grep -n superfluousWriteHeaderFilteringWriter cmd/cluster-agent/api/server.go` — expect the struct definition at the bottom of the file plus the `&superfluousWriteHeaderFilteringWriter{inner: logWriter}` wrapping near the `errorLog := stdLog.New(...)` call.

### Disk default config: `file_system_exclude` instead of deprecated `excluded_filesystems`
- **File:** `cmd/agent/dist/conf.d/disk.d/conf.yaml.default`
- **What:** The default disk-check config ships with `file_system_exclude: [tmpfs, squashfs]` instead of the upstream `excluded_filesystems` key. Both keys map to the same internal config field, but the diskv2 check emits a deprecation WARN at every node-agent startup when the old key is used.
- **Why:** Single-line deprecation noise that fires on every node, and DD will re-introduce the old key on every merge.
- **Symptom if reverted:** `WARN ... excluded_filesystems is deprecated and will be removed in a future release. Please use file_system_exclude instead.` once per node-agent startup.

### Workloadmeta startup race log severity downgrade
- **Files:**
  - `comp/core/workloadmeta/impl/store.go` — `"no workloadmeta collector ready after %s, starting pull loop anyway"` was `Warnf`, downgrade to `Infof`.
  - `comp/core/autodiscovery/autodiscoveryimpl/autoconfig.go` — `"Workloadmeta collectors are not ready after %d retries: ..., starting check scheduler controller anyway."` was `Errorf`, downgrade to `Infof`.
- **What:** On EKS/GKE/etc., the kubelet workloadmeta collector takes longer than the autoconfig retry window to become ready. Both code paths log a single message at startup and then self-recover — neither is actionable.
- **Why:** Each fires exactly once per agent startup; combined they account for one WARN + one ERROR per pod boot. Customer telemetry / paging picks them up despite being benign.
- **Symptom if reverted:** One ERROR + one WARN per agent startup mentioning workloadmeta readiness.

### Diskv2: demote permission-denied mountpoint WARN to DEBUG
- **File:** `pkg/collector/corechecks/system/disk/diskv2/disk.go` — `getPartitionUsage`.
- **What:** When `disk.Usage(mountpoint)` fails, upstream logs `WARN Unable to get disk metrics for <mountpoint>: <err>`. STS adds an `errors.Is(err, fs.ErrPermission)` branch that downgrades the EACCES case to `Debugf`. Real disk problems (NFS hangs, FS corruption, stuck mounts) surface as other errno values and continue to WARN with the original message.
- **Why:** On any Kubernetes cluster, the kubelet exposes per-pod CSI mounts and `volume-subpaths/` bind mounts in the node-agent's mount table but locks their permissions down to the owning pod's UID. The check correctly skips them (returns `nil`), but the upstream WARN fires dozens of times *per collection cycle* — hundreds per minute under steady state. A chart-side mount-path exclusion regex was considered and rejected: it would silently drop legitimate metrics on customers running CSI drivers with looser permissions, and it would fail to match on RKE2/K3s clusters where kubelet lives at a different root path. A permission-aware demotion in the source is path- and topology-agnostic.
- **Symptom if reverted:** 400+ disk-metric WARN lines per node-agent per minute on any Kubernetes deployment with CSI-backed PVCs or `subPath:` mounts.

### Test stability patches we carry on top of upstream

`pkg/logs/client/http/worker_pool_test.go` carries an STS-specific `driveUntil` helper plus an `absDuration` utility, used to absorb a goroutine-scheduling race in `TestRetryableError`, `TestNonRetryableError`, and `TestWorkerCounts`. Without these, the tests flake on busy CI runners with off-by-one worker counts and millisecond-level `assert.InDelta` mismatches on `virtualLatency`. **An upstream merge into `pkg/logs/client/http/` may overwrite this patch — verify the helpers are still present and the `Test*` functions still call `driveUntil(...)` rather than the original fixed-iteration loops.** The original assertions (`for i := 0; i < 100; i++ { pool.run(...) }; require.Equal(t, 8, pool.inUseWorkers)`) compile but flake in CI.

### Monitor identity: volatile metric labels
- **Not an agent code issue** — this is a stackpacks/platform concern, but triggered by agent version changes.
- **What:** The threshold monitor function (`urn:stackpack:common:monitor-function:threshold`) derives `healthStateId` from ALL metric label values. If the agent version adds, removes, or changes any label (e.g., `orch_cluster_id` appearing, `status` flip-flopping), the platform creates duplicate monitor instances for the same component.
- **Affected monitors:** Node Disk/Memory/PID Pressure, Node Readiness, Available Endpoints (fixed in stackpacks MR 1332 by adding `max by (...)` aggregation). Desired-replicas monitors (daemonset/deployment/replicaset/statefulset) are theoretically vulnerable but not currently affected.
- **After merge:** Check if new KSM metrics add labels that differ from the labels used in monitor `urnTemplate` fields. If so, the monitor queries in the kubernetes-v2 stackpack need `by (...)` aggregation to strip volatile labels.

## Pre-merge: branch setup

Before any conflict resolution, set up the branches that the merge will run on. The strategy is to give git a meaningful three-way merge base by replaying StackState's changes onto the upstream commit that the source and target DD versions share. Without this, git treats every line of every file StackState ever touched as a potential conflict.

There is no `upstream` remote in this repo. Pristine DataDog code is fetched from a separate DataDog clone and pushed to `origin` as `base-*` branches.

### The branch graph

For a merge from current DD version `<CURRENT>` to target DD version `<NEXT>`:

```
base-<CURRENT>                            ← pristine DD <CURRENT> upstream (no STS code)
  ↓
stackstate-<CURRENT>                      ← current fork main = base-<CURRENT> + STS changes
  ↓
common-ancestor-<CURRENT>-<NEXT>          ← upstream commit shared by both DD tags
  ↓
backport-<CURRENT>-common-ancestor-<NEXT> ← STS changes replayed onto the common ancestor
  ↓
base-<NEXT>                               ← pristine DD <NEXT> upstream
  ↓
merged-<CURRENT>-to-<NEXT>                ← merge of backport into base-<NEXT>;
                                            conflict resolution and fix-ups land here
  ↓ (merged into via MR at cutover)
stackstate-<NEXT>                         ← new fork main; CREATED EARLY off base-<NEXT>
                                            so SUSE/customization commits can land while
                                            the merge work happens
```

For a solo merge, fix-up commits go directly on `merged-<CURRENT>-to-<NEXT>`. When more than one person is contributing, open per-developer feature branches (any naming) off `merged-<CURRENT>-to-<NEXT>` and merge them back via MR.

### The branches and what each one contains

| Branch | Contents | Created when |
|---|---|---|
| `base-<CURRENT>` | Pristine DD `<CURRENT>` upstream commit, no StackState code | Already exists from the previous merge |
| `stackstate-<CURRENT>` | Current fork main (= `base-<CURRENT>` + all STS changes) | Already exists; this is the live main branch |
| `common-ancestor-<CURRENT>-<NEXT>` | Output of `git merge-base base-<CURRENT> base-<NEXT>` — the upstream commit shared by both DD versions | New, this merge |
| `backport-<CURRENT>-common-ancestor-<NEXT>` | `common-ancestor-...` + every StackState change from `stackstate-<CURRENT>` replayed on top | New, this merge |
| `base-<NEXT>` | Tip of DD's `<MAJOR.MINOR>.x` release branch at prep time, named after the latest released patch (NOT the version tag — see Prep commands note below) | New, this merge |
| `merged-<CURRENT>-to-<NEXT>` | Result of merging `backport-...` into `base-<NEXT>` plus all conflict-resolution and fix-up commits | New, this merge |
| `stackstate-<NEXT>` | Created early off `base-<NEXT>`. Carries the SUSE/customization commit (a single big commit historically — see prep step 6 below) plus any backports owners want to land in parallel with the merge work. Receives `merged-<CURRENT>-to-<NEXT>` via MR at cutover, becomes the new fork main thereafter. | New, this merge |

### Prep commands

These assume a separate DataDog clone exists somewhere on disk (e.g., a clone of `https://github.com/DataDog/datadog-agent`). If you don't have one, clone it once — it's a large repo, treat it as a long-lived workspace.

**Why the release branch tip and not the tag:** DataDog's release tags often point at release-prep commits that are off the `<MAJOR.MINOR>.x` branch line of history (changelog generators, version bumpers, etc.). Using a tag commit as `base-<NEXT>` can push `git merge-base base-<CURRENT> base-<NEXT>` further back in history than necessary — sometimes to the previous DD minor version's branch point — yielding a less useful three-way merge base. The `<MAJOR.MINOR>.x` branch tip is on the "real" line of history and matches what the previous merge cycle did (verify by `git branch -r --contains <previous base-* tip>`).

```bash
# 1. In the DataDog clone: push the DD release-branch tip as base-<NEXT>.
#    Name the branch after the latest released patch version (e.g., base-7.78.2,
#    even when origin/7.78.x has moved a few backports past the 7.78.2 tag).
cd /path/to/datadog-agent
git fetch origin
git push <stackstate-gitlab-remote> origin/<MAJOR.MINOR>.x:refs/heads/base-<NEXT>

# 2. Back in the StackState fork: get the new base branch locally
cd /path/to/stackstate-agent
git fetch origin
git checkout base-<NEXT>

# 3. Compute and push the common ancestor
COMMON=$(git merge-base base-<CURRENT> base-<NEXT>)
git push origin "$COMMON":refs/heads/common-ancestor-<CURRENT>-<NEXT>
git fetch origin

# 4. Build the backport branch: STS changes replayed onto the common ancestor
git checkout -b backport-<CURRENT>-common-ancestor-<NEXT> common-ancestor-<CURRENT>-<NEXT>

#    Bring over every file StackState changed vs. base-<CURRENT>:
git diff --name-only base-<CURRENT>..stackstate-<CURRENT> > /tmp/sts-files.txt
git checkout stackstate-<CURRENT> -- $(cat /tmp/sts-files.txt)
git commit -m "All StackState changes replayed on top of common-ancestor-<CURRENT>-<NEXT>"
git push -u origin backport-<CURRENT>-common-ancestor-<NEXT>

# 5. Open the merge branch and do the actual merge
git checkout -b merged-<CURRENT>-to-<NEXT> base-<NEXT>
git merge backport-<CURRENT>-common-ancestor-<NEXT>
# resolve conflicts (this is the big sit-down), then commit
git push -u origin merged-<CURRENT>-to-<NEXT>

# 6. Create the eventual main branch EARLY off base-<NEXT>, so SUSE/customization
#    commits and parallel backports can land while the merge work is in progress.
#    The previous cycle did this as commit 4b19fc7b01 ("SUSE customization added
#    to DataDog 7.71.2 agent.") on top of base-7.71.2 — a single ~786-file commit
#    adding .assistance/, .ci-builders/, .cerberus/, etc. Bram (or whoever owns
#    SUSE infra) is the typical author; coordinate with them on timing. Some of
#    those files may already be on base-<NEXT> if DD has absorbed them upstream.
git push origin base-<NEXT>:refs/heads/stackstate-<NEXT>
# Then Bram opens an MR adding the SUSE customization commit (target stackstate-<NEXT>).
```

After step 5, the merge tree is in place and you move on to the workflow below. Fix-ups can be committed directly on `merged-<CURRENT>-to-<NEXT>`, or via feature branches if multiple people are working in parallel. Step 6 can be done in parallel with steps 5/post-5 — they're on independent branches.

### Optional: compare copy

Worth setting up at this point: a second clone at a sibling path checked out at the merge commit (the tip of `merged-<CURRENT>-to-<NEXT>` *before* any fix-up commits land). Useful for diffing "what have I changed since the merge point" without polluting the working tree. Naming convention: `<repo-path>-compare`.

## Typical Merge Workflow

Pre-merge branch setup (above) is a prerequisite. By the time you're here, `merged-<CURRENT>-to-<NEXT>` exists with the conflict-resolution merge committed.

1. Fix compilation errors on `merged-<CURRENT>-to-<NEXT>` (or on feature branches off it)
2. Update `fix_branding.sh` to handle new branding patterns
3. Iterate on CI until `branded_unit_tests` and `unbranded_unit_tests` pass on both x86 and ARM
4. **Verify all StackState-specific code blocks** (see "StackState-Specific Code That Can Be Lost During Merge" above)
5. **If this merge bumped the embedded Python** (`omnibus/config/software/python3.rb` `default_version`), bump and rebuild the `stackstate-agent-integrations` CI runner image + venv to match — see "Integrations repo: CI runner image (embedded Python bump)" above. Do this before integration testing so checks run on the interpreter the agent ships.
6. Run integration tests (beest) against the produced container images
7. **Deploy to sandbox and verify metric enrichment + log noise floor** (see "Sandbox soak validation" below)
8. Fix any runtime issues
9. **Open the cutover MR**: source `merged-<CURRENT>-to-<NEXT>` → target `stackstate-<NEXT>` (which already exists from prep step 6 with the SUSE customization commit on it). Title pattern from MR !387: "Our changes on top of <NEXT> into base". Once that merges, `stackstate-<NEXT>` becomes the new fork main and the four-repo coordination in "Cutover" below kicks in.

## Sandbox soak validation

The merge-branch image is validated against a real customer-shaped cluster (sandbox-main) before cutover. This catches three classes of issue that CI + beest don't surface:

- **Live workload noise**: helm-deployed agents on a busy cluster reveal log spam that idle integration tests miss (disk EACCES on per-pod CSI mounts, repeated cluster-checks dispatch loops, etc.).
- **Vestigial config**: per-cluster values.yaml overrides that have been silently broken for years (in the 7.78.2 cycle, sandbox-main's `aws_topology` override had been failing every collection since ~2023, hidden behind log severity).
- **Production receiver path validation**: sandbox-main forwards to a real receiver, so payload-rejection issues (compression, `internalHostname`, v1/v2 API gates) surface as 400/404 rather than being masked by mock receivers.

### Deploying the soak

Sandbox-main is ArgoCD-managed. **As of mid-June 2026 the soak vehicle is `github.com/StackVista/argocd-apps`** (private; SSH read works if you're in the StackVista GH org). This is a recent migration — predecessor was `github.com/StackVista/agent-promoter` (deprecated in commit `a5ccc994`), which itself was a successor to the original GitLab `stackvista/devops/agent-promoter` (archived). A separate GitLab `stackvista/devops/argocd-apps` exists but is **infra-only** — agent manifests are NOT there; don't be misled by the matching name.

To soak a merge-branch image:

1. **Wait for a successful merged-results pipeline on the agent MR.** GitLab's `merge_request_event` pipelines build images tagged with the MR-merge commit SHA (10 chars). Capture the SHA from the pipeline page or via the GitLab API (`/api/v4/projects/<id>/merge_requests/<iid>/pipelines`). Do NOT use the source-branch tip SHA — that image won't exist in the registry because no source-branch pipeline runs.
2. **Open a soak PR on `StackVista/argocd-apps`** that uncomments + sets `image.tag` in `cluster_definitions/sandbox-main/apps/suse-observability-agent/values.yaml` for `checksAgent`, `clusterAgent`, and `nodeAgent.containers.agent` (three places). Leave `processAgent` at chart default — separate release cadence and excluded from the soak-vehicle scope. Branch naming convention: `STAC-NNNNN-sandbox-pin-<sha10>`.
3. **Merge the soak PR.** ArgoCD picks up the change within its sync window (typically a few minutes); helm chart's checksum annotation on the ConfigMap triggers cluster-agent pod rotation; checks-agent and node-agents follow as their image tags resolve.
4. **Verify pod rotation:**
   ```bash
   kubectl get pods -n monitoring -l app.kubernetes.io/instance=suse-observability-dev-agent \
     -o jsonpath='{range .items[*]}{.metadata.name}{"\t"}{.metadata.creationTimestamp}{"\t"}{.spec.containers[0].image}{"\n"}{end}' \
     | grep -E 'cluster-agent|checks-agent|node-agent'
   ```
   Confirm all three components show the new tag and recent `creationTimestamp`.

**Self-healing pin reset (transient by design):** the argocd-apps repo runs `.github/workflows/drop-sandbox-agent-image-pins.yml` daily at 00:07 UTC. It `yq del()`s any developer-set `.image.tag` on the three soak-vehicle sites and restores the commented placeholders. Idempotent, GitHub-App-signed, commits straight to main. Consequences:

- A soak that runs past 00:07 UTC needs **daily re-pinning** OR coordination with whoever owns the workflow to pause it.
- Soak PRs no longer need an explicit revert commit when done — just let the next daily run clean up. (This is a change from the deprecated agent-promoter era, when soak transitions used a revert-then-re-pin commit pair for audit-trail hygiene.)
- Steady-state on `main` = no tag override. The chart's default agent image flows when no soak is active.

### Soak triage

After ~30 minutes of soak (enough for the disk check, kubernetes events, cluster-checks polling, and a process-agent flush cycle to have run multiple times), run the three triage one-liners:

```bash
# cluster-agent — deduplicated WARN/ERROR (excluding known event-mapper noise)
kubectl logs -n monitoring <cluster-agent-pod> \
  | grep -E ' (WARN|ERROR|FATAL) ' \
  | awk -F '|' '{print $3"|"$5}' | awk -F ' in ' '{print $1}' \
  | sed 's/^[[:space:]]*//' | sort | uniq -c | sort -rn | head -15

# checks-agent (same shape)
kubectl logs -n monitoring <checks-agent-pod> -c suse-observability-agent | <same pipeline>

# node-agent main container
kubectl logs -n monitoring <node-agent-pod> -c node-agent | <same pipeline>
```

**Expected baseline** with all log-noise patches applied (the four entries in "StackState-Specific Code That Can Be Lost During Merge" labelled cluster-agent server.go, disk default config, workloadmeta startup race, and diskv2 EACCES demotion):

| Pod | Expected non-startup WARN/ERROR | Acceptable |
|---|---|---|
| cluster-agent | 0 | Single-occurrence event-mapper "unknown reason" WARNs as Kubernetes emits novel event types are expected and fine. |
| checks-agent | 0 (silent after startup) | One startup-race line per pod boot is acceptable. |
| node-agent main | 0 | Anything from `pkg/collector/corechecks/system/disk/diskv2/disk.go:602` should be EACCES-demoted to DEBUG; if it shows up at WARN something regressed. |
| node-agent process-agent | 2 startup lines (CPU/memory threshold WARN — known disused feature) | Transient receiver 502s on rolling deploys are OK. |

If a noise pattern fires repeatedly during steady state, treat it as a regression and check the "StackState-Specific Code That Can Be Lost During Merge" entries first — most production log noise is a regression of one of those patches, not a new issue.

## Log noise triage methodology

Lessons accumulated across cycles, generalizing the discovery workflow above.

**Triage in four buckets, in order:**
- **A. Known noise** — entries that match documented patterns (per-pod CSI EACCES, event-mapper unknown reasons, startup workloadmeta race, etc.). Verify the relevant STS patch is in place; if not, restore it.
- **B. Vestigial config** — entries pointing at checks/endpoints/integrations that no longer exist. Often the loudest production noise comes from per-cluster `values.yaml` overrides that have been broken for years (the 7.78.2 cycle's `aws_topology` was this).
- **C. Cascade noise** — a single root failure that produces many derived WARNs (the WriteHeader spam was caused by failing aws_topology cluster-check dispatch, not a generic upstream bug).
- **D. Real regressions** — anything left after the first three buckets is the real signal.

**Correlated ERROR + WARN often share a root cause.** When you see a steady ERROR and a steady WARN at similar rates, fix the ERROR first and re-measure rather than chasing both independently. In the 7.78.2 cycle, removing the `aws_topology` ERROR silenced the `superfluous WriteHeader` WARN simultaneously, with no agent code change required — they were both symptoms of the same failing cluster-check dispatch.

**Fix at the right layer.** Choices in order of preference for log-noise patches:
1. **In-agent severity demotion** when the condition is benign-by-design (workloadmeta startup race, EACCES on locked-down mounts). Path- and topology-agnostic; works on every cluster.
2. **In-agent error filter on the stdlib path** when the noise is structural and the wrapper-level fix is incomplete (errorLog filter on cluster-agent's http.Server).
3. **Default-config rename** when the warning is a deprecation that's purely cosmetic (the `excluded_filesystems` → `file_system_exclude` rename).
4. **Helm chart exclusion regex** — *avoid* unless you're confident the same path/permission pattern holds across every customer cluster topology. Mount path regexes that work on EKS often break on RKE2/K3s/OpenShift; permission-based filters miss customers with looser security contexts.
5. **Sandbox values.yaml change** for vestigial overrides — the cheapest fix when the noise is purely a sandbox configuration mistake.

## Cutover: switching the fork's main branch

Cutover is a two-stage event:

1. **The cutover MR**: source `merged-<CURRENT>-to-<NEXT>` → target `stackstate-<NEXT>`. The target branch already exists (created in prep step 6 with the SUSE customization commit on it). Once the MR merges, `stackstate-<NEXT>` IS the post-merge state.
2. **Multi-repo coordination** (below) — flips helm-charts-internal and beest references over from `stackstate-<CURRENT>` to `stackstate-<NEXT>` and sets the new branch as GitLab default in this repo. As of the 7.78.2 cycle, neither argocd-apps nor agent-promoter requires a cutover-time change (see §2 — both deprecated as a cutover touchpoint).

Once the merge branch has clean CI, sandbox verification is healthy, and the team is ready to retire the previous main, **the four-repo changes need to land in lockstep**. Without coordination, the nightly promoter pipeline, beest CI gating, and the helm chart appVersion silently desynchronize and you end up debugging "why did my dev tag get clobbered overnight?" the morning after.

### 1. stackstate-agent (this repo)

- Set the new branch as the GitLab default branch (Settings → Repository → Branch defaults).
- Update protected branches: add the new branch, optionally remove the old one (or keep it for a grace period).
- The branch name pattern `stackstate-<DD-version>` is the convention; keep it.
- **`stackstate-deps.json` — flip `STACKSTATE_INTEGRATIONS_VERSION` from the transition branch to the released integrations tag.** During the merge cycle, this field holds the work-in-progress integrations branch (e.g., `transition-7.71.2-7.78.2`) so iterative changes flow through to agent builds without retagging. For the cutover-built release, it must hold a **TAG**, not a branch — tags pin reproducibly while branches drift. Tag naming convention from the integrations repo: `<DD-major>.<DD-minor>.<DD-patch>-<release>`, e.g., `7.78.2-1` (where `-1` is the integrations release-cut counter for that DD patch). Verify the tag exists in `stackstate-agent-integrations` and resolves to the merge commit of the transition-branch PR before flipping. **Easy mistake to make:** pinning to the long-lived release branch (e.g., `stackstate-7.78.2`) instead of the tag — superficially "works" because the next agent build pulls the right commits, but loses reproducibility (later commits on that branch drift the agent build silently).

### 2. argocd-apps (`git@github.com:StackVista/argocd-apps.git`)

**No cutover change required as of the 7.78.2 cycle (June 2026).** The repo is pure ArgoCD declarative config; ArgoCD pulls `version: latest` from helm-internal each sync, so when the helm-charts-internal MR (§3 below) merges and helm-internal publishes, dev clusters pick up the new chart automatically. No branch reference exists anywhere in the argocd-apps tree to flip.

**Multiple deprecations to be aware of:**
- The original GitLab `stackvista/devops/agent-promoter` (Python promoter + nightly + ArgoCD config all in one) — archived. Push fails with "You can't push code to an archived project."
- The GitHub `StackVista/agent-promoter` (ArgoCD-only after the Python promoter was deleted in PR #10, 2026-05-29) — **also deprecated** in commit `a5ccc994` (mid-June 2026); sandbox manifests migrated to argocd-apps.
- The GitLab `stackvista/devops/argocd-apps` — **infra-only**, no agent app. Don't be misled by the matching name; the agent manifests live in the GitHub argocd-apps repo, not this one.
- The GitHub `StackVista/argocd-apps` (private) — **canonical for sandbox / dev / prod cluster manifests.** SSH read works for org members; HTTPS shows 404.

The discovery sequence cost an afternoon during the 7.78.2 cycle — Copilot CLI session started on the old agent-promoter pattern, then tried the GitLab argocd-apps (wrong), finally landed on the GitHub one.

Optional during a cycle: `cluster_definitions/<cluster>/apps/suse-observability-agent/values.yaml` can be hand-edited to pin a specific agent image tag for sandbox soak — but **pins are transient by design**, auto-dropped daily at 00:07 UTC by `.github/workflows/drop-sandbox-agent-image-pins.yml`. See [[sandbox-soak-validation]] in Claude memory for the full workflow including the soak PR naming convention and self-healing model.

### 3. helm-charts-internal (`git@github.com:StackVista/helm-charts-internal.git`), chart `stable/suse-observability-agent`

**Three repos, one canonical:**
- `git@gitlab.com:stackvista/devops/helm-charts.git` — **archived** as of 2026. Push fails with "You can't push code to an archived project."
- `git@github.com:StackVista/helm-charts.git` — **public read-only mirror**. PRs filed here are in the wrong place; even if a maintainer merges one, the change doesn't reach production (the internal repo is the source of truth and mirrors out, not the reverse). Confusing because the URL pattern matches the agent-promoter GitHub-only migration; helm-charts went GitHub-too-but-private.
- `git@github.com:StackVista/helm-charts-internal.git` — **canonical, private**. This is where the cutover PR must land. Read access via SSH is fine if you're in the StackVista GitHub org; HTTPS shows 404 because the repo is private.

Cutover-MR mis-targeting cost an afternoon during the 7.78.2 cycle when STAC-25069 was opened on the public mirror; Bram flagged it on review. Always verify `git remote -v` shows `helm-charts-internal` before pushing.

- Bump `Chart.yaml`'s `appVersion`. **Convention:** `<STS-major>.<DD-minor>.<DD-patch>`. The StackState major (currently `3`) tracks DD's major-version family — DD v5/v6/v7 mapped to STS v1/v2/v3 historically. So DD `7.78.2` → STS appVersion `3.78.2`.
- Bump `Chart.yaml`'s `version` (the chart's own semver, separate from `appVersion`). On long-lived branches the live master may have already auto-bumped past your branch base — expect a rebase conflict on this line, take the live version + 1. The `verify_versions_bumped.sh` script gates MRs on this being strictly greater than the target branch's value.
- **Bump the 3 agent image tags in `values.yaml` AND the updatecli `agentHash` source branch ref.** The agent-tag automation in helm-charts post-PR-#10 (agent-promoter) is **updatecli**, not the deleted Python promoter. It lives at `updatecli/updatecli.d/update-docker-images/update.yaml` — the `agentHash` source curl-pings `https://api.github.com/repos/StackVista/stackstate-agent/commits/<BRANCH>` and writes `.sha[0:8]` into the chart's image tags via the matching targets. The cutover MR must update BOTH:
  - The 3 image tag sites in `values.yaml`:
    - `nodeAgent.containers.agent.image.tag`
    - `clusterAgent.image.tag`
    - `checksAgent.image.tag`
    All three carry the same agent tag — the new agent's latest built hash from the post-cutover `stackstate-<NEXT>` branch.
  - The `agentHash` source's branch ref in `updatecli/updatecli.d/update-docker-images/update.yaml`: change `commits/stackstate-<PREV>` → `commits/stackstate-<NEXT>`.

  Without flipping the updatecli source, every scheduled updatecli run (driven by `.github/workflows/updatecli.yml`) would resolve the (now stale) `stackstate-<PREV>` tip and silently rewrite `values.yaml` back to the old agent. The manual `values.yaml` bump in the cutover MR is needed because updatecli won't run until after the MR merges; the source-branch flip is what gets the automation pointing at the right reference going forward.

  **Leave alone** (separate release cadences): `nodeAgent.containers.processAgent.image.tag` and `logsAgent.image.tag` (promtail).
- Audit `templates/_container-agent.yaml` and `templates/checks-agent-deployment.yaml` for env vars deprecated by the new agent. Concrete example from the 7.71.2 cutover: removed `STS_PROCESS_AGENT_ENABLED` (the deprecated `process_config.enabled` key) — the replacement pair `STS_PROCESS_CONFIG_PROCESS_COLLECTION_ENABLED` + `STS_PROCESS_CONFIG_CONTAINER_COLLECTION_ENABLED` had been added alongside it earlier so the removal was a no-op deletion. Look for similar deprecation pairs introduced upstream during the merge.
- `nodes/stats` RBAC entry must be present in `templates/node-agent-clusterrole.yaml` (was missing pre-cutover; verify it's still there).
- Pre-commit hooks must run for every commit in this repo (helm-docs, shellcheck, helm-lint). Don't squash commits past hook runs.
- **helm-docs README regeneration gotcha**: the `helm-docs-built` pre-commit hook regenerates `README.md` from `Chart.yaml` + `values.yaml` and rejects the commit if the regenerated file isn't staged. Expect two commit attempts: first fails with README modified-but-unstaged, second succeeds after `git add README.md`.
- **Rebase-conflict gotcha** (long-lived branches): if live master has auto-bumped the `version:` or the dev image tag has rolled while your MR was in flight, expect 3-way conflicts on `Chart.yaml`, `values.yaml`, AND `README.md`. Resolution recipe: take live's `version:` (bump +1), take your `appVersion:`, take your image tags, take **either** side of `README.md` then re-run `pre-commit run helm-docs-built --files stable/<chart>/values.yaml stable/<chart>/Chart.yaml` and amend.

### 4. beest (`git@gitlab.com:stackvista/integrations/beest.git`)

The agent's main branch name is referenced in roughly 30 places, all needing the same find-and-replace:

- 5 CI rule files use `merge_train_always` rules pinned to the agent's main branch:
  - `.gitlab-ci-rancher-tests.yml`
  - `.gitlab-ci-suse-observability-cli-tests.yml`
  - `.gitlab-ci-agent-x86-tests.yml`
  - `.gitlab-ci-agent-arm-tests.yml`
  - `.gitlab-ci-suse-observability-ui-inspection.yml`
- `Makefile:21` last-resort `GIT_BRANCH ?= ... echo "<branch>"` fallback
- `helpers/resolve-agent-hashes.sh:48` `AGENT_DEFAULT_BRANCH` fallback (and the matching comment on line 47)
- `README.md:143` example value for the `AGENT_BRANCH_UNDER_TEST` env var
- `docs/setup-locally.md` references a non-existent `beest/` subfolder of the agent repo at the old branch — that link has been dead for a while; either fix or delete it as a separate cleanup.

### Sequencing

Order matters slightly. Recommended (as of the 7.78.2 cycle, post-PR-#10):

1. Flip the agent fork's default branch (§1) and the integrations fork's default branch — these are the prerequisites; everything else assumes the agent ref is live.
2. Wait for one post-cutover build on the new main to land in the registry — capture the agent image hash for §3.
3. Merge the beest CI sweep (§4) and the helm-charts cutover (§3) on roughly the same day. They're independent of each other; both reference the new agent ref.
4. **No agent-promoter merge required** — see §2. ArgoCD picks up the new chart from helm-internal automatically once §3 publishes.

Don't merge the helm-charts cutover before the new agent image hash exists in the registry — chart consumers would pull a tag that resolves nowhere.
