package opentelemetry

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/aws"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/http"
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
			return fmt.Sprintf("%s%s", api.OpenTelemetrySource, aws.OpenTelemetryLambdaEntryServiceIdentifier)

		case "@opentelemetry/instrumentation-http":
			return fmt.Sprintf("%s%s", api.OpenTelemetrySource, http.OpenTelemetryHTTPServiceIdentifier)

		case "@opentelemetry/instrumentation-aws-sdk":
			return InterpretBuilderForAwsSdkInstrumentation(span, source)

		default:
			log.Debugf("[OTEL] Unknown instrumentation library: %s", instrumentationLibrary)
		}
	}

	return source
}

// InterpretBuilderForAwsSdkInstrumentation an Open Telemetry mapping specific for aws services
// It is separate to the one above as these services might grow to a massive list of items
func InterpretBuilderForAwsSdkInstrumentation(span *pb.Span, source string) string {
	serviceIdentifier := span.Meta["aws.service.identifier"]
	log.Debugf("[OTEL] Mapping service: %s", serviceIdentifier)

	switch serviceIdentifier {
	case aws.OpenTelemetrySQSAwsIdentifier:
		sqsInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, aws.OpenTelemetrySQSServiceIdentifier)
		log.Debugf("[OTEL] Mapped service '%s' to '%s'", serviceIdentifier, sqsInterpreterIdentifier)
		return sqsInterpreterIdentifier

	case aws.OpenTelemetryLambdaEntryAwsIdentifier:
		lambdaInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, aws.OpenTelemetryLambdaEntryServiceIdentifier)
		log.Debugf("[OTEL] Mapped service '%s' to '%s'", serviceIdentifier, lambdaInterpreterIdentifier)
		return lambdaInterpreterIdentifier

	case aws.OpenTelemetryLambdaAwsIdentifier:
		lambdaInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, aws.OpenTelemetryLambdaServiceIdentifier)
		log.Debugf("[OTEL] Mapped service '%s' to '%s'", serviceIdentifier, lambdaInterpreterIdentifier)
		return lambdaInterpreterIdentifier

	case aws.OpenTelemetrySNSAwsIdentifier:
		snsInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, aws.OpenTelemetrySNSServiceIdentifier)
		log.Debugf("[OTEL] Mapped service '%s' to '%s'", serviceIdentifier, snsInterpreterIdentifier)
		return snsInterpreterIdentifier

	case aws.OpenTelemetryS3AwsIdentifier:
		s3InterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, aws.OpenTelemetryS3ServiceIdentifier)
		log.Debugf("[OTEL] Mapped service '%s' to '%s'", serviceIdentifier, s3InterpreterIdentifier)
		return s3InterpreterIdentifier

	case aws.OpenTelemetrySFNAwsIdentifier:
		sfnInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, aws.OpenTelemetrySFNServiceIdentifier)
		log.Debugf("[OTEL] Mapped service '%s' to '%s'", serviceIdentifier, sfnInterpreterIdentifier)
		return sfnInterpreterIdentifier

	default:
		log.Debugf("[OTEL] Unable to map the service: %s", serviceIdentifier)
		return source
	}
}
