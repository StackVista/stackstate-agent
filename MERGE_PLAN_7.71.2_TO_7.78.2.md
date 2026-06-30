# Upstream Merge Plan: DD 7.71.2 → 7.78.2

Working plan for the upstream merge from DataDog Agent 7.71.2 to 7.78.2. This is the concrete, version-specific companion to [UPSTREAM_MERGE.md](UPSTREAM_MERGE.md), which contains the durable process documentation. Read that first.

## Inputs

- **Current fork main:** `stackstate-7.71.2`
- **Target DataDog version:** `7.78.2`
- **DataDog clone (for fetching pristine upstream):** `/home/louis/go/src/github.com/DataDog/datadog-agent`, currently behind `origin/main` by ~1300 commits, has only the `origin → github.com:DataDog/datadog-agent` remote
- **StackState GitLab remote (to be added in the DataDog clone):** `stackstate → git@gitlab.com:stackvista/agent/stackstate-agent.git`

## Phase 0 — One-time setup

- [x] In the DataDog clone, fast-forwarded `main` and fetched tags (`c22c35f45a4`, ~1300 commits caught up).
- [x] Confirmed `origin/7.78.x` exists and inspected the tip — 8 commits past the `7.78.2` tag, includes 7.78.3-rc.1/rc.2 release.json bumps. Decision: take the branch tip (option C) to inherit the OTel CVE backport and other stability fixes; the RC release.json bumps are inert metadata.
- [x] Added `stackstate` remote in the DataDog clone (`git@gitlab.com:stackvista/agent/stackstate-agent.git`) and fetched.
- [x] Integrations repo on `stackstate-7.71.2`; will branch off when Python dep work is needed.

## Phase 1 — Pre-merge branch setup

Following the recipe in UPSTREAM_MERGE.md "Pre-merge: branch setup".

- [x] **Pushed `origin/7.78.x` tip as `base-7.78.2`** (commit `6e93cda2cb0`, 8 commits past the `7.78.2` tag — includes a CVE-2026-39883 OTel SDK backport, CWS retry queue fix, Windows code-signing cert thumbprint update, autoscaling burstable-mode backport, 7.78.3-rc.1/rc.2 release.json bumps, and an installer logging fix):
  ```bash
  cd /home/louis/go/src/github.com/DataDog/datadog-agent
  git push stackstate origin/7.78.x:refs/heads/base-7.78.2
  ```
- [x] **Fetched `base-7.78.2` into the StackState fork** and confirmed tip `6e93cda2cb`.
- [x] **Pushed `common-ancestor-7.71.2-7.78.2`** at commit `20c9c848e4` ("dyninst/irgen: include type information for interface implementors", 2025-09-05). Sits on `origin/main` and on every DD release branch from `7.71.x` through `7.78.x` — i.e., the natural fork point of `7.71.x` from main.
- [x] **Pushed `backport-7.71.2-common-ancestor-7.78.2`** at commit `eafccba808`. Replayed 1028 changes (648 M + 372 A + 6 D + 2 R) from `stackstate-7.71.2` onto the common ancestor. Mechanism:
  ```bash
  git checkout -b backport-7.71.2-common-ancestor-7.78.2 origin/common-ancestor-7.71.2-7.78.2
  git diff --name-status origin/base-7.71.2..origin/stackstate-7.71.2 \
    | awk -F'\t' '$1 ~ /^[AM]$/ {print $2} $1 ~ /^R/ {print $3}' > /tmp/sts-checkout.txt
  git diff --name-status origin/base-7.71.2..origin/stackstate-7.71.2 \
    | awk -F'\t' '$1=="D" {print $2} $1 ~ /^R/ {print $2}' > /tmp/sts-rm.txt
  xargs -a /tmp/sts-checkout.txt -d '\n' git checkout origin/stackstate-7.71.2 --
  xargs -a /tmp/sts-rm.txt -d '\n' git rm
  git commit -m "All StackState changes replayed on top of common-ancestor-7.71.2-7.78.2"
  git push -u origin backport-7.71.2-common-ancestor-7.78.2
  ```
  Invariant verified: `xargs -a /tmp/sts-checkout.txt -d '\n' git diff backport-... origin/stackstate-7.71.2 --` produces zero diff lines (all checked-out files are byte-identical to `stackstate-7.71.2`).
