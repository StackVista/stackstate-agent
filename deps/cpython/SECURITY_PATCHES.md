# Embedded CPython security patches

The agent packages CPython 3.13.15. `cpython.MODULE.bazel` applies the following
upstream fixes during the source build on every supported platform. The version
string stays 3.13.15; version-only vulnerability scanners can still report these
CVEs. These patches do not change VEX or exception decisions.

| CVE | Upstream source | Local adaptation |
| --- | --- | --- |
| CVE-2026-15806 | [3.13 commit a2773a34](https://github.com/python/cpython/commit/a2773a34183b7d94a243bb98fd658926cc5348ce) | None |
| CVE-2026-19672 | [3.13 commit c7979f3a](https://github.com/python/cpython/commit/c7979f3a819011a3222bd16e671264b1e34282cb) | None; test hunk has a line offset |
| CVE-2026-15310 | [3.13 PR156738](https://github.com/python/cpython/pull/156738), commits `7a2b7388233ea0c647168c5792dc31f3051fe177`, `d53a41e8cf0148af073c46abc36c2969adcb70c2`, `f3f5234facd2ba491b3bff9e0ac9bd54378314ce` | None; includes removal of unsupported Zstandard tests and the third-party decompressor compatibility correction |
| CVE-2026-87910 | [3.13 PR157308](https://github.com/python/cpython/pull/157308), commits `05f187c81a6e2d90d627d96f2ad0b6724f044aef`, `f93473e550cc2e75f3ebc5cc890ca1af318efe36` | None; includes the Windows regression-test correction |
| CVE-2026-17084 | [3.13 commit c28b121a](https://github.com/python/cpython/commit/c28b121a4f0b975937c8b5a1b4934bb361d84296) | None; preserves the 3.13-specific generated Unicode tables |
| CVE-2025-15367 | [commit b234a2b6](https://github.com/python/cpython/commit/b234a2b67539f787e191d2ef19a7cbdce32874e7) | Test import context accounts for 3.13's existing `ExtraAssertions` import; production fix and test logic unchanged |

Each patch retains upstream regression tests. With a patched CPython source build,
run `./python -m test test_urllib2 test_tarfile test_poplib test_codecs test_unicodedata test_stringprep test_zipfile`. The agent image CI
also runs `scripts/test_embedded_python_security.py` with the actual packaged
interpreter on AMD64 and ARM64. This checks the six security boundaries and
valid operations without external network access. Image vulnerability and secret
scans remain separate steps. Remove a backport only when an upstream release
includes it and the packaged-interpreter tests pass.

Upstream status checked 2026-09-18:

The 3.13 zipfile and tar-link backport PRs remain open upstream. Their main-branch
fixes are merged. The copied zipfile series includes the compatibility correction
from [GH-157180](https://github.com/python/cpython/pull/157180); the tar-link series
includes the Windows test correction from
[GH-157334](https://github.com/python/cpython/pull/157334). Their open status is
retained for independent review; they are not represented as released Python
3.13 fixes. The other 3.13 backports are merged; POP3 is adapted from its merged
main-branch fix. Runtime regression tests are mandatory for all six fixes.

The latest upstream 3.13 tag checked was 3.13.15. The 3.15.0rc2 scanner leads are
not used as a runtime upgrade. Version-only findings for these six CVEs may remain
until upstream release/scanner metadata or a separately approved applicability
decision accounts for the backports. No applicability decision, suppression, or
exception renewal is part of this work. PR511 and GO-2026-5932 remain outside scope.
