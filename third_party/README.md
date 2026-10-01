# Temporary owning-library YAML backports

These independently addressable modules retain the selected upstream APIs and
source commits. Each module's `PROVENANCE.md` records its source and removal
condition. They are temporarily owned under
https://github.com/StackVista/stackstate/issues/717.

The root module and workspace explicitly replace all four owners with these sources; the independent
network-device profile module also selects ordered-map. External Go consumers
must repeat the appropriate source selections because dependency `replace`
directives do not propagate. OPA retains the Datadog lightweight source rather
than switching to upstream OPA.

The workspace uses owner replacements rather than registering these as main
modules, so `go work sync` does not rewrite their independent dependency minima.
The modules retain their original logical paths, so they are outside the agent's
module release registry. Validate each with `GOWORK=off go mod tidy` and upstream
unit fixtures separately, then run the agent module consistency and consumer
checks. Upstream CI configuration and vendored snapshots are excluded; upstream
source, tests, fixtures, licenses and attribution are retained.
