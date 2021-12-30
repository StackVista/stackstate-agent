package interpreters

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
)

func openTelemetryMappings (span *pb.Span, kind string, urn string, arn string, service string, serviceType string, spanType string) {
	_ = log.Warnf(fmt.Sprintf("Open Telemetry - Mapping Service '%s'", service))

	var mappingKey = "open.telemetry"

	// Kind
	span.Meta["span.kind"] = kind

	// Service
	span.Service = fmt.Sprintf("%s.%s", mappingKey, service)
	span.Meta["service"] = fmt.Sprintf("%s.%s", mappingKey, service)
	span.Meta["span.serviceName"] = fmt.Sprintf("%s.%s", mappingKey, service)
	span.Meta["span.serviceType"] = serviceType

	// Resource
	span.Resource = "aws.service"

	// Type
	span.Type = spanType

	// Identifiers
	span.Meta["span.serviceURN"] = urn
	span.Meta["sts.service.identifiers"] = arn
}

func OpenTelemetryConsumerMappings (span *pb.Span, urn string, arn string, service string, serviceType string, spanType string) {
	openTelemetryMappings(span, "consumer", urn, arn, service, serviceType, spanType)
}

func OpenTelemetryProducerMappings (span *pb.Span, urn string, arn string, service string, serviceType string, spanType string) {
	openTelemetryMappings(span, "producer", urn, arn, service, serviceType, spanType)
}

func UpdateOpenTelemetrySpanSource(source string, span *pb.Span) string {
	var openTelemetryPrefix = "openTelemetry"

	if source == api.OpenTelemetrySource {
		switch span.Meta["instrumentation_library"] {

		// Open Telemetry - Lambda Function Mapping (Core Function being used)
		// Reference: https://www.npmjs.com/package/@opentelemetry/instrumentation-aws-lambda
		case "@opentelemetry/instrumentation-aws-lambda":
			return fmt.Sprintf("%s%s", openTelemetryPrefix, "Lambda")

		// Open Telemetry - Http Requests Mapping
		// Reference: https://www.npmjs.com/package/@opentelemetry/instrumentation-http
		case "@opentelemetry/instrumentation-http":
			return fmt.Sprintf("%s%s", openTelemetryPrefix, "Http")

		// Open Telemetry - AWS SDK NodeJS Library Mappings
		// Reference: https://www.npmjs.com/package/@opentelemetry/instrumentation-aws-sdk
		case "@opentelemetry/instrumentation-aws-sdk":
			// We explicitly map certain services to know what we support as each service needs manual mapping
			switch span.Meta["aws.service.identifier"] {
				case "sqs":
					return fmt.Sprintf("%s%s", openTelemetryPrefix, "SQS")

				case "lambda":
					return fmt.Sprintf("%s%s", openTelemetryPrefix, "Lambda")

				case "sns":
					return fmt.Sprintf("%s%s", openTelemetryPrefix, "SNS")

				case "s3":
					return fmt.Sprintf("%s%s", openTelemetryPrefix, "S3")

				case "stepfunctions":
					return fmt.Sprintf("%s%s", openTelemetryPrefix, "StepFunctions")

				default:
					fmt.Printf("[WARNING] Unknown AWS identifier for Open Telemetry: %v\n", span.Meta["aws.service.identifier"])
			}
			break

		default:
			fmt.Printf("[WARNING] Unknown Open Telemetry instrumentation library: %v\n", span.Meta["instrumentation_library"])
		}
	}

	return source
}