- [ ] **Open the merge branch and resolve conflicts.**
  - Local-only branch `merged-7.71.2-to-7.78.2` already exists, checked out at `base-7.78.2` tip; never pushed, no merge committed.
  - First merge attempt (`git merge --no-commit --no-ff`) was aborted after capturing conflict topology — see "Conflict topology" below.
  - Next session: re-run the merge, work through conflicts, commit, push.
  ```bash
  git checkout merged-7.71.2-to-7.78.2  # local branch should still exist
  git merge origin/backport-7.71.2-common-ancestor-7.78.2
  # ... long conflict-resolution session ...
  git push -u origin merged-7.71.2-to-7.78.2
  ```
- [ ] **(Optional) Stand up a compare copy** at `/home/louis/go/src/github.com/StackVista/stackstate-agent-compare` once the initial merge commit lands; check it out at that commit so you can diff fix-ups against the raw merge point.

### Conflict topology (captured 2026-05-06, before merge resolution)

First merge attempt produced **363 conflicts** plus ~211 auto-merges to spot-check:

| Status | Count | Meaning |
|---|---|---|
| UU | 340 | Content conflict — hunk-level merge needed |
| DU | 22 | DD deleted upstream, STS still has changes — per-file decision |
| UD | 1 | STS deleted, DD modified — keep STS deletion |
| A | 372 | New STS files, no conflict |
| M | 211 | Auto-merged successfully — needs spot-check in heavy-STS files |
| D | 5 | Clean deletes |
| R | 2 | Renames preserved |

**UU conflicts by top-level directory (heaviest first):**

| Directory | UU count | Notes |
|---|---|---|
| `pkg/collector/` | 99 | STS's check-handler / topology customization — bulk of careful merge work goes here |
| `comp/otelcol/` | 34 | DD restructured otelcol heavily between 7.71 and 7.78 |
| `comp/core/` | 32 | DD continued component-architecture refactor |
| `pkg/util/` | 28 | DD utility churn |
| `pkg/config/` | 17 | STS overrides + DD config refactor |
| `pkg/logs/` | 13 | |
| `omnibus/config/` | 13 | Build system |
| `pkg/opentelemetry-mapping-go/` | 11 | |
| `pkg/serverless/` | 7 | DD likely refactored serverless heavily — see DU list too |
| `pkg/security/` | 6 | |

**Modify/delete (DU) — 22 files, decisions to make:**

- `pkg/serverless/*` (5 files) — DD restructured serverless. Likely drop STS mods wholesale; STS doesn't surface serverless features.
- `omnibus/config/software/{file.rb,freetds.rb,libtool.rb,pip3.rb}` — DD dropped these build deps. Investigate whether STS still needs them; if not, drop.
- `comp/otelcol/otlp/components/.../go.{mod,sum}`, `pkg/trace/stats/oteltest/go.sum`, `test/integration/serverless/src/go.{mod,sum}` (5 files) — DD restructured Go modules; STS files need to follow the new layout.
- `comp/{core/remoteagentregistry,metadata/inventoryotel,metadata/softwareinventory,trace/config}/.../*_test.go` and `component_mock.go` (5 files) — DD removed/renamed components STS had test mods for. Drop the test files.
- `comp/otelcol/otlp/map_provider_serverless_test.go`, `omnibus/config/software/datadog-cf-finalize.rb` — single removals; case-by-case.

**Update/delete (UD) — 1 file:**

- `Dockerfiles/agent/s6-services/trace/run` — STS deleted s6 services entirely; keep the deletion.

