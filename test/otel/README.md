# OTel Integration Test Files

This folder contains test files used as fixtures for the tests in `.gitlab/integration_test/otel.yml`. This integration test calls the following invoke tasks: `check-otel-build`, `check-otel-module-versions`.

## `check-otel-build`
This test attempts to build the files in `test/otel` with the same build args used in the make command of opentelemetry-collector-contrib. If an incompatibility is found, such as a dependency that requires CGO, this test will fail.

`dependencies.go` is a skeleton Go file that imports all direct imports that are used in the upstream [datadogexporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/exporter/datadogexporter/go.mod) and [datadogconnector](https://github.com/open-telemetry/opentelemetry-collector-contrib/blob/main/connector/datadogconnector/go.mod).

## `check-otel-module-versions`
This test reads the modules in `modules.py` with the `used_by_otel` flag and verifies that all of their `go.mod` versions match the same version that opentelemetry-collector-contrib uses. If these versions are out of sync, the modules can no longer be imported in upstream and this test will fail.

## Exporter security regressions

Run `go test ./test/otel/security -count=1 -v` from the repository root.
These local tests cover verbose trace diagnostics omitting exporter endpoints
(CVE-2026-81870) and the log gRPC exporter honoring both generic and log-specific
OTLP environment CA/client certificate settings (CVE-2026-81871). The TLS test
starts a loopback collector requiring a private CA and mutual authentication;
it does not provision infrastructure or send telemetry to an external collector.
