package opentelemetry

import (
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	instrumentationAwsLambda "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/instrumentations/instrumentation-aws-lambda"
	instrumentationAwsSdk "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/instrumentations/instrumentation-aws-sdk"
	instrumentationHttp "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/instrumentations/instrumentation-http"
	instrumentationStackState "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/instrumentations/instrumentation-stackstate"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// InterpretBasedOnInstrumentationLibrary Open Telemetry mappings per instrumentation library
// This allows us to tag certain resources in a way that we can interpret them
func InterpretBasedOnInstrumentationLibrary(span *pb.Span, source string) string {
	if source == api.OpenTelemetrySource {
		instrumentationLibrary := span.Meta["instrumentation_library"]

		switch instrumentationLibrary {
		case instrumentationAwsSdk.InterpreterInstrumentationIdentifier:
			return instrumentationAwsSdk.ComposeInstrumentationIdentifier(span, source)

		case instrumentationAwsLambda.InterpreterInstrumentationIdentifier:
			return instrumentationAwsLambda.ComposeInstrumentationIdentifier()

		case instrumentationHttp.InterpreterInstrumentationIdentifier:
			return instrumentationHttp.ComposeInstrumentationIdentifier()

		case instrumentationStackState.InterpreterInstrumentationIdentifier:
			return instrumentationStackState.ComposeInstrumentationIdentifier()

		default:
			log.Debugf("[OTEL] [INSTRUMENTATION] Unknown instrumentation library: %s", instrumentationLibrary)
		}
	}

	return source
}