**Conflict lists captured to** (volatile, may not survive reboot):
- `/tmp/merge-7.71.2-to-7.78.2-conflicts.txt` (363 entries)
- `/tmp/merge-7.71.2-to-7.78.2-modify-delete.txt` (22)
- `/tmp/merge-7.71.2-to-7.78.2-update-delete.txt` (1)
- `/tmp/merge-7.71.2-to-7.78.2-automerged.txt` (211)

If these are gone by next session, regenerate by checking out `merged-7.71.2-to-7.78.2`, running `git merge --no-commit --no-ff origin/backport-7.71.2-common-ancestor-7.78.2`, capturing with `git diff --name-only --diff-filter=U` and `git status --short | awk '$1=="DU"'` etc., then `git merge --abort`.

**Exit criteria for Phase 1:** `merged-7.71.2-to-7.78.2` pushed, all merge conflicts resolved (compilation may still be broken — that's Phase 2).

## Phase 2 — Compile, brand, and pass CI

- [ ] Get the agent compiling on `merged-7.71.2-to-7.78.2` (or on feature branches off it). Expect this to surface dependency renames, removed packages, signature changes.
- [ ] Update `fix_branding.sh` for any new `datadoghq.com` references introduced upstream — see UPSTREAM_MERGE.md "Branding" section for the patterns that need reverts vs. what gets caught automatically.
- [ ] Run `branded_unit_tests` and `unbranded_unit_tests` on x86 and ARM until green.

**Exit criteria for Phase 2:** all four CI test jobs (`branded_unit_tests` x86/ARM, `unbranded_unit_tests` x86/ARM) green.

## Phase 3 — Verify StackState-specific code is intact

Walk every item in UPSTREAM_MERGE.md "StackState-Specific Code That Can Be Lost During Merge". Expect at least some of these to need re-application; upstream merges silently drop `// sts begin/end` blocks when surrounding code shifts.

- [ ] `kube_cluster_name` tagger tag (`comp/core/tagger/collectors/workloadmeta_extract.go`)
- [ ] `DefaultCompressorKind` zstd → zlib (must be in `fix_branding.sh`, not `stackstate()`)
- [ ] StackState defaults in `stackstate()` (`pkg/config/setup/config.go`)
- [ ] Resources metadata provider disabled (`fx.Supply(resourcesimpl.Disabled())` in `cmd/agent/subcommands/run/command.go`)
- [ ] `use_v2_api.series` forced to `false`
- [ ] `EventContext` field in event payloads (`pkg/serializer/internal/metrics/events.go`)
- [ ] Connectivity checker disabled by default
- [ ] RTLoader C++ branding in `fix_branding.sh`
- [ ] `forwarder_max_concurrent_requests` defaults to `1`
- [ ] `pkg/logs/client/http/worker_pool_test.go` — `driveUntil` helper and `absDuration` utility still present, tests still call them
- [ ] Grep for `SendMetadata\|SendProcessesMetadata\|SendHostMetadata\|SendAgentchecksMetadata` to find any new metadata payload call sites added upstream that may need the same `internalHostname`-or-disable treatment

## Phase 4 — Integration testing

- [ ] If this merge bumped the embedded Python (`omnibus/config/software/python3.rb` `default_version`), bump the `stackstate-agent-integrations` CI runner image + venv to the same version and rebuild/publish it — see UPSTREAM_MERGE.md "Integrations repo: CI runner image (embedded Python bump)". Otherwise integration checks test on a different Python than the agent ships (cf. STAC-25137).
- [ ] Build container images via the agent CI pipeline.
- [ ] Trigger beest pipeline against the new images. Use the `AGENT_BRANCH_UNDER_TEST` mechanism (see beest README).
- [ ] Investigate any failures; loop back to Phase 2/3 fixes as needed.

**Exit criteria for Phase 4:** beest pipeline green.

## Phase 5 — Sandbox verification

- [ ] Deploy the new agent images to a sandbox cluster (via agent-promoter sandbox flow — see [memory: agent-promoter / sandbox-testing-ticket](../../.claude/projects/-home-louis-go-src-github-com-StackVista-stackstate-agent/memory/sandbox-testing-ticket.md)).
- [ ] Verify in StackState UI:
  - Container CPU/memory columns populated (confirms `kube_cluster_name` tag flowing).
  - Topology components/relations being created and updated.
  - No `DuplicateSnapshotItem` or `ComponentForRelationMissing` errors in receiver logs.
- [ ] Check receiver logs for `400` responses on `/intake/` (would indicate a metadata payload regression — Phase 3 item).

**Exit criteria for Phase 5:** sandbox cluster shows healthy topology + metric enrichment for ≥24h with no receiver-side errors traceable to the new agent.

## Phase 6 — Cutover

Following UPSTREAM_MERGE.md "Cutover: switching the fork's main branch" — four repos in coordinated lockstep.

- [ ] **stackstate-agent (this repo):**
  - [ ] Create branch `stackstate-7.78.2` from `merged-7.71.2-to-7.78.2`.
  - [ ] Set `stackstate-7.78.2` as GitLab default branch.
  - [ ] Add to protected branches (optionally remove `stackstate-7.71.2` after grace period).
- [ ] **agent-promoter (`stackvista/devops/agent-promoter`):**
  - [ ] Update third arg of `AgentOps(...)` at `main.py:103` from `stackstate-7.71.2` to `stackstate-7.78.2`.
  - [ ] Update `.github/copilot-instructions.md` if it references the branch name.
- [ ] **helm-charts (`stackvista/devops/helm-charts`), chart `stable/suse-observability-agent`:**
  - [ ] Bump `Chart.yaml` `appVersion` to `3.78.2` (convention: `<STS-major>.<DD-minor>.<DD-patch>`).
  - [ ] Audit `templates/_container-agent.yaml` and `templates/checks-agent-deployment.yaml` for env vars deprecated in DD 7.72-7.78.
  - [ ] Confirm `nodes/stats` RBAC entry still in `templates/node-agent-clusterrole.yaml`.
  - [ ] Leave image tags in `values.yaml` alone — the next agent-promoter nightly will rewrite them.
- [ ] **beest (`stackvista/integrations/beest`):** find-and-replace `stackstate-7.71.2` → `stackstate-7.78.2` in:
  - [ ] `.gitlab-ci-rancher-tests.yml`
  - [ ] `.gitlab-ci-suse-observability-cli-tests.yml`
  - [ ] `.gitlab-ci-agent-x86-tests.yml`
  - [ ] `.gitlab-ci-agent-arm-tests.yml`
  - [ ] `.gitlab-ci-suse-observability-ui-inspection.yml`
  - [ ] `Makefile:21` (`GIT_BRANCH ?= ...` fallback)
  - [ ] `helpers/resolve-agent-hashes.sh:48` (`AGENT_DEFAULT_BRANCH` fallback) and the comment on line 47
  - [ ] `README.md:143` (`AGENT_BRANCH_UNDER_TEST` example)

**Cutover sequencing:**

1. Day 1: Merge agent default-branch change AND beest CI changes.
2. Day 1 (after #1 confirmed): Merge agent-promoter change. The next overnight run will start writing tags from `stackstate-7.78.2`.
3. Whenever convenient: Merge helm-charts `appVersion` bump.

Do **not** merge the agent-promoter change before the agent default branch flips, or the next nightly will fail to find commits to promote.

## Phase 7 — Post-cutover

- [ ] After the first successful nightly run from `stackstate-7.78.2`, verify the new tag landed in `agent-promoter/config.yml` and in the helm-charts `values.yaml`.
- [ ] Update memory with anything surprising or non-obvious from this merge cycle that future merges should know about. Record in `~/.claude/projects/-home-louis-go-src-github-com-StackVista-stackstate-agent/memory/`.
- [ ] Update `UPSTREAM_MERGE.md` with any newly discovered patterns (new STS-specific code blocks, new branding edge cases, new deprecated env vars).
- [ ] If this plan file (`MERGE_PLAN_7.71.2_TO_7.78.2.md`) was committed, delete it — the durable lessons should now live in `UPSTREAM_MERGE.md` and memory.
