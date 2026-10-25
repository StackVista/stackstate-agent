# Embedded CPython security regressions

Python 3.13.18 includes the upstream temporary-directory cleanup race fix,
as well as the StringPrep Unicode-table and tarfile hard-link fixes previously
carried as local backports. Their patches remain in git history.

Keep `scripts/test_embedded_python_security.py`: package CI runs it with the
installed interpreter on native AMD64 and ARM64. It checks IDNA behavior,
both tarfile extraction filters and the packaged `build.safe_extractall`
wrapper, including unchanged outside-file contents and metadata. It also
checks temporary-directory cleanup during a symlink race. Run as a non-root
user to exercise the permission-error recovery path.

Omnibus must invoke Bazel even at the same Python version so restored caches
cannot bypass changed source or patch inputs. Existing reviewed VEX and
exception decisions remain separate from these runtime regressions.
