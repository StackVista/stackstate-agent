module example.com/testmodule

go 1.26.6

require (
	github.com/DataDog/datadog-agent/comp/api/api/def v0.0.0
	github.com/DataDog/datadog-agent/comp/core/agenttelemetry/def v0.0.0
	github.com/DataDog/datadog-agent/comp/core/agenttelemetry/fx v0.0.0
	github.com/DataDog/datadog-agent/comp/core/agenttelemetry/impl v0.0.0
	github.com/DataDog/datadog-agent/comp/core/config v0.0.0
	github.com/DataDog/datadog-agent/comp/core/configsync v0.0.0
	github.com/DataDog/datadog-agent/comp/core/delegatedauth v0.0.0
	github.com/DataDog/datadog-agent/comp/core/delegatedauth/api/cloudauth/aws v0.0.0
	github.com/DataDog/datadog-agent/comp/core/flare/builder v0.0.0
	github.com/DataDog/datadog-agent/comp/core/flare/types v0.0.0
	github.com/DataDog/datadog-agent/comp/core/hostname/hostnameinterface v0.0.0
	github.com/DataDog/datadog-agent/comp/core/ipc/def v0.0.0
	github.com/DataDog/datadog-agent/comp/core/ipc/httphelpers v0.0.0
	github.com/DataDog/datadog-agent/comp/core/ipc/impl v0.0.0
	github.com/DataDog/datadog-agent/comp/core/ipc/mock v0.0.0
	github.com/DataDog/datadog-agent/comp/core/log/def v0.0.0
	github.com/DataDog/datadog-agent/comp/core/log/fx v0.0.0
	github.com/DataDog/datadog-agent/comp/core/log/impl v0.0.0
	github.com/DataDog/datadog-agent/comp/core/log/impl-trace v0.0.0
	github.com/DataDog/datadog-agent/comp/core/log/mock v0.0.0
	github.com/DataDog/datadog-agent/comp/core/secrets/def v0.0.0
	github.com/DataDog/datadog-agent/comp/core/secrets/fx v0.0.0
	github.com/DataDog/datadog-agent/comp/core/secrets/impl v0.0.0
	github.com/DataDog/datadog-agent/comp/core/secrets/mock v0.0.0
	github.com/DataDog/datadog-agent/comp/core/secrets/noop-impl v0.0.0
	github.com/DataDog/datadog-agent/comp/core/secrets/utils v0.0.0
	github.com/DataDog/datadog-agent/comp/core/status v0.0.0
	github.com/DataDog/datadog-agent/comp/core/status/statusimpl v0.0.0
	github.com/DataDog/datadog-agent/comp/core/tagger/def v0.0.0
	github.com/DataDog/datadog-agent/comp/core/tagger/fx-remote v0.0.0
	github.com/DataDog/datadog-agent/comp/core/tagger/generic_store v0.0.0
	github.com/DataDog/datadog-agent/comp/core/tagger/impl-remote v0.0.0
	github.com/DataDog/datadog-agent/comp/core/tagger/origindetection v0.0.0
	github.com/DataDog/datadog-agent/comp/core/tagger/subscriber v0.0.0
	github.com/DataDog/datadog-agent/comp/core/tagger/tags v0.0.0
	github.com/DataDog/datadog-agent/comp/core/tagger/telemetry v0.0.0
	github.com/DataDog/datadog-agent/comp/core/tagger/types v0.0.0
	github.com/DataDog/datadog-agent/comp/core/tagger/utils v0.0.0
	github.com/DataDog/datadog-agent/comp/core/telemetry v0.0.0
	github.com/DataDog/datadog-agent/comp/def v0.0.0
	github.com/DataDog/datadog-agent/comp/forwarder/defaultforwarder v0.0.0
	github.com/DataDog/datadog-agent/comp/forwarder/orchestrator/orchestratorinterface v0.0.0
	github.com/DataDog/datadog-agent/comp/logs-library v0.0.0
	github.com/DataDog/datadog-agent/comp/logs/agent/config v0.0.0
	github.com/DataDog/datadog-agent/comp/netflow/payload v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/collector-contrib/def v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/collector-contrib/impl v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/converter/def v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/converter/impl v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/ddflareextension/def v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/ddflareextension/impl v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/ddflareextension/types v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/ddprofilingextension/def v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/ddprofilingextension/impl v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/logsagentpipeline v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/logsagentpipeline/logsagentpipelineimpl v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/otlp/components/exporter/datadogexporter v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/otlp/components/exporter/logsagentexporter v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/otlp/components/exporter/serializerexporter v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/otlp/components/metricsclient v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/otlp/components/processor/infraattributesprocessor v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/otlp/testutil v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/status/def v0.0.0
	github.com/DataDog/datadog-agent/comp/otelcol/status/impl v0.0.0
	github.com/DataDog/datadog-agent/comp/serializer/logscompression v0.0.0
	github.com/DataDog/datadog-agent/comp/serializer/metricscompression v0.0.0
	github.com/DataDog/datadog-agent/comp/trace/agent/def v0.0.0
	github.com/DataDog/datadog-agent/comp/trace/compression/def v0.0.0
	github.com/DataDog/datadog-agent/comp/trace/compression/impl-gzip v0.0.0
	github.com/DataDog/datadog-agent/comp/trace/compression/impl-zstd v0.0.0
	github.com/DataDog/datadog-agent/internal/tools v0.0.0
	github.com/DataDog/datadog-agent/internal/tools/gotest-custom v0.0.0
	github.com/DataDog/datadog-agent/internal/tools/independent-lint v0.0.0
	github.com/DataDog/datadog-agent/internal/tools/modformatter v0.0.0
	github.com/DataDog/datadog-agent/internal/tools/modparser v0.0.0
	github.com/DataDog/datadog-agent/internal/tools/proto v0.0.0
	github.com/DataDog/datadog-agent/internal/tools/worksynchronizer v0.0.0
	github.com/DataDog/datadog-agent/pkg/aggregator/ckey v0.0.0
	github.com/DataDog/datadog-agent/pkg/api v0.0.0
	github.com/DataDog/datadog-agent/pkg/collector/check/defaults v0.0.0
	github.com/DataDog/datadog-agent/pkg/config/basic v0.0.0
	github.com/DataDog/datadog-agent/pkg/config/create v0.0.0
	github.com/DataDog/datadog-agent/pkg/config/env v0.0.0
	github.com/DataDog/datadog-agent/pkg/config/helper v0.0.0
	github.com/DataDog/datadog-agent/pkg/config/mock v0.0.0
	github.com/DataDog/datadog-agent/pkg/config/model v0.0.0
	github.com/DataDog/datadog-agent/pkg/config/nodetreemodel v0.0.0
	github.com/DataDog/datadog-agent/pkg/config/remote v0.0.0
	github.com/DataDog/datadog-agent/pkg/config/render_config v0.0.0
	github.com/DataDog/datadog-agent/pkg/config/setup v0.0.0
	github.com/DataDog/datadog-agent/pkg/config/structure v0.0.0
	github.com/DataDog/datadog-agent/pkg/config/teeconfig v0.0.0
	github.com/DataDog/datadog-agent/pkg/config/utils v0.0.0
	github.com/DataDog/datadog-agent/pkg/config/viperconfig v0.0.0
	github.com/DataDog/datadog-agent/pkg/errors v0.0.0
	github.com/DataDog/datadog-agent/pkg/fips v0.0.0
	github.com/DataDog/datadog-agent/pkg/fleet/installer v0.0.0
	github.com/DataDog/datadog-agent/pkg/gohai v0.0.0
	github.com/DataDog/datadog-agent/pkg/logs/client v0.0.0
	github.com/DataDog/datadog-agent/pkg/logs/diagnostic v0.0.0
	github.com/DataDog/datadog-agent/pkg/logs/message v0.0.0
	github.com/DataDog/datadog-agent/pkg/logs/metrics v0.0.0
	github.com/DataDog/datadog-agent/pkg/logs/pipeline v0.0.0
	github.com/DataDog/datadog-agent/pkg/logs/processor v0.0.0
	github.com/DataDog/datadog-agent/pkg/logs/sender v0.0.0
	github.com/DataDog/datadog-agent/pkg/logs/sources v0.0.0
	github.com/DataDog/datadog-agent/pkg/logs/status/statusinterface v0.0.0
	github.com/DataDog/datadog-agent/pkg/logs/status/utils v0.0.0
	github.com/DataDog/datadog-agent/pkg/logs/types v0.0.0
	github.com/DataDog/datadog-agent/pkg/logs/util/testutils v0.0.0
	github.com/DataDog/datadog-agent/pkg/metrics v0.0.0
	github.com/DataDog/datadog-agent/pkg/network/driver v0.0.0
	github.com/DataDog/datadog-agent/pkg/network/payload v0.0.0
	github.com/DataDog/datadog-agent/pkg/networkdevice/profile v0.0.0
	github.com/DataDog/datadog-agent/pkg/networkpath/payload v0.0.0
	github.com/DataDog/datadog-agent/pkg/obfuscate v0.0.0
	github.com/DataDog/datadog-agent/pkg/opentelemetry-mapping-go/inframetadata v0.0.0
	github.com/DataDog/datadog-agent/pkg/opentelemetry-mapping-go/inframetadata/gohai/internal/gohaitest v0.0.0
	github.com/DataDog/datadog-agent/pkg/opentelemetry-mapping-go/otlp/attributes v0.0.0
	github.com/DataDog/datadog-agent/pkg/opentelemetry-mapping-go/otlp/logs v0.0.0
	github.com/DataDog/datadog-agent/pkg/opentelemetry-mapping-go/otlp/metrics v0.0.0
	github.com/DataDog/datadog-agent/pkg/opentelemetry-mapping-go/otlp/rum v0.0.0
	github.com/DataDog/datadog-agent/pkg/orchestrator/model v0.0.0
	github.com/DataDog/datadog-agent/pkg/orchestrator/util v0.0.0
	github.com/DataDog/datadog-agent/pkg/process/util/api v0.0.0
	github.com/DataDog/datadog-agent/pkg/proto v0.0.0
	github.com/DataDog/datadog-agent/pkg/remoteconfig/state v0.0.0
	github.com/DataDog/datadog-agent/pkg/security/secl v0.0.0
	github.com/DataDog/datadog-agent/pkg/security/seclwin v0.0.0
	github.com/DataDog/datadog-agent/pkg/serializer v0.0.0
	github.com/DataDog/datadog-agent/pkg/ssi/testutils v0.0.0
	github.com/DataDog/datadog-agent/pkg/status/health v0.0.0
	github.com/DataDog/datadog-agent/pkg/tagger/types v0.0.0
	github.com/DataDog/datadog-agent/pkg/tagset v0.0.0
	github.com/DataDog/datadog-agent/pkg/telemetry v0.0.0
	github.com/DataDog/datadog-agent/pkg/template v0.0.0
	github.com/DataDog/datadog-agent/pkg/trace v0.0.0
	github.com/DataDog/datadog-agent/pkg/trace/log v0.0.0
	github.com/DataDog/datadog-agent/pkg/trace/otel v0.0.0
	github.com/DataDog/datadog-agent/pkg/trace/stats v0.0.0
	github.com/DataDog/datadog-agent/pkg/trace/traceutil v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/aws/creds v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/backoff v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/buf v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/cache v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/cgroups v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/common v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/compression v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/containers/image v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/defaultpaths v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/executable v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/filesystem v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/flavor v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/fxutil v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/grpc v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/hostinfo v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/hostname/validate v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/http v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/json v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/jsonquery v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver/common/namespace v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/log v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/log/setup v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/option v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/otel v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/pointer v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/prometheus v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/quantile v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/quantile/sketchtest v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/scrubber v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/sort v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/startstop v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/statstracker v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/system v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/testutil v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/utilizationtracker v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/uuid v0.0.0
	github.com/DataDog/datadog-agent/pkg/util/winutil v0.0.0
	github.com/DataDog/datadog-agent/pkg/version v0.0.0
	github.com/DataDog/datadog-agent/test/e2e-framework v0.0.0
	github.com/DataDog/datadog-agent/test/fakeintake v0.0.0
	github.com/DataDog/datadog-agent/test/new-e2e v0.0.0
	github.com/DataDog/datadog-agent/test/otel v0.0.0
	github.com/DataDog/datadog-agent/tools/build-ddot-byoc v0.0.0
	github.com/DataDog/datadog-agent/tools/retry_file_dump v0.0.0
)

