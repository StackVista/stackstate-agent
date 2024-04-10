package instrumentations

import (
	awsLambdaInstrumentation "github.com/DataDog/datadog-agent/pkg/trace/interpreter/interpreters/instrumentations/aws-lambda"
	awsSdkInstrumentation "github.com/DataDog/datadog-agent/pkg/trace/interpreter/interpreters/instrumentations/aws-sdk"
	httpInstrumentation "github.com/DataDog/datadog-agent/pkg/trace/interpreter/interpreters/instrumentations/http"
	stackStateInstrumentation "github.com/DataDog/datadog-agent/pkg/trace/interpreter/interpreters/instrumentations/stackstate"
	"github.com/DataDog/datadog-agent/pkg/trace/pb"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

// InterpretBasedOnInstrumentationLibrary OpenTelemetry contains multiple instrumentation libraries in the wild
// To make sure we route those libraries to the correct interpreter for the instrumentation we attempt to get the
// instrumentation library passed from the protobuf decoded block. Based from that we can determine the route
func InterpretBasedOnInstrumentationLibrary(span *pb.Span, source string) string {
	instrumentationLibrary := span.Meta["instrumentation_library"]

	// Switches based on the top-level instrumentation library id
	switch instrumentationLibrary {
	// @opentelemetry/instrumentation-aws-lambda
	case awsLambdaInstrumentation.InstrumentationIdentifier:
		return awsLambdaInstrumentation.InterpretBuilderForAwsLambdaInstrumentation()

	// @opentelemetry/instrumentation-http
	case httpInstrumentation.InstrumentationIdentifier:
		return httpInstrumentation.InterpretBuilderForHTTPInstrumentation()

	// @opentelemetry/instrumentation-aws-sdk
	case awsSdkInstrumentation.InstrumentationIdentifier:
		return awsSdkInstrumentation.InterpretBuilderForAwsSdkInstrumentation(span, source)

	// @opentelemetry/instrumentation-stackstate
	case stackStateInstrumentation.InstrumentationIdentifier:
		return stackStateInstrumentation.InterpretBuilderForStackStateInstrumentation()

	default:
		log.Debugf("[OTEL] [INSTRUMENTATION] Unknown instrumentation library: %s.", instrumentationLibrary)
	}

	// If we do not find an instrumentation route then we return a blank value to drop the span value
	return source
}
