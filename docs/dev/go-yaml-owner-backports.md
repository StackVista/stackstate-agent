# Maintained YAML owner backports

Tracking: https://github.com/StackVista/stackstate/issues/717.

The agent and cluster-agent select temporary independently addressable source
modules through explicit root and workspace replacements. The independent
network-device profile module also selects ordered-map. Dependency replacements
do not propagate to external consumers: repeat the relevant owner selections
when consuming these sources outside this checkout.

| Logical owner | Retained exact source |
| --- | --- |
| `github.com/netsampler/goflow2` v1.3.3 | `69a6eaf99e205ed3cd02d29f044a01659107dd02` |
| `github.com/wk8/go-ordered-map/v2` v2.1.8 | `85ca4a2b29d3241fa4513f82be3d38fe2392a791` |
| `github.com/cloudfoundry-community/go-cfclient/v2` | `3d15366c582080a7cafa6aeb28d2f5fe9c3146d1` |
| `github.com/open-policy-agent/opa` v1.7.1 | Datadog lightweight source `d2e1e78e081663d4b11109032715b68eb9b9d17a` |

The sources retain their logical paths, licenses, tests and attribution. The
only Go source changes select matching maintained YAML v2/v3 imports. Ordered-map
retains coherent typed node marshal/unmarshal interfaces. Datadog OPA patches are
preserved. Compatible releases checked on October 1, 2026 still use legacy YAML
or predate the selected source; their metadata/evidence is recorded in the
handoff. Standalone test requirements use maintained Testify. OPA's documentation
compiler uses compatible Cobra v1.10.2, already selected by the agent.

Do not classify the owners from their old requested versions alone: inspect the
explicit replacements, source commits, actual parser imports and binary build
info. The owned parser lines are `go.yaml.in/yaml/v2` v2.4.4 and
`go.yaml.in/yaml/v3` v3.0.5. No blanket old-parser module replacement is used.

The modules are not workspace main modules; this prevents workspace sync from
rewriting their independent dependency minima. They are outside the agent module
release registry and require separate standalone checks. Restore and verify long
upstream fixture/document paths before upstream fixtures:

```sh
(cd third_party/go-ordered-map && tar -xzf LONG_PATHS.tar.gz && sha256sum -c LONG_PATHS.sha256)
(cd third_party/opa && tar -xzf LONG_PATHS.tar.gz && sha256sum -c LONG_PATHS.sha256)
for owner in goflow2 go-ordered-map go-cfclient; do
  (cd "third_party/$owner" && GOWORK=off go mod tidy && GOWORK=off go test ./...)
done
(cd third_party/opa && GOWORK=off go mod tidy && GOWORK=off go test ./v1/ast ./v1/util ./v1/loader ./internal/config ./internal/strvals)
```

The archives retain 124 assets byte-for-byte without weakening the agent's Windows
filename limit. No Go source or embedded production resource is archived. OPA
bundle fixtures pass with vet disabled; two formatting assertions in untouched
upstream bundle tests fail Go 1.26 vet and reproduce on the original source.

Consumer qualification includes NetFlow adapter/server, profiledefinition
contracts in workspace and standalone mode, Cloud Foundry/autodiscovery,
compliance, and Kubernetes event/configuration tests under their actual tags.
For an unbranded profiledefinition test set `AGENT_GITHUB_ORG=DataDog` and
`AGENT_REPO_NAME=datadog-agent`; the broader schema tests' repository-ID mismatch
reproduces on the previous candidate. Compliance fixture file-mode assertions
require a process-local umask that permits the requested fixture permissions.
Run the normal module consistency and native/CI build checks as well.

The original candidate and baseline run 36773106564 failed the cacerts checksum
check before package compilation: mutable `https://curl.se/ca/cacert.pem` returned
September 25 bytes against the older pinned checksum. The recipe now pins
`https://curl.se/ca/cacert-2026-09-25.pem`, independently verified against curl's
published `cacert-2026-09-25.pem.sha256` and CA-extract publication. SHA256:
`a41b5d356aea97a529fe27e0f7316d2f9d946d75927476cf9cf1b90637d00505`
(188900 bytes, 121 certificates). Target filename, installed trust paths, modes
and checksum enforcement are preserved. Fresh package/image CI qualification
remains necessary; repairing the source prerequisite does not establish adoption.
A local agent build needs systemd development headers or the documented systemd
exclusion. Inherited Windows API failures and the bundled-events timeout were
qualified separately in the original handoff.

Residual test/tooling graphs remain distinct from the selected Linux product
binaries: Windows E2E uses Pulumi/ESC and needs test-infra-definitions coordination;
standalone OPA CLI tooling still uses legacy v3 through Viper. Other inherited
commands must be qualified individually rather than inferred from agent graphs.

Remove each backport after a compatible owning release (or lightweight Datadog
OPA patch release) selects maintained YAML and passes the same consumer fixtures.
Approved release rebuilds still require both architectures, artifact provenance,
image/secret/VEX checks and downstream chart/GitOps adoption. A source PR or local
binary does not establish that adoption.

Bazel metadata follows `tasks/go.py:_bazel_tidy`: prune imports, sync/tidy
registered modules, reconcile imports, then infer BUILD rules. The initial
`bazel mod tidy` removes stale ghodss and old YAML imports and selects ordered-map.
The filtered Bazel workspace links the local owner sources, while Gazelle excludes
the retained third_party trees from root BUILD generation. Original owner identities
and explicit replacements remain unchanged. Package qualification must use fresh
CI, including the Omnibus `@openssl//:install` analysis that previously failed on
the stale module-extension import; a mirror warning alone is not that failure.