replace github.com/DataDog/datadog-agent/comp/api/api/def => ../comp/api/api/def

replace github.com/DataDog/datadog-agent/comp/core/agenttelemetry/def => ../comp/core/agenttelemetry/def

replace github.com/DataDog/datadog-agent/comp/core/agenttelemetry/fx => ../comp/core/agenttelemetry/fx

replace github.com/DataDog/datadog-agent/comp/core/agenttelemetry/impl => ../comp/core/agenttelemetry/impl

replace github.com/DataDog/datadog-agent/comp/core/config => ../comp/core/config

replace github.com/DataDog/datadog-agent/comp/core/configsync => ../comp/core/configsync

replace github.com/DataDog/datadog-agent/comp/core/delegatedauth => ../comp/core/delegatedauth

replace github.com/DataDog/datadog-agent/comp/core/delegatedauth/api/cloudauth/aws => ../comp/core/delegatedauth/api/cloudauth/aws

replace github.com/DataDog/datadog-agent/comp/core/flare/builder => ../comp/core/flare/builder

replace github.com/DataDog/datadog-agent/comp/core/flare/types => ../comp/core/flare/types

replace github.com/DataDog/datadog-agent/comp/core/hostname/hostnameinterface => ../comp/core/hostname/hostnameinterface

replace github.com/DataDog/datadog-agent/comp/core/ipc/def => ../comp/core/ipc/def

