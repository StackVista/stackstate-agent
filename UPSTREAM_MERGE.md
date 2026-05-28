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

## Path Relocation (fix_package_paths.sh)

When `RELOCATED=true`, the source is moved from the Datadog import path to the StackState path:
- `github.com/DataDog/datadog-agent` → `github.com/StackVista/stackstate-agent`

This involves rewriting Go import paths, cleaning the module cache, removing `go.sum` and `vendor`, then re-syncing `go work` and re-vendoring.

## StackState-Specific Code That Can Be Lost During Merge

Upstream Datadog merges can silently drop StackState-specific code blocks (usually marked with `// sts begin` / `// sts end` or `// [sts]` comments). These are modifications to upstream files that don't exist in Datadog's codebase. **After every merge, verify these are still present:**

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
5. Run integration tests (beest) against the produced container images
6. **Deploy to sandbox and verify metric enrichment** (cluster_name, _k8s_cluster_, _scope_ labels present)
7. Fix any runtime issues
8. **Open the cutover MR**: source `merged-<CURRENT>-to-<NEXT>` → target `stackstate-<NEXT>` (which already exists from prep step 6 with the SUSE customization commit on it). Title pattern from MR !387: "Our changes on top of <NEXT> into base". Once that merges, `stackstate-<NEXT>` becomes the new fork main and the four-repo coordination in "Cutover" below kicks in.

## Cutover: switching the fork's main branch

Cutover is a two-stage event:

1. **The cutover MR**: source `merged-<CURRENT>-to-<NEXT>` → target `stackstate-<NEXT>`. The target branch already exists (created in prep step 6 with the SUSE customization commit on it). Once the MR merges, `stackstate-<NEXT>` IS the post-merge state.
2. **Four-repo coordination** (below) — flips agent-promoter, helm-charts, and beest references over from `stackstate-<CURRENT>` to `stackstate-<NEXT>` and sets the new branch as GitLab default in this repo.

Once the merge branch has clean CI, sandbox verification is healthy, and the team is ready to retire the previous main, **the four-repo changes need to land in lockstep**. Without coordination, the nightly promoter pipeline, beest CI gating, and the helm chart appVersion silently desynchronize and you end up debugging "why did my dev tag get clobbered overnight?" the morning after.

### 1. stackstate-agent (this repo)

- Set the new branch as the GitLab default branch (Settings → Repository → Branch defaults).
- Update protected branches: add the new branch, optionally remove the old one (or keep it for a grace period).
- The branch name pattern `stackstate-<DD-version>` is the convention; keep it.

### 2. agent-promoter (`stackvista/devops/agent-promoter`)

- `main.py:103` hard-codes the agent's main branch: `AgentOps("stackvista/agent/stackstate-agent", "stackstate-agent", "stackstate-7.51.1")`. Update the third argument to the new branch.
- The nightly `promote_agent_master_to_promoter_dev` pipeline reads this and rewrites `config.yml` and `deploy/argocd/common/apps/dev-agent/values.yaml` with the latest commit on that branch. **If you skip this step, every overnight run will continue setting dev tags from the old branch, clobbering whatever dev verification you're trying to do on the new one.**
- `.github/copilot-instructions.md` (if tracked in your local copy) also references the branch name; check and update.

### 3. helm-charts (`stackvista/devops/helm-charts`), chart `stable/suse-observability-agent`

- Bump `Chart.yaml`'s `appVersion`. **Convention:** `<STS-major>.<DD-minor>.<DD-patch>`. The StackState major (currently `3`) tracks DD's major-version family — DD v5/v6/v7 mapped to STS v1/v2/v3 historically. So DD `7.71.2` → STS appVersion `3.71.2`. The chart `version` (separate from `appVersion`) follows its own SemVer cadence and is bumped by `verify_versions_bumped.sh` rules whenever any file under `stable/<chart>/` changes.
- Audit `templates/_container-agent.yaml` and `templates/checks-agent-deployment.yaml` for env vars deprecated by the new agent. Concrete example from the 7.71.2 cutover: removed `STS_PROCESS_AGENT_ENABLED` (the deprecated `process_config.enabled` key) — the replacement pair `STS_PROCESS_CONFIG_PROCESS_COLLECTION_ENABLED` + `STS_PROCESS_CONFIG_CONTAINER_COLLECTION_ENABLED` had been added alongside it earlier so the removal was a no-op deletion. Look for similar deprecation pairs introduced upstream during the merge.
- `nodes/stats` RBAC entry must be present in `templates/node-agent-clusterrole.yaml` (was missing pre-cutover; verify it's still there).
- Image tags in `values.yaml` are managed by the agent-promoter nightly — leave them alone in the cutover MR; once #2 above is merged, the next nightly will write a tag from the new branch.
- Pre-commit hooks must run for every commit in this repo (helm-docs, shellcheck, helm-lint). Don't squash commits past hook runs.

### 4. beest (`stackvista/integrations/beest`)

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

Order matters slightly. Recommended:

1. Merge the agent default-branch change AND the beest CI change on roughly the same day.
2. Merge the agent-promoter change next — its nightly run that night will start writing tags from the new branch.
3. Merge the helm-charts appVersion bump whenever convenient (independent).

Don't merge the agent-promoter change *before* the agent default branch flips, or the next nightly will fail to find commits to promote.
