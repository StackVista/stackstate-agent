package opentelemetry

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/aws"
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
			return fmt.Sprintf("%s%s", api.OpenTelemetrySource, aws.OpenTelemetryLambdaServiceIdentifier)

		case "@opentelemetry/instrumentation-http":
			return fmt.Sprintf("%s%s", api.OpenTelemetrySource, aws.OpenTelemetryHTTPServiceIdentifier)

		case "@opentelemetry/instrumentation-aws-sdk":
			InterpretBuilderForAwsSdkInstrumentation(span, source)
			break

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
	log.Debugf("[OTEL] AWS-SDK instrumentation mapping service: %s", serviceIdentifier)

	switch serviceIdentifier {
	case aws.OpenTelemetrySQSAwsIdentifier:
		return fmt.Sprintf("%s%s", api.OpenTelemetrySource, aws.OpenTelemetrySQSServiceIdentifier)

	case aws.OpenTelemetryLambdaAwsIdentifier:
		return fmt.Sprintf("%s%s", api.OpenTelemetrySource, aws.OpenTelemetryLambdaServiceIdentifier)

	case aws.OpenTelemetrySNSAwsIdentifier:
		return fmt.Sprintf("%s%s", api.OpenTelemetrySource, aws.OpenTelemetrySNSServiceIdentifier)

	case aws.OpenTelemetryS3AwsIdentifier:
		return fmt.Sprintf("%s%s", api.OpenTelemetrySource, aws.OpenTelemetryS3ServiceIdentifier)

	case aws.OpenTelemetrySFNAwsIdentifier:
		return fmt.Sprintf("%s%s", api.OpenTelemetrySource, aws.OpenTelemetrySFNServiceIdentifier)

	default:
		log.Debugf("[OTEL] AWS-SDK instrumentation unable to map the service: %s", serviceIdentifier)
		return source
	}
}
