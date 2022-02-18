package aws

import (
	"fmt"
	"github.com/StackVista/stackstate-agent/pkg/trace/api"
	config "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/config"
	interpreter "github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters"
	"github.com/StackVista/stackstate-agent/pkg/trace/interpreter/interpreters/open-telemetry/instrumentations"
	"github.com/StackVista/stackstate-agent/pkg/trace/pb"
	"github.com/StackVista/stackstate-agent/pkg/util/log"
	"strings"
)

// OpenTelemetrySQSInterpreter default span interpreter for this data structure
type OpenTelemetrySQSInterpreter struct {
	interpreter.Interpreter
}

// OpenTelemetrySQSServiceIdentifier The base identifier for this interpreter, Is also used in identifying AWS services
const OpenTelemetrySQSServiceIdentifier = "SQS"

// OpenTelemetrySQSInterpreterSpan An identifier used to direct Open Telemetry interprets to this Interpreter
var OpenTelemetrySQSInterpreterSpan = fmt.Sprintf("%s%s", api.OpenTelemetrySource, OpenTelemetrySQSServiceIdentifier)

// OpenTelemetrySQSAwsIdentifier An identifier used to map the AWS Service to the STS InterpreterServiceIdentifier
var OpenTelemetrySQSAwsIdentifier = strings.ToLower(OpenTelemetrySQSServiceIdentifier)

// MakeOpenTelemetrySQSInterpreter creates an instance of the OpenTelemetrySQS span interpreter
func MakeOpenTelemetrySQSInterpreter(config *config.Config) *OpenTelemetrySQSInterpreter {
	return &OpenTelemetrySQSInterpreter{interpreter.Interpreter{Config: config}}
}

// Interpret performs the interpretation for the OpenTelemetrySQSInterpreter
func (t *OpenTelemetrySQSInterpreter) Interpret(spans []*pb.Span) []*pb.Span {
	log.Debugf("[OTEL] [SQS] Interpreting and mapping Open Telemetry data")

	for _, span := range spans {
		// no meta, add a empty map
		if span.Meta == nil {
			span.Meta = map[string]string{}
		}

		// awsService, awsServiceOk := span.Meta["aws.service.api"]
		awsRegion, awsRegionOk := span.Meta["aws.region"]
		awsOperation, awsOperationOk := span.Meta["aws.operation"]
		sqsEndpoint, sqsEndpointOk := span.Meta["messaging.url"]
		sqsQueueName, sqsQueueNameOk := span.Meta["messaging.destination"]

		if sqsQueueNameOk && sqsEndpointOk && awsOperationOk && awsRegionOk {
			sqsEndpointPieces := strings.Split(sqsEndpoint, "/") // Example Input: https://sqs.<region>.amazonaws.com/<account-id>/<queue-name>

			if len(sqsEndpointPieces) >= 3 {
				var accountID = sqsEndpointPieces[3]
				var urn = t.CreateServiceURN(sqsEndpoint)
				var arn = strings.ToLower(
					fmt.Sprintf("https://%s.queue.amazonaws.com/%s/%s", awsRegion, accountID, sqsQueueName))

				instrumentations.SpanBuilder(
					span,
					"consumer",
					"sqs",
					awsOperation,
					urn,
					arn,
				)
			} else {
				_ = log.Errorf("[OTEL] [SQS]: The SQS Endpoint URL is incorrect, Unable to parse %s.", sqsEndpointPieces)
			}
		} else {
			_ = log.Errorf("[OTEL] [SQS]: Unable to map the SQS request")

			if !awsRegionOk {
				_ = log.Errorf("[OTEL] [SQS]: 'aws.region' is not found in the span meta data, this value is required.")
			}
			if !awsOperationOk {
				_ = log.Errorf("[OTEL] [SQS]: 'aws.operation' is not found in the span meta data, this value is required.")
			}
			if !sqsEndpointOk {
				_ = log.Errorf("[OTEL] [SQS]: 'messaging.url' is not found in the span meta data, this value is required.")
			}
			if !sqsQueueNameOk {
				_ = log.Errorf("[OTEL] [SQS]: 'messaging.destination' is not found in the span meta data, this value is required.")
			}

			return nil
		}

		t.interpretHTTPError(span)
	}

	return spans
}

func (t *OpenTelemetrySQSInterpreter) interpretHTTPError(span *pb.Span) {
	if span.Error != 0 {
		if httpStatus, found := span.Metrics["http.status_code"]; found {
			if httpStatus >= 400 && httpStatus < 500 {
				span.Meta["span.errorClass"] = "4xx"
			} else if httpStatus >= 500 {
				span.Meta["span.errorClass"] = "5xx"
			}
		}
	}
}
