# STAC-24773 — Omnibus → Bazel migration handoff

Handoff for humans and AI pair-programming on the StackState Agent fork.  
**Ticket:** STAC-24773 · **MR:** !426 · **Branch:** `STAC-24773-bazel-migration`

## Status (June 2026)

| Milestone | State |
|-----------|--------|
| Phase A (initial Bazel deps: libpcap, snmp-traps, jmxfetch, systemd, …) | Done (earlier commits on branch) |
| Phase B (openscap chain, curl/nghttp2) | Done |
| Phase C (orphan `.rb` sweep + integrations-chain deps) | Done (C1–C3, C5, C6, C-E) |
| Phase D1 (`python3.rb` → `@cpython`, pip 26, pip entrypoints) | **Done** — [pipeline 2628270529](https://gitlab.com/stackvista/agent/stackstate-agent/-/pipelines/2628270529) + Beest green |
| Phase D2 (drop python-build `.rb` orphans) | **Pushed** — delete `bzip2`, `liblzma`, `libsqlite3`, `libdb`, `libiconv` (replaced by Bazel `@bzip2`/`@xz`/`@sqlite3`) |
| Phase D3+ (`libffi`/`libtool`/`zlib` for arm integrations) | **Deferred** — `libffi.rb` still required by `stackstate-agent-integrations-py3.rb` on arm |
| Non-Bazel cleanup MR | **Not started** — separate ticket (not Bazel migration) |

Commit history on the branch may be **squashed**; use `git log --grep=STAC-24773` and file contents (grep `STAC-24773` in `omnibus/`) rather than assuming one commit per phase.

## Why this exists

Datadog Agent 7.78 moved most native dependencies from `omnibus/config/software/*.rb` to Bazel (`bazelisk run @dep//:install`, `//packages/agent/dependencies:install`). STS stayed hybrid after the 7.71.2 → 7.78.2 merge, maintaining dozens of extra Ruby recipes. STAC-24773 closes that gap in waves so each wave is CI-gated.

**Driver:** upstream alignment and lower merge cost — not housekeeping alone.

## Resume protocol (AI or human)

1. **Branch:** `git checkout STAC-24773-bazel-migration` (not `STAC-25035` CVE work unless explicitly switching tickets).
2. **Read:** this file + `omnibus/config/software/datadog-agent-dependencies.rb` (Bazel install blocks).
3. **Do not use:** `~/where-do-I-go-from-here.txt` (archived May 2026 merge handoff).
4. **Claude session name:** `bazel-migration-phase-b-c` (session `792c5837-…`) — optional; repo docs are canonical.
5. **Plan file (historical):** `~/.claude/plans/reactive-enchanting-hamming.md`.

## CI gating cadence

- Push one logical commit → wait for `build_deb` **x86 + arm** green → push next.
- **Beest** extra gate: after C2 (libxml2/libxslt) and before relying on integrations ODBC paths; not required between every phase.

## Required Bazel flag on STS runners

All `bazelisk run` invocations from omnibus must include:

```text
--downloader_config=/dev/null
```

Also pass `flavor_flag` (`--//packages/agent:flavor=fips` or heroku) where upstream does. Without `/dev/null`, fetches go through DD's internal mirror and time out on GitLab STS runners.

## Core patterns

### Shared libraries → `install` + `replace_prefix`

**Do not** ship Bazel-built `.so` files only via `//packages/agent/dependencies:install` / `pkg_filegroup all_files` if they need to live under `/opt/stackstate-agent/embedded/lib` with embedded RPATH. That failed omnibus health check (C2 pipeline 2619669055): `DT_NEEDED` pointed at system `/lib` for `libz` / `liblzma`.

**Do** mirror curl/nghttp2/libxml2:

```ruby
command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- @dep//:install --destdir='#{install_dir}'"
command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded' #{install_dir}/embedded/lib/lib….so"
```

Canonical templates: `git show origin/base-7.78.2:omnibus/config/software/datadog-agent-dependencies.rb` (Linux block).

### Pre-delete audit (Pitfall #12)

Before deleting `omnibus/config/software/foo.rb`, grep **both** quote styles:

```bash
rg "dependency ['\"]foo['\"]" omnibus/
```

Remove every consumer `dependency` line in the **same commit**, or omnibus parse fails.

### STS shebang branding

`deps/gstatus/BUILD.bazel` and `deps/nfsiostat/BUILD.bazel` hardcode `/opt/datadog-agent/embedded/bin/python`. After Bazel install, re-stamp in omnibus:

```ruby
command "sed -i '1s|.*|#!#{install_dir}/embedded/bin/python|' #{install_dir}/embedded/sbin/gstatus"
```

`@cpython` console scripts (`pip3.*`) have the same problem: Bazel bakes `//:install_dir` (`/opt/datadog-agent`) while `fix_branding.sh` sets omnibus `install_dir` to `/opt/stackstate-agent`. Re-stamp in `python3.rb` after pip upgrade; integrations invoke pip via `python -m pip` (upstream 7.78.2).

Local probe: `scripts/dev/test-python3-pip-entrypoint.sh` (run inside build container via `local.sh cmd`).

(Long-term: ABLD-302-style `config_flag` in Bazel; until then, omnibus `sed` is intentional.)

## Where Bazel installs live (post B+C)

| Deps | Omnibus hook |
|------|----------------|
| jmxfetch, libpcap, snmp-traps, systemd (top-level) | `datadog-agent-dependencies.rb` → `//packages/agent/dependencies:install` |
| nghttp2, curl, libxml2, libxslt | `datadog-agent-dependencies.rb` Linux `build do` |
| secret-generic-connector | `datadog-agent-dependencies.rb` → `//deps/secret_connector:install` |
| unixodbc, freetds, msodbcsql18 | `datadog-agent-integrations-py3-dependencies.rb` |
| gstatus, nfsiostat | same (plus shebang `sed`) |

`stackstate-agent-integrations-py3.rb` depends on `datadog-agent-integrations-py3-dependencies` (after `datadog-agent`).

## Phase map (what each wave dropped)

| Phase | Omnibus recipes removed / notes |
|-------|----------------------------------|
| B1 | openscap → Bazel; attr, libacl, libgcrypt, libgpg-error, libyaml, popt, rpm, xmlsec, … |
| B2 | curl.rb, nghttp2.rb |
| C1 | dbus, systemd, libselinux, libsepol, pcre2, util-linux, expat (+ iot-agent.rb systemd dep) |
| C2 | libxml2, libxslt only (libffi, unixodbc, nfsiostat deferred — see below) |
| C3 | file, elfutils, m4, libxcrypt (+ drop libxcrypt from python3.rb) |
| C5 | freetds, msodbcsql18, unixodbc, secret-generic-connector.rb |
| C6 | gstatus, nfsiostat, lua, sysstat |
| C-E | mac-app, cf-finalize, buildpack-finalize, cacerts_py{2,3}_local (+ agent-binaries buildpack dep) |
| D1 | `python3.rb` → Bazel `@bzip2`/`@xz`/`@sqlite3`/`@cpython`; drop `pip3.rb`; pip 26.0.1; `python -m pip` in integrations; re-stamp `pip3.*` shebangs |
| D2 | `bzip2.rb`, `liblzma.rb`, `libsqlite3.rb`, `libdb.rb`, `libiconv.rb` (orphaned after D1 Bazel installs) |

**Deferred from D2:** `libffi` + `libtool` (arm integrations `cffi`/`lxml` wheels), `zlib.rb` (still `openssl3` dep).

## Phase D (same ticket — STAC-24773)

**Goal:** Replace omnibus `python3.rb` source build with upstream pattern: `@bzip2`, `@xz`, `@sqlite3`, `@cpython//:install`, pip 26.0.1 bump; drop orphaned python-build recipes.

**D1 complete (June 2026):** `build_deb` [pipeline 2628270529](https://gitlab.com/stackvista/agent/stackstate-agent/-/pipelines/2628270529) + Beest green after pip entrypoint fixes (fix-ups 5–6).

**D2:** Delete omnibus recipes with zero `dependency` consumers, superseded by Bazel targets wired in `python3.rb`.

**Still deferred:** `libffi.rb`, `libtool.rb` — arm integrations recipe; `zlib.rb` — `openssl3.rb`.

## Parallel workstreams (do not confuse)

| Ticket | Branch / focus |
|--------|----------------|
| STAC-24773 | `STAC-24773-bazel-migration` — this doc |
| STAC-25035 | CVE remediation, `stackstate-deps.json` integrations pin `7.78.2-2`, prometheus/OTel |
| STAC-25069 | Cutover docs, beest/helm-charts branches |
| Sandbox soak | `github.com/StackVista/argocd-apps` — `cluster_definitions/sandbox-main/apps/suse-observability-agent/values.yaml`; pins auto-drop 00:07 UTC |

## Integrations coupling

- Integrations git: `github.com/StackVista/stackstate-agent-integrations`
- Pin: `STACKSTATE_INTEGRATIONS_VERSION` in `stackstate-deps.json` (e.g. tag `7.78.2-2`)
- Keep `python_version = "3.13"` in `stackstate-agent-integrations-py3.rb` in sync with `python3.rb` `default_version`
- Arm builds: `dependency 'libffi'` in integrations recipe until Bazel libffi is wired (Phase D3+)

## After upstream merge

1. Do **not** run the old “restore all `.rb` from `stackstate-<prev>`” script without review.
2. For each DD-deleted software def, check whether STAC-24773 already Bazel-installs it (grep `bazelisk` + this doc).
3. Prefer taking DD's Bazel wiring from `origin/base-7.78.2` over restoring pre-7.78 STS recipes.

## References

- [UPSTREAM_MERGE.md](../../UPSTREAM_MERGE.md) — merge workflow, sandbox soak, cutover
- [.claude/skills/omnibus-to-bazel/SKILL.md](../../.claude/skills/omnibus-to-bazel/SKILL.md) — single-dep migration procedure
- Claude memory: `omnibus-bazel-migration-gap.md` (short index, points here)

## MR !426 description (keep in sync with branch tip)

Copy into [MR !426](https://gitlab.com/stackvista/agent/stackstate-agent/-/merge_requests/426) when status changes.

```markdown
## STAC-24773 — Omnibus → Bazel migration (7.78.2)

Migrates StackState Agent native/python dependencies from omnibus Ruby recipes to Bazel (`bazelisk run @dep//:install`), aligned with Datadog Agent 7.78.2. All `bazelisk` invocations use `--downloader_config=/dev/null` for STS runner egress.

### Phases complete

| Phase | Summary |
|-------|---------|
| A | libpcap, snmp-traps, jmxfetch, systemd, … |
| B | openscap chain, curl, nghttp2 |
| C | Orphan `.rb` sweep + integrations-chain deps (C1–C3, C5, C6, C-E) |
| D1 | `python3.rb` → `@cpython`; pip 26.0.1; `python -m pip` in integrations; pip3 shebang re-stamp |
| D2 | Drop `bzip2`, `liblzma`, `libsqlite3`, `libdb`, `libiconv` (Bazel replaces them) |

### CI / validation

- `build_deb` x86 + arm: [pipeline 2628270529](https://gitlab.com/stackvista/agent/stackstate-agent/-/pipelines/2628270529) ✅
- Beest: ✅ (test job succeeded)

### Deferred (follow-up commits)

- `libffi.rb` / `libtool.rb` — arm integrations still depend on omnibus libffi
- `zlib.rb` — still required by `openssl3.rb`

### Test plan

- [x] `build_deb` green (x86 + arm)
- [x] Beest on branch tip
- [ ] Sandbox soak (optional before merge to `stackstate-7.78.2`)
- [ ] Reviewer smoke: metrics/topology/checks in StackState UI

### Docs

- Handoff: `docs/dev/stac-24773-bazel-migration.md`
- Local pip probe: `scripts/dev/test-python3-pip-entrypoint.sh`
```
