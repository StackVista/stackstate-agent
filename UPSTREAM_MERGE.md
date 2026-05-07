# Upstream Merge Guide

This document covers the periodic task of merging upstream Datadog Agent changes into the StackState Agent fork. This is not day-to-day work — see [CLAUDE.md](CLAUDE.md) for normal development workflows.

## Overview

The StackState Agent is a fork of the Datadog Agent. Periodically, upstream Datadog releases are merged into the fork to pick up new features, bug fixes, and dependency updates. This is a large, intensive task that touches most of the codebase.

**Fork structure:**
- Main branch: `stackstate-7.51.1` (named after the last synced DD version)
- Upstream merges happen on feature branches (e.g., `stac-22523-main-agent-docker`)
- A clean "compare copy" at a separate path can be useful for diffing post-merge fix-ups against the raw merge point

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

### Monitor identity: volatile metric labels
- **Not an agent code issue** — this is a stackpacks/platform concern, but triggered by agent version changes.
- **What:** The threshold monitor function (`urn:stackpack:common:monitor-function:threshold`) derives `healthStateId` from ALL metric label values. If the agent version adds, removes, or changes any label (e.g., `orch_cluster_id` appearing, `status` flip-flopping), the platform creates duplicate monitor instances for the same component.
- **Affected monitors:** Node Disk/Memory/PID Pressure, Node Readiness, Available Endpoints (fixed in stackpacks MR 1332 by adding `max by (...)` aggregation). Desired-replicas monitors (daemonset/deployment/replicaset/statefulset) are theoretically vulnerable but not currently affected.
- **After merge:** Check if new KSM metrics add labels that differ from the labels used in monitor `urnTemplate` fields. If so, the monitor queries in the kubernetes-v2 stackpack need `by (...)` aggregation to strip volatile labels.

## Typical Merge Workflow

1. Create a feature branch from the fork's main branch
2. Merge the target upstream DD tag (e.g., `7.71.2`)
3. Resolve merge conflicts
4. Fix compilation errors
5. Update `fix_branding.sh` to handle new branding patterns
6. Iterate on CI until `branded_unit_tests` and `unbranded_unit_tests` pass on both x86 and ARM
7. **Verify all StackState-specific code blocks** (see section above)
8. Run integration tests (beest) against the produced container images
9. **Deploy to sandbox and verify metric enrichment** (cluster_name, _k8s_cluster_, _scope_ labels present)
10. Fix any runtime issues
11. Merge to the fork's main branch (and rename it to reflect the new upstream version)
