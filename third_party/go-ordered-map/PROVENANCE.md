# Temporary maintained-YAML backport

Source: https://github.com/wk8/go-ordered-map
Exact source commit: `85ca4a2b29d3241fa4513f82be3d38fe2392a791` (v2.1.8).
Logical module identity: `github.com/wk8/go-ordered-map/v2`.

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

Long upstream fixture/document paths are stored byte-for-byte in
`LONG_PATHS.tar.gz` with `LONG_PATHS.sha256` verification, preserving the agent's
Windows checkout path limit without weakening its filename gate. Before running
all upstream fixtures, restore them from this module directory with:

```sh
tar -xzf LONG_PATHS.tar.gz
sha256sum -c LONG_PATHS.sha256
```

These are fixture/docs assets, not Go source or embedded production resources.
