# CVE exceptions

One YAML file per `(image, CVE)`, under a directory named for the consuming image.
The schema is owned by [image-pipeline](https://github.com/StackVista/image-pipeline)
(`schemas/exception.schema.json`); `schema_version: '1'` is the only version the
evaluator accepts, and it rejects duplicate `(image, CVE)` pairs.

An entry records a *deferral with a deadline*, not an acceptance. `expires` is a
short review date by which someone re-verifies upstream and either renews it with a
fresh date or deletes the file because a fix shipped.

## What enforces what

`build-deb.yml` passes this tree to the `scan-image` action, which suppresses a
matching finding while the exception is current and turns it back into a live
finding once `expires` has passed or cannot be parsed. That scan runs in
`mode: inform`, so it reports but never fails, and the gating chart scan in
[cve-reporter](https://github.com/StackVista/cve-reporter) does not read this tree
at all.

`scripts/check_cve_exception_expiry.py` is therefore what makes the dates real: it
fails the `CI success (lint and unit tests)` check when an entry has expired or
carries an unparseable date, and warns for two weeks beforehand. Run it locally
with `python3 scripts/check_cve_exception_expiry.py`.

## When the check goes red

Re-verify the advisory upstream first — that is the whole point of the date. Then
either drop the file if a compatible fix now exists, or renew `expires` and say in
`statement` what you checked and when. Do not extend a date without re-checking,
and do not silence a finding here that a version bump could fix instead.