replace github.com/DataDog/datadog-agent/comp/core/ipc/httphelpers => ../comp/core/ipc/httphelpers

replace github.com/DataDog/datadog-agent/comp/core/ipc/impl => ../comp/core/ipc/impl

replace github.com/DataDog/datadog-agent/comp/core/ipc/mock => ../comp/core/ipc/mock

replace github.com/DataDog/datadog-agent/comp/core/log/def => ../comp/core/log/def

replace github.com/DataDog/datadog-agent/comp/core/log/fx => ../comp/core/log/fx

replace github.com/DataDog/datadog-agent/comp/core/log/impl => ../comp/core/log/impl

replace github.com/DataDog/datadog-agent/comp/core/log/impl-trace => ../comp/core/log/impl-trace

replace github.com/DataDog/datadog-agent/comp/core/log/mock => ../comp/core/log/mock

replace github.com/DataDog/datadog-agent/comp/core/secrets/def => ../comp/core/secrets/def

replace github.com/DataDog/datadog-agent/comp/core/secrets/fx => ../comp/core/secrets/fx

replace github.com/DataDog/datadog-agent/comp/core/secrets/impl => ../comp/core/secrets/impl

replace github.com/DataDog/datadog-agent/comp/core/secrets/mock => ../comp/core/secrets/mock

