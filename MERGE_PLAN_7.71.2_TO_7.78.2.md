# Upstream Merge Plan: DD 7.71.2 → 7.78.2

Working plan for the upstream merge from DataDog Agent 7.71.2 to 7.78.2. This is the concrete, version-specific companion to [UPSTREAM_MERGE.md](UPSTREAM_MERGE.md), which contains the durable process documentation. Read that first.

## Inputs

- **Current fork main:** `stackstate-7.71.2`
- **Target DataDog version:** `7.78.2`
- **DataDog clone (for fetching pristine upstream):** `/home/louis/go/src/github.com/DataDog/datadog-agent`, currently behind `origin/main` by ~1300 commits, has only the `origin → github.com:DataDog/datadog-agent` remote
- **StackState GitLab remote (to be added in the DataDog clone):** `stackstate → git@gitlab.com:stackvista/agent/stackstate-agent.git`

## Phase 0 — One-time setup

- [ ] In the DataDog clone, fast-forward `main` and fetch tags:
  ```bash
  cd /home/louis/go/src/github.com/DataDog/datadog-agent
  git pull --ff-only
  git fetch origin --tags
  ```
- [ ] Confirm the `7.78.2` tag exists upstream (`git tag --list 7.78.2`). If not, the release hasn't been cut yet — pick the latest `7.78.x` available, or wait.
- [ ] Add the StackState GitLab repo as a remote in the DataDog clone:
  ```bash
  git remote add stackstate git@gitlab.com:stackvista/agent/stackstate-agent.git
  git fetch stackstate
  ```
- [ ] Confirm the integrations repo is on a feature branch ready for parallel work — this merge will need a corresponding integrations branch for Python dep bumps.

## Phase 1 — Pre-merge branch setup

Following the recipe in UPSTREAM_MERGE.md "Pre-merge: branch setup".

- [ ] **Push pristine DD 7.78.2 as `base-7.78.2`:**
  ```bash
  cd /home/louis/go/src/github.com/DataDog/datadog-agent
  git push stackstate 7.78.2:refs/heads/base-7.78.2
  ```
  (The `git fetch origin --tags` from Phase 0 already pulled the tag locally.)
- [ ] **Fetch the new branch into the StackState fork:**
  ```bash
  cd /home/louis/go/src/github.com/StackVista/stackstate-agent
  git fetch origin
  git checkout base-7.78.2
  ```
- [ ] **Compute and push the common ancestor:**
  ```bash
  COMMON=$(git merge-base base-7.71.2 base-7.78.2)
  echo "Common ancestor: $COMMON"
  git push origin "$COMMON":refs/heads/common-ancestor-7.71.2-7.78.2
  git fetch origin
  ```
- [ ] **Build the backport branch:**
  ```bash
  git checkout -b backport-7.71.2-common-ancestor-7.78.2 common-ancestor-7.71.2-7.78.2
  git diff --name-only base-7.71.2..stackstate-7.71.2 > /tmp/sts-files.txt
  wc -l /tmp/sts-files.txt   # sanity-check the count
  git checkout stackstate-7.71.2 -- $(cat /tmp/sts-files.txt)
  git commit -m "All StackState changes replayed on top of common-ancestor-7.71.2-7.78.2"
  git push -u origin backport-7.71.2-common-ancestor-7.78.2
  ```
- [ ] **Open the merge branch:**
  ```bash
  git checkout -b merged-7.71.2-to-7.78.2 base-7.78.2
  git merge backport-7.71.2-common-ancestor-7.78.2
  # resolve conflicts; commit
  git push -u origin merged-7.71.2-to-7.78.2
  ```
- [ ] **(Optional) Stand up a compare copy:** `git clone . ../stackstate-agent-compare` or similar; check it out at the merge commit before any fix-ups land.

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
