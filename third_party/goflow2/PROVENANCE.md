# Temporary maintained-YAML backport

Source: https://github.com/netsampler/goflow2
Exact source commit: `69a6eaf99e205ed3cd02d29f044a01659107dd02` (v1.3.3).
Logical module identity: `github.com/netsampler/goflow2`.

Upstream source, tests, fixtures, licenses and attribution are retained. Upstream
CI metadata and vendored dependency snapshots are excluded; dependency sources
continue to resolve through Go modules. The only code edits select the matching
maintained YAML v2/v3 API, including typed YAML node methods and tests. Module
metadata changes select those parsers and satisfy their Go minima.

Tracking and temporary ownership: https://github.com/StackVista/stackstate/issues/717
Removal condition: adopt a compatible owning-library release (or Datadog
lightweight OPA patch release) with maintained YAML and passing consumer tests,
then remove this source and its explicit root/workspace/consumer selections.
Dependency replacements do not propagate: external consumers must explicitly
select this independently addressable nested module as well.