replace github.com/DataDog/datadog-agent/comp/core/secrets/noop-impl => ../comp/core/secrets/noop-impl

replace github.com/DataDog/datadog-agent/comp/core/secrets/utils => ../comp/core/secrets/utils

replace github.com/DataDog/datadog-agent/comp/core/status => ../comp/core/status

replace github.com/DataDog/datadog-agent/comp/core/status/statusimpl => ../comp/core/status/statusimpl

replace github.com/DataDog/datadog-agent/comp/core/tagger/def => ../comp/core/tagger/def

replace github.com/DataDog/datadog-agent/comp/core/tagger/fx-remote => ../comp/core/tagger/fx-remote

replace github.com/DataDog/datadog-agent/comp/core/tagger/generic_store => ../comp/core/tagger/generic_store

replace github.com/DataDog/datadog-agent/comp/core/tagger/impl-remote => ../comp/core/tagger/impl-remote

replace github.com/DataDog/datadog-agent/comp/core/tagger/origindetection => ../comp/core/tagger/origindetection

replace github.com/DataDog/datadog-agent/comp/core/tagger/subscriber => ../comp/core/tagger/subscriber

replace github.com/DataDog/datadog-agent/comp/core/tagger/tags => ../comp/core/tagger/tags

replace github.com/DataDog/datadog-agent/comp/core/tagger/telemetry => ../comp/core/tagger/telemetry

