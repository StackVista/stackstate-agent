# Embedded CPython maintenance

Update embedded CPython to the 3.13.18 maintenance release.

Keep `scripts/test_embedded_python_security.py`: package CI runs it with the
installed interpreter on native AMD64 and ARM64. Run as a non-root user so
all maintained runtime checks execute.

Omnibus must invoke Bazel even at the same Python version so restored caches
cannot bypass changed source or patch inputs.
