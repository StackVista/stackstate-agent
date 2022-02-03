package open_telemetry

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/aws"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
)

func InterpretBasedOnInstrumentationLibrary(span *pb.Span, source string) string {
	if source == api.OpenTelemetrySource {
		switch span.Meta["instrumentation_library"] {
		case "@opentelemetry/instrumentation-aws-lambda":
			return fmt.Sprintf("%s%s", api.OpenTelemetrySource, aws.OpenTelemetryLambdaServiceIdentifier)

		case "@opentelemetry/instrumentation-http":
			return fmt.Sprintf("%s%s", api.OpenTelemetrySource, aws.OpenTelemetryHTTPServiceIdentifier)

		case "@opentelemetry/instrumentation-aws-sdk":
			InterpretBuilderForAwsSdkInstrumentation(span, source)
			break
		}
	}

	return source
}

func InterpretBuilderForAwsSdkInstrumentation(span *pb.Span, source string) string {
	switch span.Meta["aws.service.identifier"] {
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
	}

	return source
}