replace github.com/DataDog/datadog-agent/comp/core/tagger/types => ../comp/core/tagger/types

replace github.com/DataDog/datadog-agent/comp/core/tagger/utils => ../comp/core/tagger/utils

replace github.com/DataDog/datadog-agent/comp/core/telemetry => ../comp/core/telemetry

replace github.com/DataDog/datadog-agent/comp/def => ../comp/def

replace github.com/DataDog/datadog-agent/comp/forwarder/defaultforwarder => ../comp/forwarder/defaultforwarder

replace github.com/DataDog/datadog-agent/comp/forwarder/orchestrator/orchestratorinterface => ../comp/forwarder/orchestrator/orchestratorinterface

replace github.com/DataDog/datadog-agent/comp/logs-library => ../comp/logs-library

replace github.com/DataDog/datadog-agent/comp/logs/agent/config => ../comp/logs/agent/config

replace github.com/DataDog/datadog-agent/comp/netflow/payload => ../comp/netflow/payload

replace github.com/DataDog/datadog-agent/comp/otelcol/collector-contrib/def => ../comp/otelcol/collector-contrib/def

replace github.com/DataDog/datadog-agent/comp/otelcol/collector-contrib/impl => ../comp/otelcol/collector-contrib/impl

replace github.com/DataDog/datadog-agent/comp/otelcol/converter/def => ../comp/otelcol/converter/def

replace github.com/DataDog/datadog-agent/comp/otelcol/converter/impl => ../comp/otelcol/converter/impl

replace github.com/DataDog/datadog-agent/comp/otelcol/ddflareextension/def => ../comp/otelcol/ddflareextension/def

replace github.com/DataDog/datadog-agent/comp/otelcol/ddflareextension/impl => ../comp/otelcol/ddflareextension/impl

replace github.com/DataDog/datadog-agent/comp/otelcol/ddflareextension/types => ../comp/otelcol/ddflareextension/types

replace github.com/DataDog/datadog-agent/comp/otelcol/ddprofilingextension/def => ../comp/otelcol/ddprofilingextension/def

replace github.com/DataDog/datadog-agent/comp/otelcol/ddprofilingextension/impl => ../comp/otelcol/ddprofilingextension/impl

replace github.com/DataDog/datadog-agent/comp/otelcol/logsagentpipeline => ../comp/otelcol/logsagentpipeline

replace github.com/DataDog/datadog-agent/comp/otelcol/logsagentpipeline/logsagentpipelineimpl => ../comp/otelcol/logsagentpipeline/logsagentpipelineimpl

replace github.com/DataDog/datadog-agent/comp/otelcol/otlp/components/exporter/datadogexporter => ../comp/otelcol/otlp/components/exporter/datadogexporter

replace github.com/DataDog/datadog-agent/comp/otelcol/otlp/components/exporter/logsagentexporter => ../comp/otelcol/otlp/components/exporter/logsagentexporter

replace github.com/DataDog/datadog-agent/comp/otelcol/otlp/components/exporter/serializerexporter => ../comp/otelcol/otlp/components/exporter/serializerexporter

replace github.com/DataDog/datadog-agent/comp/otelcol/otlp/components/metricsclient => ../comp/otelcol/otlp/components/metricsclient

replace github.com/DataDog/datadog-agent/comp/otelcol/otlp/components/processor/infraattributesprocessor => ../comp/otelcol/otlp/components/processor/infraattributesprocessor

replace github.com/DataDog/datadog-agent/comp/otelcol/otlp/testutil => ../comp/otelcol/otlp/testutil

replace github.com/DataDog/datadog-agent/comp/otelcol/status/def => ../comp/otelcol/status/def

replace github.com/DataDog/datadog-agent/comp/otelcol/status/impl => ../comp/otelcol/status/impl

replace github.com/DataDog/datadog-agent/comp/serializer/logscompression => ../comp/serializer/logscompression

replace github.com/DataDog/datadog-agent/comp/serializer/metricscompression => ../comp/serializer/metricscompression

replace github.com/DataDog/datadog-agent/comp/trace/agent/def => ../comp/trace/agent/def

