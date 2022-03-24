package instrumentation_aws_sdk

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/instrumentations/instrumentation-aws-sdk/modules"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

var InterpreterInstrumentationIdentifier = "@opentelemetry/instrumentation-aws-sdk"

// ComposeInstrumentationIdentifier an Open Telemetry mapping specific for aws services
// It is separate to the one above as these services might grow to a massive list of items
func ComposeInstrumentationIdentifier(span *pb.Span, source string) string {
	serviceIdentifier := span.Meta["aws.service.identifier"]

	switch serviceIdentifier {
	case modules.OpenTelemetrySQSAwsIdentifier:
		sqsInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, modules.OpenTelemetrySQSServiceIdentifier)
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Mapped service '%s' to '%s'", serviceIdentifier, sqsInterpreterIdentifier)
		return sqsInterpreterIdentifier

	case modules.OpenTelemetryLambdaEntryAwsIdentifier:
		lambdaInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, modules.OpenTelemetryLambdaEntryServiceIdentifier)
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Mapped service '%s' to '%s'", serviceIdentifier, lambdaInterpreterIdentifier)
		return lambdaInterpreterIdentifier

	case modules.OpenTelemetryLambdaAwsIdentifier:
		lambdaInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, modules.OpenTelemetryLambdaServiceIdentifier)
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Mapped service '%s' to '%s'", serviceIdentifier, lambdaInterpreterIdentifier)
		return lambdaInterpreterIdentifier

	case modules.OpenTelemetrySNSAwsIdentifier:
		snsInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, modules.OpenTelemetrySNSServiceIdentifier)
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Mapped service '%s' to '%s'", serviceIdentifier, snsInterpreterIdentifier)
		return snsInterpreterIdentifier

	case modules.OpenTelemetryS3AwsIdentifier:
		s3InterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, modules.OpenTelemetryS3ServiceIdentifier)
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Mapped service '%s' to '%s'", serviceIdentifier, s3InterpreterIdentifier)
		return s3InterpreterIdentifier

	case modules.OpenTelemetrySFNAwsIdentifier:
		sfnInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, modules.OpenTelemetrySFNServiceIdentifier)
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Mapped service '%s' to '%s'", serviceIdentifier, sfnInterpreterIdentifier)
		return sfnInterpreterIdentifier

	default:
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Unable to map the service: %s", serviceIdentifier)
	}

	return source
}
