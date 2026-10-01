# Temporary maintained-YAML backport

Source: https://github.com/DataDog/opa
Exact source commit: `d2e1e78e081663d4b11109032715b68eb9b9d17a` (v0.0.0-20251126100856-d2e1e78e0816).
Logical module identity: `github.com/open-policy-agent/opa`.

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