replace github.com/DataDog/datadog-agent/comp/trace/compression/def => ../comp/trace/compression/def

replace github.com/DataDog/datadog-agent/comp/trace/compression/impl-gzip => ../comp/trace/compression/impl-gzip

replace github.com/DataDog/datadog-agent/comp/trace/compression/impl-zstd => ../comp/trace/compression/impl-zstd

replace github.com/DataDog/datadog-agent/internal/tools => ../internal/tools

replace github.com/DataDog/datadog-agent/internal/tools/gotest-custom => ../internal/tools/gotest-custom

replace github.com/DataDog/datadog-agent/internal/tools/independent-lint => ../internal/tools/independent-lint

replace github.com/DataDog/datadog-agent/internal/tools/modformatter => ../internal/tools/modformatter

replace github.com/DataDog/datadog-agent/internal/tools/modparser => ../internal/tools/modparser

replace github.com/DataDog/datadog-agent/internal/tools/proto => ../internal/tools/proto

replace github.com/DataDog/datadog-agent/internal/tools/worksynchronizer => ../internal/tools/worksynchronizer

replace github.com/DataDog/datadog-agent/pkg/aggregator/ckey => ../pkg/aggregator/ckey

replace github.com/DataDog/datadog-agent/pkg/api => ../pkg/api

replace github.com/DataDog/datadog-agent/pkg/collector/check/defaults => ../pkg/collector/check/defaults

replace github.com/DataDog/datadog-agent/pkg/config/basic => ../pkg/config/basic

replace github.com/DataDog/datadog-agent/pkg/config/create => ../pkg/config/create

replace github.com/DataDog/datadog-agent/pkg/config/env => ../pkg/config/env

replace github.com/DataDog/datadog-agent/pkg/config/helper => ../pkg/config/helper

replace github.com/DataDog/datadog-agent/pkg/config/mock => ../pkg/config/mock

replace github.com/DataDog/datadog-agent/pkg/config/model => ../pkg/config/model

replace github.com/DataDog/datadog-agent/pkg/config/nodetreemodel => ../pkg/config/nodetreemodel

replace github.com/DataDog/datadog-agent/pkg/config/remote => ../pkg/config/remote

replace github.com/DataDog/datadog-agent/pkg/config/render_config => ../pkg/config/render_config

replace github.com/DataDog/datadog-agent/pkg/config/setup => ../pkg/config/setup

replace github.com/DataDog/datadog-agent/pkg/config/structure => ../pkg/config/structure

replace github.com/DataDog/datadog-agent/pkg/config/teeconfig => ../pkg/config/teeconfig

replace github.com/DataDog/datadog-agent/pkg/config/utils => ../pkg/config/utils

replace github.com/DataDog/datadog-agent/pkg/config/viperconfig => ../pkg/config/viperconfig

replace github.com/DataDog/datadog-agent/pkg/errors => ../pkg/errors

replace github.com/DataDog/datadog-agent/pkg/fips => ../pkg/fips

replace github.com/DataDog/datadog-agent/pkg/fleet/installer => ../pkg/fleet/installer

replace github.com/DataDog/datadog-agent/pkg/gohai => ../pkg/gohai

replace github.com/DataDog/datadog-agent/pkg/logs/client => ../pkg/logs/client

replace github.com/DataDog/datadog-agent/pkg/logs/diagnostic => ../pkg/logs/diagnostic

replace github.com/DataDog/datadog-agent/pkg/logs/message => ../pkg/logs/message

replace github.com/DataDog/datadog-agent/pkg/logs/metrics => ../pkg/logs/metrics

replace github.com/DataDog/datadog-agent/pkg/logs/pipeline => ../pkg/logs/pipeline

replace github.com/DataDog/datadog-agent/pkg/logs/processor => ../pkg/logs/processor

replace github.com/DataDog/datadog-agent/pkg/logs/sender => ../pkg/logs/sender

replace github.com/DataDog/datadog-agent/pkg/logs/sources => ../pkg/logs/sources

replace github.com/DataDog/datadog-agent/pkg/logs/status/statusinterface => ../pkg/logs/status/statusinterface

