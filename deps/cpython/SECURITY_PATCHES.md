# Embedded CPython security patch

CVE-2026-17084 is fixed by the unchanged upstream Python 3.13 backport
[c28b121a](https://github.com/python/cpython/commit/c28b121a4f0b975937c8b5a1b4934bb361d84296).
It corrects StringPrep's Unicode tables used by the shipped IDNA call paths.
Python continues to report 3.13.15; image-specific `fixed` VEX must identify
artifacts containing this patch, never all Python 3.13.15 installations.

The patch includes upstream regression tests. Existing package CI also checks
IDNA behavior using each architecture's packaged interpreter. Remove the patch
when upgrading to a release containing the fix and keep the runtime regression.
Omnibus must invoke Bazel even at the same Python version so restored caches
cannot bypass changed patch inputs.

The other five backports from the original candidate remain in git history.
Four findings already have reviewed image-scoped VEX in
[StackVista/vexhub](https://github.com/StackVista/vexhub/blob/main/pkg/oci/stackstate-k8s-agent/scan.openvex.json).
Tar-link CVE-2026-87910 is handled by a separate image-scoped applicability
assessment; it is not declared fixed by this patch. Existing VEX and exception
decisions outside this change remain intact.
