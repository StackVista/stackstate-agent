# Embedded CPython security patches

The agent packages CPython 3.13.15. `cpython.MODULE.bazel` applies the following
upstream fixes during the source build on every supported platform. The version
string stays 3.13.15; version-only vulnerability scanners can still report these
CVEs. These patches do not change VEX or exception decisions.

| CVE | Upstream source | Local adaptation |
| --- | --- | --- |
| CVE-2026-15806 | [3.13 commit a2773a34](https://github.com/python/cpython/commit/a2773a34183b7d94a243bb98fd658926cc5348ce) | None |
| CVE-2026-19672 | [3.13 commit c7979f3a](https://github.com/python/cpython/commit/c7979f3a819011a3222bd16e671264b1e34282cb) | None; test hunk has a line offset |
| CVE-2025-15367 | [commit b234a2b6](https://github.com/python/cpython/commit/b234a2b67539f787e191d2ef19a7cbdce32874e7) | Test import context accounts for 3.13's existing `ExtraAssertions` import; production fix and test logic unchanged |

Each patch retains upstream regression tests. With a patched CPython source build,
run `./python -m test test_urllib2 test_tarfile test_poplib`. The agent image CI
also runs `scripts/test_embedded_python_security.py` with the actual packaged
interpreter on AMD64 and ARM64. This checks the three security boundaries and
valid operations without external network access. Image vulnerability and secret
scans remain separate steps. Remove a backport only when an upstream release
includes it and the packaged-interpreter tests pass.

Remaining rows as of 2026-09-18:

- **CVE-2026-15310:** the [3.13 zipfile backport](https://github.com/python/cpython/pull/156738)
  remains open. The original fix also needed a [third-party decompressor
  compatibility correction](https://github.com/python/cpython/pull/157180).
  Reconsider when the 3.13 backport and applicable compatibility correction are
  accepted upstream; the scanner's 3.15.0rc2 lead is not a compatible 3.13 bump.
- **CVE-2026-17084:** retain the existing no-stable-3.13-fix assessment; the known
  scanner lead is 3.15.0rc2. Reconsider on a supported 3.13 fix/backport or changed
  upstream evidence in [issue 155292](https://github.com/python/cpython/issues/155292).
- **CVE-2026-87910:** retain the existing no-scanner-fix assessment. Reconsider
  when [issue 157265](https://github.com/python/cpython/issues/157265) provides a
  supported 3.13 correction or scanner data changes.

The latest upstream 3.13 tag checked was 3.13.15. No runtime upgrade to 3.15,
applicability decision, suppression, or exception renewal is part of this work.
PR511 and GO-2026-5932 remain outside this patch's scope.
