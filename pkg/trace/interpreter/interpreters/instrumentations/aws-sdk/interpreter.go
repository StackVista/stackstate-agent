package awssdkinstrumentation

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/instrumentations/aws-sdk/modules"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

// InstrumentationIdentifier Identifier for this instrumentation library
// This identifier is based on the one used within the library using within the wild
// https://www.npmjs.com/package/@opentelemetry/instrumentation-aws-sdk
var InstrumentationIdentifier = "@opentelemetry/instrumentation-aws-sdk"

// InterpretBuilderForAwsSdkInstrumentation Mapping for modules used within the Instrumentation library
// Modules are basically sub parts that have certain functionality. Based on that functionality different context
// values needs to be mapped to be part of the span response
func InterpretBuilderForAwsSdkInstrumentation(span *pb.Span, source string) string {
	serviceIdentifier := span.Meta["aws.service.identifier"]

	// Mapping internal services that the instrumentation library uses to the correct modules inside the instrumentation
	// Example of these submodules would be two services that has unique context that needs to be mapped to the span
	// in a certain way, even the context data might be different per service
	switch serviceIdentifier {

	// AWS Service: Lambda
	// Description: This is a Lambda that is being invoked, Not the main Lambda doing the execution but rather the
	// Lambda that the main Lambda is attempting to invoke
	case modules.OpenTelemetryLambdaAwsIdentifier:
		lambdaInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, modules.OpenTelemetryLambdaServiceIdentifier)
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Mapped service '%s' to '%s'", serviceIdentifier, lambdaInterpreterIdentifier)
		return lambdaInterpreterIdentifier

	// AWS Service: SQS
	case modules.OpenTelemetrySQSAwsIdentifier:
		sqsInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, modules.OpenTelemetrySQSServiceIdentifier)
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Mapped service '%s' to '%s'", serviceIdentifier, sqsInterpreterIdentifier)
		return sqsInterpreterIdentifier

	// AWS Service: SNS
	case modules.OpenTelemetrySNSAwsIdentifier:
		snsInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, modules.OpenTelemetrySNSServiceIdentifier)
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Mapped service '%s' to '%s'", serviceIdentifier, snsInterpreterIdentifier)
		return snsInterpreterIdentifier

	// AWS Service: S3
	case modules.OpenTelemetryS3AwsIdentifier:
		s3InterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, modules.OpenTelemetryS3ServiceIdentifier)
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Mapped service '%s' to '%s'", serviceIdentifier, s3InterpreterIdentifier)
		return s3InterpreterIdentifier

	// AWS Service: SFN
	case modules.OpenTelemetrySFNAwsIdentifier:
		sfnInterpreterIdentifier := fmt.Sprintf("%s%s", api.OpenTelemetrySource, modules.OpenTelemetrySFNServiceIdentifier)
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Mapped service '%s' to '%s'", serviceIdentifier, sfnInterpreterIdentifier)
		return sfnInterpreterIdentifier

	default:
		log.Debugf("[OTEL] [INSTRUMENTATION-AWS-SDK] Unable to map the service: %s", serviceIdentifier)
	}

	return source
}