replace github.com/DataDog/datadog-agent/pkg/logs/status/utils => ../pkg/logs/status/utils

replace github.com/DataDog/datadog-agent/pkg/logs/types => ../pkg/logs/types

replace github.com/DataDog/datadog-agent/pkg/logs/util/testutils => ../pkg/logs/util/testutils

replace github.com/DataDog/datadog-agent/pkg/metrics => ../pkg/metrics

replace github.com/DataDog/datadog-agent/pkg/network/driver => ../pkg/network/driver

replace github.com/DataDog/datadog-agent/pkg/network/payload => ../pkg/network/payload

replace github.com/DataDog/datadog-agent/pkg/networkdevice/profile => ../pkg/networkdevice/profile

replace github.com/DataDog/datadog-agent/pkg/networkpath/payload => ../pkg/networkpath/payload

replace github.com/DataDog/datadog-agent/pkg/obfuscate => ../pkg/obfuscate

replace github.com/DataDog/datadog-agent/pkg/opentelemetry-mapping-go/inframetadata => ../pkg/opentelemetry-mapping-go/inframetadata

replace github.com/DataDog/datadog-agent/pkg/opentelemetry-mapping-go/inframetadata/gohai/internal/gohaitest => ../pkg/opentelemetry-mapping-go/inframetadata/gohai/internal/gohaitest

replace github.com/DataDog/datadog-agent/pkg/opentelemetry-mapping-go/otlp/attributes => ../pkg/opentelemetry-mapping-go/otlp/attributes

replace github.com/DataDog/datadog-agent/pkg/opentelemetry-mapping-go/otlp/logs => ../pkg/opentelemetry-mapping-go/otlp/logs

replace github.com/DataDog/datadog-agent/pkg/opentelemetry-mapping-go/otlp/metrics => ../pkg/opentelemetry-mapping-go/otlp/metrics

replace github.com/DataDog/datadog-agent/pkg/opentelemetry-mapping-go/otlp/rum => ../pkg/opentelemetry-mapping-go/otlp/rum

replace github.com/DataDog/datadog-agent/pkg/orchestrator/model => ../pkg/orchestrator/model

replace github.com/DataDog/datadog-agent/pkg/orchestrator/util => ../pkg/orchestrator/util

replace github.com/DataDog/datadog-agent/pkg/process/util/api => ../pkg/process/util/api

replace github.com/DataDog/datadog-agent/pkg/proto => ../pkg/proto

replace github.com/DataDog/datadog-agent/pkg/remoteconfig/state => ../pkg/remoteconfig/state

replace github.com/DataDog/datadog-agent/pkg/security/secl => ../pkg/security/secl

replace github.com/DataDog/datadog-agent/pkg/security/seclwin => ../pkg/security/seclwin

replace github.com/DataDog/datadog-agent/pkg/serializer => ../pkg/serializer

replace github.com/DataDog/datadog-agent/pkg/ssi/testutils => ../pkg/ssi/testutils

replace github.com/DataDog/datadog-agent/pkg/status/health => ../pkg/status/health

replace github.com/DataDog/datadog-agent/pkg/tagger/types => ../pkg/tagger/types

replace github.com/DataDog/datadog-agent/pkg/tagset => ../pkg/tagset

replace github.com/DataDog/datadog-agent/pkg/telemetry => ../pkg/telemetry

replace github.com/DataDog/datadog-agent/pkg/template => ../pkg/template

replace github.com/DataDog/datadog-agent/pkg/trace => ../pkg/trace

replace github.com/DataDog/datadog-agent/pkg/trace/log => ../pkg/trace/log

replace github.com/DataDog/datadog-agent/pkg/trace/otel => ../pkg/trace/otel

replace github.com/DataDog/datadog-agent/pkg/trace/stats => ../pkg/trace/stats

replace github.com/DataDog/datadog-agent/pkg/trace/traceutil => ../pkg/trace/traceutil

replace github.com/DataDog/datadog-agent/pkg/util/aws/creds => ../pkg/util/aws/creds

replace github.com/DataDog/datadog-agent/pkg/util/backoff => ../pkg/util/backoff

replace github.com/DataDog/datadog-agent/pkg/util/buf => ../pkg/util/buf

