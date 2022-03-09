package aws

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// InterpretBuilderForAwsSdkInstrumentation an Open Telemetry mapping specific for aws services
// It is separate to the one above as these services might grow to a massive list of items
func InterpretBuilderForAwsSdkInstrumentation(span *pb.Span, source string) string {
	serviceIdentifier := span.Meta["aws.service.identifier"]

	switch serviceIdentifier {
	case OpenTelemetrySQSAwsIdentifier:
		sqsInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, OpenTelemetrySQSServiceIdentifier)
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Mapped service '%s' to '%s'", serviceIdentifier, sqsInterpreterIdentifier)
		return sqsInterpreterIdentifier

	case OpenTelemetryLambdaEntryAwsIdentifier:
		lambdaInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, OpenTelemetryLambdaEntryServiceIdentifier)
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Mapped service '%s' to '%s'", serviceIdentifier, lambdaInterpreterIdentifier)
		return lambdaInterpreterIdentifier

	case OpenTelemetryLambdaAwsIdentifier:
		lambdaInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, OpenTelemetryLambdaServiceIdentifier)
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Mapped service '%s' to '%s'", serviceIdentifier, lambdaInterpreterIdentifier)
		return lambdaInterpreterIdentifier

	case OpenTelemetrySNSAwsIdentifier:
		snsInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, OpenTelemetrySNSServiceIdentifier)
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Mapped service '%s' to '%s'", serviceIdentifier, snsInterpreterIdentifier)
		return snsInterpreterIdentifier

	case OpenTelemetryS3AwsIdentifier:
		s3InterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, OpenTelemetryS3ServiceIdentifier)
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Mapped service '%s' to '%s'", serviceIdentifier, s3InterpreterIdentifier)
		return s3InterpreterIdentifier

	case OpenTelemetrySFNAwsIdentifier:
		sfnInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, OpenTelemetrySFNServiceIdentifier)
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Mapped service '%s' to '%s'", serviceIdentifier, sfnInterpreterIdentifier)
		return sfnInterpreterIdentifier

	default:
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Unable to map the service: %s", serviceIdentifier)
		return source
	}
}
