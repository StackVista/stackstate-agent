package opentelemetry

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/modules/aws"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/modules/http"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// InterpretBasedOnInstrumentationLibrary Open Telemetry mappings per instrumentation library
// This allows us to tag certain resources in a way that we can interpret them
func InterpretBasedOnInstrumentationLibrary(span *pb.Span, source string) string {
	if source == api.OpenTelemetrySource {
		instrumentationLibrary := span.Meta["instrumentation_library"]
		switch instrumentationLibrary {
		case "@opentelemetry/instrumentation-aws-lambda":
			lambdaInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, aws.OpenTelemetryLambdaEntryServiceIdentifier)
			log.Debugf("[OTEL] [INSTRUMENTATION-AWS-LAMBDA] Mapping service: %s", lambdaInterpreterIdentifier)
			return lambdaInterpreterIdentifier

		case "@opentelemetry/instrumentation-http":
			httpInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, http.OpenTelemetryHTTPServiceIdentifier)
			log.Debugf("[OTEL] [INSTRUMENTATION-HTTP] Mapping service: %s", httpInterpreterIdentifier)
			return httpInterpreterIdentifier

		case "@opentelemetry/instrumentation-aws-sdk":
			return aws.InterpretBuilderForAwsSdkInstrumentation(span, source)

		default:
			log.Debugf("[OTEL] [INSTRUMENTATION] Unknown instrumentation library: %s", instrumentationLibrary)
		}
	}

	return source
}