replace github.com/DataDog/datadog-agent/pkg/util/cache => ../pkg/util/cache

replace github.com/DataDog/datadog-agent/pkg/util/cgroups => ../pkg/util/cgroups

replace github.com/DataDog/datadog-agent/pkg/util/common => ../pkg/util/common

replace github.com/DataDog/datadog-agent/pkg/util/compression => ../pkg/util/compression

replace github.com/DataDog/datadog-agent/pkg/util/containers/image => ../pkg/util/containers/image

replace github.com/DataDog/datadog-agent/pkg/util/defaultpaths => ../pkg/util/defaultpaths

replace github.com/DataDog/datadog-agent/pkg/util/executable => ../pkg/util/executable

replace github.com/DataDog/datadog-agent/pkg/util/filesystem => ../pkg/util/filesystem

replace github.com/DataDog/datadog-agent/pkg/util/flavor => ../pkg/util/flavor

replace github.com/DataDog/datadog-agent/pkg/util/fxutil => ../pkg/util/fxutil

replace github.com/DataDog/datadog-agent/pkg/util/grpc => ../pkg/util/grpc

replace github.com/DataDog/datadog-agent/pkg/util/hostinfo => ../pkg/util/hostinfo

replace github.com/DataDog/datadog-agent/pkg/util/hostname/validate => ../pkg/util/hostname/validate

replace github.com/DataDog/datadog-agent/pkg/util/http => ../pkg/util/http

replace github.com/DataDog/datadog-agent/pkg/util/json => ../pkg/util/json

replace github.com/DataDog/datadog-agent/pkg/util/jsonquery => ../pkg/util/jsonquery

replace github.com/DataDog/datadog-agent/pkg/util/kubernetes/apiserver/common/namespace => ../pkg/util/kubernetes/apiserver/common/namespace

replace github.com/DataDog/datadog-agent/pkg/util/log => ../pkg/util/log

replace github.com/DataDog/datadog-agent/pkg/util/log/setup => ../pkg/util/log/setup

replace github.com/DataDog/datadog-agent/pkg/util/option => ../pkg/util/option

replace github.com/DataDog/datadog-agent/pkg/util/otel => ../pkg/util/otel

replace github.com/DataDog/datadog-agent/pkg/util/pointer => ../pkg/util/pointer

replace github.com/DataDog/datadog-agent/pkg/util/prometheus => ../pkg/util/prometheus

replace github.com/DataDog/datadog-agent/pkg/util/quantile => ../pkg/util/quantile

replace github.com/DataDog/datadog-agent/pkg/util/quantile/sketchtest => ../pkg/util/quantile/sketchtest

replace github.com/DataDog/datadog-agent/pkg/util/scrubber => ../pkg/util/scrubber

replace github.com/DataDog/datadog-agent/pkg/util/sort => ../pkg/util/sort

replace github.com/DataDog/datadog-agent/pkg/util/startstop => ../pkg/util/startstop

replace github.com/DataDog/datadog-agent/pkg/util/statstracker => ../pkg/util/statstracker

replace github.com/DataDog/datadog-agent/pkg/util/system => ../pkg/util/system

replace github.com/DataDog/datadog-agent/pkg/util/testutil => ../pkg/util/testutil

replace github.com/DataDog/datadog-agent/pkg/util/utilizationtracker => ../pkg/util/utilizationtracker

replace github.com/DataDog/datadog-agent/pkg/util/uuid => ../pkg/util/uuid

replace github.com/DataDog/datadog-agent/pkg/util/winutil => ../pkg/util/winutil

replace github.com/DataDog/datadog-agent/pkg/version => ../pkg/version

replace github.com/DataDog/datadog-agent/test/e2e-framework => ../test/e2e-framework

replace github.com/DataDog/datadog-agent/test/fakeintake => ../test/fakeintake

replace github.com/DataDog/datadog-agent/test/new-e2e => ../test/new-e2e

replace github.com/DataDog/datadog-agent/test/otel => ../test/otel

replace github.com/DataDog/datadog-agent/tools/build-ddot-byoc => ../tools/build-ddot-byoc

replace github.com/DataDog/datadog-agent/tools/retry_file_dump => ../tools/retry_file_dump
